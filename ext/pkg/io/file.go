package io

import (
	"io"
	"os"
)

type nopReadWriteCloser struct {
	io.ReadWriter
}

// Open opens a file in the file system. It handles special files and normal files.
func OpenFile(name string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
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
}

func (c *nopReadWriteCloser) Close() error {
	return nil
}
