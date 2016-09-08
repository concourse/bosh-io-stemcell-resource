package boshio

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/cheggaaa/pb.v1"

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

func (c *Client) WriteMetadata(name string, version string, locations map[string]*os.File) error {
	for _, stemcell := range c.GetStemcells(name) {
		if stemcell.Version == version {
			_, err := locations["url"].Write([]byte(stemcell.Details().URL))
			if err != nil {
				return err
			}

			_, err = locations["sha1"].Write([]byte(stemcell.Details().SHA1))
			if err != nil {
				return err
			}

			_, err = locations["version"].Write([]byte(stemcell.Version))
			if err != nil {
				return err
			}

			break
		}
	}
	return nil
}

func (c *Client) DownloadStemcell(name string, version string, location *os.File) error {
	stemcellURL := c.Host + c.stemcellDownloadPath
	resp, err := http.Head(fmt.Sprintf(stemcellURL, name, version))
	if err != nil {
		log.Fatalln(err)
	}

	stemcellURL = resp.Request.URL.String()

	ranger := content.NewRanger(routines)
	ranges, err := ranger.BuildRange(resp.ContentLength)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Downloading stemcell from '%s'...\n", stemcellURL)

	var wg sync.WaitGroup
	bar := pb.New(int(resp.ContentLength))
	bar.ShowTimeLeft = false
	bar.Start()
	for _, r := range ranges {
		wg.Add(1)
		go func(byteRange string) {
			defer wg.Done()
			req, err := http.NewRequest("GET", stemcellURL, nil)
			if err != nil {
				log.Fatalln(err)
			}

			byteRangeHeader := fmt.Sprintf("bytes=%s", byteRange)
			req.Header.Add("Range", byteRangeHeader)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Fatalln(err)
			}

			defer resp.Body.Close()

			respBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatalln(err)
			}

			offset, err := strconv.Atoi(strings.Split(byteRange, "-")[0])
			if err != nil {
				log.Fatalln(err)
			}

			bytesWritten, err := location.WriteAt(respBytes, int64(offset))
			if err != nil {
				log.Fatalln(err)
			}

			bar.Add(bytesWritten)
		}(r)
	}

	wg.Wait()
	bar.Finish()

	return nil
}
