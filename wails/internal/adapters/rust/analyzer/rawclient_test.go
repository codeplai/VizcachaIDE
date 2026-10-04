package analyzer

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/uri"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/rusttest"
)

type pipes struct {
	io.Reader
	io.WriteCloser
}

// rawClient is a minimal LSP client that talks to the real rust-analyzer to record fixtures.
type rawClient struct {
	t           *testing.T
	conn        jsonrpc2.Conn
	diagnostics chan json.RawMessage
	initResult  json.RawMessage
	pulled      atomic.Value // what answers workspace/configuration; nil leaves the capability off
}

// locatorForTests finds rust-analyzer in the test toolchain (or skips the test).
func locatorForTests(t *testing.T) *rust.Locator {
	t.Helper()
	return rust.NewLocator(rust.Options{AppDir: os.TempDir(), BaseEnvironment: rusttest.Environment(t)})
}

// startRaw starts the real rust-analyzer for file and sends initialize with the options of the
// flavor, exactly as the lsp client does.
func startRaw(t *testing.T, file string) *rawClient { return startRawWith(t, file, nil) }

func startRawWith(t *testing.T, file string, pulled any) *rawClient {
	t.Helper()
	flavor := NewFlavor(Config{Locator: locatorForTests(t)})
	executable, args, err := flavor.Command(nil)
	if err != nil {
		t.Fatal(err)
	}
	root := flavor.RootOf(file)
	cmd := exec.Command(executable, args...)
	cmd.Dir = root
	cmd.Env = flavor.cfg.Locator.Environment()
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill() })
	c := &rawClient{t: t, diagnostics: make(chan json.RawMessage, 64)}
	if pulled != nil {
		c.pulled.Store(pulled)
	}
	c.conn = jsonrpc2.NewConn(jsonrpc2.NewStream(pipes{Reader: stdout, WriteCloser: stdin}))
	c.conn.Go(context.Background(), func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
		if req.Method() == "textDocument/publishDiagnostics" {
			c.diagnostics <- req.Params()
		}
		if _, isCall := req.(*jsonrpc2.Call); isCall {
			if req.Method() == "workspace/configuration" && c.pulled.Load() != nil {
				return reply(ctx, []any{c.pulled.Load()}, nil)
			}
			return reply(ctx, nil, nil)
		}
		return nil
	})
	c.initResult = c.call("initialize", map[string]any{
		"processId": os.Getpid(), "rootUri": string(uri.File(root)), "capabilities": capabilitiesFor(pulled != nil),
		"initializationOptions": flavor.InitializationOptions(),
		"workspaceFolders":      []any{map[string]any{"uri": string(uri.File(root)), "name": filepath.Base(root)}},
	})
	_ = c.conn.Notify(context.Background(), "initialized", map[string]any{})
	return c
}

func (c *rawClient) call(method string, params any) json.RawMessage {
	c.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	var result json.RawMessage
	if _, err := c.conn.Call(ctx, method, params, &result); err != nil {
		c.t.Fatalf("%s: %v", method, err)
	}
	return result
}

func (c *rawClient) notify(method string, params any) {
	_ = c.conn.Notify(context.Background(), method, params)
}

func (c *rawClient) open(file, text string) {
	c.notify("textDocument/didOpen", map[string]any{"textDocument": map[string]any{
		"uri": string(uri.File(file)), "languageId": languageID, "version": 1, "text": text,
	}})
}

func capabilitiesFor(pullsConfiguration bool) map[string]any {
	capabilities := recordedCapabilities()
	if pullsConfiguration {
		capabilities["workspace"] = map[string]any{"configuration": true}
	}
	return capabilities
}

func recordedCapabilities() map[string]any {
	plain := []string{"plaintext"}
	return map[string]any{"textDocument": map[string]any{
		"publishDiagnostics": map[string]any{},
		"completion":         map[string]any{"completionItem": map[string]any{"snippetSupport": false, "documentationFormat": plain}},
		"hover":              map[string]any{"contentFormat": plain},
		"signatureHelp": map[string]any{"signatureInformation": map[string]any{
			"documentationFormat": plain, "parameterInformation": map[string]any{"labelOffsetSupport": true}, "activeParameterSupport": true,
		}},
		"documentSymbol": map[string]any{"hierarchicalDocumentSymbolSupport": true},
	}}
}

func (c *rawClient) tryCall(method string, params any) (json.RawMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	var result json.RawMessage
	_, err := c.conn.Call(ctx, method, params, &result)
	return result, err
}
