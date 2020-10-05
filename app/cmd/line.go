package cmd

import (
	"fmt"
	"os"

	"github.com/icexin/eggos/app"
	"github.com/peterh/liner"
)

func linemain(ctx *app.Context) error {
	fmt.Println(os.Environ())
	fmt.Println(liner.TerminalSupported())

	r := liner.NewLiner()
	defer r.Close()
	for {
		line, err := r.Prompt(">>> ")
		if err != nil {
			break
		}
		fmt.Println(line)
		r.AppendHistory(line)
	}
	return nil
}

func init() {
	app.Register("line", linemain)
}
