package io

import (
	"io"
	"os"
)

// FileSystem interface abstracts the file operations.
type FileSystem interface {
	Open(name string, flag int) (io.ReadWriteCloser, error)
}

// FileOpenFunc is a function type that matches the signature of os.OpenFile.
type FileOpenFunc func(name string, flag int) (io.ReadWriteCloser, error)

type nopReadWriteCloser struct {
	io.ReadWriter
}

var _ FileSystem = (FileOpenFunc)(nil)

// NewOSFileSystem creates a new FileSystem that wraps os.OpenFile with special cases for stdin, stdout, and stderr.
func NewOSFileSystem() FileSystem {
	return FileOpenFunc(func(name string, flag int) (io.ReadWriteCloser, error) {
		switch name {
		case "/dev/stdin", "stdin":
			return &nopReadWriteCloser{os.Stdin}, nil
		case "/dev/stdout", "stdout":
			return &nopReadWriteCloser{os.Stdout}, nil
		case "/dev/stderr", "stderr":
			return &nopReadWriteCloser{os.Stderr}, nil
		default:
			return os.OpenFile(name, flag, 0666)
		}
	})
}

// Open opens a file with the given name, flag, and permissions.
func (f FileOpenFunc) Open(name string, flag int) (io.ReadWriteCloser, error) {
	return f(name, flag)
}

// Close is a no-op for ReadWriteCloserWrapper since stdin, stdout, and stderr shouldn't be closed.
func (c *nopReadWriteCloser) Close() error {
	return nil
}
