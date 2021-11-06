package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "hello eggos")
	})
	fmt.Println("http server listen on :8000")
	http.ListenAndServe(":8000", nil)
}
