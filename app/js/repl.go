package js

import (
	"fmt"
	"strings"

	"github.com/icexin/eggos/app"
	"github.com/robertkrimen/otto"
)

func repl(ctx *app.Context, vm *otto.Otto) error {
	const promptstr = ">>> "
	var d []string
	prompt := promptstr
	r := ctx.LineReader()
	defer r.Close()

	setAutoComplete(r, vm)
	for {
		cmd, err := r.Prompt(prompt)
		if err != nil {
			break
		}
		if cmd == "" {
			d = nil
			prompt = promptstr
			continue
		}
		r.AppendHistory(cmd)

		d = append(d, cmd)
		s, err := vm.Compile("repl", strings.Join(d, "\n"))
		if err != nil {
			prompt = strings.Repeat(" ", len(promptstr))
			continue
		}
		prompt = promptstr
		d = nil
		_, err = vm.Eval(s)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "%s\n", err)
			continue
		}
		// fmt.Fprintf(ctx.Stdout, "%s\n", v.String())
	}
	return nil
}
