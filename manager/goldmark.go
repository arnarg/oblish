package manager

import (
	wikilink "github.com/13rac1/goldmark-wikilink"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func newMarkdown() goldmark.Markdown {
	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			meta.Meta,
			wikilink.New(),
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
	return gm
}
