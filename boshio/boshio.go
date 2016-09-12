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

type bar interface {
	SetTotal(contentLength int64)
	Add(totalWritten int) int
	Kickoff()
	Finish()
}

type ranger interface {
	BuildRange(contentLength int64) ([]string, error)
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
	ranger               ranger
	stemcellMetadataPath string
	stemcellDownloadPath string
}

func NewClient(b bar, r ranger) *Client {
	return &Client{
		Host:                 "https://bosh.io/",
		bar:                  b,
		ranger:               r,
		stemcellMetadataPath: "api/v1/stemcells/%s",
		stemcellDownloadPath: "d/stemcells/%s?v=%s",
	}
}

func (c *Client) GetStemcells(name string) ([]Stemcell, error) {
	metadataURL := c.Host + c.stemcellMetadataPath

	resp, err := http.Get(fmt.Sprintf(metadataURL, name))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed fetching metadata - boshio returned: %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var stemcells []Stemcell
	err = json.Unmarshal(bodyBytes, &stemcells)
	if err != nil {
		return nil, err
	}

	return stemcells, nil
}

func (c *Client) WriteMetadata(name string, version string, metadataKey string, metadataFile io.Writer) error {
	var stemcell Stemcell

	stemcells, err := c.GetStemcells(name)
	if err != nil {
		return err
	}

	for _, s := range stemcells {
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
		return err
	}

	stemcellURL = resp.Request.URL.String()

	ranges, err := c.ranger.BuildRange(resp.ContentLength)
	if err != nil {
		return err
	}

	stemcellFileName := "stemcell.tgz"
	if preserveFileName {
		stemcellFileName = filepath.Base(resp.Request.URL.Path)
	}

	stemcell, err := os.Create(filepath.Join(location, stemcellFileName))
	if err != nil {
		return err
	}
	defer stemcell.Close()

	c.bar.SetTotal(int64(resp.ContentLength))
	c.bar.Kickoff()

	var wg sync.WaitGroup
	finish := make(chan error)
	broken := make(chan error)
	for _, r := range ranges {
		wg.Add(1)
		go func(byteRange string, errChan chan<- error) {
			defer wg.Done()

			req, err := http.NewRequest("GET", stemcellURL, nil)
			if err != nil {
				errChan <- err
				return
			}

			byteRangeHeader := fmt.Sprintf("bytes=%s", byteRange)
			req.Header.Add("Range", byteRangeHeader)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				errChan <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusPartialContent {
				errChan <- fmt.Errorf("failed to download stemcell - boshio returned %d", resp.StatusCode)
				return
			}

			respBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				errChan <- err
				return
			}

			offset, err := strconv.Atoi(strings.Split(byteRange, "-")[0])
			if err != nil {
				errChan <- err
				return
			}

			bytesWritten, err := stemcell.WriteAt(respBytes, int64(offset))
			if err != nil {
				errChan <- err
				return
			}

			c.bar.Add(bytesWritten)
		}(r, broken)
	}

	go func() {
		wg.Wait()
		close(finish)
	}()

	select {
	case <-finish:
		c.bar.Finish()
	case err := <-broken:
		if err != nil {
			return err
		}
	}

	return nil
}
