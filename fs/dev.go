package fs

import "math/rand"

type zero struct{}

func (z zero) Read(b []byte) (int, error) {
	for i := range b {
		b[i] = 0
	}
	return len(b), nil
}

type random struct{}

func (r random) Read(b []byte) (int, error) {
	return rand.Read(b)
}
