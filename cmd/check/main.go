package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/concourse/bosh-io-stemcell-resource/boshio"
	"github.com/concourse/bosh-io-stemcell-resource/versions"
)

type concourseCheck struct {
	Source struct {
		Name         string `json:"name"`
		ForceRegular bool   `json:"force_regular"`
	}
	Version struct {
		Version string `json:"version"`
	}
}

func main() {
	var checkRequest concourseCheck
	err := json.NewDecoder(os.Stdin).Decode(&checkRequest)
	if err != nil {
		log.Fatalf("failed reading json: %s", err)
	}

	httpClient := boshio.HTTPClient{
		Host:   "https://bosh.io",
		Wait:   800 * time.Millisecond,
		Client: http.DefaultClient,
	}

	client := boshio.NewClient(httpClient, nil, nil, checkRequest.Source.ForceRegular)
	stemcells, err := client.GetStemcells(checkRequest.Source.Name)
	if err != nil {
		log.Fatalf("failed getting stemcell: %s", err)
	}

	stemcells = stemcells.FilterByType()
	filter := versions.NewFilter(checkRequest.Version.Version, stemcells)

	filteredVersions, err := filter.Versions()
	if err != nil {
		log.Fatalf("failed filtering versions: %s", err)
	}

	content, err := json.Marshal(filteredVersions)
	if err != nil {
		log.Fatalf("failed to marshal: %s", err)
	}

	fmt.Println(string(content))
}
