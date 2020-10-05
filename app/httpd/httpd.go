package httpd

import (
	"net/http"

	"github.com/icexin/eggos/app"
	"github.com/icexin/eggos/fs"
	"github.com/spf13/afero"
)

func main(ctx *app.Context) error {
	addr := ":80"
	if len(ctx.Args) > 1 {
		addr = ctx.Args[1]
	}
	httpfs := afero.NewHttpFs(fs.Root)
	http.Handle("/fs/", http.StripPrefix("/fs", http.FileServer(httpfs)))
	return http.ListenAndServe(addr, nil)
}

func init() {
	app.Register("httpd", main)
}
