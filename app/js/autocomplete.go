package js

import (
	"regexp"
	"strings"

	"github.com/icexin/eggos/app"
	"github.com/peterh/liner"
	"github.com/robertkrimen/otto"
)

var lastExpressionRegex = regexp.MustCompile(`[a-zA-Z0-9]([a-zA-Z0-9\.]*[a-zA-Z0-9])?\.?$`)

func setAutoComplete(r app.LineReader, vm *otto.Otto) {
	r.SetAutoComplete(jsAutocompleteWrapper(vm))
}

func jsAutocompleteWrapper(vm *otto.Otto) liner.Completer {
	return func(line string) []string {
		return jsAutocomplete(vm, line)
	}
}
func jsAutocomplete(vm *otto.Otto, line string) []string {
	lastExpression := lastExpressionRegex.FindString(line)

	bits := strings.Split(lastExpression, ".")

	first := bits[:len(bits)-1]
	last := bits[len(bits)-1]

	var l []string

	if len(first) == 0 {
		c := vm.Context()

		l = make([]string, len(c.Symbols))

		i := 0
		for k := range c.Symbols {
			l[i] = k
			i++
		}
	} else {
		r, err := vm.Eval(strings.Join(bits[:len(bits)-1], "."))
		if err != nil {
			return nil
		}

		if o := r.Object(); o != nil {
			for _, v := range o.KeysByParent() {
				l = append(l, v...)
			}
		}
	}

	var r []string
	for _, s := range l {
		if strings.HasPrefix(s, last) {
			r = append(r, line+strings.TrimPrefix(s, last))
		}
	}

	return r
}
