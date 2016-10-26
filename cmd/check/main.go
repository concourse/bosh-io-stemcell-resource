package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/concourse/bosh-io-stemcell-resource/boshio"
	"github.com/concourse/bosh-io-stemcell-resource/versions"
)

type concourseCheck struct {
	Source struct {
		Name          string `json:"name"`
		ForceRegular  bool   `json:"force_regular"`
		VersionFamily string `json:"version_family"`
	}
	Version struct {
		Version string `json:"version"`
	}
}

func main() {
	rawJSON, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("failed reading json: %s", err)
	}

	var checkRequest concourseCheck
	err = json.Unmarshal(rawJSON, &checkRequest)
	if err != nil {
		log.Fatalf("failed unmarshalling: %s", err)
	}

	httpClient := boshio.NewHTTPClient("https://bosh.io", 5*time.Minute)

	client := boshio.NewClient(httpClient, nil, nil, checkRequest.Source.ForceRegular)
	stemcells, err := client.GetStemcells(checkRequest.Source.Name)
	if err != nil {
		log.Fatalf("failed getting stemcell: %s", err)
	}

	stemcells = stemcells.FilterByType()
	filter := versions.NewFilter(
		checkRequest.Version.Version,
		stemcells,
		checkRequest.Source.VersionFamily,
	)

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
