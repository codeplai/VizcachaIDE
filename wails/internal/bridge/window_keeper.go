package bridge

import (
	"context"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// WindowKeeper restores the window geometry at startup and saves it when closing.
type WindowKeeper struct {
	store app.WindowStateStore
	state domain.WindowState
}

// NewWindowKeeper loads the saved state (defaults on the first run or if unreadable).
func NewWindowKeeper(store app.WindowStateStore) *WindowKeeper {
	state, _ := store.Load()
	return &WindowKeeper{store: store, state: state.Clamped()}
}

// State is the state to open the window with (size already clamped).
func (k *WindowKeeper) State() domain.WindowState { return k.state }

// Restore moves the window to the saved position, or centers it when that position is
// not on any screen. Call it from OnDomReady.
func (k *WindowKeeper) Restore(ctx context.Context) {
	if k.state.Maximised {
		return // the position of a maximised window is decided by the system
	}
	if !k.state.PositionPlausible(screenSizes(ctx)) {
		runtime.WindowCenter(ctx)
		return
	}
	runtime.WindowSetPosition(ctx, k.state.X, k.state.Y)
}

// Save stores the geometry; it is the OnBeforeClose hook and never prevents closing.
// A maximised window keeps the previous normal size and position; a minimised one is
// not recorded.
func (k *WindowKeeper) Save(ctx context.Context) (prevent bool) {
	if runtime.WindowIsMinimised(ctx) {
		return false
	}
	state := k.state
	state.Maximised = runtime.WindowIsMaximised(ctx)
	if !state.Maximised {
		state.Width, state.Height = runtime.WindowGetSize(ctx)
		state.X, state.Y = runtime.WindowGetPosition(ctx)
		state.HasPosition = true
		state = state.Clamped()
	}
	_ = k.store.Save(state) // best effort: losing the geometry is not worth an error
	return false
}

// screenSizes lists the monitors with the primary one first.
func screenSizes(ctx context.Context) []domain.ScreenSize {
	screens, err := runtime.ScreenGetAll(ctx)
	if err != nil {
		return nil
	}
	sizes := make([]domain.ScreenSize, 0, len(screens))
	for _, s := range screens {
		size := domain.ScreenSize{Width: s.Size.Width, Height: s.Size.Height}
		if s.IsPrimary {
			sizes = append([]domain.ScreenSize{size}, sizes...)
			continue
		}
		sizes = append(sizes, size)
	}
	return sizes
}
