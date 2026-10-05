package vcpkg

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

// installingConfigurer behaves like CMake with vcpkg: it "installs" the libraries of vcpkg.json
// by writing the usage text and the status file, or fails when asked.
type installingConfigurer struct {
	calls  int
	failOn string
	usage  map[string]string
	block  chan struct{}
}

func (c *installingConfigurer) Configure(_ context.Context, root string) (string, error) {
	c.calls++
	if c.block != nil {
		<-c.block
	}
	ports, _ := Dependencies(root)
	var status strings.Builder
	for _, port := range ports {
		if port == c.failOn {
			return "error: building " + port + ":x64-linux failed", errors.New("configure failed")
		}
		installed := InstalledDirectory(root, "x64-linux")
		_ = os.MkdirAll(filepath.Join(installed, "share", port), 0o755)
		_ = os.WriteFile(filepath.Join(installed, "share", port, "usage"), []byte(c.usage[port]), 0o644)
		status.WriteString("Package: " + port + "\nVersion: 1.2.3\nStatus: install ok installed\n\n")
	}
	_ = os.MkdirAll(filepath.Join(root, "build", "vcpkg_installed", "vcpkg"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "build", "vcpkg_installed", "vcpkg", "status"), []byte(status.String()), 0o644)
	return "configured", nil
}

func newManager(t *testing.T, configurer Configurer) (*Manager, *recordedEvents, string) {
	t.Helper()
	events := newEvents()
	locator := NewLocator(LocatorOptions{AppDir: t.TempDir(), BaseEnvironment: []string{}, CacheDir: t.TempDir(), GOOS: "linux", GOARCH: "amd64"})
	manager := NewManager(locator, ManagerOptions{Configurer: configurer, Events: events})
	project := filepath.Join(t.TempDir(), "Ñandú")
	mustWrite(t, filepath.Join(project, "CMakeLists.txt"), template)
	mustWrite(t, filepath.Join(project, "vcpkg.json"), "{\n  \"name\": \"nandu\",\n  \"version\": \"0.1.0\",\n  \"dependencies\": []\n}\n")
	return manager, events, project
}

func TestAddInstallsAndWritesTheBlock(t *testing.T) {
	configurer := &installingConfigurer{usage: map[string]string{"fmt": fmtUsage}}
	manager, events, project := newManager(t, configurer)

	if err := manager.Add(context.Background(), project, "fmt"); err != nil {
		t.Fatal(err)
	}
	if code := events.wait(t); code != 0 {
		t.Fatalf("exit code %d, output %s", code, events.text())
	}
	if !strings.Contains(events.text(), "first time") {
		t.Errorf("no slow-install notice in %q", events.text())
	}
	if len(events.started) != 1 || events.started[0].Target != "vcpkg add fmt" || events.started[0].WorkingDir != project {
		t.Errorf("started = %+v", events.started)
	}
	cmake := readText(t, filepath.Join(project, "CMakeLists.txt"))
	if !strings.Contains(cmake, "# fmt\nfind_package(fmt CONFIG REQUIRED)\ntarget_link_libraries(nandu PRIVATE fmt::fmt)\n") {
		t.Errorf("CMakeLists:\n%s", cmake)
	}
	if configurer.calls != 2 {
		t.Errorf("configure ran %d times, want 2 (install, then with the lines)", configurer.calls)
	}
}

func TestListWritesDependenciesWithTheirInstalledVersion(t *testing.T) {
	manager, events, project := newManager(t, &installingConfigurer{})
	if _, err := AddDependency(project, "fmt"); err != nil {
		t.Fatal(err)
	}
	if _, err := AddDependency(project, "zlib"); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(project, "build", "vcpkg_installed", "vcpkg", "status"), "Package: fmt\nVersion: 12.2.0\nStatus: install ok installed\n")
	if err := manager.List(context.Background(), project); err != nil {
		t.Fatal(err)
	}
	events.wait(t)
	if got := events.text(); got != "fmt 12.2.0\nzlib \n" {
		t.Errorf("list = %q", got)
	}
}

func TestRemoveCleansBothFilesAndConfigures(t *testing.T) {
	configurer := &installingConfigurer{usage: map[string]string{"fmt": fmtUsage}}
	manager, events, project := newManager(t, configurer)
	_ = manager.Add(context.Background(), project, "fmt")
	events.wait(t)
	if err := manager.Remove(context.Background(), project, "fmt"); err != nil {
		t.Fatal(err)
	}
	if code := events.wait(t); code != 0 {
		t.Fatalf("exit %d: %s", code, events.text())
	}
	if got := readText(t, filepath.Join(project, "CMakeLists.txt")); got != template {
		t.Errorf("CMakeLists:\n%s", got)
	}
	if names, _ := Dependencies(project); len(names) != 0 {
		t.Errorf("dependencies = %v", names)
	}
}

func TestAFailedInstallRestoresTheProjectAndExplainsInTheOutput(t *testing.T) {
	manager, events, project := newManager(t, &installingConfigurer{failOn: "broken"})
	before := readText(t, filepath.Join(project, "vcpkg.json"))
	if err := manager.Add(context.Background(), project, "broken"); err != nil {
		t.Fatal(err)
	}
	if code := events.wait(t); code != 1 {
		t.Fatalf("exit code %d", code)
	}
	if got := readText(t, filepath.Join(project, "vcpkg.json")); got != before {
		t.Errorf("vcpkg.json was left changed:\n%s", got)
	}
	if !strings.Contains(events.text(), "error: building broken:x64-linux failed") {
		t.Errorf("output = %q", events.text())
	}
}

func TestOnlyOneOperationRunsAtATime(t *testing.T) {
	configurer := &installingConfigurer{block: make(chan struct{})}
	manager, events, project := newManager(t, configurer)
	if err := manager.Add(context.Background(), project, "fmt"); err != nil {
		t.Fatal(err)
	}
	if err := manager.Add(context.Background(), project, "zlib"); !errors.Is(err, app.ErrBusy) {
		t.Errorf("second operation: %v", err)
	}
	close(configurer.block)
	events.wait(t)
}

func TestStopCancelsTheOperationInProgress(t *testing.T) {
	manager, events, project := newManager(t, cancellable{})
	_ = manager.Add(context.Background(), project, "fmt")
	manager.Stop()
	if code := events.wait(t); code != 1 {
		t.Errorf("exit code %d", code)
	}
}

type cancellable struct{}

func (c cancellable) Configure(ctx context.Context, _ string) (string, error) {
	<-ctx.Done()
	return "", ctx.Err()
}

func TestVerbsNeedACMakeProjectAndRejectBadNames(t *testing.T) {
	manager, _, _ := newManager(t, &installingConfigurer{})
	if err := manager.Add(context.Background(), t.TempDir(), "fmt"); !errors.Is(err, ErrNeedsProject) || !errors.Is(err, app.ErrUnsupported) {
		t.Errorf("outside a project: %v", err)
	}
	if err := manager.Add(context.Background(), t.TempDir(), "--help"); !errors.Is(err, app.ErrInvalidArgument) {
		t.Errorf("bad name: %v", err)
	}
	if err := manager.Init(context.Background(), "x", "y"); !errors.Is(err, app.ErrUnsupported) {
		t.Errorf("init: %v", err)
	}
	if err := manager.Tidy(context.Background(), "x"); !errors.Is(err, app.ErrUnsupported) {
		t.Errorf("tidy: %v", err)
	}
}
