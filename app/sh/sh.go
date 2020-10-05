package sh

import (
	"fmt"
	"log"
	"strings"

	"github.com/icexin/eggos/app"
	"github.com/icexin/eggos/console"

	"github.com/mattn/go-shellwords"
)

const prompt = "root@eggos# "

func main(ctx *app.Context) error {
	r := ctx.LineReader()
	defer r.Close()
	r.SetAutoComplete(autocomplete)
	for {
		line, err := r.Prompt(prompt)
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		r.AppendHistory(line)
		err = doline(ctx, line)
		if err != nil {
			fmt.Fprintf(ctx.Stderr, "%s\n", err)
		}
	}
	fmt.Fprintf(ctx.Stdout, "exit\n")
	return nil
}

func doline(ctx *app.Context, line string) error {
	list, err := shellwords.Parse(line)
	if err != nil {
		return err
	}
	var bg bool
	if list[0] == "go" {
		bg = true
		list = list[1:]
	}
	name, args := list[0], list[1:]
	err = runApp(ctx, name, args, bg)
	if err != nil {
		return err
	}
	return nil
}

func runApp(ctx *app.Context, name string, args []string, bg bool) error {
	entry := app.Get(name)
	if entry == nil {
		return fmt.Errorf("%s not found", name)
	}
	nctx := *ctx
	nctx.Args = append([]string{name}, args...)
	if bg {
		go func() {
			entry(&nctx)
			fmt.Fprintf(ctx.Stderr, "job %s done\n", name)
		}()
		return nil
	}
	return entry(&nctx)
}

func Bootstrap() {
	con := console.Console()
	log.SetOutput(con)
	ctx := &app.Context{
		Args:   []string{"sh"},
		Stdin:  con,
		Stdout: con,
		Stderr: con,
	}
	main(ctx)
}

func init() {
	app.Register("sh", main)
}
