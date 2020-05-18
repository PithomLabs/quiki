package quikirenderer

// Modeled from the html renderer at
// https://github.com/yuin/goldmark/blob/master/renderer/html/html.go
//
// Copyright (c) 2020 Mitchell Cooper
// Copyright (c) 2019 Yusuke Inuzuka
//
// See LICENSE

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// A Renderer struct is an implementation of renderer.NodeRenderer that renders
// nodes as quiki markup.
type Renderer struct {
	Config
	headingLevel int
	braceEscape  bool
	linkDest     string
}

// NewRenderer returns a new Renderer with given options.
func NewRenderer(opts ...Option) renderer.NodeRenderer {
	r := &Renderer{
		Config: NewConfig(),
	}

	for _, opt := range opts {
		opt.SetQuikiOption(&r.Config)
	}
	return r
}

// RegisterFuncs implements NodeRenderer.RegisterFuncs .
func (r *Renderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// blocks

	reg.Register(ast.KindDocument, r.renderDocument)
	reg.Register(ast.KindHeading, r.renderHeading)
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	reg.Register(ast.KindHTMLBlock, r.renderHTMLBlock)
	reg.Register(ast.KindList, r.renderList)
	reg.Register(ast.KindListItem, r.renderListItem)
	reg.Register(ast.KindParagraph, r.renderParagraph)
	reg.Register(ast.KindTextBlock, r.renderTextBlock)
	reg.Register(ast.KindThematicBreak, r.renderThematicBreak)

	// inlines

	reg.Register(ast.KindAutoLink, r.renderAutoLink)
	reg.Register(ast.KindCodeSpan, r.renderCodeSpan)
	reg.Register(ast.KindEmphasis, r.renderEmphasis)
	reg.Register(ast.KindImage, r.renderImage)
	reg.Register(ast.KindLink, r.renderLink)
	reg.Register(ast.KindRawHTML, r.renderRawHTML)
	reg.Register(ast.KindText, r.renderText)
	reg.Register(ast.KindString, r.renderString)
}

func (r *Renderer) renderDocument(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		// head of page

		// partial page means don't include vars
		if r.PartialPage {
			return ast.WalkContinue, nil
		}

		// all the vars minus the title (has to be postposed til it is extracted)
		w.WriteString("@page.author:       Markdown;\n")
		w.WriteString("@page.generator:    quiki/markdown/goldmark;\n")
		w.WriteString("@page.generated;\n\n")

		// table of contents
		if r.TableOfContents {
			w.WriteString("toc{}\n\n")
		}
	} else {
		// foot of page

		// close open sections
		for ; r.headingLevel > 0; r.headingLevel-- {
			w.WriteString("\n}")
		}

		// page title
		if r.PageTitle != "" {
			w.WriteString("\n@page.title: " + quikiEscFmt(r.PageTitle) + ";\n")
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderHeading(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	if entering {

		// if we already have a header of this level open, this
		// terminates it. if we have a header of a lower level (higher
		// number) open, this terminates it and all others up to the
		// biggest level.
		for i := n.Level; i <= r.headingLevel; i++ {
			w.WriteString("\n}\n")
		}

		// e.g. going from # to ###
		if n.Level > r.headingLevel+1 {
			for i := r.headingLevel + 2; i <= n.Level; i++ {
				w.WriteString("~sec{\n")
			}
		}

		// set level, start the section with name opening tag.
		r.headingLevel = n.Level
		w.WriteString("~sec[")

	} else {
		w.WriteString("]")

		// TODO: assume page title as first heading
		// if r.PageTitle == "" {
		// 	r.PageTitle = r.heading
		// }

		// TODO: figure the anchor for github compatibility
		// id := node.HeadingID
		// if node.HeadingID == "" {
		// 	// https://github.com/jch/html-pipeline/blob/master/lib/html/pipeline/toc_filter.rb
		// 	// $section_id =~ tr/A-Z/a-z/;                 # ASCII downcase
		// 	id = strings.ToLower(r.heading)                // downcase
		// 	id = punctuationRegex.ReplaceAllString(id, "") // remove punctuation
		// 	id = strings.Replace(id, " ", "-", -1)         // replace spaces with dashes
		// 	r.heading = ""
		// }

		// // heading ID
		// id = r.ensureUniqueHeadingID(id)
		// if r.HeadingIDPrefix != "" {
		// 	id = r.HeadingIDPrefix + id
		// }
		// if r.HeadingIDSuffix != "" {
		// 	id = id + r.HeadingIDSuffix
		// }
		// r.addText(w, " "+id+"# ")
		w.WriteString("{\n")
	}

	return ast.WalkContinue, nil
}

func (r *Renderer) renderBlockquote(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("~quote{\n")
	} else {
		w.WriteString("\n}\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderCodeBlockLang(lang string, w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {

		// language
		if lang != "" {
			lang = "[" + quikiEscFmt(lang) + "]"
		}

		// extract the code
		var code []byte
		l := n.Lines().Len()
		for i := 0; i < l; i++ {
			line := n.Lines().At(i)
			code = append(code, line.Value(source)...)
		}

		// if there is a closing brace for every opening brace, we can use brace-escape
		braceLevel, braceCount := 0, 0
		for _, c := range code {
			if c == '{' {
				braceLevel++
				braceCount++
			} else if c == '}' {
				braceLevel--
				if braceLevel < 0 {
					break
				}
			}
		}
		r.braceEscape = braceLevel == 0 && braceCount != 0

		if r.braceEscape {
			// use brace-escape
			w.WriteString("~code" + lang + "{{\n")
			w.Write(code)
		} else {
			// can't use brace-escape; escape the code
			w.WriteString("~code" + lang + "{\n")
			w.WriteString(quikiEsc(string(code)))
		}
	} else {
		// closing

		if r.braceEscape {
			w.WriteString("\n}}\n")
			r.braceEscape = false
		} else {
			w.WriteString("\n}\n")
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	return r.renderCodeBlockLang("", w, source, node, entering)
}

func (r *Renderer) renderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	lang := ""
	if entering {
		language := node.(*ast.FencedCodeBlock).Language(source)
		if language != nil {
			lang = string(language)
		}
	}
	return r.renderCodeBlockLang(lang, w, source, node, entering)
}

func (r *Renderer) renderHTMLBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.HTMLBlock)
	if entering {
		if r.Unsafe {
			l := n.Lines().Len()
			for i := 0; i < l; i++ {
				line := n.Lines().At(i)
				w.WriteString("~html{")
				w.WriteString(quikiEsc(string(line.Value(source))))
			}
		} else {
			w.WriteString("/* raw HTML omitted */\n")
		}
	} else {
		if n.HasClosure() {
			if r.Unsafe {
				closure := n.ClosureLine
				w.WriteString(quikiEsc(string(closure.Value(source))))
				w.WriteByte('}')
			} else {
				w.WriteString("/* raw HTML omitted */\n")
			}
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderList(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.List)
	if entering && n.IsOrdered() {
		// TODO: n.IsOrdered() && n.Start != 1
		w.WriteString("~numlist{")
	} else if entering {
		w.WriteString("~list{")
	} else {
		w.WriteString("\n}\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderListItem(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteByte('\n')
	} else {
		w.WriteByte(';')
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderParagraph(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("~p{\n")
	} else {
		w.WriteString("\n}\n")
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTextBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		if _, ok := n.NextSibling().(ast.Node); ok && n.FirstChild() != nil {
			w.WriteByte('\n')
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderThematicBreak(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	// TODO
	return ast.WalkContinue, nil
}

func (r *Renderer) renderAutoLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.AutoLink)
	if !entering {
		return ast.WalkContinue, nil
	}

	// extract label and url
	url := n.URL(source)
	label := n.Label(source)

	// escape link
	if n.AutoLinkType != ast.AutoLinkEmail {
		url = r.addAbsPrefix(url)
	}
	link := quikiEscLink(string(url))

	w.WriteString("[[ ")
	w.WriteString(quikiEscFmt(string(label)))
	w.WriteString(" | ")
	w.WriteString(link)
	w.WriteString(" ]]")
	return ast.WalkContinue, nil
}

func (r *Renderer) renderCodeSpan(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("[c]")
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			segment := c.(*ast.Text).Segment
			value := segment.Value(source)
			if bytes.HasSuffix(value, []byte("\n")) {
				r.Writer.RawWrite(w, value[:len(value)-1])
				if c != n.LastChild() {
					r.Writer.RawWrite(w, []byte(" "))
				}
			} else {
				r.Writer.RawWrite(w, value)
			}
		}
		return ast.WalkSkipChildren, nil
	}
	w.WriteString("[/c]")
	return ast.WalkContinue, nil
}

func (r *Renderer) renderEmphasis(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Emphasis)
	var tag byte = 'i'
	if n.Level == 2 {
		tag = 'b'
	}
	if entering {
		w.Write([]byte{'[', tag, ']'})
	} else {
		w.Write([]byte{'[', '/', tag, ']'})
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderLink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	if entering {
		// TODO: anything we can do with n.Title? quiki doesn't support this atm
		link := string(r.addAbsPrefix(n.Destination))
		link = quikiEscLink(link)
		if hashIdx := strings.IndexByte(link, '#'); hashIdx != -1 {
			r.linkDest = strings.TrimSuffix(link[:hashIdx], ".md") + link[hashIdx:]
		} else {
			r.linkDest = strings.TrimSuffix(link, ".md")
		}
		w.WriteString("[[ ")
	} else {
		w.WriteString(" | " + r.linkDest + " ]]")
		r.linkDest = ""
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)

	dest := n.Destination
	dest = r.addAbsPrefix(dest)
	w.WriteString("~image{\n    file: ")
	if r.Unsafe || !IsDangerousURL(dest) {
		w.WriteString(quikiEscListMapValue(string(util.URLEscape(dest, true))))
	}
	w.WriteString(";\n    alt: ")
	w.WriteString(quikiEscListMapValue(string(n.Text(source))))
	// TODO: can we do anything with n.Title?
	w.WriteString(";\n}\n")

	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderRawHTML(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	if r.Unsafe {
		n := node.(*ast.RawHTML)
		l := n.Segments.Len()
		for i := 0; i < l; i++ {
			segment := n.Segments.At(i)
			w.WriteString("[html:")
			w.WriteString(quikiEscFmt(string(segment.Value(source))))
			w.WriteByte(']')
		}
		return ast.WalkSkipChildren, nil
	}
	w.WriteString("/* raw HTML omitted */")
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Text)
	segment := n.Segment
	if n.IsRaw() {
		r.Writer.RawWrite(w, segment.Value(source))
	} else {
		r.Writer.Write(w, segment.Value(source))
		if n.HardLineBreak() || (n.SoftLineBreak() && r.HardWraps) {
			w.WriteString("[nl]\n")
		} else if n.SoftLineBreak() {
			w.WriteByte('\n')
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderString(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.String)
	if n.IsCode() {
		w.Write(n.Value)
	} else {
		if n.IsRaw() {
			r.Writer.RawWrite(w, n.Value)
		} else {
			r.Writer.Write(w, n.Value)
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) addAbsPrefix(link []byte) []byte {
	if r.AbsolutePrefix != "" && isRelativeLink(link) && link[0] != '.' {
		newDest := r.AbsolutePrefix
		if link[0] != '/' {
			newDest += "/"
		}
		newDest += string(link)
		return []byte(newDest)
	}
	return link
}