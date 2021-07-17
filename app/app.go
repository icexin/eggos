package app

import (
	"fmt"
	"runtime/debug"
	"sort"
)

type AppEntry func(ctx *Context) error

var apps = map[string]AppEntry{}

func Register(name string, app AppEntry) {
	apps[name] = app
}

func Get(name string) AppEntry {
	return apps[name]
}

func AppNames() []string {
	var l []string
	for name := range apps {
		l = append(l, name)
	}
	sort.Strings(l)
	return l
}

func Run(name string, ctx *Context) error {
	entry := Get(name)
	if entry == nil {
		return fmt.Errorf("command not found: %s", name)
	}
	defer func() {
		err := recover()
		if err == nil {
			return
		}
		stack := debug.Stack()
		fmt.Fprintf(ctx.Stderr, "panic:%s\n", err)
		ctx.Stderr.Write(stack)
	}()
	return entry(ctx)
}
