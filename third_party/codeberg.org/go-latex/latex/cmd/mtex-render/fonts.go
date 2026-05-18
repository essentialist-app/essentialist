// Copyright ©2020 The go-latex Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"codeberg.org/go-latex/latex/font/liberation"
	"codeberg.org/go-latex/latex/font/lm"
	"codeberg.org/go-latex/latex/font/ttf"
)

const useLiberation = true

func liberationFonts() *ttf.Fonts {
	return liberation.Fonts()
}

func lmromanFonts() *ttf.Fonts {
	return lm.Fonts()
}
