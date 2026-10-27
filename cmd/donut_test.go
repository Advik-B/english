package cmd

import (
	"strings"
	"testing"
)

func TestRenderDonutFrameShape(t *testing.T) {
	const width, height = 40, 12
	frame := renderDonutFrame(width, height, 0.3, 0.7)

	if !strings.HasPrefix(frame, "\x1b[H") {
		t.Fatalf("frame should start with cursor-home escape sequence")
	}

	body := strings.TrimPrefix(frame, "\x1b[H")
	lines := strings.Split(body, "\n")
	if got, want := len(lines), height+1; got != want { // trailing newline
		t.Fatalf("line count mismatch: got %d want %d", got, want)
	}
	for i := 0; i < height; i++ {
		if got := len(lines[i]); got != width {
			t.Fatalf("line %d width mismatch: got %d want %d", i, got, width)
		}
	}
}

func TestRenderDonutFrameContainsShading(t *testing.T) {
	frame := renderDonutFrame(donutWidth, donutHeight, 0, 0)
	body := strings.TrimPrefix(frame, "\x1b[H")

	if !strings.ContainsAny(body, ".,-~:;=!*#$@") {
		t.Fatalf("frame did not contain any doughnut shading characters")
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		value int
		want  int
	}{
		{value: -5, want: 0},
		{value: 3, want: 3},
		{value: 99, want: 11},
	}

	for _, tc := range tests {
		if got := clamp(tc.value, 0, 11); got != tc.want {
			t.Fatalf("clamp(%d)=%d, want %d", tc.value, got, tc.want)
		}
	}
}
