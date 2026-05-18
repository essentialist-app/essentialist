package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"net/url"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	emoast "github.com/yuin/goldmark-emoji/ast"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/renderer"
)

type fyneRenderer struct{}

// Render given AST node to given writer.
func (f *fyneRenderer) Render(w io.Writer, source []byte, n ast.Node) error {
	return fmt.Errorf("Not yet implemented.")
}

// AddOptions adds given option to this renderer.
func (f *fyneRenderer) AddOptions(...renderer.Option) {
}

// NewRichTextFromMarkdownAt configures a RichText widget by parsing the provided markdown content.
// Complete image path with specified pathPrefix.
func NewRichTextFromMarkdownAt(content string, pathPrefix string) *widget.RichText {
	return widget.NewRichText(parseMarkdown(content, pathPrefix)...)
}

// NewRichTextFromMarkdown configures a RichText widget by parsing the provided markdown content.
func NewRichTextFromMarkdown(content string) *widget.RichText {
	return widget.NewRichText(parseMarkdown(content, "")...)
}

type markdownRenderer struct {
	path     string
	segments []widget.RichTextSegment
}

func (m *markdownRenderer) AddOptions(...renderer.Option) {}

func (m *markdownRenderer) Render(_ io.Writer, source []byte, n ast.Node) error {
	segs, err := renderNode(source, n, m.path, false)
	m.segments = segs
	return err
}

func renderNode(source []byte, n ast.Node, path string, blockquote bool) ([]widget.RichTextSegment, error) {
	switch t := n.(type) {
	case *ast.Document:
		return renderChildren(source, n, path, blockquote)
	case *ast.Paragraph:
		children, err := renderChildren(source, n, path, blockquote)
		if !blockquote {
			linebreak := &widget.TextSegment{Style: widget.RichTextStyleParagraph}
			children = append(children, linebreak)
		}
		return children, err
	case *ast.List:
		items, err := renderChildren(source, n, path, blockquote)
		indentation := 0
		for parent := n.Parent(); parent != nil; parent = parent.Parent() {
			if _, ok := parent.(*ast.List); ok {
				indentation++
			}
		}
		return []widget.RichTextSegment{
			&widget.TextSegment{Style: widget.RichTextStyleParagraph, Text: ""},
			&ListSegment{Items: items, IndentationLevel: indentation, Ordered: t.Marker != '*' && t.Marker != '-' && t.Marker != '+'},
		}, err
	case *ast.ListItem:
		texts, err := renderChildren(source, n, path, blockquote)
		return []widget.RichTextSegment{&widget.ParagraphSegment{Texts: texts}}, err
	case *ast.TextBlock:
		return renderChildren(source, n, path, blockquote)
	case *ast.Heading:
		text := forceIntoHeadingText(source, n)
		switch t.Level {
		case 1:
			return []widget.RichTextSegment{&widget.TextSegment{Style: widget.RichTextStyleHeading, Text: text}}, nil
		case 2:
			return []widget.RichTextSegment{&widget.TextSegment{Style: widget.RichTextStyleSubHeading, Text: text}}, nil
		default:
			textSegment := widget.TextSegment{Style: widget.RichTextStyleParagraph, Text: text}
			textSegment.Style.TextStyle.Bold = true
			return []widget.RichTextSegment{&textSegment}, nil
		}
	case *ast.ThematicBreak:
		return []widget.RichTextSegment{&widget.SeparatorSegment{}}, nil
	case *ast.Link:
		link, _ := url.Parse(string(t.Destination))
		text := forceIntoText(source, n)
		return []widget.RichTextSegment{&widget.HyperlinkSegment{Alignment: fyne.TextAlignLeading, Text: text, URL: link}}, nil
	case *ast.CodeSpan:
		text := forceIntoText(source, n)
		return []widget.RichTextSegment{&widget.TextSegment{Style: widget.RichTextStyleCodeInline, Text: text}}, nil
	case *ast.CodeBlock, *ast.FencedCodeBlock:
		var data []byte
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			data = append(data, line.Value(source)...)
		}
		if len(data) == 0 {
			return nil, nil
		}
		if data[len(data)-1] == '\n' {
			data = data[:len(data)-1]
		}
		return []widget.RichTextSegment{&widget.TextSegment{Style: widget.RichTextStyleCodeBlock, Text: string(data)}}, nil
	case *ast.Emphasis:
		text := string(forceIntoText(source, n))
		switch t.Level {
		case 2:
			return []widget.RichTextSegment{&widget.TextSegment{Style: widget.RichTextStyleStrong, Text: text}}, nil
		default:
			return []widget.RichTextSegment{&widget.TextSegment{Style: widget.RichTextStyleEmphasis, Text: text}}, nil
		}
	case *ast.Text:
		text := string(t.Text(source))
		if text == "" {
			// These empty text elements indicate single line breaks after non-text elements in goldmark.
			return []widget.RichTextSegment{&widget.TextSegment{Style: widget.RichTextStyleInline, Text: " "}}, nil
		}
		text = suffixSpaceIfAppropriate(text, n)
		if blockquote {
			return []widget.RichTextSegment{&widget.TextSegment{Style: widget.RichTextStyleBlockquote, Text: text}}, nil
		}
		return []widget.RichTextSegment{&widget.TextSegment{Style: widget.RichTextStyleInline, Text: text}}, nil
	case *ast.Blockquote:
		return renderChildren(source, n, path, true)
	case *ast.Image:
		dest := string(t.Destination)
		u, err := storage.ParseURI(dest)
		if err != nil {
			u = storage.NewFileURI(dest)
		}
		return []widget.RichTextSegment{&widget.ImageSegment{Source: u, Title: string(t.Title), Alignment: fyne.TextAlignCenter}}, nil
	case *east.TableCell:
		segs, err := renderChildren(source, n, path, blockquote)
		if err != nil {
			return nil, err
		}
		return []widget.RichTextSegment{NewTableCell(widget.NewRichText(segs...))}, nil

	case *east.TableHeader:
		segs, err := renderChildren(source, n, path, blockquote)
		if err != nil {
			return nil, err
		}
		cells := make([]*TableCell, len(segs))
		for i, seg := range segs {
			cell, ok := seg.(*TableCell)
			if !ok {
				return nil, fmt.Errorf("Unable to cast element %d to TableCell", i)
			}
			cells[i] = cell
		}
		return []widget.RichTextSegment{&TableRow{cells: cells}}, nil
	case *east.TableRow:
		segs, err := renderChildren(source, n, path, blockquote)
		if err != nil {
			return nil, err
		}
		cells := make([]*TableCell, len(segs))
		for i, seg := range segs {
			cell, ok := seg.(*TableCell)
			if !ok {
				return nil, fmt.Errorf("Unable to cast element %d to TableCell", i)
			}
			cells[i] = cell
		}
		return []widget.RichTextSegment{&TableRow{cells: cells}}, nil
	case *east.Table:
		segs, err := renderChildren(source, n, path, blockquote)
		if err != nil {
			return nil, err
		}
		rows := make([]*TableRow, len(segs))
		for i, seg := range segs {
			row, ok := seg.(*TableRow)
			if !ok {
				return nil, fmt.Errorf("Unable to cast element %d to TableCell", i)
			}
			rows[i] = row
		}
		return []widget.RichTextSegment{NewTableSegment(rows)}, nil
	case *emoast.Emoji:
		text := string(t.Value.Unicode)
		return []widget.RichTextSegment{&widget.TextSegment{Style: widget.RichTextStyleInline, Text: text}}, nil
	case *MathNode:
		dark := isDarkTheme()
		res, err := renderMath(string(t.LaTeX), t.Display, dark)
		if err != nil {
			fallbackText := "$" + string(t.LaTeX) + "$"
			if t.Display {
				fallbackText = "$$" + string(t.LaTeX) + "$$"
			}
			return []widget.RichTextSegment{&widget.TextSegment{Style: widget.RichTextStyleParagraph, Text: fallbackText}}, nil
		}
		return []widget.RichTextSegment{&MathSegment{LaTeX: string(t.LaTeX), Display: t.Display, Resource: res}}, nil
	}
	return nil, nil
}

func suffixSpaceIfAppropriate(text string, n ast.Node) string {
	next := n.NextSibling()
	if next != nil && next.Type() == ast.TypeInline && !strings.HasSuffix(text, " ") {
		return text + " "
	}
	return text
}

func renderChildren(source []byte, n ast.Node, path string, blockquote bool) ([]widget.RichTextSegment, error) {
	children := make([]widget.RichTextSegment, 0, n.ChildCount())
	for childCount, child := n.ChildCount(), n.FirstChild(); childCount > 0; childCount-- {
		if child == nil {
			continue
		}
		segs, err := renderNode(source, child, path, blockquote)
		if err != nil {
			return children, err
		}
		children = append(children, segs...)
		child = child.NextSibling()
	}
	return children, nil
}

func forceIntoText(source []byte, n ast.Node) string {
	texts := make([]string, 0)
	ast.Walk(n, func(n2 ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch t := n2.(type) {
			case *ast.Text:
				texts = append(texts, string(t.Text(source)))
			}
		}
		return ast.WalkContinue, nil
	})
	return strings.Join(texts, " ")
}

func forceIntoHeadingText(source []byte, n ast.Node) string {
	var text strings.Builder
	ast.Walk(n, func(n2 ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch t := n2.(type) {
			case *ast.Text:
				text.Write(t.Text(source))
			}
		}
		return ast.WalkContinue, nil
	})
	return text.String()
}

func parseMarkdown(content string, path string) []widget.RichTextSegment {
	r := markdownRenderer{path: path}
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
			emoji.Emoji,
			&mathExtension{},
		),
		goldmark.WithRenderer(&r))
	err := md.Convert([]byte(content), nil)
	if err != nil {
		fyne.LogError("Failed to parse markdown", err)
	}
	return r.segments
}

type (
	DummyRichTextSegment struct{}
	TableCell            struct {
		widget.BaseWidget
		DummyRichTextSegment
		content  *widget.RichText
		renderer cellRenderer
	}
	TableRow struct {
		DummyRichTextSegment
		cells []*TableCell
	}
	TableSegment struct {
		widget.Table
		DummyRichTextSegment
		rows []*TableRow
		size fyne.Size
	}
)

// DummyRichTextSegment is used by TableRow and TableCell to conform with RichTextSegment.
func (c *DummyRichTextSegment) Inline() bool                    { return false }
func (c *DummyRichTextSegment) Textual() string                 { panic("not implemented") }
func (c *DummyRichTextSegment) Update(fyne.CanvasObject)        { panic("not implemented") }
func (c *DummyRichTextSegment) Visual() fyne.CanvasObject       { panic("not implemented") }
func (c *DummyRichTextSegment) Select(pos1, pos2 fyne.Position) { panic("not implemented") }
func (c *DummyRichTextSegment) SelectedText() string            { panic("not implemented") }
func (c *DummyRichTextSegment) Unselect()                       { panic("not implemented") }

// Cell implements CreateRenderer and draw the underlaying RichTextSegments using RichText.
func (c *TableCell) CreateRenderer() fyne.WidgetRenderer {
	return c.renderer
}

func NewTableCell(content *widget.RichText) *TableCell {
	cell := &TableCell{
		content:  content,
		renderer: NewCellRenderer(content),
	}
	cell.ExtendBaseWidget(cell)
	return cell
}

func (c *TableCell) updateSegment(content *widget.RichText) {
	c.content = content
	c.renderer.setObject(c.content)
}

func NewTableSegment(rows []*TableRow) *TableSegment {
	length := func() (int, int) {
		if len(rows) > 0 {
			return len(rows), len(rows[0].cells)
		}
		return 0, 0
	}
	create := func() fyne.CanvasObject {
		return NewTableCell(widget.NewRichText(&widget.TextSegment{}))
	}
	update := func(pos widget.TableCellID, o fyne.CanvasObject) {
		if pos.Row >= len(rows) || pos.Col >= len(rows[pos.Row].cells) {
			return
		}
		cell := o.(*TableCell)
		cell.updateSegment(rows[pos.Row].cells[pos.Col].content)
	}
	table := &TableSegment{
		Table: widget.Table{
			Length:     length,
			CreateCell: create,
			UpdateCell: update,
		},
		rows: rows,
	}
	table.ExtendBaseWidget(table)
	table.resize()
	return table
}

func (l *TableSegment) resize() {
	// Compute the size of the columns and rows
	widths := []float32{}
	heights := []float32{}
	for i, row := range l.rows {
		for j, cell := range row.cells {
			width := cell.content.MinSize().Width
			height := cell.content.MinSize().Height
			if len(heights) < i+1 {
				heights = append(heights, height)
			} else if heights[i] < height {
				heights[i] = height
			}
			if len(widths) < j+1 {
				widths = append(widths, width)
			} else if widths[j] < width {
				widths[j] = width
			}
		}
	}
	l.size.Height = 0
	for i, height := range heights {
		l.SetRowHeight(i, height)
		l.size.Height += height + 4
	}
	l.size.Width = 0
	for j, width := range widths {
		l.SetColumnWidth(j, width+8)
		l.size.Width += width + 16
	}
}

func (l *TableSegment) Unselect()                       { panic("not implemented") }
func (l *TableSegment) Select(pos1, pos2 fyne.Position) { panic("not implemented") }
func (l *TableSegment) SelectedText() string            { panic("not implemented") }

// MinSize returns the table size otherwise is it minimzed.
func (l *TableSegment) MinSize() fyne.Size {
	return l.size
}

// Visual returns the graphical elements required to render this segment.
func (l *TableSegment) Visual() fyne.CanvasObject {
	return l
}

// Update applies the current state of this table segment to an existing visual.
func (l *TableSegment) Update(o fyne.CanvasObject) {}

// cellRenderer implements fyne.WidgetRenderer. It contains exactly one canvas object.
type cellRenderer []fyne.CanvasObject

func NewCellRenderer(object fyne.CanvasObject) cellRenderer {
	return cellRenderer([]fyne.CanvasObject{object})
}

func (r cellRenderer) setObject(object fyne.CanvasObject) {
	r[0] = object
}

// Destroy does nothing in this implementation.
func (r cellRenderer) Destroy() {
}

// Layout updates the contained object to be the requested size.
func (r cellRenderer) Layout(s fyne.Size) {
	r[0].Resize(s)
}

// MinSize returns the smallest size that this render can use, returned from the underlying object.
func (r cellRenderer) MinSize() fyne.Size {
	return r[0].MinSize()
}

// Objects returns the objects that should be rendered.
func (r cellRenderer) Objects() []fyne.CanvasObject {
	return r
}

// Refresh requests the underlying object to redraw.
func (r cellRenderer) Refresh() {
	r[0].Refresh()
}

// ListSegment includes an itemised list with the content set using the Items field.
//
// Since: 2.1
type ListSegment struct {
	Items            []widget.RichTextSegment
	Ordered          bool
	IndentationLevel int
}

// Inline returns false as a list should be in a block.
func (l *ListSegment) Inline() bool {
	return false
}

// Segments returns the segments required to draw bullets before each item
func (l *ListSegment) Segments() []widget.RichTextSegment {
	out := make([]widget.RichTextSegment, len(l.Items))
	for i, in := range l.Items {
		var txt string
		for j := 0; j < l.IndentationLevel; j++ {
			txt = txt + "\t"
		}
		txt = txt + "• "
		if l.Ordered {
			txt = strconv.Itoa(i+1) + "."
		}
		bullet := &widget.TextSegment{Text: txt + " ", Style: widget.RichTextStyleStrong}
		out[i] = &widget.ParagraphSegment{Texts: []widget.RichTextSegment{
			bullet,
			in,
		}}
	}
	return out
}

// Textual returns no content for a list as the content is in sub-segments.
func (l *ListSegment) Textual() string {
	return ""
}

// Visual returns no additional elements for this segment.
func (l *ListSegment) Visual() fyne.CanvasObject {
	return nil
}

// Update doesnt need to change a list visual.
func (l *ListSegment) Update(fyne.CanvasObject) {
}

// Select does nothing for a list container.
func (l *ListSegment) Select(_, _ fyne.Position) {
}

// SelectedText returns the empty string for this list.
func (l *ListSegment) SelectedText() string {
	return ""
}

// Unselect does nothing for a list container.
func (l *ListSegment) Unselect() {
}

type MathSegment struct {
	LaTeX    string
	Display  bool
	Resource fyne.Resource
}

func (m *MathSegment) Inline() bool {
	return !m.Display
}

func (m *MathSegment) Textual() string {
	if m.Display {
		return "$$" + m.LaTeX + "$$"
	}
	return "$" + m.LaTeX + "$"
}

func (m *MathSegment) Visual() fyne.CanvasObject {
	return newMathWidget(m.Resource)
}

func (m *MathSegment) Update(obj fyne.CanvasObject) {
	if w, ok := obj.(*mathWidget); ok {
		w.SetResource(m.Resource)
	}
}

func (m *MathSegment) Select(pos1, pos2 fyne.Position) {}

func (m *MathSegment) SelectedText() string {
	if m.Display {
		return "$$" + m.LaTeX + "$$"
	}
	return "$" + m.LaTeX + "$"
}

func (m *MathSegment) Unselect() {}

type mathWidget struct {
	widget.BaseWidget
	img  *canvas.Image
	size fyne.Size
}

func newMathWidget(res fyne.Resource) *mathWidget {
	w := &mathWidget{}
	w.ExtendBaseWidget(w)
	w.img = canvas.NewImageFromResource(res)
	w.img.FillMode = canvas.ImageFillContain

	// Decode size
	cfg, _, err := image.DecodeConfig(bytes.NewReader(res.Content()))
	if err == nil {
		scale := float32(mathDPI / 72.0)
		w.size = fyne.NewSize(float32(cfg.Width)/scale, float32(cfg.Height)/scale)
	} else {
		w.size = fyne.NewSize(100, 20)
	}
	return w
}

func (w *mathWidget) CreateRenderer() fyne.WidgetRenderer {
	return &mathWidgetRenderer{img: w.img}
}

func (w *mathWidget) MinSize() fyne.Size {
	return w.size
}

func (w *mathWidget) SetResource(res fyne.Resource) {
	w.img.Resource = res
	cfg, _, err := image.DecodeConfig(bytes.NewReader(res.Content()))
	if err == nil {
		scale := float32(mathDPI / 72.0)
		w.size = fyne.NewSize(float32(cfg.Width)/scale, float32(cfg.Height)/scale)
	}
	w.img.Refresh()
	w.Refresh()
}

type mathWidgetRenderer struct {
	img *canvas.Image
}

func (r *mathWidgetRenderer) Destroy() {}

func (r *mathWidgetRenderer) Layout(s fyne.Size) {
	r.img.Resize(s)
}

func (r *mathWidgetRenderer) MinSize() fyne.Size {
	return r.img.MinSize()
}

func (r *mathWidgetRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.img}
}

func (r *mathWidgetRenderer) Refresh() {
	r.img.Refresh()
}
