package process_test

import (
	"context"
	"runtime"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// promptProgram prints a prompt without newline, reads a line, answers and exits with code 7.
func promptProgram() process.Job {
	job := process.Job{Config: domain.RunConfiguration{Echo: true}, Mode: process.Terminal}
	if runtime.GOOS == "windows" {
		job.Command, job.Args = "cmd", []string{"/c", "set /p name=Name: & call echo hello %name% & exit /b 7"}
		return job
	}
	job.Command, job.Args = "sh", []string{"-c", "printf 'Name: '; read name; echo hello $name; exit 7"}
	return job
}

func TestTerminalModeRunsAnInteractiveProgram(t *testing.T) {
	sink := newSink()
	supervisor := process.New(sink)
	if err := supervisor.Start(context.Background(), promptProgram()); err != nil {
		t.Skipf("no pseudoterminal on this machine: %v", err)
	}
	sink.waitStdout(t, "Name:") // the prompt has no newline: it must still be shown

	if err := supervisor.WriteInput("Ana\n"); err != nil {
		t.Fatal(err)
	}
	event := sink.waitFinished(t)

	out := sink.Stdout()
	if !strings.Contains(out, "hello Ana") || strings.Contains(out, "\x1b") {
		t.Errorf("stdout = %q, want the answer without escape sequences", out)
	}
	if event.ExitCode != 7 || sink.startedCount() != 1 || supervisor.IsRunning() {
		t.Errorf("exit %d, started %d, running %v", event.ExitCode, sink.startedCount(), supervisor.IsRunning())
	}
}

func TestTerminalModeStopKillsTheProgram(t *testing.T) {
	sink := newSink()
	supervisor := process.New(sink)
	job := helperJob("beat", t.TempDir()+"/beat.txt")
	job.Mode = process.Terminal
	if err := supervisor.Start(context.Background(), job); err != nil {
		t.Skipf("no pseudoterminal on this machine: %v", err)
	}
	sink.waitStdout(t, "ready")

	_ = supervisor.Stop()
	sink.waitFinished(t)

	if supervisor.IsRunning() {
		t.Error("IsRunning must be false after Stop")
	}
}
