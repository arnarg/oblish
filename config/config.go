package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type configFile struct {
	ThemeDirectory *string                `yaml:"themeDirectory"`
	NoteTemplate   *string                `yaml:"noteTemplate"`
	TagsTemplate   *string                `yaml:"tagsTemplate"`
	Vars           map[string]interface{} `yaml:"vars"`
	Copy           []string               `yaml:"copy"`
}

type Config struct {
	NoteTemplate string
	TagsTemplate string
	Vars         map[string]interface{}
	Copy         []copyFile
}

type copyFile struct {
	Base     string
	Relative string
}

func ComputePath(v, c string) string {
	if c != "" {
		return c
	}

	return strings.TrimSuffix(v, "/") + "/.oblish/config.yml"
}

func Load(p string) (*Config, error) {
	oblishConfig := &Config{
		Vars: map[string]interface{}{},
	}
	oblishConfigFile := &configFile{}

	// Get absolute path of config file
	absPath, err := filepath.Abs(p)
	if err != nil {
		return nil, err
	}

	// Check if the file already exists
	_, err = os.Stat(absPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	// Read config file
	yamlString, err := ioutil.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(yamlString, oblishConfigFile)
	if err != nil {
		return nil, err
	}

	if oblishConfigFile.ThemeDirectory != nil {
		themeDirPath, err := computeRelativeToConfigFile(absPath, *oblishConfigFile.ThemeDirectory)
		if err != nil {
			return nil, err
		}
		themeConfigFile := strings.TrimSuffix(themeDirPath, "/") + "/config.yml"
		// Check if theme directory is the same as the current directory to prevent infinite recursion
		if absPath != themeConfigFile {
			oblishConfig, err = Load(themeConfigFile)
			if err != nil {
				return nil, err
			}
		}
	}

	if oblishConfigFile.NoteTemplate != nil {
		noteTemplateFile, err := computeRelativeToConfigFile(absPath, *oblishConfigFile.NoteTemplate)
		if err != nil {
			return nil, err
		}

		noteTemplateString, err := ioutil.ReadFile(noteTemplateFile)
		if err != nil {
			return nil, err
		}

		oblishConfig.NoteTemplate = string(noteTemplateString)
	}

	if oblishConfigFile.TagsTemplate != nil {
		tagsTemplateFile, err := computeRelativeToConfigFile(absPath, *oblishConfigFile.TagsTemplate)
		if err != nil {
			return nil, err
		}

		tagsTemplateString, err := ioutil.ReadFile(tagsTemplateFile)
		if err != nil {
			return nil, err
		}

		oblishConfig.TagsTemplate = string(tagsTemplateString)
	}

	for k, v := range oblishConfigFile.Vars {
		oblishConfig.Vars[k] = v
	}

	for _, path := range oblishConfigFile.Copy {
		oblishConfig.Copy = append(oblishConfig.Copy, copyFile{
			Base:     filepath.Dir(absPath),
			Relative: path,
		})
	}

	return oblishConfig, nil
}

func computeRelativeToConfigFile(c, r string) (string, error) {
	configDir := filepath.Dir(c)
	relativeFile, err := filepath.Abs(configDir + "/" + r)
	if err != nil {
		return "", err
	}
	return relativeFile, nil
}
