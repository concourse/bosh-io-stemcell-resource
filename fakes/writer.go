package fakes

import "errors"

type NoopWriter struct{}

func (no NoopWriter) Write(b []byte) (n int, err error) {
	return 0, errors.New("explosions")
}
