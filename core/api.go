package core

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	simplejson "github.com/bitly/go-simplejson"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"
)

const (
	DefaultAPIBaseURL = "http://localhost:8080"
	DefaultAPIVersion = 1
)

type APIClient struct {
	Server    string
	Id        string
	AuthToken string
	Client    *http.Client
}

func NewAPIClient(server string, id string, authToken string) *APIClient {
	return &APIClient{
		Client:    &http.Client{},
		Id:        id,
		AuthToken: authToken,
		Server:    server,
	}
}

type CaseResponse struct {
	Id          int
	Description string
	Created     string `json:",omitempty"`
	IsPrivate   bool
	Signed      string
	Config      string
	Token       string
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

type UploadFile struct {
	Filename string
	Content  string
}

type ConfigResponse struct {
	Signed string
	Config string
	Files  []string
}

func NewConfigResponse(j *simplejson.Json) (*ConfigResponse, error) {
	c := ConfigResponse{}

	config, err := j.Get("Config").String()
	if err != nil {
		return nil, err
	}

	signed, err := j.Get("Signed").String()
	if err != nil {
		return nil, err
	}

	files, err := j.Get("Files").Array()
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		q := file.(map[string]interface{})
		c.Files = append(c.Files, q["Id"].(json.Number).String())
	}

	c.Config = config
	c.Signed = signed

	return &c, nil
}

func (c *ConfigResponse) GetRawDecoded() string {
	rawDecoded, _ := base64.StdEncoding.DecodeString(c.Config)
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
	if api.AuthToken != "" {
		url = fmt.Sprintf("%s?token=%s", url, api.AuthToken)
	}

	request, err := http.NewRequest(method, url, bytes.NewReader(params))
	if err != nil {
		return nil, err
	}

	if method == "POST" {
		request.Header.Set("Content-Type", "application/json")
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

func (api *APIClient) Config() (*ConfigResponse, error) {
	response, err := api.NewRequest("GET", api.GetFormattedURL("case", api.Id), nil, []int{200})
	if err != nil {
		return nil, err
	}

	config, err := NewConfigResponse(response)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (api *APIClient) Create(description string, private bool, config *Config) (*CaseResponse, error) {
	c, err := json.Marshal(CaseResponse{
		IsPrivate:   private,
		Description: description,
		Config:      config.Raw,
		Signed:      config.Signed,
	})

	if err != nil {
		return nil, err
	}

	response, err := api.NewRequest("POST", api.GetFormattedURL("case"), c, []int{201})
	if err != nil {
		return nil, err
	}

	new_case, err := NewCaseResponse(response)
	if err != nil {
		return nil, err
	}

	return new_case, nil
}

func (api *APIClient) Pull(fileId string) (*UploadFile, error) {
	f, err := api.NewRequest("GET", api.GetFormattedURL("case", api.Id, "file", fileId), nil, []int{200})
	if err != nil {
		return nil, err
	}

	name, err := f.Get("Filename").String()
	if err != nil {
		return nil, err
	}

	content, err := f.Get("Content").String()
	if err != nil {
		return nil, err
	}

	decoded, _ := base64.StdEncoding.DecodeString(content)

	return &UploadFile{
		Filename: name,
		Content:  string(decoded),
	}, nil

}

func (api *APIClient) Upload(filename string) error {
	readed, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(readed)
	c, err := json.Marshal(UploadFile{
		Filename: path.Base(filename),
		Content:  encoded,
	})

	if err != nil {
		return err
	}

	_, err = api.NewRequest("POST", api.GetFormattedURL("case", api.Id, "file"), c, []int{200, 201})
	if err != nil {
		return err
	}

	return nil
}
