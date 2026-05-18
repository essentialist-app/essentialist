package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image/color"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"codeberg.org/go-latex/latex/font/liberation"
	"codeberg.org/go-latex/latex/font/ttf"
	"codeberg.org/go-latex/latex/mtex"
)

const (
	mathFontSize = 20.0
	mathDPI      = 72.0
)

var (
	mathFontsOnce sync.Once
	mathFonts     *ttf.Fonts

	mathCache sync.Map
)

func loadMathFonts() *ttf.Fonts {
	mathFontsOnce.Do(func() {
		mathFonts = liberation.Fonts()
	})
	return mathFonts
}

func mathCacheKey(expr string, display, dark bool) string {
	var b [2]byte
	if display {
		b[0] = 'D'
	} else {
		b[0] = 'I'
	}
	if dark {
		b[1] = 'd'
	} else {
		b[1] = 'l'
	}
	h := sha1.New()
	h.Write(b[:])
	h.Write([]byte{'|'})
	h.Write([]byte(expr))
	return hex.EncodeToString(h.Sum(nil))
}

func renderMath(expr string, display, dark bool) (res *fyne.StaticResource, err error) {
	expr = strings.TrimSpace(expr)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("math render panic: %v", r)
		}
	}()

	key := mathCacheKey(expr, display, dark)
	if v, ok := mathCache.Load(key); ok {
		return v.(*fyne.StaticResource), nil
	}

	wrapped := "$" + expr + "$"

	ink := color.Black
	if dark {
		ink = color.White
	}

	var buf bytes.Buffer
	dst := newColoredRenderer(&buf, ink)
	if err = mtex.Render(dst, wrapped, mathFontSize, mathDPI, loadMathFonts()); err != nil {
		return nil, err
	}

	res = fyne.NewStaticResource(key+".png", buf.Bytes())
	actual, _ := mathCache.LoadOrStore(key, res)
	return actual.(*fyne.StaticResource), nil
}

func isDarkTheme() bool {
	app := fyne.CurrentApp()
	if app == nil {
		return false
	}
	themeName := getThemeName()
	if themeName == "light" {
		return false
	}
	if themeName == "dark" {
		return true
	}
	return app.Settings().ThemeVariant() == theme.VariantDark
}
