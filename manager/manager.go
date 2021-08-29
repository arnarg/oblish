package manager

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
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
	tags:     map[string][]*Note{},
	markdown: newMarkdown(),
}

type noteManager struct {
	notes    map[string]*Note
	tags     map[string][]*Note
	markdown goldmark.Markdown
	index    *Note
}

func (nm *noteManager) GetNote(f string) *Note {
	title := strings.TrimSuffix(f, ".md")
	if _, ok := nm.notes[title]; !ok {
		nm.notes[title] = NewNote(title)
	}
	note := nm.notes[title]

	return note
}

func (nm *noteManager) processTagsForNote(note *Note, tags interface{}) {
	var noteTags []string
	// If tags is just a string we split it on commas and/or space as
	// delimiter
	if tags, ok := tags.(string); ok {
		delim := regexp.MustCompile(`[, ]+`)
		noteTags = delim.Split(tags, -1)
	}
	// If tags is an slice of interface{} we check if each item is a string
	// if so we add it to the final noteTags slice
	if tags, ok := tags.([]interface{}); ok {
		for _, tag := range tags {
			if tag, ok := tag.(string); ok {
				noteTags = append(noteTags, tag)
			}
		}
	}

	for _, tag := range noteTags {
		nm.tags[tag] = append(nm.tags[tag], note)
	}
}

func (nm *noteManager) ParseNote(f, p string) (*Note, error) {
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
	node := nm.markdown.Parser().Parse(
		text.NewReader(content),
		parser.WithContext(context),
	)

	// Render markdown
	buf := &bytes.Buffer{}
	err = nm.markdown.Renderer().Render(buf, content, node)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	note := nm.GetNote(f)
	note.Markdown = content
	note.Node = &node
	note.Meta = meta.Get(context)

	if index, ok := note.Meta["index"]; ok {
		if isIndex, ok := index.(bool); ok && isIndex {
			nm.index = note
		}
	}

	if tags, ok := note.Meta["tags"]; ok {
		nm.processTagsForNote(note, tags)
	}

	return nil, nil
}

func (nm *noteManager) ComputeLinks() error {
	for title, note := range nm.notes {
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
					otherNote := nm.GetNote(string(link.Title))
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
