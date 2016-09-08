package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/concourse/bosh-io-stemcell-resource/boshio"
)

type concourseIn struct {
	Source struct {
		Name string
	}
	Params struct {
		Tarball bool
	}
	Version struct {
		Version string
	}
}

func main() {
	rawJSON, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	var inRequest concourseIn
	inRequest.Params.Tarball = true

	err = json.Unmarshal(rawJSON, &inRequest)
	if err != nil {
		panic(err)
	}

	location := os.Args[1]

	client := boshio.NewClient()

	dataLocations := map[string]*os.File{
		"version": nil,
		"sha1":    nil,
		"url":     nil,
	}

	if inRequest.Params.Tarball {
		dataLocations["stemcell.tgz"] = nil
	}

	for key := range dataLocations {
		fileLocation, err := os.Create(filepath.Join(location, key))
		if err != nil {
			panic(err)
		}
		defer fileLocation.Close()

		dataLocations[key] = fileLocation
	}

	err = client.WriteMetadata(inRequest.Source.Name, inRequest.Version.Version, dataLocations)
	if err != nil {
		panic(err)
	}

	if inRequest.Params.Tarball {
		err = client.DownloadStemcell(inRequest.Source.Name, inRequest.Version.Version, dataLocations["stemcell.tgz"])
		if err != nil {
			panic(err)
		}
	}
}
