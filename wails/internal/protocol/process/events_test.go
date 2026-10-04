package process_test

import (
	"context"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// A job with its own Events (a debug adapter's runInTerminal) reports there, not to the sink.
func TestJobEventsReplaceTheSink(t *testing.T) {
	sink := newSink()
	supervisor := process.New(sink)
	events := newSink()
	job := helperJob("print", "debugged output", "4")
	job.Events = events

	if err := supervisor.Start(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	event := events.waitFinished(t)

	if !strings.Contains(events.Stdout(), "debugged output") || event.ExitCode != 4 || events.startedCount() != 1 {
		t.Errorf("job events: stdout %q, exit %d, started %d", events.Stdout(), event.ExitCode, events.startedCount())
	}
	if sink.Stdout() != "" || sink.startedCount() != 0 || len(sink.finished) != 0 {
		t.Errorf("the supervisor's sink got events: stdout %q, started %d", sink.Stdout(), sink.startedCount())
	}
	if supervisor.IsRunning() {
		t.Error("the slot is still taken after the job ended")
	}
}

// Job.Finished adds the closing line of the last stage (a crash) before run:finished.
func TestFinishedLineComesBeforeTheEnd(t *testing.T) {
	sink := newSink()
	supervisor := process.New(sink)
	job := helperJob("print", "program output", "3")
	job.Finished = func(exitCode int) string {
		if exitCode == 3 {
			return "Segmentation fault"
		}
		return ""
	}
	if err := supervisor.Start(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	if event := sink.waitFinished(t); event.ExitCode != 3 {
		t.Fatalf("exit %d", event.ExitCode)
	}
	if !strings.Contains(sink.Stderr(), "Segmentation fault\n") {
		t.Errorf("stderr = %q, want the crash line", sink.Stderr())
	}
}

func TestOutputFilterRewritesAndDropsWholeLines(t *testing.T) {
	sink := newSink()
	supervisor := process.New(sink)
	job := helperJob("print", "json:error E0382\nnoise\nplain text", "0")
	job.OutputFilter = func(stream, line string) (string, string, bool) {
		if line == "noise" {
			return stream, "", false
		}
		if strings.HasPrefix(line, "json:") {
			return "stderr", strings.TrimPrefix(line, "json:"), true
		}
		return stream, line, true
	}
	if err := supervisor.Start(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	sink.waitFinished(t)
	if got := strings.ReplaceAll(sink.Stdout(), "\r", ""); got != "plain text\n" {
		t.Errorf("stdout = %q", got)
	}
	if got := strings.ReplaceAll(sink.Stderr(), "\r", ""); got != "error E0382\n" {
		t.Errorf("stderr = %q", got)
	}
}
