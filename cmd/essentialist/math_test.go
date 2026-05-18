package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"fyne.io/fyne/v2"
)

func TestRenderMath_ProducesPNG(t *testing.T) {
	res, err := renderMath("E = mc", false, false)
	if err != nil {
		t.Fatalf("renderMath returned error: %v", err)
	}
	if res == nil || len(res.StaticContent) == 0 {
		t.Fatalf("renderMath produced empty resource")
	}
	if !isPNG(res.StaticContent) {
		t.Errorf("renderMath output does not start with PNG magic bytes")
	}
}

func TestRenderMath_CacheHit(t *testing.T) {
	a, err := renderMath("a + b", false, false)
	if err != nil {
		t.Fatalf("renderMath(a, false, false) err: %v", err)
	}
	b, err := renderMath("a + b", false, false)
	if err != nil {
		t.Fatalf("renderMath(a, false, false) err: %v", err)
	}
	if a != b {
		t.Errorf("expected identical pointer from cache on second call, got distinct resources")
	}
}

func TestRenderMath_ThemeSeparated(t *testing.T) {
	light, _ := renderMath("x + y", false, false)
	dark, _ := renderMath("x + y", false, true)
	if light == nil || dark == nil {
		t.Fatalf("renderMath returned nil resource")
	}
	if light == dark {
		t.Errorf("expected distinct resources for light vs dark theme")
	}
}

func TestRenderMath_DisplaySeparated(t *testing.T) {
	inline, _ := renderMath("\\sum x", false, false)
	display, _ := renderMath("\\sum x", true, false)
	if inline == display {
		t.Errorf("expected distinct resources for inline vs display mode")
	}
}

func TestRenderMath_InkColorMatchesTheme(t *testing.T) {
	light, err := renderMath("X", false, false)
	if err != nil {
		t.Fatalf("renderMath light err: %v", err)
	}
	dark, err := renderMath("X", false, true)
	if err != nil {
		t.Fatalf("renderMath dark err: %v", err)
	}
	if !hasInk(t, light.StaticContent, color.RGBA{0, 0, 0, 255}) {
		t.Errorf("light theme output should contain black pixels")
	}
	if !hasInk(t, dark.StaticContent, color.RGBA{255, 255, 255, 255}) {
		t.Errorf("dark theme output should contain white pixels")
	}
}

func isPNG(b []byte) bool {
	if len(b) < 8 {
		return false
	}
	magic := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
	for i, m := range magic {
		if b[i] != m {
			return false
		}
	}
	return true
}

func hasInk(t *testing.T, raw []byte, want color.Color) bool {
	t.Helper()
	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("png decode err: %v", err)
	}
	return imageContains(img, want)
}

func imageContains(img image.Image, want color.Color) bool {
	wr, wg, wb, wa := want.RGBA()
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			if a == 0 {
				continue
			}
			if r == wr && g == wg && bl == wb && a == wa {
				return true
			}
		}
	}
	return false
}

func TestParseMarkdown_ProducesMathSegment(t *testing.T) {
	segments := parseMarkdown("Here is $a + b$ now.", "")
	var mathSeg *MathSegment
	for _, seg := range segments {
		if ms, ok := seg.(*MathSegment); ok {
			mathSeg = ms
			break
		}
	}

	if mathSeg == nil {
		t.Fatalf("markdown parsing did not produce a MathSegment")
	}

	if mathSeg.LaTeX != "a + b" {
		t.Errorf("expected MathSegment LaTeX to be 'a + b', got %q", mathSeg.LaTeX)
	}

	if mathSeg.Inline() != true {
		t.Errorf("expected inline MathSegment to be Inline() == true")
	}

	if mathSeg.SelectedText() != "$a + b$" {
		t.Errorf("expected SelectedText() to be '$a + b$', got %q", mathSeg.SelectedText())
	}

	// Verify selection methods are safe no-ops and don't panic
	mathSeg.Select(fyne.NewPos(0, 0), fyne.NewPos(10, 10))
	mathSeg.Unselect()
}

func TestRenderMath_RecoversPanic(t *testing.T) {
	_, err := renderMath("\\substack{0<i<m}", false, false)
	if err == nil {
		t.Errorf("expected error from unknown macro substack, got nil")
	}
}

func TestRenderMathsDeck(t *testing.T) {
	equations := []struct {
		expr    string
		display bool
	}{
		{expr: "a + b = c", display: false},
		{expr: "c = \\pm\\sqrt{a^2 + b^2}", display: false},
		{expr: "\\begin{Bmatrix} a & b \\\\ c & d \\end{Bmatrix}", display: true},
		{expr: "\\pi=\\frac c d", display: false},
		{expr: "\\pi=\\frac{c}{d}", display: false},
		{expr: "\n\\pi=\\frac{c}{d}\n", display: true},
		{expr: "\\sum_{\\substack{0<i<m\\\\0<j<n}}", display: true},
		{expr: "\\sum_{0 < i < m,\\ 0 < j < n}", display: true},
	}

	for _, eq := range equations {
		res, err := renderMath(eq.expr, eq.display, false)
		if err != nil {
			t.Logf("EXPR %q (display: %v) -> FAILED: %v", eq.expr, eq.display, err)
		} else {
			t.Logf("EXPR %q (display: %v) -> SUCCESS (resource size: %d bytes)", eq.expr, eq.display, len(res.StaticContent))
		}
	}
}

func TestRenderMath_SummationLimits(t *testing.T) {
	testCases := []struct {
		name string
		expr string
	}{
		{name: "Subscript limit only", expr: "\\sum_{i=1}"},
		{name: "Superscript limit only", expr: "\\sum^{\\infty}"},
		{name: "Both limits sub-first", expr: "\\sum_{i=1}^{\\infty}"},
		{name: "Both limits sup-first", expr: "\\sum^{\\infty}_{i=1}"},
		{name: "Function limit lim", expr: "\\lim_{x \\to \\infty}"},
		{name: "Function limit max", expr: "\\max_{s \\in S}"},
		{name: "Full equation", expr: "\\sum_{i=1}^{\\infty} \\frac{1}{n^2} = \\frac{\\pi^2}{6}"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := renderMath(tc.expr, true, false)
			if err != nil {
				t.Fatalf("Failed to render math expression %q: %v", tc.expr, err)
			}
			if res == nil || len(res.StaticContent) == 0 {
				t.Fatalf("Produced empty resource for %q", tc.expr)
			}
			if !isPNG(res.StaticContent) {
				t.Errorf("Result for %q is not a valid PNG", tc.expr)
			}
		})
	}
}

