package cga

import (
	"reflect"
	"testing"
)

func TestAnsi(t *testing.T) {
	var cases = []struct {
		str    string
		action byte
		params []string
		err    error
	}{
		{
			str:    "\x1b[12;24G",
			action: 'G',
			params: []string{"12", "24"},
		},
		{
			str:    "\x1b[12;;24G",
			action: 'G',
			params: []string{"12", "", "24"},
		},
		{
			str:    "\x1b[G",
			action: 'G',
			params: []string{},
		},
		{
			str: "X\x1b[G",
			err: errNormalChar,
		},
		{
			str: "\x1bX[G",
			err: errInvalidChar,
		},
		{
			str: "\x1b[\x00G",
			err: errInvalidChar,
		},
		{
			str: "\x1b[12\x00",
			err: errInvalidChar,
		},
	}

	p := ansiParser{}
	for _, test := range cases {
		p.Reset()
	runcase:
		for i := range test.str {
			err := p.step(test.str[i])
			if err == nil {
				continue
			}
			if err != errCSIDone {
				if err == test.err {
					break runcase
				} else {
					t.Fatalf("%q[%d] expect %v got %v", test.str, i, test.err, err)
				}
			}
			if test.action != p.Action() {
				t.Fatalf("%q[%d] expect %v got %v", test.str, i, test.action, p.Action())
			}
			if !reflect.DeepEqual(test.params, p.Params()) {
				t.Fatalf("%q[%d] expect %q got %q", test.str, i, test.params, p.Params())
			}
		}
	}
}
