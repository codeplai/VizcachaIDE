//go:build windows

package filesystem

import (
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"golang.org/x/sys/windows"
)

// Flags of SHFileOperationW: delete to the Recycle Bin, with no dialogs of any kind.
const (
	foDelete     = 0x0003
	fofSilent    = 0x0004
	fofNoConfirm = 0x0010
	fofAllowUndo = 0x0040
	fofNoErrorUI = 0x0400
	trashFlags   = fofAllowUndo | fofNoConfirm | fofSilent | fofNoErrorUI
)

// shFileOpStruct is SHFILEOPSTRUCTW. Go aligns the fields like the C compiler does on
// 64-bit Windows (the only target of VizcachaIDE).
type shFileOpStruct struct {
	Hwnd                 uintptr
	Func                 uint32
	From                 *uint16
	To                   *uint16
	Flags                uint16
	AnyOperationsAborted int32
	NameMappings         uintptr
	ProgressTitle        *uint16
}

// doubleNulTerminated builds the list SHFileOperationW expects: UTF-16 path + two NULs.
func doubleNulTerminated(path string) ([]uint16, error) {
	units, err := windows.UTF16FromString(path) // already ends with one NUL
	if err != nil {
		return nil, fmt.Errorf("path %q: %w", path, err)
	}
	return append(units, 0), nil
}

func trash(path string) error {
	from, err := doubleNulTerminated(path)
	if err != nil {
		return err
	}
	op := shFileOpStruct{Func: foDelete, From: &from[0], Flags: trashFlags}
	proc := windows.NewLazySystemDLL("shell32.dll").NewProc("SHFileOperationW")
	code, _, _ := proc.Call(uintptr(unsafe.Pointer(&op)))
	if code != 0 || op.AnyOperationsAborted != 0 {
		return fmt.Errorf("%w: SHFileOperationW code 0x%X", app.ErrTrashUnavailable, code)
	}
	return nil
}

func reveal(path string) error {
	command := exec.Command("explorer.exe")
	// explorer.exe needs the raw argument line: /select,"C:\a b\c.go".
	command.SysProcAttr = &syscall.SysProcAttr{CmdLine: revealCommandLine(path)}
	if err := command.Start(); err != nil {
		return fmt.Errorf("explorer: %w", err)
	}
	go func() { _ = command.Wait() }()
	return nil
}

func revealCommandLine(path string) string { return `explorer.exe /select,"` + path + `"` }
