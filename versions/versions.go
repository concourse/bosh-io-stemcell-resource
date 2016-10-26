package versions

import (
	"fmt"
	"sort"
	"strings"

	"github.com/blang/semver"
	"github.com/concourse/bosh-io-stemcell-resource/boshio"
)

type StemcellVersions []map[string]string

type Filter struct {
	initialVersion string
	stemcells      []boshio.Stemcell
	versionFamily  string
}

func NewFilter(initialVersion string, stemcells []boshio.Stemcell, versionFamily string) Filter {
	return Filter{
		initialVersion: initialVersion,
		stemcells:      stemcells,
		versionFamily:  versionFamily,
	}
}

func (f Filter) Versions() (StemcellVersions, error) {
	if len(f.stemcells) == 0 {
		return StemcellVersions{}, nil
	}

	stemcellVersions := f.mapStemcellsToVersions(f.stemcells)
	if len(f.versionFamily) > 0 {
		var err error
		stemcellVersions, err = f.filterStemcellsByVersionFamily(stemcellVersions)
		if err != nil {
			return StemcellVersions{}, err
		}
	}

	if len(stemcellVersions) == 0 {
		return StemcellVersions{}, nil
	}

	sort.Sort(stemcellVersions)

	if f.initialVersion == "" {
		return stemcellVersions[len(stemcellVersions)-1:], nil
	}

	return f.selectVersionsGreaterThanInitial(stemcellVersions)
}

func (f Filter) mapStemcellsToVersions(stemcells []boshio.Stemcell) StemcellVersions {
	versions := StemcellVersions{}
	for _, s := range stemcells {
		versions = append(versions, map[string]string{"version": s.Version})
	}
	return versions
}

func (f Filter) filterStemcellsByVersionFamily(stemcells StemcellVersions) (StemcellVersions, error) {
	parsedVersion, err := semver.ParseTolerant(f.versionFamily)
	if err != nil {
		return StemcellVersions{}, err
	}

	parsedVersionCeiling := parsedVersion
	numberOfSignificantDigits := strings.Count(f.versionFamily, ".") + 1
	switch numberOfSignificantDigits {
	case 1:
		parsedVersionCeiling.Major += 1
	case 2:
		parsedVersionCeiling.Minor += 1
	default:
		parsedVersionCeiling.Patch += 1
	}
	versionFamilyRange, err := semver.ParseRange(fmt.Sprintf(">=%s <%s", parsedVersion.String(), parsedVersionCeiling.String()))
	if err != nil {
		return StemcellVersions{}, err
	}

	filteredStemcells := StemcellVersions{}
	for _, s := range stemcells {
		v, err := semver.ParseTolerant(s["version"])
		if err != nil {
			return StemcellVersions{}, err
		}
		if versionFamilyRange(v) {
			filteredStemcells = append(filteredStemcells, s)
		}
	}

	return filteredStemcells, nil
}

func (f Filter) selectVersionsGreaterThanInitial(stemcells StemcellVersions) (StemcellVersions, error) {
	parsedInitialVersion, err := semver.ParseTolerant(f.initialVersion)
	if err != nil {
		return StemcellVersions{}, nil
	}
	greaterThanInitial, err := semver.ParseRange(fmt.Sprintf(">=%s", parsedInitialVersion.String()))
	if err != nil {
		return StemcellVersions{}, nil
	}

	filteredStemcells := StemcellVersions{}
	for _, s := range stemcells {
		v, err := semver.ParseTolerant(s["version"])
		if err != nil {
			return StemcellVersions{}, nil
		}

		if greaterThanInitial(v) {
			filteredStemcells = append(filteredStemcells, s)
		}
	}
	return filteredStemcells, nil
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
