package core

import (
	"encoding/base64"
	"fmt"
	simplejson "github.com/bitly/go-simplejson"
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

type Response interface {
	GetData()
}

type ConfigResponse struct {
	Signed string
	Raw    string
}

func NewConfigResponse(json *simplejson.Json) (*ConfigResponse, error) {
	raw, err := json.Get("raw").String()
	if err != nil {
		return nil, err
	}

	signed, err := json.Get("signed").String()
	if err != nil {
		return nil, err
	}

	return &ConfigResponse{
		Raw:    raw,
		Signed: signed,
	}, nil
}

func (c *ConfigResponse) GetRawDecoded() string {
	rawDecoded, _ := base64.StdEncoding.DecodeString(c.Raw)
	return string(rawDecoded)
}

func (c *ConfigResponse) GetSignedDecoded() string {
	signedDecoded, _ := base64.StdEncoding.DecodeString(c.Signed)
	return string(signedDecoded)
}

func (api *APIClient) GetFormattedURL(prefix ...string) string {
	return fmt.Sprintf("%s/%s/%s", api.Server, strconv.Itoa(DefaultAPIVersion),
		strings.Join(prefix, "/"))
}

func (api *APIClient) NewRequest(method string, url string) (*simplejson.Json, error) {
	request, err := http.NewRequest(method, url, nil)
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

	defer response.Body.Close()
	reader, err := simplejson.NewFromReader(response.Body)

	if err != nil {
		return nil, err
	}

	return reader, nil
}

func (api *APIClient) GetConfig() (*ConfigResponse, error) {
	response, err := api.NewRequest("GET", api.GetFormattedURL(api.UUID))

	if err != nil {
		return nil, err
	}

	config, err := NewConfigResponse(response)
	if err != nil {
		return nil, err
	}

	return config, nil
}
