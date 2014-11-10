package core

import (
	"errors"
	"fmt"
	goyaml "gopkg.in/yaml.v1"
	"io/ioutil"
	"os"
	"path/filepath"
)

type File struct {
	Path string
}

type Command struct {
	Executable string
}

type Config struct {
	Path          string
	Files         []File
	Commands      []Command
	FilesField    []string `yaml:"copy"`
	CommandsField []string `yaml:"run"`
}

func NewConfig(path string) (*Config, error) {
	config := Config{}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("Cannot find configuration path: %s", path)
	}

	readed, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Cannot read configuration file")
	}

	err = goyaml.Unmarshal(readed, &config)
	if err != nil {
		return nil, fmt.Errorf("cannot read configuration: %v", err)
	} else {
		config.Path = path
	}

	err = ValidateConfig(&config)

	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) GetFiles() ([]File, error) {
	if len(c.FilesField) == 0 {
		return nil, errors.New("Not defined Files")
	}

	for _, file := range c.FilesField {
		files, err := filepath.Glob(file)
		if err != nil {
			c.Files = append(c.Files, File{Path: file})
		} else {
			for _, ff := range files {
				c.Files = append(c.Files, File{Path: ff})
			}
		}
	}

	return c.Files, nil
}

func (c *Config) GetCommands() ([]Command, error) {
	if len(c.CommandsField) == 0 {
		return nil, errors.New("Not defined commands")
	}

	for _, command := range c.CommandsField {
		c.Commands = append(c.Commands, Command{Executable: command})
	}

	return c.Commands, nil
}

func ValidateConfig(c *Config) error {
	var err error

	if _, err = c.GetFiles(); err != nil {
		return err
	}

	if _, err = c.GetCommands(); err != nil {
		return err
	}

	return nil
}
