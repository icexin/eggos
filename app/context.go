package app

import (
	"flag"
	"fmt"
	"io"

	"github.com/jspc/eggos/fs"
	"github.com/jspc/eggos/fs/chdir"
	"github.com/peterh/liner"
)

type Context struct {
	Args   []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	*chdir.Chdirfs

	flag  *flag.FlagSet
	liner *liner.State
}

func (c *Context) Init() {
	c.Chdirfs = chdir.New(fs.Root)
}

func (c *Context) Printf(fmtstr string, args ...interface{}) {
	fmt.Fprintf(c.Stdout, fmtstr, args...)
}

func (c *Context) Flag() *flag.FlagSet {
	if c.flag != nil {
		return c.flag
	}
	c.flag = flag.NewFlagSet(c.Args[0], flag.ContinueOnError)
	return c.flag
}

func (c *Context) ParseFlags() error {
	return c.Flag().Parse(c.Args[1:])
}

func (c *Context) LineReader() LineReader {
	_, ok := c.Stdin.(fs.Ioctler)
	if !ok {
		return newSimpleLineReader(c.Stdin, c.Stdout)
	}
	return newLineEditor()
}
