package cmd

import (
	"errors"
	"net/url"

	"github.com/icexin/eggos/app"
	"github.com/icexin/eggos/fs"
	"github.com/icexin/eggos/fs/smb"
	"github.com/icexin/eggos/fs/stripprefix"
)

func mountmain(ctx *app.Context) error {
	if len(ctx.Args) < 3 {
		return errors.New("usage: mount $uri target")
	}
	uristr, target := ctx.Args[1], ctx.Args[2]
	uri, err := url.Parse(uristr)
	if err != nil {
		return err
	}
	switch uri.Scheme {
	case "smb":
		return mountsmb(uri, target)
	default:
		return errors.New("unsupported scheme " + uri.Scheme)
	}
}

func mountsmb(uri *url.URL, target string) error {
	passwd, _ := uri.User.Password()
	smbfs, err := smb.New(&smb.Config{
		Host:     uri.Host,
		User:     uri.User.Username(),
		Password: passwd,
		Mount:    uri.Path[1:],
	})
	if err != nil {
		return err
	}
	return fs.Mount(target, stripprefix.New("/", smbfs))
}

func init() {
	app.Register("mount", mountmain)
}
