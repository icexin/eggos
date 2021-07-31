package cga

import (
	"bytes"
	"errors"
)

const (
	stateBegin = iota
	stateESC
	stateLeft
	stateParam
	stateDone
)

const (
	_ESC = 0x1b
)

var (
	errInvalidChar = errors.New("invalid char")
	errNormalChar  = errors.New("normal char")
	errCSIDone     = errors.New("done")
)

type ansiParser struct {
	state int

	action   byte
	parambuf []byte
	params   []string
}

func (p *ansiParser) step(ch byte) error {
	switch p.state {
	case stateBegin:
		if ch != _ESC {
			return errNormalChar
		}
		p.state = stateESC
	case stateESC:
		if ch != '[' {
			return errInvalidChar
		}
		p.state = stateLeft
	case stateLeft:
		switch {
		case ch >= 0x30 && ch <= 0x3f:
			p.state = stateParam
			p.parambuf = append(p.parambuf, ch)
		case ch >= 0x40 && ch <= 0x7f:
			p.action = ch
			p.state = stateDone
		default:
			return errInvalidChar
		}
	case stateParam:
		switch {
		case ch >= 0x30 && ch <= 0x3f:
			p.parambuf = append(p.parambuf, ch)
		case ch >= 0x40 && ch <= 0x7f:
			p.action = ch
			p.state = stateDone
		default:
			return errInvalidChar
		}
	}
	if p.state == stateDone {
		p.decodePram()
		return errCSIDone
	}
	return nil
}

func (p *ansiParser) decodePram() {
	if len(p.parambuf) == 0 {
		return
	}
	params := bytes.Split(p.parambuf, []byte(";"))
	for _, param := range params {
		p.params = append(p.params, string(param))
	}
}

func (p *ansiParser) Action() byte {
	return p.action
}

func (p *ansiParser) Params() []string {
	return p.params
}

func (p *ansiParser) Reset() {
	p.state = stateBegin
	p.action = 0
	p.parambuf = p.parambuf[:0]
	p.params = p.params[:0]
}
