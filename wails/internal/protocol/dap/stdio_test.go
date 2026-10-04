package dap_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/dap"
)

const echoAdapterEnv = "VIZCACHA_ECHO_ADAPTER"

// TestEchoAdapter is not a test: TestStdioTransport runs it as an adapter that echoes lines.
func TestEchoAdapter(t *testing.T) {
	if os.Getenv(echoAdapterEnv) != "1" {
		t.Skip("helper process")
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Println("echo:", scanner.Text())
	}
	os.Exit(0)
}

func TestStdioTransportTalksToTheAdapterAndCloses(t *testing.T) {
	transport := &dap.StdioTransport{
		Command: os.Args[0],
		Args:    []string{"-test.run=^TestEchoAdapter$"},
		Env:     append(os.Environ(), echoAdapterEnv+"=1"),
	}
	conn, err := transport.Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(conn, "hello\n"); err != nil {
		t.Fatal(err)
	}
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil || strings.TrimSpace(line) != "echo: hello" {
		t.Fatalf("read %q, %v", line, err)
	}
	if _, err := transport.Open(context.Background()); err == nil {
		t.Error("a second Open while the adapter runs must fail")
	}
	if err := transport.Close(); err != nil {
		t.Fatal(err)
	}
	if err := transport.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestStdioTransportReportsAMissingAdapter(t *testing.T) {
	transport := &dap.StdioTransport{Command: "vizcacha-no-such-adapter"}
	if _, err := transport.Open(context.Background()); err == nil {
		t.Fatal("Open of a missing command must fail")
	}
}
