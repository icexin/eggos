package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/afero"
)

type logger struct {
	w       io.Writer
	backend afero.Fs
}

func New(w io.Writer, fs afero.Fs) afero.Fs {
	return &logger{
		w:       w,
		backend: fs,
	}
}

func (l *logger) logf(fmtstr string, args ...interface{}) {
	fmtstr = fmt.Sprintf("[%s] %s\n", l.backend.Name(), fmtstr)
	fmt.Fprintf(l.w, fmtstr, args...)
}

// Create creates a file in the filesystem, returning the file and an
// error, if any happens.
func (l *logger) Create(name string) (afero.File, error) {
	ret, err := l.backend.Create(name)
	l.logf("Create(%s) %v", name, err)
	return ret, err
}

// Mkdir creates a directory in the filesystem, return an error if any
// happens.
func (l *logger) Mkdir(name string, perm os.FileMode) error {
	err := l.backend.Mkdir(name, perm)
	l.logf("Mkdir(%s, %s) %v", name, perm, err)
	return err
}

// MkdirAll creates a directory path and all parents that does not exist
// yet.
func (l *logger) MkdirAll(path string, perm os.FileMode) error {
	err := l.backend.MkdirAll(path, perm)
	l.logf("MkdirAll(%s, %s) %v", path, perm, err)
	return err
}

// Open opens a file, returning it or an error, if any happens.
func (l *logger) Open(name string) (afero.File, error) {
	ret, err := l.backend.Open(name)
	l.logf("Open(%s) %v", name, err)
	return ret, err
}

// OpenFile opens a file using the given flags and the given mode.
func (l *logger) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	ret, err := l.backend.OpenFile(name, flag, perm)
	l.logf("OpenFile(%s, %x, %s) %v", name, flag, perm, err)
	return ret, err
}

// Remove removes a file identified by name, returning an error, if any
// happens.
func (l *logger) Remove(name string) error {
	err := l.backend.Remove(name)
	l.logf("Remove(%s) %v", name, err)
	return err
}

// RemoveAll removes a directory path and any children it contains. It
// does not fail if the path does not exist (return nil).
func (l *logger) RemoveAll(path string) error {
	err := l.backend.RemoveAll(path)
	l.logf("RemoveAll(%s) %v", path, err)
	return err
}

// Rename renames a file.
func (l *logger) Rename(oldname string, newname string) error {
	err := l.backend.Rename(oldname, newname)
	l.logf("Rename(%s, %s) %v", oldname, newname, err)
	return err
}

// Stat returns a FileInfo describing the named file, or an error, if any
// happens.
func (l *logger) Stat(name string) (os.FileInfo, error) {
	ret, err := l.backend.Stat(name)
	l.logf("Stat(%s) %v", name, err)
	return ret, err
}

// The name of this FileSystem
func (l *logger) Name() string {
	return "logger"
}

//Chmod changes the mode of the named file to mode.
func (l *logger) Chmod(name string, mode os.FileMode) error {
	err := l.backend.Chmod(name, mode)
	l.logf("Chmod(%s, %s) %v", name, mode, err)
	return err
}

//Chtimes changes the access and modification times of the named file
func (l *logger) Chtimes(name string, atime time.Time, mtime time.Time) error {
	err := l.backend.Chtimes(name, atime, mtime)
	l.logf("Chtimes(%s, %d, %d) %v", name, atime.Unix(), mtime.Unix(), err)
	return err
}
