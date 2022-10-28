package util

import (
	"os"
	"path/filepath"
)

func GoBin() string {
	gr, ok := os.LookupEnv("GOROOT")
	if !ok {
		return "go"
	}

	return filepath.Join(gr, "bin", "go")
}
