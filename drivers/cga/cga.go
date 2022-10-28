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

func setCursorHome() {
	getbackend().SetPos(0)
}

func eraseLine(method EraseMethod) {
	backend := getbackend()
	pos := backend.GetPos()
	switch method {
	case EraseMethod_Line:
		end := (pos/80 + 1) * 80
		for i := pos; i < end; i++ {
			backend.WritePos(i, ' ')
		}

	case EraseMethod_All:
		for i := 0; i < 25*80; i++ {
			backend.WritePos(i, ' ')
		}

	default:
		panic("unsupported erase line method")
	}
}

func writeCSI(action byte, params []string) {
	// fmt.Fprintf(os.Stderr, "action:%c, params:%v\n", action, params)
	switch action {
	// set cursor column
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
			eraseLine(EraseMethod_Line)
		} else {
			//n, _ := strconv.Atoi(params[0])
			eraseLine(EraseMethod_Unknown)
		}

	// Erase screen - note; this action *looks* like it ought
	// to be just <ESC>[J - which is 'erase from current line to
	// bottom of screen', but I actually want it to mimic <ESC>[2J
	// which is 'erase screen and return to top'
	//
	// Hopefully nobody ever uses my fork if they expect that functionality
	case 'J':
		eraseLine(EraseMethod_All)
		setCursorHome()

	default:
		// ignore
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
			} else {
				backend.WriteByte('?')
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
		getbackend().WriteByte(ch)
		// ignore
	}
}
