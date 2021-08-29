package manager

import (
	"os"
	"strings"
	"text/template"

	"github.com/arnarg/oblish/util"
)

type TagInventory struct {
	Tags map[string][]TagLink
	Vars map[string]interface{}
}

type TagLink struct {
	Title string
	Path  string
}

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

func (nm *noteManager) RenderTags(dest, tpl string, extraVars map[string]interface{}) error {
	if !strings.HasSuffix(dest, "/") {
		dest = dest + "/"
	}

	err := util.CreateDirIfNotExist(dest + "tags/")
	if err != nil {
		return err
	}
	dest = dest + "tags/index.html"

	inv := TagInventory{
		Tags: map[string][]TagLink{},
		Vars: extraVars,
	}

	for tag, pages := range nm.tags {
		inv.Tags[tag] = []TagLink{}
		for _, page := range pages {
			p := TagLink{
				Title: page.Title,
				Path:  page.GetSlug(),
			}
			inv.Tags[tag] = append(inv.Tags[tag], p)
		}
	}

	tmpl, err := template.New("Tags").Parse(tpl)
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

func (nm *noteManager) RenderNotes(dest, tpl string, extraVars map[string]interface{}) error {
	if !strings.HasSuffix(dest, "/") {
		dest = dest + "/"
	}
	if nm.index != nil {
		err := nm.renderNote(dest+"index.html", tpl, nm.index, extraVars)
		if err != nil {
			return err
		}
	}
	for _, note := range nm.notes {
		slug := note.GetSlug()

		err := util.CreateDirIfNotExist(dest + slug)
		if err != nil {
			return err
		}
		err = nm.renderNote(dest+slug+"/index.html", tpl, note, extraVars)
		if err != nil {
			return err
		}
	}
	return nil
}

func (nm *noteManager) renderNote(dest, tpl string, note *Note, extraVars map[string]interface{}) error {

	// Get backlinks to note
	bls := []Backlink{}
	for k, n := range note.Backlinks {
		bls = append(bls, Backlink{
			Title: k,
			Path:  n.GetSlug(),
		})
	}

	body, err := note.Render(nm.markdown.Renderer())
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
