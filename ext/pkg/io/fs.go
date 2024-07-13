package io

import (
	"io"
	"os"
)

type FS interface {
	OpenFile(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error)
}

type OpenFileFunc func(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error)

type nopReadWriteCloser struct {
	io.ReadWriter
}

var _ FS = (OpenFileFunc)(nil)

func NewOsFs() FS {
	return OpenFileFunc(func(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
		switch name {
		case "/dev/stdin", "stdin":
			return &nopReadWriteCloser{os.Stdin}, nil
		case "/dev/stdout", "stdout":
			return &nopReadWriteCloser{os.Stdout}, nil
		case "/dev/stderr", "stderr":
			return &nopReadWriteCloser{os.Stderr}, nil
		default:
			return os.OpenFile(name, flag, perm)
		}
	})
}

func (f OpenFileFunc) OpenFile(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
	return f(name, flag, perm)
}

func (c *nopReadWriteCloser) Close() error {
	return nil
}
