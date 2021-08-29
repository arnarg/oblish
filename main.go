package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/arnarg/oblish/config"
	"github.com/arnarg/oblish/manager"
	"github.com/arnarg/oblish/util"
	"github.com/urfave/cli/v2"
)

func run(c *cli.Context) error {
	noteManager := manager.NoteManager
	vaultDir := c.String("vault")
	configPath := c.String("config")
	destinationPath := c.String("destination")

	fi, err := os.Stat(vaultDir)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", vaultDir)
	}
	vaultDir, err = filepath.Abs(vaultDir)
	if err != nil {
		return err
	}

	conf, err := config.Load(config.ComputePath(vaultDir, configPath))
	if err != nil {
		return err
	}

	err = util.CreateDirIfNotExist(destinationPath)
	if err != nil {
		return err
	}

	filepath.WalkDir(vaultDir, func(p string, d fs.DirEntry, err error) error {
		// Skip all dot dirs
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return fs.SkipDir
		}

		// Process a markdown file
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			noteManager.ParseNote(d.Name(), p)
			return nil
		}

		// Process other files
		if !d.IsDir() {
			relativePath := strings.TrimPrefix(p, vaultDir)
			err := util.CopyFile(p, destinationPath+relativePath)
			if err != nil {
				fmt.Println(err)
			}
		}

		return nil
	})

	for _, path := range conf.Copy {
		src, err := filepath.Abs(path.Base + "/" + path.Relative)
		if err != nil {
			return err
		}
		dst, err := filepath.Abs(destinationPath + "/" + path.Relative)
		if err != nil {
			return err
		}
		err = util.RecursiveCopy(src, dst)
		if err != nil {
			return err
		}
	}

	noteManager.ComputeLinks()

	err = noteManager.RenderNotes(destinationPath, conf.NoteTemplate, conf.Vars)
	if err != nil {
		return err
	}

	if conf.TagsTemplate != "" {
		err = noteManager.RenderTags(destinationPath, conf.TagsTemplate, conf.Vars)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	app := cli.App{
		Name:   "oblish",
		Usage:  "Obsidian digital garden static site generator",
		Action: run,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "vault",
				Aliases: []string{"v"},
				Value:   ".",
				Usage:   "filesystem path to vault",
			},
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Value:       "",
				DefaultText: "$vault/.obsidian/config.yml",
				Usage:       "config file",
			},
			&cli.StringFlag{
				Name:    "destination",
				Aliases: []string{"d"},
				Value:   "public",
				Usage:   "filesystem path to write files to",
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
