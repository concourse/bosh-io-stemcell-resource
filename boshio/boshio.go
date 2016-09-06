package boshio

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Stemcell struct {
	Name    string
	Version string
	Light   *Metadata `json:"light"`
	Regular *Metadata `json:"regular"`
}

type Metadata struct {
	URL  string
	Size int64
	MD5  string
	SHA1 string
}

func (s Stemcell) Details() Metadata {
	if s.Light != nil {
		return *s.Light
	}

	return *s.Regular
}

type Client struct {
	Host                 string
	stemcellMetadataPath string
	stemcellDownloadPath string
}

func NewClient() *Client {
	return &Client{
		Host:                 "https://bosh.io/",
		stemcellMetadataPath: "api/v1/stemcells/%s",
		stemcellDownloadPath: "d/stemcells/%s?v=%s",
	}
}

func (c *Client) GetStemcells(name string) []Stemcell {
	metadataURL := c.Host + c.stemcellMetadataPath

	resp, err := http.Get(fmt.Sprintf(metadataURL, name))
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != http.StatusOK {
		panic("wrong code")
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var stemcells []Stemcell
	err = json.Unmarshal(bodyBytes, &stemcells)
	if err != nil {
		panic(err)
	}

	return stemcells
}
