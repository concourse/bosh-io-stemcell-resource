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
		log.Fatalln(err)
	}

	var inRequest concourseIn
	inRequest.Params.Tarball = true

	err = json.Unmarshal(rawJSON, &inRequest)
	if err != nil {
		log.Fatalln(err)
	}

	location := os.Args[1]

	client := boshio.NewClient(progress.NewBar(), content.NewRanger(routines))

	dataLocations := []string{"version", "sha1", "url"}

	for _, name := range dataLocations {
		fileLocation, err := os.Create(filepath.Join(location, name))
		if err != nil {
			log.Fatalln(err)
		}
		defer fileLocation.Close()

		err = client.WriteMetadata(inRequest.Source.Name, inRequest.Version.Version, name, fileLocation)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if inRequest.Params.Tarball {
		err = client.DownloadStemcell(inRequest.Source.Name, inRequest.Version.Version, location, inRequest.Params.PreserveFileName)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
