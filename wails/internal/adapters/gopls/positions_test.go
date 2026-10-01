package gopls

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"go.lsp.dev/protocol"
)

// nopCloser lets a buffer or reader act as the connection of a jsonrpc2 stream.
type nopCloser struct{ io.ReadWriter }

func (nopCloser) Close() error { return nil }

// oneByteReader delivers its data one byte per Read, to exercise message splitting.
type oneByteReader struct {
	data []byte
}

func (r *oneByteReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	p[0] = r.data[0]
	r.data = r.data[1:]
	return 1, nil
}

func (r *oneByteReader) Write(p []byte) (int, error) { return len(p), nil }

const accented = "\tx := \"canción ñandú\" + \"😀\" + y\n"

func TestColumnsConvertBetweenRunesAndUTF16(t *testing.T) {
	cases := []struct {
		name      string
		column    int // 1-based, in runes
		character int // 0-based, in UTF-16 units
	}{
		{"start", 1, 0},
		{"accents count once", 16, 15},
		{"before emoji", 26, 25},
		{"after emoji", 27, 27}, // the emoji takes two UTF-16 units
		{"the y after it", 31, 31},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			position := toLSPPosition(accented, 1, c.column)
			if int(position.Character) != c.character || position.Line != 0 {
				t.Errorf("toLSPPosition = %+v, want character %d", position, c.character)
			}
			line, column := fromLSPPosition(accented, position)
			if line != 1 || column != c.column {
				t.Errorf("fromLSPPosition = %d:%d, want 1:%d", line, column, c.column)
			}
		})
	}
}

func TestPositionInsideASurrogatePairRoundsDown(t *testing.T) {
	_, column := fromLSPPosition(accented, protocol.Position{Line: 0, Character: 26})
	if column != 26 {
		t.Errorf("column = %d, want 26 (the emoji start)", column)
	}
}

func TestLinesAreSplitWithoutCarriageReturns(t *testing.T) {
	text := "uno\r\ndos ñ\r\ntres"
	if got := lineOf(text, 1); got != "dos ñ" {
		t.Errorf("lineOf = %q", got)
	}
	if got := lineOf(text, 9); got != "" {
		t.Errorf("lineOf past the end = %q", got)
	}
}

func TestModuleRootIsTheNearestGoMod(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "cmd", "app")
	if err := os.MkdirAll(nested, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(nested, "main.go")
	if got := moduleRoot(file); pathKey(got) != pathKey(root) {
		t.Errorf("moduleRoot = %q, want %q", got, root)
	}
	loose := filepath.Join(t.TempDir(), "solo.go")
	if got := moduleRoot(loose); pathKey(got) != pathKey(filepath.Dir(loose)) {
		t.Errorf("without go.mod the root is the folder, got %q", got)
	}
}

func TestURIsRoundTripThroughPaths(t *testing.T) {
	path := filepath.Join(t.TempDir(), "carpeta ñ", "main.go")
	if got := uriToPath(pathToURI(path)); pathKey(got) != pathKey(path) {
		t.Errorf("round trip = %q, want %q", got, path)
	}
}
