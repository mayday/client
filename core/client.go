package core

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"sync"
	//	"net/http"
	"os"
	"os/exec"
	"os/user"
	"time"
)

const (
	DefaultAPIBaseURL = "https://Client.api/"
)

type Client struct {
	Configuration *Config
	Hostname      string
	ReportsPath   string
}

func NewClient(config *Config) (*Client, error) {
	client := Client{
		Configuration: config,
	}

	reportsPath, err := client.GetDefaultReportsPath()

	if err != nil {
		return nil, err
	} else {
		client.ReportsPath = reportsPath
	}

	return &client, nil
}

func (m *Client) GetDefaultReportsPath() (string, error) {
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

func (m *Client) CreateReportTempDirectory() (string, error) {
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

func (m *Client) RunCommand(reportPath string, command Command, wg *sync.WaitGroup) {
	//TODO: Refactor this to handle properly all the error cases
	log.Printf("Running %s", command.Executable)
	ran, _ := exec.Command("/bin/bash", "-c", command.Executable).Output()
	outfile, _ := os.Create(command.GetFileName(reportPath))
	outfile.WriteString(string(ran))
	defer outfile.Close()
	defer wg.Done()
}

func (m *Client) ConfirmPGP() string {
	if m.Configuration.Signature == nil {
		return "y"

	}
	var answer string

	for _, key := range m.Configuration.Signature.Keys {
		fmt.Printf("Configuration file Signed-off by PGP Key: %s\n", key.PublicKey.KeyIdShortString())
		for _, identity := range key.Entity.Identities {
			fmt.Printf(" - %s\n", identity.UserId.Id)
		}
	}

	fmt.Printf("Proceed (y/n)? ")
	fmt.Scanf("%s", &answer)
	return answer
}

func (m *Client) Run() error {
	reportPath, err := m.CreateReportTempDirectory()

	if err != nil {
		return err
	}

	if m.ConfirmPGP() != "y" {
		return fmt.Errorf("PGP validation not confirmed")
	}

	wg := new(sync.WaitGroup)
	log.Printf("Starting a new report on: %s", reportPath)

	for _, command := range m.Configuration.Commands {
		wg.Add(1)
		go m.RunCommand(reportPath, command, wg)
	}

	for _, file := range m.Configuration.Files {
		finfo, err := os.Stat(file.Path)
		if err != nil {
			log.Printf("Cannot stat file:%s", file.Path)
		} else {
			log.Printf("Archiving file:%s", file.Path)
			if finfo.IsDir() {
				CopyDir(file.Path, path.Join(reportPath, file.Path))
			} else {
				CopyFile(file.Path, reportPath)
			}
		}
	}

	wg.Wait()

	return nil
}

func (c *Command) GetFileName(Base string) string {
	return path.Join(Base, MangleCommand(c.Executable))
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
