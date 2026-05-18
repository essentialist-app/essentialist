// Copyright ©2020 The go-latex Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

package drawimg_test

import (
	"bytes"
	"os"
	"testing"

	"codeberg.org/go-fonts/latin-modern/lmroman12bold"
	"codeberg.org/go-fonts/latin-modern/lmroman12italic"
	"codeberg.org/go-fonts/latin-modern/lmroman12regular"
	"codeberg.org/go-fonts/liberation/liberationsansbold"
	"codeberg.org/go-fonts/liberation/liberationsansbolditalic"
	"codeberg.org/go-fonts/liberation/liberationsansitalic"
	"codeberg.org/go-fonts/liberation/liberationsansregular"
	"codeberg.org/go-fonts/stix/stix2mathregular"
	"codeberg.org/go-fonts/stix/stix2textbold"
	"codeberg.org/go-fonts/stix/stix2textbolditalic"
	"codeberg.org/go-fonts/stix/stix2textitalic"
	"codeberg.org/go-fonts/stix/stix2textregular"
	"codeberg.org/go-latex/latex/drawtex/drawimg"
	"codeberg.org/go-latex/latex/font/ttf"
	"codeberg.org/go-latex/latex/mtex"
	"golang.org/x/image/font/sfnt"
)

func TestRenderer(t *testing.T) {
	const (
		size = 12
		dpi  = 256
	)

	load := func(name string) []byte {
		name = "testdata/" + name + "_golden.png"
		raw, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("could not read file %q: %+v", name, err)
		}
		return raw
	}

	fonts := map[string]*ttf.Fonts{
		"gofont":     nil,
		"lmroman":    lmromanFonts(t),
		"stix":       stixFonts(t),
		"liberation": liberationFonts(t),
	}

	for _, tc := range []struct {
		name string
		expr string
	}{
		{
			name: "func",
			expr: `$f(x)=ax+b$`,
		},
		{
			name: "sqrt",
			expr: `$\sqrt{x}$`,
		},
		{
			name: "sqrt_over_2pi",
			expr: `$\frac{\sqrt{x+20}}{2\pi}$`,
		},
		{
			name: "delta",
			expr: `$\delta x \neq \frac{\sqrt{x+20}}{2\Delta}$`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, font := range []string{
				"gofont",
				"lmroman",
				"stix",
				"liberation",
			} {
				t.Run(font, func(t *testing.T) {
					out := new(bytes.Buffer)
					dst := drawimg.NewRenderer(out)
					err := mtex.Render(dst, tc.expr, size, dpi, fonts[font])
					if err != nil {
						t.Fatalf("could not render expression %q: %+v", tc.expr, err)
					}

					name := font + "_" + tc.name
					if got, want := out.Bytes(), load(name); !bytes.Equal(got, want) {
						err := os.WriteFile("testdata/"+name+".png", got, 0644)
						if err != nil {
							t.Fatalf("could not create output file: %+v", err)
						}
						t.Fatal("files differ")
					}
				})
			}
		})
	}
}

func lmromanFonts(t *testing.T) *ttf.Fonts {
	rm, err := sfnt.Parse(lmroman12regular.TTF)
	if err != nil {
		t.Fatalf("could not parse fonts: %+v", err)
	}

	it, err := sfnt.Parse(lmroman12italic.TTF)
	if err != nil {
		t.Fatalf("could not parse fonts: %+v", err)
	}

	bf, err := sfnt.Parse(lmroman12bold.TTF)
	if err != nil {
		t.Fatalf("could not parse fonts: %+v", err)
	}

	return &ttf.Fonts{
		Default: rm,
		Rm:      rm,
		It:      it,
		Bf:      bf,
		BfIt:    nil,
	}
}

func stixFonts(t *testing.T) *ttf.Fonts {
	rm, err := sfnt.Parse(stix2mathregular.TTF)
	if err != nil {
		t.Fatalf("could not parse fonts: %+v", err)
	}

	def, err := sfnt.Parse(stix2textregular.TTF)
	if err != nil {
		t.Fatalf("could not parse fonts: %+v", err)
	}

	it, err := sfnt.Parse(stix2textitalic.TTF)
	if err != nil {
		t.Fatalf("could not parse fonts: %+v", err)
	}

	bf, err := sfnt.Parse(stix2textbold.TTF)
	if err != nil {
		t.Fatalf("could not parse fonts: %+v", err)
	}

	bfit, err := sfnt.Parse(stix2textbolditalic.TTF)
	if err != nil {
		t.Fatalf("could not parse fonts: %+v", err)
	}

	return &ttf.Fonts{
		Default: def,
		Rm:      rm,
		It:      it,
		Bf:      bf,
		BfIt:    bfit,
	}
}

func liberationFonts(t *testing.T) *ttf.Fonts {
	rm, err := sfnt.Parse(liberationsansregular.TTF)
	if err != nil {
		t.Fatalf("could not parse fonts: %+v", err)
	}

	it, err := sfnt.Parse(liberationsansitalic.TTF)
	if err != nil {
		t.Fatalf("could not parse fonts: %+v", err)
	}

	bf, err := sfnt.Parse(liberationsansbold.TTF)
	if err != nil {
		t.Fatalf("could not parse fonts: %+v", err)
	}

	bfit, err := sfnt.Parse(liberationsansbolditalic.TTF)
	if err != nil {
		t.Fatalf("could not parse fonts: %+v", err)
	}

	return &ttf.Fonts{
		Default: rm,
		Rm:      rm,
		It:      it,
		Bf:      bf,
		BfIt:    bfit,
	}
}
