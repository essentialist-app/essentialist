// math_renderer.go provides a colorable mtex.Renderer.
//
// The upstream `codeberg.org/go-latex/latex/drawtex/drawimg` package hardcodes
// black ink, which is invisible on Fyne's dark theme. This file ports the
// minimum surface of drawimg and parameterises the glyph color so that the
// math expressions can be rendered in white (dark theme) or black (light
// theme).
package main

import (
	"fmt"
	"image/color"
	"image/png"
	"io"
	"math"

	"codeberg.org/go-latex/latex/drawtex"
	"codeberg.org/go-latex/latex/mtex"
	"git.sr.ht/~sbinet/gg"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

// coloredRenderer implements mtex.Renderer with a configurable foreground
// color. It mirrors `drawimg.Renderer` but exposes the ink color, allowing the
// rendered PNG to use whatever shade the current Fyne theme expects.
type coloredRenderer struct {
	w     io.Writer
	color color.Color
}

func newColoredRenderer(w io.Writer, c color.Color) *coloredRenderer {
	return &coloredRenderer{w: w, color: c}
}

func (r *coloredRenderer) Render(width, height, dpi float64, c *drawtex.Canvas) error {
	w := width * dpi
	h := height * dpi
	ctx := gg.NewContext(int(math.Ceil(w)), int(math.Ceil(h)))

	ctx.SetColor(r.color)

	for _, op := range c.Ops() {
		switch op := op.(type) {
		case drawtex.GlyphOp:
			r.drawGlyph(ctx, dpi, op)
		case drawtex.RectOp:
			drawRect(ctx, dpi, op)
		default:
			return fmt.Errorf("unknown drawtex op %T", op)
		}
	}

	return png.Encode(r.w, ctx.Image())
}

func (r *coloredRenderer) drawGlyph(ctx *gg.Context, dpi float64, op drawtex.GlyphOp) {
	face, err := opentype.NewFace(op.Glyph.Font, &opentype.FaceOptions{
		DPI:     dpi,
		Size:    op.Glyph.Size,
		Hinting: font.HintingNone,
	})
	if err != nil {
		// Glyph fonts are bundled at compile time so this should be
		// unreachable in production; bail silently instead of panicking
		// so a malformed expression cannot crash the app.
		return
	}
	defer face.Close()
	ctx.SetFontFace(face)

	dpiScale := dpi / 72
	x := op.X * dpiScale
	y := op.Y * dpiScale
	ctx.DrawString(op.Glyph.Symbol, x, y)
}

func drawRect(ctx *gg.Context, dpi float64, op drawtex.RectOp) {
	dpiScale := dpi / 72
	ctx.NewSubPath()
	ctx.MoveTo(op.X1*dpiScale, op.Y1*dpiScale)
	ctx.LineTo(op.X2*dpiScale, op.Y1*dpiScale)
	ctx.LineTo(op.X2*dpiScale, op.Y2*dpiScale)
	ctx.LineTo(op.X1*dpiScale, op.Y2*dpiScale)
	ctx.LineTo(op.X1*dpiScale, op.Y1*dpiScale)
	ctx.ClosePath()
	ctx.Fill()
}

var _ mtex.Renderer = (*coloredRenderer)(nil)
