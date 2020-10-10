package app

import (
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
