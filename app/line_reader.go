package app

import (
	"bufio"
	"fmt"
	"io"

	"github.com/peterh/liner"
)

type LineReader interface {
	Prompt(string) (string, error)
	AppendHistory(string)
	SetAutoComplete(func(string) []string)
	Close() error
}

type simpleLineReader struct {
	r *bufio.Scanner
	w io.Writer
}

func newSimpleLineReader(r io.Reader, w io.Writer) LineReader {
	return &simpleLineReader{
		r: bufio.NewScanner(r),
		w: w,
	}
}

func (r *simpleLineReader) Prompt(prompt string) (string, error) {
	fmt.Fprintf(r.w, "%s", prompt)
	ok := r.r.Scan()
	if !ok {
		return "", io.EOF
	}
	return r.r.Text(), nil
}

func (r *simpleLineReader) AppendHistory(string) {
}

func (r *simpleLineReader) SetAutoComplete(f func(string) []string) {
}

func (r *simpleLineReader) Close() error {
	return nil
}

type lineEditor struct {
	*liner.State
}

func (l lineEditor) SetAutoComplete(f func(string) []string) {
	l.SetCompleter(f)
}

func newLineEditor() LineReader {
	r := liner.NewLiner()
	return lineEditor{
		State: r,
	}
}
