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
		Tarball          bool
		PreserveFileName bool
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

	dataLocations := []string{"version", "sha1", "url"}

	for _, name := range dataLocations {
		fileLocation, err := os.Create(filepath.Join(location, name))
		if err != nil {
			panic(err)
		}
		defer fileLocation.Close()

		err = client.WriteMetadata(inRequest.Source.Name, inRequest.Version.Version, name, fileLocation)
		if err != nil {
			panic(err)
		}
	}

	if inRequest.Params.Tarball {
		err = client.DownloadStemcell(inRequest.Source.Name, inRequest.Version.Version, location, inRequest.Params.PreserveFileName)
		if err != nil {
			panic(err)
		}
	}
}
