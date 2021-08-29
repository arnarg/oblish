package manager

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	wikilink "github.com/13rac1/goldmark-wikilink"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var NoteManager = &noteManager{
	notes:    map[string]*Note{},
	markdown: newMarkdown(),
}

type noteManager struct {
	notes    map[string]*Note // Use unmodified filename (minus md) as key
	markdown goldmark.Markdown
	index    *Note
}

func (fm *noteManager) GetNote(f string) *Note {
	title := strings.TrimSuffix(f, ".md")
	if _, ok := fm.notes[title]; !ok {
		fm.notes[title] = NewNote(title)
	}
	note := fm.notes[title]

	return note
}

func (fm *noteManager) addBacklink(to, from string, note *Note) {
	toNote := fm.GetNote(to)
	toNote.AddBacklink(from, note)
}

func (fm *noteManager) ParseNote(f, p string) (*Note, error) {
	// Check if file exists
	info, err := os.Stat(p)
	if err != nil && info.IsDir() {
		return nil, err
	}

	// Read the file
	content, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}

	// Parse markdown
	context := parser.NewContext()
	node := fm.markdown.Parser().Parse(
		text.NewReader(content),
		parser.WithContext(context),
	)

	// Render markdown
	buf := &bytes.Buffer{}
	err = fm.markdown.Renderer().Render(buf, content, node)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	note := fm.GetNote(f)
	note.Markdown = content
	note.Node = &node
	note.Meta = meta.Get(context)

	if index, ok := note.Meta["index"]; ok {
		if isIndex, ok := index.(bool); ok && isIndex {
			fm.index = note
		}
	}

	return nil, nil
}

func (fm *noteManager) ComputeLinks() error {
	for title, note := range fm.notes {
		if note.Node == nil {
			continue
		}
		// Walk AST to find all backlinks and add them to relevant notes
		err := ast.Walk(
			*note.Node,
			func(n ast.Node, e bool) (ast.WalkStatus, error) {
				// Look for any wikilink in the AST
				if n.Kind() == wikilink.KindWikilink {
					link := n.(*wikilink.Wikilink)
					otherNote := fm.GetNote(string(link.Title))
					// Add backlink to the note being linked
					otherNote.AddBacklink(title, note)

					// Correct wikilink's destination if the
					// other node has a custom slug
					link.Destination = []byte(otherNote.GetSlug())

					// Set a special class to backlinks that don't go
					// anywhere (except the placeholder page)
					if otherNote.IsPlaceholder() {
						// Currently wikilink extensions doesn't render the attributes
						// https://github.com/13rac1/goldmark-wikilink/issues/3
						// TODO review
						link.SetAttributeString("class", "placeholder")
					}
					return ast.WalkSkipChildren, nil
				}

				return ast.WalkContinue, nil
			},
		)
		if err != nil {
			return err
		}
	}
	return nil
}
