package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
)

const (
	DefaultAPIBaseURL = "http://demo6970933.mockable.io"
	DefaultAPIVersion = 1
)

type APIClient struct {
	Server    string
	UUID      string
	AuthToken string
	Client    *http.Client
}

func NewAPIClient(server string, uuid string, authToken string) *APIClient {
	return &APIClient{
		Client:    &http.Client{},
		UUID:      uuid,
		AuthToken: authToken,
		Server:    server,
	}
}

type ConfigResponse struct {
	Signed string `json:"signed"`
	Raw    string `json:"raw"`
}

func (api *APIClient) GetFormattedURL(prefix ...string) string {
	return fmt.Sprintf("%s/%s/%s", api.Server, strconv.Itoa(DefaultAPIVersion),
		strings.Join(prefix, "/"))
}

func (api *APIClient) GetConfig() (*ConfigResponse, error) {
	request, err := http.NewRequest("GET",
		api.GetFormattedURL(api.UUID), nil)

	if err != nil {
		return nil, err
	}

	if api.AuthToken != "" {
		request.Header.Add("Auth-Token", api.AuthToken)
	}

	response, err := api.Client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("Invalid server response: %s", response.Status)
	}

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, err
	}

	config := new(ConfigResponse)
	err = json.Unmarshal(body, &config)

	if err != nil {
		return nil, err
	}

	return config, nil
}

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

func (client *Client) PromptPGPConfirmation(config *Config) bool {
	var answer string

	for _, key := range config.Signature.Keys {
		fmt.Printf("Configuration file Signed-off by PGP Key: %s\n", key.PublicKey.KeyIdShortString())
		for _, identity := range key.Entity.Identities {
			fmt.Printf(" - %s\n", identity.UserId.Id)
		}
	}

	fmt.Printf("Proceed (y/n)? ")
	fmt.Scanf("%s", &answer)

	return answer == "y"
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

	raw, err := base64.StdEncoding.DecodeString(apiConfig.Raw)
	if err != nil {
		return err
	}

	signed, err := base64.StdEncoding.DecodeString(apiConfig.Signed)

	if err != nil {
		return err
	}

	config, err := NewConfig(string(raw), string(signed))

	if pgp {
		err := config.CheckPGPSignature()
		if err != nil {
			return err
		}
		answer := client.PromptPGPConfirmation(config)
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
