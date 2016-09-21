package progress

import (
	"os"

	"gopkg.in/cheggaaa/pb.v1"
)

type Bar struct {
	*pb.ProgressBar
}

func NewBar() Bar {
	bar := pb.New(0)
	bar.SetUnits(pb.U_BYTES)
	bar.Output = os.Stderr
	return Bar{bar} // shop
}

func (b Bar) SetTotal(contentLength int64) {
	b.Total = contentLength
}

func (b Bar) Kickoff() {
	b.Start()
}
