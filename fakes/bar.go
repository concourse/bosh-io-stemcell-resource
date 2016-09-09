package fakes

type Bar struct {
}

func (b Bar) SetTotal(contentLength int64) {
}

func (b Bar) Kickoff() {
}

func (b Bar) Add(int) int {
	return 0
}

func (b Bar) Finish() {
}
