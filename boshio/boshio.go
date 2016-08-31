package boshio

type Stemcell struct {
	Name    string
	Version string
	Details Details `json:"light"`
}

type Details struct {
	URL  string
	Size int64
	MD5  string
	SHA1 string
}
