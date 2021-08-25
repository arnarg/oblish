package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type configFile struct {
	NoteTemplate string                 `yaml:"noteTemplate"`
	Vars         map[string]interface{} `yaml:"vars"`
}

type Config struct {
	NoteTemplate string
	Vars         map[string]interface{}
}

func ComputePath(v, c string) string {
	if c != "" {
		return c
	}

	return strings.TrimSuffix(v, "/") + "/.oblish/config.yml"
}

func Load(p string) (*Config, error) {
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

	noteTemplateFile, err := computeRelativeToConfigFile(absPath, oblishConfigFile.NoteTemplate)
	if err != nil {
		return nil, err
	}

	noteTemplateString, err := ioutil.ReadFile(noteTemplateFile)
	if err != nil {
		return nil, err
	}

	return &Config{
		NoteTemplate: string(noteTemplateString),
		Vars:         oblishConfigFile.Vars,
	}, nil
}

func computeRelativeToConfigFile(c, r string) (string, error) {
	configDir := filepath.Dir(c)
	relativeFile, err := filepath.Abs(configDir + "/" + r)
	if err != nil {
		return "", err
	}
	return relativeFile, nil
}
