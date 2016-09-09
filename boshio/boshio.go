package boshio

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/concourse/bosh-io-stemcell-resource/content"
)

const routines = 10

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

type bar interface {
	SetTotal(contentLength int64)
	Add(totalWritten int) int
	Kickoff()
	Finish()
}

func (s Stemcell) Details() Metadata {
	if s.Light != nil {
		return *s.Light
	}

	return *s.Regular
}

type Client struct {
	Host                 string
	bar                  bar
	stemcellMetadataPath string
	stemcellDownloadPath string
}

func NewClient(b bar) *Client {
	return &Client{
		Host:                 "https://bosh.io/",
		bar:                  b,
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

func (c *Client) WriteMetadata(name string, version string, metadataKey string, metadataFile io.Writer) error {
	var stemcell Stemcell

	for _, s := range c.GetStemcells(name) {
		if s.Version == version {
			stemcell = s
			break
		}
	}

	if stemcell.Name == "" {
		return fmt.Errorf("Failed to find stemcell: %q", name)
	}

	switch metadataKey {
	case "url":
		_, err := metadataFile.Write([]byte(stemcell.Details().URL))
		if err != nil {
			return err
		}
	case "sha1":
		_, err := metadataFile.Write([]byte(stemcell.Details().SHA1))
		if err != nil {
			return err
		}
	case "version":
		_, err := metadataFile.Write([]byte(stemcell.Version))
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) DownloadStemcell(name string, version string, location string, preserveFileName bool) error {
	stemcellURL := c.Host + c.stemcellDownloadPath
	resp, err := http.Head(fmt.Sprintf(stemcellURL, name, version))
	if err != nil {
		panic(err)
	}

	stemcellURL = resp.Request.URL.String()

	ranger := content.NewRanger(routines)
	ranges, err := ranger.BuildRange(resp.ContentLength)
	if err != nil {
		panic(err)
	}

	stemcellFileName := "stemcell.tgz"
	if preserveFileName {
		stemcellFileName = filepath.Base(resp.Request.URL.Path)
	}

	stemcell, err := os.Create(filepath.Join(location, stemcellFileName))
	if err != nil {
		panic(err)
	}
	defer stemcell.Close()

	c.bar.SetTotal(int64(resp.ContentLength))
	c.bar.Kickoff()

	var wg sync.WaitGroup
	for _, r := range ranges {
		wg.Add(1)
		go func(byteRange string) {
			defer wg.Done()

			req, err := http.NewRequest("GET", stemcellURL, nil)
			if err != nil {
				panic(err)
			}

			byteRangeHeader := fmt.Sprintf("bytes=%s", byteRange)
			req.Header.Add("Range", byteRangeHeader)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				panic(err)
			}

			defer resp.Body.Close()

			respBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			offset, err := strconv.Atoi(strings.Split(byteRange, "-")[0])
			if err != nil {
				panic(err)
			}

			bytesWritten, err := stemcell.WriteAt(respBytes, int64(offset))
			if err != nil {
				panic(err)
			}

			c.bar.Add(bytesWritten)
		}(r)
	}

	wg.Wait()

	c.bar.Finish()

	return nil
}
