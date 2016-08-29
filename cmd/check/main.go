package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type concourseCheck struct {
	Source struct {
		Name string
	}
	Version struct {
	}
}

type stemcell struct {
	Name    string
	Version string
	Details struct {
		URL  string
		Size int64
		MD5  string
		SHA1 string
	} `json:"light"`
}

type version map[string]string

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

	var stemcells []stemcell
	err = json.Unmarshal(bodyBytes, &stemcells)
	if err != nil {
		panic(err)
	}

	var versions []version
	for _, s := range stemcells {
		versions = append(versions, version{"version": s.Version})
	}

	content, err := json.Marshal(versions)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(content))
}
