package process_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

func fileSize(t *testing.T, path string) int64 {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info.Size()
}

func TestOneSlotIsSharedByTwoCallers(t *testing.T) {
	sink := newSink()
	supervisor := process.New(sink)
	beatFile := filepath.Join(t.TempDir(), "beat.txt")
	if err := supervisor.Start(context.Background(), helperJob("beat", beatFile)); err != nil {
		t.Fatal(err)
	}
	sink.waitStdout(t, "ready")

	cleaned := false
	second := helperJob("print", "other")
	second.Cleanup = func() { cleaned = true }
	err := supervisor.Start(context.Background(), second)

	if !errors.Is(err, app.ErrBusy) || !cleaned {
		t.Errorf("second Start: error = %v, cleaned = %v; want ErrBusy and cleanup", err, cleaned)
	}
	if !supervisor.IsRunning() {
		t.Error("IsRunning must be true while the first program lives")
	}
	_ = supervisor.Stop()
	sink.waitFinished(t)
	if supervisor.IsRunning() {
		t.Error("IsRunning must be false after Stop")
	}
	if err := supervisor.Start(context.Background(), helperJob("print", "again")); err != nil {
		t.Errorf("the slot must be free after the run: %v", err)
	}
	sink.waitFinished(t)
}

func TestThenChainsAStageAfterTheFirst(t *testing.T) {
	sink := newSink()
	supervisor := process.New(sink)
	var seenExit []int
	compile := helperJob("print", "compiling")
	compile.Then = func(exitCode int) (*process.Job, bool) {
		seenExit = append(seenExit, exitCode)
		program := helperJob("print", "running", "3")
		return &program, true
	}

	if err := supervisor.Start(context.Background(), compile); err != nil {
		t.Fatal(err)
	}
	event := sink.waitFinished(t)

	out := sink.Stdout()
	if !strings.Contains(out, "compiling") || strings.Index(out, "running") < strings.Index(out, "compiling") {
		t.Errorf("stdout = %q, want both stages in order", out)
	}
	if event.ExitCode != 3 || len(seenExit) != 1 || seenExit[0] != 0 || sink.startedCount() != 1 {
		t.Errorf("exit %d, Then saw %v, started %d; want the last exit code and one run:started", event.ExitCode, seenExit, sink.startedCount())
	}
}

func TestFinishedCarriesTheCompileFailureWhenThereIsNoNextStage(t *testing.T) {
	sink := newSink()
	compile := helperJob("print", "error here", "2")
	compile.Then = func(exitCode int) (*process.Job, bool) { return nil, exitCode == 0 }

	if err := process.New(sink).Start(context.Background(), compile); err != nil {
		t.Fatal(err)
	}

	if event := sink.waitFinished(t); event.ExitCode != 2 {
		t.Errorf("exit = %d, want 2", event.ExitCode)
	}
}

func TestSilentNoticeAppearsOnlyWhenNothingWasPrinted(t *testing.T) {
	sink := newSink()
	silent := helperJob("silent")
	silent.Notice = &process.SilentNotice{Text: func() string { return "preparing" }, Delay: 50 * time.Millisecond}
	if err := process.New(sink).Start(context.Background(), silent); err != nil {
		t.Fatal(err)
	}
	sink.waitFinished(t)
	if !strings.Contains(sink.Stdout(), "preparing") || !strings.Contains(sink.Stdout(), "done") {
		t.Errorf("stdout = %q, want the notice and the output", sink.Stdout())
	}

	chatty := newSink()
	talking := helperJob("chatty")
	talking.Notice = &process.SilentNotice{Text: func() string { return "preparing" }, Delay: time.Second}
	_ = process.New(chatty).Start(context.Background(), talking)
	chatty.waitFinished(t)
	if strings.Contains(chatty.Stdout(), "preparing") {
		t.Errorf("stdout = %q: a program that already printed needs no notice", chatty.Stdout())
	}
}

func TestStopEndsTheWholeProcessTree(t *testing.T) {
	sink := newSink()
	supervisor := process.New(sink)
	beatFile := filepath.Join(t.TempDir(), "beat.txt")
	if err := supervisor.Start(context.Background(), helperJob("parent", beatFile)); err != nil {
		t.Fatal(err)
	}
	sink.waitStdout(t, "ready")
	began := time.Now()

	if err := supervisor.Stop(); err != nil {
		t.Fatal(err)
	}
	sink.waitFinished(t)

	if took := time.Since(began); took < time.Second {
		t.Errorf("ended after %v: the programs ignore the interrupt, so Stop must wait its grace period", took)
	}
	// If the grandchild survived, the file would keep growing.
	before := fileSize(t, beatFile)
	time.Sleep(500 * time.Millisecond)
	if after := fileSize(t, beatFile); after != before {
		t.Errorf("a process of the tree is alive: beat file grew from %d to %d bytes", before, after)
	}
}

func TestWriteInputReachesStdin(t *testing.T) {
	sink := newSink()
	supervisor := process.New(sink)
	if err := supervisor.Start(context.Background(), helperJob("echo")); err != nil {
		t.Fatal(err)
	}
	if err := supervisor.WriteInput("Ana\n"); err != nil {
		t.Fatal(err)
	}
	sink.waitFinished(t)
	if !strings.Contains(sink.Stdout(), "got:Ana") {
		t.Errorf("stdout = %q", sink.Stdout())
	}
}

func TestMissingCommandReportsToolNotFound(t *testing.T) {
	job := helperJob("print", "x")
	job.Command = "vizcacha-no-such-tool"
	cleaned := false
	job.Cleanup = func() { cleaned = true }

	err := process.New(newSink()).Start(context.Background(), job)

	if !errors.Is(err, app.ErrToolNotFound) || !strings.Contains(err.Error(), "vizcacha-no-such-tool") || !cleaned {
		t.Errorf("error = %v, cleaned = %v; want ErrToolNotFound naming the tool, and cleanup", err, cleaned)
	}
}
