package assets

import (
	"net/http"
	"sync"

	"github.com/rakyll/statik/fs"
)

//go:generate statik -m -src=./files -dest .. -p assets

var (
	rootfs http.FileSystem
	once   sync.Once
)

func FS() http.FileSystem {
	var err error
	once.Do(func() {
		rootfs, err = fs.New()
	})
	if err != nil {
		panic(err)
	}
	return rootfs
}

func Open(fname string) (http.File, error) {
	fs := FS()
	return fs.Open(fname)
}
