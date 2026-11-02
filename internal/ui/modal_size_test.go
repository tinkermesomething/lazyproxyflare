package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestModalDimensions(t *testing.T) {
	tests := []struct {
		name          string
		width, height int
		large         bool
		wantW, wantH  int
	}{
		{"normal wide terminal is capped at 80", 179, 49, false, 80, 36},
		{"large uses most of a wide terminal", 179, 49, true, 161, 44},
		{"normal on 80x24", 80, 24, false, 45, 18},
		{"large on 80x24", 80, 24, true, 72, 21},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h := modalDimensions(tt.width, tt.height, tt.large)
			if w != tt.wantW || h != tt.wantH {
				t.Errorf("modalDimensions(%d,%d,%v) = %dx%d, want %dx%d", tt.width, tt.height, tt.large, w, h, tt.wantW, tt.wantH)
			}
		})
	}
}

// Modals must never be larger than the screen, even on tiny terminals.
func TestModalDimensionsNeverExceedScreen(t *testing.T) {
	for _, large := range []bool{false, true} {
		for width := 1; width <= 200; width += 7 {
			for height := 1; height <= 60; height += 5 {
				w, h := modalDimensions(width, height, large)
				if w < 1 || h < 1 {
					t.Fatalf("non-positive size %dx%d for screen %dx%d large=%v", w, h, width, height, large)
				}
				if width >= 20 && height >= 10 && (w > width || h > height) {
					t.Fatalf("modal %dx%d exceeds screen %dx%d large=%v", w, h, width, height, large)
				}
			}
		}
	}
}

func TestRenderModalOverlayLargeUsesWidth(t *testing.T) {
	long := strings.Repeat("x", 150)
	out := RenderModalOverlayLarge("", "T", long, 179, 49)
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, long) {
			return
		}
	}
	t.Errorf("150-char line was truncated in the large modal (max line width %d)", lipgloss.Width(out))
}
