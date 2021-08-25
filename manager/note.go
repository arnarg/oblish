package manager

import (
	"bytes"
	"fmt"

	"github.com/arnarg/oblish/util"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
)

type Note struct {
	Title     string
	Markdown  []byte
	Node      *ast.Node
	Body      string
	Meta      map[string]interface{}
	Backlinks map[string]*Note
}

func NewNote(t string) *Note {
	return &Note{
		Title:     t,
		Backlinks: map[string]*Note{},
	}
}

func (n *Note) AddBacklink(t string, note *Note) {
	if _, ok := n.Backlinks[t]; !ok {
		n.Backlinks[t] = note
	}
}

func (n *Note) GetSlug() string {
	if n.Meta != nil {
		if slugData, ok := n.Meta["slug"]; ok {
			if slug, ok := slugData.(string); ok {
				return slug
			}
		}
	}
	return "/" + util.CreateNoteSlug(n.Title)
}

func (n *Note) Render(r renderer.Renderer) (string, error) {
	buf := &bytes.Buffer{}
	if n.Node != nil {
		err := r.Render(buf, n.Markdown, *n.Node)
		if err != nil {
			fmt.Println(err)
			return "", err
		}
	}

	return buf.String(), nil
}
