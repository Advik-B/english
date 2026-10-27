package cmd

import (
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

const (
	donutWidth     = 80
	donutHeight    = 22
	donutFrameTime = 33 * time.Millisecond
)

var donutFrames int

var donutCmd = &cobra.Command{
	Use:   "donut",
	Short: "Render a spinning ASCII doughnut animation in the terminal",
	Long: `Render a spinning ASCII doughnut animation directly in the terminal.

The animation updates in-place until you stop it (Ctrl+C).`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		runDonut(cmd.OutOrStdout(), donutFrames)
	},
}

func runDonut(out io.Writer, frames int) {
	fmt.Fprint(out, "\x1b[2J\x1b[H\x1b[?25l")
	defer fmt.Fprint(out, "\x1b[?25h\n")

	ticker := time.NewTicker(donutFrameTime)
	defer ticker.Stop()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sig)

	A, B := 0.0, 0.0
	for frame := 0; ; frame++ {
		fmt.Fprint(out, renderDonutFrame(donutWidth, donutHeight, A, B))
		A += 0.04
		B += 0.02

		if frames > 0 && frame+1 >= frames {
			return
		}

		select {
		case <-ticker.C:
		case <-sig:
			return
		}
	}
}

func renderDonutFrame(width, height int, A, B float64) string {
	if width <= 0 || height <= 0 {
		return "\x1b[H"
	}

	output := make([]rune, width*height)
	for i := range output {
		output[i] = ' '
	}
	zbuf := make([]float64, width*height)

	sinA, cosA := math.Sin(A), math.Cos(A)
	sinB, cosB := math.Sin(B), math.Cos(B)
	chars := []rune(".,-~:;=!*#$@")

	for theta := 0.0; theta < 2*math.Pi; theta += 0.07 {
		sinTheta, cosTheta := math.Sin(theta), math.Cos(theta)

		for phi := 0.0; phi < 2*math.Pi; phi += 0.02 {
			sinPhi, cosPhi := math.Sin(phi), math.Cos(phi)

			h := cosTheta + 2
			D := 1 / (sinPhi*h*sinA + sinTheta*cosA + 5)
			t := sinPhi*h*cosA - sinTheta*sinA

			x := int(float64(width)/2 + 30*D*(cosPhi*h*cosB-t*sinB))
			y := int(float64(height)/2 + 15*D*(cosPhi*h*sinB+t*cosB))

			if x < 0 || x >= width || y < 0 || y >= height {
				continue
			}

			o := x + width*y
			if D > zbuf[o] {
				zbuf[o] = D
				N := int(8 * ((sinTheta*sinA-sinPhi*cosTheta*cosA)*cosB - sinPhi*cosTheta*sinA - sinTheta*cosA - cosPhi*cosTheta*sinB))
				output[o] = chars[clamp(N, 0, len(chars)-1)]
			}
		}
	}

	var b strings.Builder
	b.Grow(width*height + height + 2)
	b.WriteString("\x1b[H")
	for y := 0; y < height; y++ {
		start := y * width
		b.WriteString(string(output[start : start+width]))
		b.WriteByte('\n')
	}
	return b.String()
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func init() {
	donutCmd.Flags().IntVar(&donutFrames, "frames", 0, "Render this many frames and exit (default: run until interrupted)")
	rootCmd.AddCommand(donutCmd)
}
