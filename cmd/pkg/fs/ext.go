package fs

import (
	"io/fs"
	"os"
	"path/filepath"
)

type extFS struct{}

var _ fs.FS = (*extFS)(nil)

func ExtFS() fs.FS {
	return &extFS{}
}

func (f *extFS) Open(name string) (fs.File, error) {
	if filepath.IsLocal(name) {
		dir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		name = filepath.Join(dir, name)
	}
	return os.Open(name)
}
