package main

import (
	"context"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/console"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/delve"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/errorcatalog"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/gopls"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/toolchain"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/bridge"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// goProfile is the profile of Go.
//
// Transitional (M0): track N3 declares it in adapters/golang/profile.go; the integration deletes
// this copy and uses that one.
var goProfile = domain.LanguageProfile{
	ID: domain.CodeLanguageGo, NameKey: "codeLanguage.go", Extensions: []string{".go"},
	Indent: domain.IndentStyle{UseTabs: true, Size: 4},
	Capabilities: domain.Capabilities{
		Build: true, Console: true, Format: true, Check: true,
		PackageActions: []domain.PackageAction{domain.PackageInit, domain.PackageAdd, domain.PackageTidy},
		ThreadsLabel:   "debug.goroutines",
	},
	Tools: []domain.ToolSpec{
		{ID: "go", Role: domain.RoleRuntime, LabelKey: "settings.toolGo", MissingKey: "errors.goNotFound", InstallURL: "https://go.dev/dl/"},
		{ID: "dlv", Role: domain.RoleDebugAdapter, LabelKey: "settings.toolDelve", MissingKey: "errors.delveNotFound",
			InstallCommand: "go install github.com/go-delve/delve/cmd/dlv@latest"},
		{ID: "gopls", Role: domain.RoleLanguageServer, LabelKey: "settings.toolGopls", MissingKey: "errors.goplsNotFound",
			InstallCommand: "go install golang.org/x/tools/gopls@latest"},
	},
}

// newGoSupport creates the adapters of Go and gathers them in one LanguageSupport. Tool paths
// are resolved when a debug session or gopls starts (configured -> bundled -> PATH), so changing
// them in Settings needs no restart. The returned function releases what the adapters hold
// beyond what shutdown already stops (runner, debugger and language server): Go holds nothing.
func newGoSupport(sink *bridge.WailsEventSink, store app.SettingsStore, texts *backendTexts) (app.LanguageSupport, func(context.Context), error) {
	explainer, err := errorcatalog.NewExplainer()
	if err != nil {
		return app.LanguageSupport{}, nil, err
	}
	goToolchain := toolchain.New(toolchain.Options{
		Sink:             sink,
		Settings:         store,
		FirstBuildNotice: func() string { return texts.text("run.firstBuild") },
	})
	ports := newGoToolchainPorts(goToolchain)
	support := app.LanguageSupport{
		Profile: goProfile,
		Runner:  ports,
		Debugger: delve.New(sink, delve.Options{
			DelvePath:   func() string { return goToolchain.Locate(toolchain.ToolDelve).Path },
			Environment: goToolchain.Environment,
			Translate:   texts.text,
		}),
		LanguageServer: gopls.New(sink, gopls.Config{
			Executable:  func() string { return goToolchain.Locate(toolchain.ToolGopls).Path },
			Environment: goToolchain.Environment,
		}),
		Explainer: explainer,
		Console:   console.New(0),
		Formatter: ports,
		Checker:   ports,
		Packages:  ports,
	}
	return support, func(context.Context) {}, nil
}
