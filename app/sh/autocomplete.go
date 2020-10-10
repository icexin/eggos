package sh

import (
	"path"
	"strings"

	"github.com/icexin/eggos/app"
	"github.com/spf13/afero"
)

func autocompleteWrapper(ctx *app.Context) func(line string) []string {
	return func(line string) []string {
		return autocomplete(ctx, line)
	}
}

func autocomplete(ctx *app.Context, line string) []string {
	line = strings.TrimLeft(line, " ")

	list := strings.Split(line, " ")

	var (
		last   = ""
		hascmd bool
		l      []string
	)

	if len(list) != 0 && list[0] == "go" {
		list = list[1:]
	}

	switch len(list) {
	case 0:
	case 1:
		last = list[0]
	default:
		hascmd = true
		last = list[len(list)-1]
	}

	if !hascmd {
		l = app.AppNames()
	} else {
		l = completeFile(ctx, last)
	}

	var r []string
	for _, s := range l {
		if strings.HasPrefix(s, last) {
			r = append(r, line+strings.TrimPrefix(s, last))
		}
	}

	return r
}

func completeFile(fs afero.Fs, prefix string) []string {
	if prefix == "" {
		prefix = "."
	}

	joinPrefix := func(dir string, l []string) []string {
		for i := range l {
			l[i] = path.Join(dir, l[i])
		}
		return l
	}

	f, err := fs.Open(prefix)
	// user input a complete file name
	if err == nil {
		defer f.Close()

		stat, err := f.Stat()
		if err != nil {
			return nil
		}
		if !stat.IsDir() {
			return nil
		}
		names, err := f.Readdirnames(-1)
		if err != nil {
			return nil
		}
		return joinPrefix(prefix, names)
	}

	// complete dir entries
	dir := path.Dir(prefix)
	f, err = fs.Open(dir)
	if err != nil {
		return nil
	}
	defer f.Close()

	names, err := f.Readdirnames(-1)
	if err != nil {
		return nil
	}
	return joinPrefix(dir, names)
}
