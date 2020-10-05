package cmd

import (
	"fmt"

	"github.com/icexin/eggos/app"
	"github.com/icexin/eggos/inet"
)

func nsshmain(ctx *app.Context) error {
	l, err := inet.Listen("tcp", "0.0.0.0:22")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go func() {
			fmt.Fprintf(ctx.Stdout, "conn from:%s\n", conn.RemoteAddr())
			shell := app.Get("sh")
			ctx := &app.Context{
				Stdin:  conn,
				Stdout: conn,
				Stderr: conn,
			}
			shell(ctx)
			conn.Close()
			fmt.Fprintf(ctx.Stdout, "conn %s closed\n", conn.RemoteAddr())
		}()
	}
}

func init() {
	app.Register("nssh", nsshmain)
}
