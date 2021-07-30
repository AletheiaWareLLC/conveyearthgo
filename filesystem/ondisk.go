package filesystem

import (
	"io"
	"io/fs"
	"os"
	"path"
)

func NewOnDisk(path string) *OnDisk {
	return &OnDisk{
		path: path,
	}
}

type OnDisk struct {
	path string
}

func (fs *OnDisk) Open(name string) (fs.File, error) {
	return os.Open(path.Join(fs.path, name))
}

func (fs *OnDisk) Create(name string) (io.WriteCloser, error) {
	return os.Create(path.Join(fs.path, name))
}

func (fs *OnDisk) Rename(old, new string) error {
	return os.Rename(path.Join(fs.path, old), path.Join(fs.path, new))
}
