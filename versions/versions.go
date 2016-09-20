package versions

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/concourse/bosh-io-stemcell-resource/boshio"
)

type List map[string]string

type Filter struct {
	initialVersion string
	stemcells      []boshio.Stemcell
}

func NewFilter(initialVersion string, stemcells []boshio.Stemcell) Filter {
	return Filter{
		initialVersion: initialVersion,
		stemcells:      stemcells,
	}
}

func (f Filter) Versions() ([]List, error) {
	var stemcellVersions []List

	if f.initialVersion != "" {
		parsedVersion, err := semver.ParseTolerant(f.initialVersion)
		if err != nil {
			panic(err)
		}

		r, err := semver.ParseRange(fmt.Sprintf(">%s", parsedVersion.String()))
		if err != nil {
			panic(err)
		}

		for _, s := range f.stemcells {
			v, err := semver.ParseTolerant(s.Version)
			if err != nil {
				panic(err)
			}

			if r(v) {
				stemcellVersions = append(stemcellVersions, List{"version": s.Version})
			}
		}
	} else {
		stemcellVersions = append(stemcellVersions, List{"version": f.stemcells[0].Version})
	}

	return stemcellVersions, nil
}
