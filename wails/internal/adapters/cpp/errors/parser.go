package errors

import (
	"regexp"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

// Diagnostic sources.
const (
	sourceCompiler = "compiler"
	sourceLinker   = "linker"
	sourceRuntime  = "runtime"
)

// maxRawLines caps the context a diagnostic keeps (GCC prints one note per overload candidate).
const maxRawLines = 40

// noiseLine matches the lines that explain nothing by themselves: where a diagnostic was found
// ("In function ...", include chains), totals ("2 errors generated.") and the linker's summary
// ("collect2: error: ld returned 1 exit status", "clang: error: linker command failed ...").
var noiseLine = regexp.MustCompile(`^(?:In (?:file included from|instantiation of|[\w ]*function|[\w ]*constructor|[\w ]*destructor)|At global scope:|\s+from \S|compilation terminated\.|\d+ (?:errors?|warnings?)(?: and \d+ errors?)? generated\.|collect2(?:\.exe)?: error: ld returned|\S+: error: linker command failed|.*: in function )`)

// parser turns raw g++/clang++ output into diagnostics. identify returns the catalog id of a message.
type parser struct {
	identify    func(message string) string
	workingDir  string
	diagnostics []domain.Diagnostic
	raw         [][]string
	// context is the diagnostic that receives the lines that follow its first line (notes, code
	// excerpts, "candidate" lists), or -1.
	context int
}

// lineHandler recognises the diagnostic that starts at lines[0]. It returns how many further
// lines it consumed and whether it recognised the line.
type lineHandler func(p *parser, lines []string) (extra int, ok bool)

var handlers = []lineHandler{
	(*parser).cmakeLine,
	(*parser).compilerLine,
	(*parser).undefinedSymbolLine,
	(*parser).duplicateSymbolLine,
	(*parser).multipleDefinitionLine,
	(*parser).undefinedReferenceLine,
	(*parser).gnuTerminateLine,
	(*parser).libcxxTerminateLine,
	(*parser).crashLine,
}

// ParseOutput is the errorcatalog.OutputParser of C++. The raw output is what the compiler printed
// (or the checker, `-fsyntax-only`), what the linker printed, or what the program printed ending
// with the crash line the runner adds.
func ParseOutput(rawOutput, workingDir string, identify func(message string) string) []domain.Diagnostic {
	p := &parser{identify: identify, workingDir: workingDir, context: -1}
	lines := errorcatalog.SplitLines(rawOutput)
	for index := 0; index < len(lines); index++ {
		index += p.consume(lines[index:])
	}
	return p.finish()
}

// consume handles lines[0] and returns how many further lines belong to it.
func (p *parser) consume(lines []string) int {
	line := strings.TrimRight(lines[0], " \t\r")
	if noiseLine.MatchString(line) {
		return 0
	}
	for _, handle := range handlers {
		if extra, ok := handle(p, lines); ok {
			return extra
		}
	}
	p.addContext(line)
	return 0
}

// addContext appends an indented line (code excerpt, caret, candidate) to the current diagnostic;
// any other line (the program's own output) closes it.
func (p *parser) addContext(line string) {
	if p.context < 0 {
		return
	}
	if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, ">>>") {
		p.context = -1
		return
	}
	p.appendRaw(p.context, line)
	if p.diagnostics[p.context].Location == nil && p.diagnostics[p.context].Source == sourceLinker {
		p.diagnostics[p.context].Location = p.referencedBy(line)
	}
}

// add records a diagnostic; context says whether the lines after it belong to it.
func (p *parser) add(diagnostic domain.Diagnostic, rawLines []string, context bool) {
	diagnostic.Code = p.identify(diagnostic.Message)
	p.diagnostics = append(p.diagnostics, diagnostic)
	p.raw = append(p.raw, rawLines)
	p.context = -1
	if context {
		p.context = len(p.diagnostics) - 1
	}
}

func (p *parser) appendRaw(index int, line string) {
	if len(p.raw[index]) < maxRawLines {
		p.raw[index] = append(p.raw[index], line)
	}
}

func (p *parser) finish() []domain.Diagnostic {
	for index := range p.diagnostics {
		p.diagnostics[index].RawText = strings.Join(p.raw[index], "\n")
	}
	return p.diagnostics
}
