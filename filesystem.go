package conveyearthgo

import (
	"io"
	"io/fs"
)

type Filesystem interface {
	Open(string) (fs.File, error)
	Create(string) (io.WriteCloser, error)
	Rename(string, string) error
}
