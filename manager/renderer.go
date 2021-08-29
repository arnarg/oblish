package manager

import (
	"os"
	"strings"
	"text/template"

	"github.com/arnarg/oblish/util"
)

type NoteInventory struct {
	Title       string
	Placeholder bool
	Body        string
	Backlinks   []Backlink
	Vars        map[string]interface{}
}

type Backlink struct {
	Title string
	Path  string
}

func (fm *noteManager) RenderNotes(dest, tpl string, extraVars map[string]interface{}) error {
	if !strings.HasSuffix(dest, "/") {
		dest = dest + "/"
	}
	if fm.index != nil {
		err := fm.renderNote(dest+"index.html", tpl, fm.index, extraVars)
		if err != nil {
			return err
		}
	}
	for _, note := range fm.notes {
		slug := note.GetSlug()

		err := util.CreateDirIfNotExist(dest + slug)
		if err != nil {
			return err
		}
		err = fm.renderNote(dest+slug+"/index.html", tpl, note, extraVars)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fm *noteManager) renderNote(dest, tpl string, note *Note, extraVars map[string]interface{}) error {

	// Get backlinks to note
	bls := []Backlink{}
	for k, n := range note.Backlinks {
		bls = append(bls, Backlink{
			Title: k,
			Path:  n.GetSlug(),
		})
	}

	body, err := note.Render(fm.markdown.Renderer())
	if err != nil {
		return err
	}

	// Template note
	inv := NoteInventory{
		Title:       note.Title,
		Placeholder: note.IsPlaceholder(),
		Body:        body,
		Backlinks:   bls,
		Vars:        extraVars,
	}
	tmpl, err := template.New(note.Title).Parse(tpl)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(
		dest,
		os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0644,
	)
	if err != nil {
		return err
	}
	defer file.Close()

	err = tmpl.Execute(file, inv)
	if err != nil {
		return err
	}
	return nil
}
