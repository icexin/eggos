package cmd

import (
	"fmt"
	"net"

	"github.com/jspc/eggos/app"
)

func nsshmain(ctx *app.Context) error {
	l, err := net.Listen("tcp", "0.0.0.0:22")
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
			ctx.Init()
			shell(ctx)
			conn.Close()
			fmt.Fprintf(ctx.Stdout, "conn %s closed\n", conn.RemoteAddr())
		}()
	}
}

func init() {
	app.Register("nssh", nsshmain)
}
