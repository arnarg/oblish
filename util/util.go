package util

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func CreateNoteSlug(n string) string {
	stripped := strings.TrimSuffix(n, ".md")
	lowered := strings.ToLower(stripped)
	reg := regexp.MustCompile("[^a-z0-9 ]")
	noEmoji := reg.ReplaceAllString(lowered, "")
	noSpaces := strings.ReplaceAll(strings.TrimSpace(noEmoji), " ", "-")
	return noSpaces
}

func CreateDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func CopyFile(src, dst string) error {
	dir := filepath.Dir(dst)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func RecursiveCopy(src, dst string) error {
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return CopyFile(src, dst)
	}

	filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			relativePath := strings.TrimPrefix(p, src)
			err := CopyFile(p, dst+"/"+relativePath)
			if err != nil {
				fmt.Println(err)
			}
		}
		return nil
	})
	return nil
}
