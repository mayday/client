package core

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"sync"
)

type Client struct {
	Hostname    string
	ReportsPath string
	APIClient   *APIClient
}

func NewClient(server string, uuid string, authToken string) (*Client, error) {
	reportsPath, err := GetDefaultReportsDirectory()
	if err != nil {
		return nil, err
	}

	return &Client{
		ReportsPath: reportsPath,
		APIClient:   NewAPIClient(server, uuid, authToken),
	}, nil
}

func (client *Client) Run(pgp bool, upload bool) error {
	reportPath, err := GetTempReportDirectory()

	if err != nil {
		return err
	}

	apiConfig, err := client.APIClient.GetConfig()
	if err != nil {
		return fmt.Errorf("Error getting configuration from server: %s", err)
	}

	config, err := NewConfig(apiConfig.GetRawDecoded(), apiConfig.GetSignedDecoded())

	if pgp {
		err := config.CheckPGPSignature()
		if err != nil {
			return err
		}
		answer := PromptPGPConfirmation(config)
		if answer != true {
			return fmt.Errorf("PGP key has not been accepted")
		}
	}

	wg := new(sync.WaitGroup)
	log.Printf("Starting a new report on: %s", reportPath)

	for _, command := range config.Commands {
		wg.Add(1)
		go client.RunCommand(reportPath, command, wg)
	}

	for _, file := range config.Files {
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

func (m *Client) RunCommand(reportPath string, command Command, wg *sync.WaitGroup) {
	//TODO: Refactor this to handle properly all the error cases
	log.Printf("Running %s", command.Executable)
	ran, _ := exec.Command("/bin/bash", "-c", command.Executable).Output()
	outfile, _ := os.Create(command.GetFileName(reportPath))
	outfile.WriteString(string(ran))
	defer outfile.Close()
	defer wg.Done()
}

func (c *Command) GetFileName(Base string) string {
	return path.Join(Base, MangleCommand(c.Executable))
}
