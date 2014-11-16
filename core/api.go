package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
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
