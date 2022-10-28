//go:build gin
// +build gin

package ginserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jspc/eggos/app"
)

func main(ctx *app.Context) error {
	r := gin.Default()
	r.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "hello from eggos")
	})
	ctx.Printf("run gin server on :80")
	return r.Run(":80")
}

func init() {
	app.Register("gin", main)
}
