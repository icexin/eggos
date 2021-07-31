package tests

import (
	"os"
	"testing"
)

func TestEnv(t *testing.T) {
	for _, env := range os.Environ() {
		t.Log(env)
	}
}
