package main

import (
	"errors"
	"fmt"
	goyaml "gopkg.in/yaml.v1"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
	"sync"
	//	"net/http"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"time"
)

const (
	DefaultAPIBaseURL = "https://mayday.api/"
)

type Mayday struct {
	Configuration *Config
	Hostname      string
	ReportsPath   string
}

func NewMayday(config *Config) (*Mayday, error) {
	mayday := Mayday{
		Configuration: config,
	}

	reportsPath, err := mayday.GetDefaultReportsPath()

	if err != nil {
		return nil, err
	} else {
		mayday.ReportsPath = reportsPath
	}

	return &mayday, nil
}

func (m *Mayday) GetDefaultReportsPath() (string, error) {
	usr, err := user.Current()

	if err != nil {
		return "", err
	}

	base := path.Join(usr.HomeDir, ".mayday", "reports")

	if _, err := os.Stat(base); os.IsNotExist(err) {
		err := os.MkdirAll(base, 0700)
		if err != nil {
			return "", err
		}
	}

	return base, nil
}

func (m *Mayday) CreateReportTempDirectory() (string, error) {
	hostname, err := os.Hostname()

	if err != nil {
		return "", err
	}

	current := time.Now().Local()

	//TODO: Make the report format configurable via CLI
	reportPath, err := ioutil.TempDir(m.ReportsPath,
		fmt.Sprintf("%s-%d-%d-%d-", hostname, current.Year(), current.Month(), current.Day()))

	if err != nil {
		return "", err
	}

	return reportPath, nil
}

func (m *Mayday) RunCommand(reportPath string, command Command, wg *sync.WaitGroup) {
	ran, _ := exec.Command("/bin/bash", "-c", command.Executable).Output()
	outfile, _ := os.Create(command.GetFileName(reportPath))
	outfile.WriteString(string(ran))
	defer outfile.Close()
	defer wg.Done()
}

func (m *Mayday) Run() error {
	reportPath, err := m.CreateReportTempDirectory()

	if err != nil {
		return err
	}

	wg := new(sync.WaitGroup)

	fmt.Println(reportPath)

	for _, command := range m.Configuration.Commands {
		wg.Add(1)
		go m.RunCommand(reportPath, command, wg)
	}

	wg.Wait()

	for _, file := range m.Configuration.Files {
		finfo, err := os.Stat(file.Path)
		if err != nil {
			fmt.Println("Cannot stat file")
		} else {
			if finfo.IsDir() {
				CopyDir(file.Path, path.Join(reportPath, file.Path))
			} else {
				CopyFile(file.Path, reportPath)
			}
		}
	}

	return nil
}

type File struct {
	Path string
}

func CopyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}

	}

	return
}

func CopyDir(source string, dest string) (err error) {

	// get properties of source dir
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// create dest dir

	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)

	objects, err := directory.Readdir(-1)

	for _, obj := range objects {
		sourcefilepointer := source + "/" + obj.Name()
		destinationfilepointer := dest + "/" + obj.Name()

		if obj.IsDir() {
			// create sub-directories - recursively
			err = CopyDir(sourcefilepointer, destinationfilepointer)
			if err != nil {
				//TODO: Logging
				fmt.Println(err)
			}
		} else {
			//TODO: Logging
			err = CopyFile(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		}

	}
	return
}

type Command struct {
	Executable string
}

func MangleCommand(command string) string {
	//Ported from https://github.com/sosreport/sos/blob/48a99c95078bab306cb56bb1a05420d88bf15a64/sos/plugins/__init__.py
	regexes := map[string](string){
		"^/(usr/|)(bin|sbin)/": "",
		"[^\\w\\-\\.\\/]+":     "_",
		"/":                    ".",
		"[\\.|_|-|\\t]+":       "",
	}

	for regex, replace := range regexes {
		command = regexp.MustCompile(regex).ReplaceAllLiteralString(command, replace)
	}

	return command
}

func (c *Command) GetFileName(Base string) string {
	return path.Join(Base, MangleCommand(c.Executable))
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

// func NewConfigFromURL(id string, authCode string, apiURL string) (*Config, error) {
// 	client := &http.Client{}

// 	if apiURL == "" {

// 		apiURL = DefaultAPIURL
// 	}

// 	request, err := http.NewRequest("GET",
// 		fmt.Sprintf("%s/report/config", apiURL), nil)

// 	if err != nil {
// 		return nil, err
// 	}

// 	if authCode != "" {
// 		request.Header.Add("Auth-Code", authCode)
// 	}

// 	resp, err := client.Do(request)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return &Config{}, nil
// }

func main() {
	config, err := NewConfig("./config.yaml")

	if err != nil {
		fmt.Printf("Error %s\n", err)
	}

	mayday, err := NewMayday(config)
	mayday.Run()

}
