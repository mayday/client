package core

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	simplejson "github.com/bitly/go-simplejson"
	"net/http"
	"strconv"
	"strings"
)

const (
	DefaultAPIBaseURL = "http://localhost:8080"
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

type CaseResponse struct {
	Id        int
	Created   string
	IsPrivate bool
	Token     string
}

func NewCaseResponse(json *simplejson.Json) (*CaseResponse, error) {
	isPrivate, err := json.Get("IsPrivate").Bool()
	if err != nil {
		return nil, err
	}

	id, err := json.Get("Id").Int()
	if err != nil {
		return nil, err
	}

	token, err := json.Get("Token").String()
	if err != nil {
		return nil, err
	}

	created, err := json.Get("Created").String()
	if err != nil {
		return nil, err
	}

	return &CaseResponse{
		Id:        id,
		Created:   created,
		IsPrivate: isPrivate,
		Token:     token,
	}, nil
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

func (api *APIClient) NewRequest(method string, url string, params []byte, validStatus []int) (*simplejson.Json, error) {
	request, err := http.NewRequest(method, url, bytes.NewReader(params))
	if err != nil {
		return nil, err
	}

	if method == "POST" {
		request.Header.Set("Content-Type", "application/json")
	}

	if api.AuthToken != "" {
		request.Header.Add("Auth-Token", api.AuthToken)
	}

	response, err := api.Client.Do(request)
	if err != nil {
		return nil, err
	}

	if !Contains(validStatus, response.StatusCode) {
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
	response, err := api.NewRequest("GET", api.GetFormattedURL("case", api.UUID), nil, []int{200})

	if err != nil {
		return nil, err
	}

	config, err := NewConfigResponse(response)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (api *APIClient) CreateCase() (*CaseResponse, error) {
	b, err := json.Marshal(CaseResponse{
		IsPrivate: true,
	})

	if err != nil {
		return nil, err
	}

	response, err := api.NewRequest("POST", api.GetFormattedURL("case"), b, []int{201})
	if err != nil {
		return nil, err
	}

	new_case, err := NewCaseResponse(response)
	if err != nil {
		return nil, err
	}

	return new_case, nil
}
