package cga

import (
	"strconv"
)

var (
	parser  = ansiParser{}
	backend = cgabackend{}
)

func WriteString(s string) {
	for i := range s {
		WriteByte(s[i])
	}
}

func setCursorColumn(n int) {
	pos := getbackend().GetPos()
	pos = (pos/80)*80 + n - 1
	getbackend().SetPos(pos)
}

func eraseLine(method int) {
	backend := getbackend()
	pos := backend.GetPos()
	switch method {
	case 0:
		end := (pos/80 + 1) * 80
		for i := pos; i < end; i++ {
			backend.WritePos(i, ' ')
		}
	default:
		panic("unsupported erase line method")
	}
}

func writeCSI(action byte, params []string) {
	// fmt.Fprintf(os.Stderr, "action:%c, params:%v\n", action, params)
	switch action {
	// set cursor
	case 'G':
		if len(params) == 0 {
			setCursorColumn(1)
		} else {
			n, _ := strconv.Atoi(params[0])
			setCursorColumn(n)
		}
	// erase line
	case 'K':
		if len(params) == 0 {
			eraseLine(0)
		} else {
			n, _ := strconv.Atoi(params[0])
			eraseLine(n)
		}
	default:
		panic("unsupported CSI action")
	}
}

func WriteByte(ch byte) {
	switch parser.step(ch) {
	case errNormalChar:
		backend := getbackend()

		switch ch {
		case '\n', '\r', '\b':
			backend.WriteByte(ch)
		case '\t':
			for i := 0; i < 4; i++ {
				backend.WriteByte(' ')
			}
		default:
			if ch >= 32 && ch <= 127 {
				backend.WriteByte(ch)
			}
		}

		// do normal char
	case errCSIDone:
		// do csi
		writeCSI(parser.Action(), parser.Params())
		parser.Reset()
	case errInvalidChar:
		parser.Reset()
	default:
		// ignore
	}
}
