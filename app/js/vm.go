package js

import (
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/robertkrimen/otto"
)

func NewVM() *otto.Otto {
	vm := otto.New()
	addBuiltins(vm)
	return vm
}

func addBuiltins(vm *otto.Otto) {
	vm.Set("http", map[string]interface{}{
		"Get": func(url string) string {
			resp, err := http.Get(url)
			if err != nil {
				Throw(err)
			}
			defer resp.Body.Close()
			buf, _ := ioutil.ReadAll(resp.Body)
			return string(buf)
		},
	})
}

// Throw throw go error in js vm as an Exception
func Throw(err error) {
	v, _ := otto.ToValue("Exception: " + err.Error())
	panic(v)
}

// Throws throw go string in js vm as an Exception
func Throws(msg string) {
	Throw(errors.New(msg))
}
