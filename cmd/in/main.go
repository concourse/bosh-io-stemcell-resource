package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/concourse/bosh-io-stemcell-resource/boshio"
	"github.com/concourse/bosh-io-stemcell-resource/content"
	"github.com/concourse/bosh-io-stemcell-resource/progress"
)

const routines = 10

type concourseInRequest struct {
	Source struct {
		Name string `json:"name"`
	} `json:"source"`
	Params struct {
		Tarball          bool `json:"tarball"`
		PreserveFilename bool `json:"preserve_filename"`
	} `json:"params"`
	Version struct {
		Version string `json:"version"`
	} `json:"version"`
}

type concourseInResponse struct {
	Version struct {
		Version string `json:"version"`
	} `json:"version"`
	Metadata []concourseMetadataField `json:"metadata"`
}

type concourseMetadataField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func main() {
	rawJSON, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalln(err)
	}

	var inRequest concourseInRequest
	inRequest.Params.Tarball = true

	err = json.Unmarshal(rawJSON, &inRequest)
	if err != nil {
		log.Fatalln(err)
	}

	location := os.Args[1]

	client := boshio.NewClient(progress.NewBar(), content.NewRanger(routines))

	stemcells, err := client.GetStemcells(inRequest.Source.Name)
	if err != nil {
		log.Fatalln(err)
	}

	stemcell, err := client.FilterStemcells(inRequest.Version.Version, stemcells)
	if err != nil {
		log.Fatalln(err)
	}

	dataLocations := []string{"version", "sha1", "url"}

	for _, name := range dataLocations {
		fileLocation, err := os.Create(filepath.Join(location, name))
		if err != nil {
			log.Fatalln(err)
		}
		defer fileLocation.Close()

		err = client.WriteMetadata(stemcell, name, fileLocation)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if inRequest.Params.Tarball {
		err = client.DownloadStemcell(stemcell, location, inRequest.Params.PreserveFilename)
		if err != nil {
			log.Fatalln(err)
		}
	}

	json.NewEncoder(os.Stdout).Encode(concourseInResponse{
		Version: inRequest.Version,
		Metadata: []concourseMetadataField{
			{Name: "url", Value: stemcell.Details().URL},
			{Name: "sha1", Value: stemcell.Details().SHA1},
		},
	})
}
