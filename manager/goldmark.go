package manager

import (
	wikilink "github.com/13rac1/goldmark-wikilink"
	"github.com/arnarg/oblish/util"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	gmutil "github.com/yuin/goldmark/util"
)

type noteLinkNormalizer struct{}

func (n *noteLinkNormalizer) Normalize(l string) string {
	return "/" + util.CreateNoteSlug(l) + "/"
}

func newMarkdown() goldmark.Markdown {
	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			meta.Meta,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
	// Add wikilink parser
	gm.Parser().AddOptions(
		parser.WithInlineParsers(
			gmutil.Prioritized(
				wikilink.NewParser().WithNormalizer(
					&noteLinkNormalizer{},
				),
				102,
			),
		),
	)
	// Add wikilink renderer
	gm.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			gmutil.Prioritized(
				wikilink.NewHTMLRenderer(),
				500,
			),
		),
	)
	return gm
}
