package fakes

type Ranger struct {
	BuildRangeCall struct {
		Receives struct {
			ContentLength int64
		}

		Returns struct {
			Ranges []string
			Err    error
		}
	}
}

func (r *Ranger) BuildRange(contentLength int64) ([]string, error) {
	r.BuildRangeCall.Receives.ContentLength = contentLength

	return r.BuildRangeCall.Returns.Ranges, r.BuildRangeCall.Returns.Err
}
