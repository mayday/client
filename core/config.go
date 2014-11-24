package core

import (
	"code.google.com/p/go.crypto/openpgp"
	"errors"
	"fmt"
	goyaml "gopkg.in/yaml.v1"
	"path/filepath"
)

type File struct {
	Path string
}

type Command struct {
	Executable string
}

type Config struct {
	Signed        string
	Raw           string
	Files         []File
	Commands      []Command
	FilesField    []string `yaml:"copy"`
	CommandsField []string `yaml:"run"`
}

func NewConfig(readed string) (*Config, error) {
	config := Config{Raw: readed}

	err := goyaml.Unmarshal([]byte(readed), &config)

	if err != nil {
		return nil, fmt.Errorf("cannot read configuration: %v", err)
	}

	err = ValidateConfig(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Sign(keyid string) error {
	pgp, err := NewPGP()

	if err != nil {
		return err
	}

	signed, err := pgp.Sign(c.Raw, keyid)
	if err != nil {
		return fmt.Errorf("cannot sign configuration: %s", err)
	} else {
		c.Signed = signed
	}

	return nil
}

func (c *Config) Verify(signed string) (*openpgp.Entity, error) {
	pgp, err := NewPGP()

	if err != nil {
		return nil, err
	}

	signature, err := pgp.Verify(c.Raw, signed)
	if err != nil {
		return nil, fmt.Errorf("invalid pgp signature: %s", err)
	}

	return signature, nil
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
