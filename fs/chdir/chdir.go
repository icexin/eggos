package chdir

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
)

// assert that chdir.Chdirfs implements afero.Fs.
var _ afero.Fs = (*Chdirfs)(nil)

type Chdirfs struct {
	dir     string
	backend afero.Fs
}

func New(backend afero.Fs) *Chdirfs {
	return &Chdirfs{
		dir:     "/",
		backend: backend,
	}
}

func (c *Chdirfs) Chdir(dir string) error {
	name := c.name(dir)
	fi, err := c.backend.Stat(name)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return errors.New("not a dir")
	}
	c.dir = name
	return nil
}

func (c *Chdirfs) name(name string) string {
	if filepath.IsAbs(name) {
		return name
	}
	return filepath.Join(c.dir, name)
}

// Create creates a file in the filesystem, returning the file and an
// error, if any happens.
func (c *Chdirfs) Create(name string) (afero.File, error) {
	return c.backend.Create(c.name(name))
}

// Mkdir creates a directory in the filesystem, return an error if any
// happens.
func (c *Chdirfs) Mkdir(name string, perm os.FileMode) error {
	return c.backend.Mkdir(c.name(name), perm)
}

// MkdirAll creates a directory path and all parents that does not exist
// yet.
func (c *Chdirfs) MkdirAll(path string, perm os.FileMode) error {
	return c.backend.MkdirAll(c.name(path), perm)
}

// Open opens a file, returning it or an error, if any happens.
func (c *Chdirfs) Open(name string) (afero.File, error) {
	return c.backend.Open(c.name(name))
}

// OpenFile opens a file using the given flags and the given mode.
func (c *Chdirfs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	return c.backend.OpenFile(c.name(name), flag, perm)
}

// Remove removes a file identified by name, returning an error, if any
// happens.
func (c *Chdirfs) Remove(name string) error {
	return c.backend.Remove(c.name(name))
}

// RemoveAll removes a directory path and any children it contains. It
// does not fail if the path does not exist (return nil).
func (c *Chdirfs) RemoveAll(path string) error {
	return c.backend.RemoveAll(c.name(path))
}

// Rename renames a file.
func (c *Chdirfs) Rename(oldname string, newname string) error {
	return c.backend.Rename(c.name(oldname), c.name(newname))
}

// Stat returns a FileInfo describing the named file, or an error, if any
// happens.
func (c *Chdirfs) Stat(name string) (os.FileInfo, error) {
	return c.backend.Stat(c.name(name))
}

// The name of this FileSystem
func (c *Chdirfs) Name() string {
	return "chdirfs"
}

//Chmod changes the mode of the named file to mode.
func (c *Chdirfs) Chmod(name string, mode os.FileMode) error {
	return c.backend.Chmod(c.name(name), mode)
}

// Chown changes the uid and gid of the named file.
func (c *Chdirfs) Chown(name string, uid, gid int) error {
	return c.backend.Chown(name, uid, gid)
}

//Chtimes changes the access and modification times of the named file
func (c *Chdirfs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return c.backend.Chtimes(c.name(name), atime, mtime)
}
