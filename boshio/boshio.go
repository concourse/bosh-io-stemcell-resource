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

	"golang.org/x/sync/errgroup"
)

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

//go:generate counterfeiter -o ../fakes/http_client.go --fake-name HTTPClient . httpClient
type httpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	httpClient           httpClient
	Bar                  bar
	Ranger               ranger
	StemcellMetadataPath string
	ForceRegular         bool
}

func NewClient(httpClient httpClient, b bar, r ranger, forceRegular bool) *Client {
	return &Client{
		httpClient:           httpClient,
		Bar:                  b,
		Ranger:               r,
		StemcellMetadataPath: "/api/v1/stemcells/%s",
		ForceRegular:         forceRegular,
	}
}

func (c *Client) GetStemcells(name string) (Stemcells, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(c.StemcellMetadataPath, name), nil)
	if err != nil {
		panic(err)
	}

	resp, err := c.httpClient.Do(req)
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
			stemcells[i].ForceRegular = true
		}
	}

	return stemcells, nil
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

	var g errgroup.Group
	for _, r := range ranges {
		byteRange := r
		g.Go(func() error {
			req, err := http.NewRequest("GET", stemcellURL, nil)
			if err != nil {
				return err
			}

			byteRangeHeader := fmt.Sprintf("bytes=%s", byteRange)
			req.Header.Add("Range", byteRangeHeader)

			resp, err := c.httpClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusPartialContent {
				return fmt.Errorf("failed to download stemcell - boshio returned %d", resp.StatusCode)
			}

			respBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			offset, err := strconv.Atoi(strings.Split(byteRange, "-")[0])
			if err != nil {
				return err
			}

			bytesWritten, err := stemcellData.WriteAt(respBytes, int64(offset))
			if err != nil {
				return err
			}

			c.Bar.Add(bytesWritten)

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	c.Bar.Finish()

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
