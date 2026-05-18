package main

import (
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	gma "github.com/yuin/goldmark/ast"
	gmt "github.com/yuin/goldmark/text"
)

func parseMathNodes(t *testing.T, md string) []*MathNode {
	t.Helper()
	g := goldmark.New(goldmark.WithExtensions(&mathExtension{}))
	src := []byte(md)
	doc := g.Parser().Parse(gmt.NewReader(src))

	var nodes []*MathNode
	_ = gma.Walk(doc, func(n gma.Node, entering bool) (gma.WalkStatus, error) {
		if !entering {
			return gma.WalkContinue, nil
		}
		if m, ok := n.(*MathNode); ok {
			nodes = append(nodes, m)
		}
		return gma.WalkContinue, nil
	})
	return nodes
}

func TestMathParser_Inline(t *testing.T) {
	tests := []struct {
		name string
		md   string
		want []struct {
			tex     string
			display bool
		}
	}{
		{
			name: "simple inline",
			md:   "What is $E = mc^2$ ?",
			want: []struct {
				tex     string
				display bool
			}{{tex: "E = mc^2", display: false}},
		},
		{
			name: "simple display",
			md:   "Result:\n\n$$\\int_0^\\infty e^{-x}\\,dx = 1$$",
			want: []struct {
				tex     string
				display bool
			}{{tex: "\\int_0^\\infty e^{-x}\\,dx = 1", display: true}},
		},
		{
			name: "currency false-positive: $5 and $10",
			md:   "I have $5 and $10.",
			want: nil,
		},
		{
			name: "two inline math in one paragraph",
			md:   "If $a = 1$ and $b = 2$ then sum is 3.",
			want: []struct {
				tex     string
				display bool
			}{
				{tex: "a = 1", display: false},
				{tex: "b = 2", display: false},
			},
		},
		{
			name: "escaped dollar inside math",
			md:   "The cost is $\\$100$ today.",
			want: []struct {
				tex     string
				display bool
			}{{tex: "\\$100", display: false}},
		},
		{
			name: "math in list item",
			md:   "- inline $x^2$ done\n- more $y_i$ here",
			want: []struct {
				tex     string
				display bool
			}{
				{tex: "x^2", display: false},
				{tex: "y_i", display: false},
			},
		},
		{
			name: "math in emphasis",
			md:   "This is *very $x = 1$ important*.",
			want: []struct {
				tex     string
				display bool
			}{{tex: "x = 1", display: false}},
		},
		{
			name: "multi-line display math",
			md:   "Display:\n\n$$a + b\n= c\n+ d$$\n\nDone.",
			want: []struct {
				tex     string
				display bool
			}{{tex: "a + b\n= c\n+ d", display: true}},
		},
		{
			name: "no closing delimiter is not math",
			md:   "An unclosed $expression here.",
			want: nil,
		},
		{
			name: "space immediately after $ is not math",
			md:   "Costs $ 5 dollars.",
			want: nil,
		},
		{
			name: "space immediately before closing $ is not math",
			md:   "Try $x + 1 $ here.",
			want: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseMathNodes(t, tc.md)
			if len(got) != len(tc.want) {
				t.Fatalf("got %d math nodes, want %d (nodes: %s)",
					len(got), len(tc.want), describeMathNodes(got))
			}
			for i, w := range tc.want {
				if string(got[i].LaTeX) != w.tex {
					t.Errorf("node %d tex: got %q, want %q",
						i, string(got[i].LaTeX), w.tex)
				}
				if got[i].Display != w.display {
					t.Errorf("node %d display: got %v, want %v",
						i, got[i].Display, w.display)
				}
			}
		})
	}
}

func describeMathNodes(nodes []*MathNode) string {
	parts := make([]string, len(nodes))
	for i, n := range nodes {
		mode := "inline"
		if n.Display {
			mode = "display"
		}
		parts[i] = mode + ":" + string(n.LaTeX)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func TestMathParser_DisplayOwnLine(t *testing.T) {
	md := "## Fraction\n\n$$\n\\pi=\\frac{c}{d}\n$$\n"
	nodes := parseMathNodes(t, md)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	t.Logf("latex: %q", string(nodes[0].LaTeX))
}
