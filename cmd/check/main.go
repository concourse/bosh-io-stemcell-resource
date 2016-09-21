package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

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
	rawJSON, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("failed reading json: %s", err)
	}

	var checkRequest concourseCheck
	err = json.Unmarshal(rawJSON, &checkRequest)
	if err != nil {
		log.Fatalf("failed unmarshalling: %s", err)
	}

	client := boshio.NewClient(nil, nil, checkRequest.Source.ForceRegular)
	stemcells, err := client.GetStemcells(checkRequest.Source.Name)
	if err != nil {
		log.Fatalf("failed getting stemcell: %s", err)
	}

	supportsLight := client.SupportsLight(stemcells)
	if supportsLight {
		if checkRequest.Source.ForceRegular {
			regularFilter := func(stemcell boshio.Stemcell) bool {
				return stemcell.Regular != nil
			}
			stemcells = client.FilterStemcells(regularFilter, stemcells)
		} else {
			lightFilter := func(stemcell boshio.Stemcell) bool {
				return stemcell.Light != nil
			}
			stemcells = client.FilterStemcells(lightFilter, stemcells)
		}
	}

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
