package boshio

import (
	"crypto/sha1"
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
	Name         string
	Version      string
	Light        *Metadata `json:"light"`
	Regular      *Metadata `json:"regular"`
	forceRegular bool
}

type Metadata struct {
	URL  string
	Size int64
	MD5  string
	SHA1 string
}

//go:generate counterfeiter -o ../fakes/bar.go --fake-name Bar . bar
type bar interface {
	SetTotal(contentLength int64)
	Add(totalWritten int) int
	Kickoff()
	Finish()
}

//go:generate counterfeiter -o ../fakes/ranger.go --fake-name Ranger . ranger
type ranger interface {
	BuildRange(contentLength int64) ([]string, error)
}

func (s Stemcell) Details() Metadata {
	if s.Light != nil && s.forceRegular == false {
		return *s.Light
	}

	return *s.Regular
}

type Client struct {
	Host                 string
	Bar                  bar
	Ranger               ranger
	StemcellMetadataPath string
	ForceRegular         bool
}

func NewClient(b bar, r ranger, forceRegular bool) *Client {
	return &Client{
		Host:                 "https://bosh.io/",
		Bar:                  b,
		Ranger:               r,
		StemcellMetadataPath: "api/v1/stemcells/%s",
		ForceRegular:         forceRegular,
	}
}

func (c *Client) GetStemcells(name string) ([]Stemcell, error) {
	metadataURL := c.Host + c.StemcellMetadataPath

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

	if c.ForceRegular {
		for i := 0; i < len(stemcells); i++ {
			stemcells[i].forceRegular = true
		}
	}

	return stemcells, nil
}

func (c *Client) FilterStemcells(lambdaFilter func(Stemcell) bool, stemcells []Stemcell) []Stemcell {
	var filteredStemcells []Stemcell

	for _, s := range stemcells {
		if lambdaFilter(s) {
			filteredStemcells = append(filteredStemcells, s)
		}
	}

	return filteredStemcells
}

func (c *Client) WriteMetadata(stemcell Stemcell, metadataKey string, metadataFile io.Writer) error {
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

func (c *Client) SupportsLight(stemcells []Stemcell) bool {
	for _, s := range stemcells {
		if s.Light != nil {
			return true
		}
	}
	return false
}

func (c *Client) DownloadStemcell(stemcell Stemcell, location string, preserveFileName bool) error {
	stemcellURL := stemcell.Details().URL
	resp, err := http.Head(stemcellURL)
	if err != nil {
		return err
	}

	stemcellURL = resp.Request.URL.String()

	ranges, err := c.Ranger.BuildRange(resp.ContentLength)
	if err != nil {
		return err
	}

	stemcellFileName := "stemcell.tgz"
	if preserveFileName {
		stemcellFileName = filepath.Base(resp.Request.URL.Path)
	}

	stemcellData, err := os.Create(filepath.Join(location, stemcellFileName))
	if err != nil {
		return err
	}
	defer stemcellData.Close()

	c.Bar.SetTotal(int64(resp.ContentLength))
	c.Bar.Kickoff()

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

			bytesWritten, err := stemcellData.WriteAt(respBytes, int64(offset))
			if err != nil {
				errChan <- err
				return
			}

			c.Bar.Add(bytesWritten)
		}(r, broken)
	}

	go func() {
		wg.Wait()
		close(finish)
	}()

	select {
	case <-finish:
		c.Bar.Finish()
	case err := <-broken:
		if err != nil {
			return err
		}
	}

	computedSHA := sha1.New()
	_, err = io.Copy(computedSHA, stemcellData)
	if err != nil {
		return err
	}

	if fmt.Sprintf("%x", computedSHA.Sum(nil)) != stemcell.Details().SHA1 {
		return fmt.Errorf("computed sha1 %x did not match expected sha1 of %s", computedSHA.Sum(nil), stemcell.Details().SHA1)
	}

	return nil
}
