package main

import (
	"github.com/yuin/goldmark"
	gma "github.com/yuin/goldmark/ast"
	gmp "github.com/yuin/goldmark/parser"
	gmt "github.com/yuin/goldmark/text"
	gmu "github.com/yuin/goldmark/util"
)

type MathNode struct {
	gma.BaseInline

	LaTeX   []byte
	Display bool
}

var kindMath = gma.NewNodeKind("Math")

func (n *MathNode) Kind() gma.NodeKind { return kindMath }

func (n *MathNode) Dump(source []byte, level int) {
	mode := "inline"
	if n.Display {
		mode = "display"
	}
	gma.DumpHelper(n, source, level, map[string]string{
		"mode": mode,
		"tex":  string(n.LaTeX),
	}, nil)
}

type mathInlineParser struct{}

func (p *mathInlineParser) Trigger() []byte { return []byte{'$'} }

// Parse recognises Pandoc-style TeX delimiters: `$...$` for inline math and
// `$$...$$` for display math. Rules:
//   - `$` opening must not be followed by whitespace.
//   - `$` closing must not be preceded by whitespace (unless escaped).
//   - `\$` is a literal dollar sign and must not be treated as a delimiter.
//   - `$$...$$` may span multiple lines.
//
// This mirrors the well-tested approach in `goldmark-qjs-katex/qjskatex.go`.
func (p *mathInlineParser) Parse(parent gma.Node, block gmt.Reader, pc gmp.Context) gma.Node {
	src := block.Source()
	ln, pos := block.Position()
	lStart, lEnd := pos.Start, pos.Stop
	line := src[lStart:lEnd]

	if len(line) < 2 {
		return nil
	}

	var (
		display      bool
		start, end   int
		consumeAfter int
	)

	switch {
	case line[1] == '$':
		display = true
		start = lStart + 2
		offset := 2
		for end == 0 {
			for c := offset; c < len(line); c++ {
				if line[c] == '\\' {
					c++
					continue
				}
				if line[c] == '$' {
					c++
					if c == len(line) {
						break
					}
					if line[c] == '$' {
						end = lStart + c - 1
						consumeAfter = 2
						break
					}
				}
			}
			if end != 0 {
				break
			}
			if lEnd == len(src) {
				break
			}
			rest := src[lEnd:]
			c := 1
			for c < len(rest) && rest[c] != '\n' {
				c++
			}
			if isBlankSegment(rest[:c]) {
				break
			}
			lStart = lEnd
			lEnd = lStart + c
			line = src[lStart:lEnd]
			ln++
			offset = 0
		}

	case !gmu.IsSpace(line[1]):
		display = false
		start = lStart + 1
		for end == 0 {
			for c := 1; c < len(line); c++ {
				if line[c] == '\\' {
					c++
					continue
				}
				if line[c] == '$' {
					if !gmu.IsSpace(line[c-1]) {
						end = lStart + c
						consumeAfter = 1
						break
					}
				}
			}
			if end != 0 {
				break
			}
			if lEnd == len(src) {
				break
			}
			rest := src[lEnd:]
			c := 0
			for c < len(rest) && rest[c] != '\n' {
				c++
			}
			if isBlankSegment(rest[:c]) {
				break
			}
			lStart = lEnd
			lEnd = lStart + c
			line = src[lStart:lEnd]
			ln++
		}

	default:
		return nil
	}

	if start >= end {
		return nil
	}

	tex := src[start:end]

	newPos := end + consumeAfter
	if newPos < lEnd {
		block.SetPosition(ln, gmt.NewSegment(newPos, lEnd))
	} else {
		block.Advance(end + consumeAfter - pos.Start)
	}

	n := &MathNode{Display: display}
	n.LaTeX = append(n.LaTeX, tex...)
	return n
}

func isBlankSegment(s []byte) bool {
	for _, c := range s {
		if !gmu.IsSpace(c) {
			return false
		}
	}
	return true
}

type mathExtension struct{}

func (e *mathExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		gmp.WithInlineParsers(gmu.PrioritizedValue{Value: &mathInlineParser{}, Priority: 150}),
	)
}
