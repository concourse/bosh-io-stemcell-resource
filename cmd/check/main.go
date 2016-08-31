package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

	resp, err := http.Get(fmt.Sprintf("https://bosh.io/api/v1/stemcells/%s", checkRequest.Source.Name))
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != http.StatusOK {
		panic("wrong code")
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var stemcells []boshio.Stemcell
	err = json.Unmarshal(bodyBytes, &stemcells)
	if err != nil {
		panic(err)
	}

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
