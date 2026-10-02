package domain

import "testing"

func TestClamped(t *testing.T) {
	cases := []struct {
		name     string
		in, want WindowState
	}{
		{"empty gets defaults", WindowState{}, WindowState{Width: 1280, Height: 800}},
		{"too small", WindowState{Width: 300, Height: 200}, WindowState{Width: 1024, Height: 640}},
		{"too big", WindowState{Width: 99999, Height: 20000}, WindowState{Width: 8000, Height: 8000}},
		{"valid kept", WindowState{Width: 1500, Height: 900, X: 5}, WindowState{Width: 1500, Height: 900, X: 5}},
	}
	for _, c := range cases {
		if got := c.in.Clamped(); got != c.want {
			t.Errorf("%s: got %+v, want %+v", c.name, got, c.want)
		}
	}
}

func TestPositionPlausible(t *testing.T) {
	one := []ScreenSize{{1920, 1080}}
	two := []ScreenSize{{1920, 1080}, {1920, 1080}}
	at := func(x, y int) WindowState {
		return WindowState{Width: 1280, Height: 800, X: x, Y: y, HasPosition: true}
	}
	cases := []struct {
		name    string
		state   WindowState
		screens []ScreenSize
		want    bool
	}{
		{"visible on one screen", at(100, 50), one, true},
		{"never saved", WindowState{Width: 1280, Height: 800}, one, false},
		{"no screens known", at(0, 0), nil, false},
		{"monitor unplugged to the right", at(2500, 100), one, false},
		{"monitor unplugged to the left", at(-1900, 100), one, false},
		{"second monitor to the right", at(2500, 100), two, true},
		{"second monitor to the left", at(-1900, 100), two, true},
		{"title bar above everything", at(100, -900), one, false},
		{"below the screen", at(100, 1050), one, false},
	}
	for _, c := range cases {
		if got := c.state.PositionPlausible(c.screens); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
