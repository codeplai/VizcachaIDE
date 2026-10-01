package toolchain

import (
	"context"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const versionTimeout = 5 * time.Second

var versionNumber = regexp.MustCompile(`\d+\.\d+(?:\.\d+)?`)

// Info implements app.Toolchain: the version of every tool that was found.
func (t *Toolchain) Info(ctx context.Context) domain.ToolchainInfo {
	var info domain.ToolchainInfo
	var wg sync.WaitGroup
	for tool, target := range map[string]*string{
		ToolGo: &info.GoVersion, ToolDelve: &info.DelveVersion, ToolGopls: &info.GoplsVersion,
	} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			*target = t.version(ctx, tool)
		}()
	}
	wg.Wait()
	info.GoSource = t.source(ToolGo)
	info.DelveSource = t.source(ToolDelve)
	info.GoplsSource = t.source(ToolGopls)
	return info
}

// source maps the locator origin to the domain value (same names on both sides).
func (t *Toolchain) source(tool string) domain.ToolSource {
	return domain.ToolSource(t.locator.Locate(tool).Origin)
}

// version runs "<tool> version" and returns the first version number it prints ("" if it fails).
func (t *Toolchain) version(ctx context.Context, tool string) string {
	location := t.locator.Locate(tool)
	if location.Origin == OriginMissing {
		return ""
	}
	ctx, cancel := context.WithTimeout(ctx, versionTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, location.Path, "version")
	cmd.Env = environmentList(t.Environment())
	hideConsole(cmd)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return parseVersion(string(output))
}

func parseVersion(output string) string {
	return strings.TrimSpace(versionNumber.FindString(output))
}
