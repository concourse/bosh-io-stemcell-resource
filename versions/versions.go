package versions

import (
	"fmt"
	"sort"

	"github.com/blang/semver"
	"github.com/concourse/bosh-io-stemcell-resource/boshio"
)

type StemcellVersions []map[string]string

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

func (f Filter) Versions() (StemcellVersions, error) {
	if len(f.stemcells) == 0 {
		return StemcellVersions{}, nil
	}

	if f.initialVersion == "" {
		return StemcellVersions{{"version": f.stemcells[0].Version}}, nil
	}

	var list StemcellVersions
	parsedVersion, err := semver.ParseTolerant(f.initialVersion)
	if err != nil {
		panic(err)
	}

	r, err := semver.ParseRange(fmt.Sprintf(">=%s", parsedVersion.String()))
	if err != nil {
		panic(err)
	}

	for _, s := range f.stemcells {
		v, err := semver.ParseTolerant(s.Version)
		if err != nil {
			panic(err)
		}

		if r(v) {
			list = append(list, map[string]string{"version": s.Version})
		}
	}

	sort.Sort(list)

	return list, nil
}

func (sv StemcellVersions) Len() int {
	return len(sv)
}

func (sv StemcellVersions) Swap(i, j int) {
	sv[i], sv[j] = sv[j], sv[i]
}

func (sv StemcellVersions) Less(i, j int) bool {
	first, err := semver.ParseTolerant(sv[i]["version"])
	if err != nil {
		panic(err)
	}

	second, err := semver.ParseTolerant(sv[j]["version"])
	if err != nil {
		panic(err)
	}

	return first.LT(second)
}
