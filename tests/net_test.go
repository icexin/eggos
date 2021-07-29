package tests

import (
	"context"
	"net/http"
	"sync"
	"testing"

	_ "github.com/icexin/eggos"
)

func TestNetworking(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	server := http.Server{
		Addr: ":8000",
	}

	go func() {
		defer wg.Done()

		err := server.ListenAndServe()
		if err != nil {
			t.Fatal(err)
		}
	}()

	resp, err := http.Get("http://127.0.0.1:8000/")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	server.Shutdown(context.Background())
	wg.Wait()
}
