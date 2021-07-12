package js

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/icexin/eggos/sys"
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
	vm.Set("sys", map[string]interface{}{
		"in8": func(port uint16) byte {
			return sys.Inb(port)
		},
		"out8": func(port uint16, data byte) {
			sys.Outb(port, data)
		},
	})
	vm.Set("printf", func(fmtstr string, args ...interface{}) int {
		n, _ := fmt.Printf(fmtstr, args...)
		return n
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
