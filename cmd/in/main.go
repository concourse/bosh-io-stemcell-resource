package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/concourse/bosh-io-stemcell-resource/boshio"
	"github.com/concourse/bosh-io-stemcell-resource/content"
	"github.com/concourse/bosh-io-stemcell-resource/progress"
)

const routines = 10

type concourseInRequest struct {
	Source struct {
		Name         string `json:"name"`
		ForceRegular bool   `json:"force_regular"`
		ForceLight   bool   `json:"force_light"`
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
	var inRequest concourseInRequest
	inRequest.Params.Tarball = true

	err := json.NewDecoder(os.Stdin).Decode(&inRequest)
	if err != nil {
		log.Fatalln(err)
	}

	location := os.Args[1]

	httpClient := boshio.NewHTTPClient("https://bosh.io", 800*time.Millisecond)

	client := boshio.NewClient(httpClient, progress.NewBar(), content.NewRanger(routines), inRequest.Source.ForceRegular, inRequest.Source.ForceLight)

	stemcells, err := client.GetStemcells(inRequest.Source.Name)
	if err != nil {
		log.Fatalln(err)
	}

	stemcell, ok := stemcells.FindStemcellByVersion(inRequest.Version.Version)
	if !ok {
		log.Fatalf("failed to find stemcell matching version: '%s'\n", inRequest.Version.Version)
	}

	dataLocations := []string{"version", "sha1", "sha256", "url"}

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

	metadata := []concourseMetadataField{
		{Name: "url", Value: stemcell.Details().URL},
		{Name: "sha1", Value: stemcell.Details().SHA1},
	}

	if stemcell.Details().SHA256 != "" {
		m := concourseMetadataField{Name: "sha256", Value: stemcell.Details().SHA256}
		metadata = append(metadata, m)
	}

	json.NewEncoder(os.Stdout).Encode(concourseInResponse{
		Version:  inRequest.Version,
		Metadata: metadata,
	},
	)
}
