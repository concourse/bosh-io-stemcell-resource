package boshio

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"golang.org/x/sync/errgroup"
)

//go:generate counterfeiter -o ../fakes/bar.go --fake-name Bar . bar
type bar interface {
	SetTotal(contentLength int64)
	Add(totalWritten int) int
	Kickoff()
	Finish()
}

type progressWriter struct {
	bar bar
}

func (w progressWriter) Write(p []byte) (int, error) {
	w.bar.Add(len(p))
	return len(p), nil
}

type Auth struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

type PrivateBucket struct {
	Endpoint string `json:"endpoint"`
	Bucket   string `json:"bucket"`
	Regexp   string `json:"regexp"`
}

func (b PrivateBucket) Configured() bool {
	return b.Endpoint != "" || b.Bucket != "" || b.Regexp != ""
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
	ForceLight           bool
}

func NewClient(httpClient httpClient, b bar, r ranger, forceRegular bool, forceLight bool) (*Client, error) {
	if forceRegular && forceLight {
		return nil, fmt.Errorf("cannot set both force_regular and force_light to true")
	}

	return &Client{
		httpClient:           httpClient,
		Bar:                  b,
		Ranger:               r,
		StemcellMetadataPath: "/api/v1/stemcells/%s?all=1",
		ForceRegular:         forceRegular,
		ForceLight:           forceLight,
	}, nil
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

	bodyBytes, err := io.ReadAll(resp.Body)
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

	if c.ForceLight {
		for i := 0; i < len(stemcells); i++ {
			stemcells[i].ForceLight = true
		}
	}

	return stemcells, nil
}

func (c *Client) GetPrivateBucketStemcells(name string, privateBucket PrivateBucket, auth Auth) (Stemcells, error) {
	if privateBucket.Endpoint == "" {
		return nil, fmt.Errorf("private_bucket endpoint is required")
	}
	if privateBucket.Bucket == "" {
		return nil, fmt.Errorf("private_bucket bucket is required")
	}
	if privateBucket.Regexp == "" {
		return nil, fmt.Errorf("private_bucket regexp is required")
	}
	if auth.AccessKey == "" || auth.SecretKey == "" {
		return nil, fmt.Errorf("auth access_key and secret_key are required for private_bucket")
	}

	objectRegexp, err := regexp.Compile(privateBucket.Regexp)
	if err != nil {
		return nil, err
	}

	versionIndex := privateBucketVersionIndex(objectRegexp)
	if versionIndex == -1 {
		return nil, fmt.Errorf("private_bucket regexp must include a version capture group")
	}

	minioClient, err := c.minioClientForEndpoint(privateBucket.Endpoint, auth)
	if err != nil {
		return nil, err
	}

	stemcells := Stemcells{}
	for object := range minioClient.ListObjects(context.Background(), privateBucket.Bucket, minio.ListObjectsOptions{Recursive: true}) {
		if object.Err != nil {
			return nil, object.Err
		}

		matches := objectRegexp.FindStringSubmatch(object.Key)
		if matches == nil || matches[0] != object.Key {
			continue
		}

		stemcells = append(stemcells, Stemcell{
			Name:       name,
			Version:    matches[versionIndex],
			ForceLight: c.ForceLight,
			Regular: &Metadata{
				URL:  privateBucket.URLForObject(object.Key),
				Size: object.Size,
				MD5:  strings.Trim(object.ETag, "\""),
			},
		})
	}

	return stemcells, nil
}

func privateBucketVersionIndex(objectRegexp *regexp.Regexp) int {
	for index, name := range objectRegexp.SubexpNames() {
		if name == "version" {
			return index
		}
	}

	if objectRegexp.NumSubexp() > 0 {
		return 1
	}

	return -1
}

func (b PrivateBucket) URLForObject(object string) string {
	objectURL, err := url.Parse(b.Endpoint)
	if err != nil {
		return ""
	}

	objectURL.Path = strings.TrimRight(objectURL.Path, "/") + "/" + strings.TrimLeft(b.Bucket+"/"+object, "/")
	return objectURL.String()
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
	case "sha256":
		_, err := metadataFile.Write([]byte(stemcell.Details().SHA256))
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

func (c *Client) DownloadStemcell(stemcell Stemcell, location string, preserveFileName bool, auth Auth) error {
	var contentLength int64
	var err error
	stemcellFileName := "stemcell.tgz"
	stemcellUrl := stemcell.Details().URL

	if preserveFileName {
		stemcellUrlObject, err := url.Parse(stemcellUrl)
		if err != nil {
			return err
		}
		stemcellFileName = filepath.Base(stemcellUrlObject.Path)
	}

	if auth.AccessKey != "" {
		contentLength, err = c.contentLengthWithAuth(stemcellUrl, auth)
		if err != nil {
			return fmt.Errorf("failed to fetch object metadata: %s", err)
		}
	} else {
		req, err := http.NewRequest("HEAD", stemcellUrl, nil)
		if err != nil {
			return fmt.Errorf("failed to construct HEAD request: %s", err)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		contentLength = resp.ContentLength
	}

	ranges, err := c.Ranger.BuildRange(contentLength)
	if err != nil {
		return err
	}

	stemcellData, err := os.Create(filepath.Join(location, stemcellFileName))
	if err != nil {
		return err
	}
	defer stemcellData.Close()

	c.Bar.SetTotal(contentLength)
	c.Bar.Kickoff()

	var g errgroup.Group
	for _, r := range ranges {
		byteRange := r
		g.Go(func() error {

			offset, err := strconv.Atoi(strings.Split(byteRange, "-")[0])
			offsetEnd, err := strconv.Atoi(strings.Split(byteRange, "-")[1])
			bytes := offsetEnd - offset + 1
			if err != nil {
				return err
			}

			var respBytes []byte
			if auth.AccessKey != "" {
				respBytes, err = c.fetchWithAuth(stemcellUrl, bytes, offset, auth)
			} else {
				respBytes, err = c.retryableRequest(stemcellUrl, byteRange)
				if err != nil {
					return err
				}
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

	if stemcell.Details().SHA256 == "" {
		computedSHA := sha1.New()
		_, err = io.Copy(computedSHA, stemcellData)
		if err != nil {
			return err
		}

		if fmt.Sprintf("%x", computedSHA.Sum(nil)) != stemcell.Details().SHA1 {
			return fmt.Errorf("computed sha1 %x did not match expected sha1 of %s", computedSHA.Sum(nil), stemcell.Details().SHA1)
		}
	} else {
		computedSHA256 := sha256.New()
		_, err = io.Copy(computedSHA256, stemcellData)
		if err != nil {
			return err
		}

		if fmt.Sprintf("%x", computedSHA256.Sum(nil)) != stemcell.Details().SHA256 {
			return fmt.Errorf("computed sha256 %x did not match expected sha256 of %s", computedSHA256.Sum(nil), stemcell.Details().SHA256)
		}
	}

	return nil
}

func (c *Client) DownloadPrivateBucketStemcell(stemcell Stemcell, location string, preserveFileName bool, auth Auth) (Metadata, error) {
	stemcellFileName := "stemcell.tgz"
	stemcellURL := stemcell.Details().URL

	if preserveFileName {
		stemcellURLObject, err := url.Parse(stemcellURL)
		if err != nil {
			return Metadata{}, err
		}
		stemcellFileName = filepath.Base(stemcellURLObject.Path)
	}

	reader, err := c.minioReaderForObject(stemcellURL, auth)
	if err != nil {
		return Metadata{}, fmt.Errorf("failed to fetch object: %s", err)
	}
	defer reader.Close()

	objectInfo, err := reader.Stat()
	if err != nil {
		return Metadata{}, fmt.Errorf("failed to fetch object metadata: %s", err)
	}
	contentLength := objectInfo.Size

	c.Bar.SetTotal(contentLength)
	c.Bar.Kickoff()

	stemcellData, err := os.Create(filepath.Join(location, stemcellFileName))
	if err != nil {
		return Metadata{}, err
	}
	defer stemcellData.Close()

	computedSHA1 := sha1.New()
	computedSHA256 := sha256.New()
	_, err = io.Copy(io.MultiWriter(stemcellData, computedSHA1, computedSHA256, progressWriter{bar: c.Bar}), reader)

	c.Bar.Finish()
	if err != nil {
		return Metadata{}, err
	}

	metadata := stemcell.Details()
	metadata.Size = contentLength
	metadata.SHA1 = fmt.Sprintf("%x", computedSHA1.Sum(nil))
	metadata.SHA256 = fmt.Sprintf("%x", computedSHA256.Sum(nil))

	return metadata, nil
}

func (c Client) retryableRequest(stemcellURL string, byteRange string) ([]byte, error) {
	req, err := http.NewRequest("GET", stemcellURL, nil)
	if err != nil {
		return []byte{}, err
	}

	byteRangeHeader := fmt.Sprintf("bytes=%s", byteRange)
	req.Header.Add("Range", byteRangeHeader)

	for {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return []byte{}, err
		}

		if resp.StatusCode != http.StatusPartialContent {
			resp.Body.Close()
			return []byte{}, fmt.Errorf("failed to download stemcell - boshio returned %d", resp.StatusCode)
		}

		var respBytes []byte
		respBytes, err = io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			if err == io.ErrUnexpectedEOF {
				fmt.Fprint(os.Stderr, "Retrying after server unexpectly closed connection")
				continue
			}

			return []byte{}, err
		}
		return respBytes, nil
	}
}

func (c Client) fetchWithAuth(urlString string, bytes int, offset int, auth Auth) ([]byte, error) {
	reader, err := c.minioReaderForObject(urlString, auth)
	if err != nil {
		return nil, err
	}

	byteSegment := make([]byte, bytes)
	_, err = reader.ReadAt(byteSegment, int64(offset))
	if err != nil {
		return nil, err
	}

	return byteSegment, nil
}

func (c Client) contentLengthWithAuth(urlString string, auth Auth) (int64, error) {
	reader, err := c.minioReaderForObject(urlString, auth)
	if err != nil {
		return 0, err
	}
	objectInfo, err := reader.Stat()
	if err != nil {
		return 0, err
	}
	return objectInfo.Size, nil
}

func (c Client) minioReaderForObject(urlString string, auth Auth) (*minio.Object, error) {
	parsedUrl, _ := url.Parse(urlString)
	pieces := strings.SplitN(parsedUrl.Path, "/", 3)
	bucket, object := pieces[1], pieces[2]
	client, err := c.minioClientForEndpoint(fmt.Sprintf("%s://%s", parsedUrl.Scheme, parsedUrl.Host), auth)
	if err != nil {
		return nil, err
	}

	reader, err := client.GetObject(context.Background(), bucket, object, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return reader, nil
}

func (c Client) minioClientForEndpoint(endpoint string, auth Auth) (*minio.Client, error) {
	parsedEndpoint, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	minioOptions := &minio.Options{
		Creds:  credentials.NewStaticV4(auth.AccessKey, auth.SecretKey, ""),
		Secure: parsedEndpoint.Scheme == "https",
	}

	return minio.New(parsedEndpoint.Host, minioOptions)
}
