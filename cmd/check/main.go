package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/concourse/bosh-io-stemcell-resource/boshio"
	"github.com/concourse/bosh-io-stemcell-resource/versions"
)

type concourseCheck struct {
	Source struct {
		Name string
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

	var checkRequest concourseCheck
	err = json.Unmarshal(rawJSON, &checkRequest)
	if err != nil {
		panic(err)
	}

	client := boshio.NewClient()
	stemcells := client.GetStemcells(checkRequest.Source.Name)

	filter := versions.NewFilter(checkRequest.Version.Version, stemcells)

	filteredVersions, err := filter.Versions()
	if err != nil {
		panic(err)
	}

	content, err := json.Marshal(filteredVersions)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(content))
}
