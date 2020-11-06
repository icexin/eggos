package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/icexin/eggos/app"
)

func handleEcho(w io.Writer, s string) {

	str := strings.Contains(s, "$")

	if str == true {
		//For future to print values of global variables
		res := strings.Split(s, "$")
		fmt.Fprintf(w, "%s\n", res[0])

	} else {

		fmt.Fprintf(w, "%s\n", s)

	}

}

//To echo output on the screen
func echoMain(ctx *app.Context) error {

	if len(ctx.Args) == 1 {
		return fmt.Errorf("%s", "\n")
	}

	handleEcho(ctx.Stdout, strings.Join(ctx.Args[1:], " "))

	return nil

}

func init() {
	app.Register("echo", echoMain)
}
