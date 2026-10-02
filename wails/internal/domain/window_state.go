package domain

// Window size limits (logical pixels). The minimum matches the Wails options.
const (
	WindowDefaultWidth  = 1280
	WindowDefaultHeight = 800
	WindowMinWidth      = 1024
	WindowMinHeight     = 640
	WindowMaxSize       = 8000

	// windowVisibleMargin is how much of the window must stay on a screen.
	windowVisibleMargin = 100
)

// WindowState is the window geometry remembered between sessions.
// Width and Height are the NORMAL (not maximised) size.
type WindowState struct {
	Width     int  `json:"width"`
	Height    int  `json:"height"`
	X         int  `json:"x"`
	Y         int  `json:"y"`
	Maximised bool `json:"maximised"`
	// HasPosition is false when X and Y were never saved (first run).
	HasPosition bool `json:"hasPosition"`
}

// ScreenSize is the size of one monitor in logical pixels.
type ScreenSize struct{ Width, Height int }

// DefaultWindowState is the state of a first run.
func DefaultWindowState() WindowState {
	return WindowState{Width: WindowDefaultWidth, Height: WindowDefaultHeight}
}

// Clamped returns the state with the size inside [minimum, WindowMaxSize]; a missing
// size becomes the default.
func (w WindowState) Clamped() WindowState {
	w.Width = clampSize(w.Width, WindowDefaultWidth, WindowMinWidth)
	w.Height = clampSize(w.Height, WindowDefaultHeight, WindowMinHeight)
	return w
}

func clampSize(value, fallback, minimum int) int {
	if value <= 0 {
		return fallback
	}
	return min(max(value, minimum), WindowMaxSize)
}

// PositionPlausible reports whether the saved position can be used. Wails v2 does not
// give the origin of each monitor, so the desktop is approximated: the primary screen
// is at (0,0) and the others may lie on either side of it. The window must keep at
// least windowVisibleMargin pixels inside that area and its title bar must be visible.
// When it is not plausible (for example a monitor was unplugged) the caller centers.
func (w WindowState) PositionPlausible(screens []ScreenSize) bool {
	if !w.HasPosition || len(screens) == 0 {
		return false
	}
	primary := screens[0]
	totalWidth, maxHeight := 0, 0
	for _, s := range screens {
		totalWidth += s.Width
		maxHeight = max(maxHeight, s.Height)
	}
	left := -(totalWidth - primary.Width)
	top := -(maxHeight - primary.Height)
	horizontal := w.X+w.Width >= left+windowVisibleMargin && w.X <= totalWidth-windowVisibleMargin
	vertical := w.Y >= top && w.Y <= top+maxHeight-windowVisibleMargin
	return horizontal && vertical
}
