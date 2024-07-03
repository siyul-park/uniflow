package mime

import "io"

type WriterFunc func([]byte) (int, error)

var _ io.Writer = WriterFunc(nil)

func (f WriterFunc) Write(p []byte) (int, error) {
	return f(p)
}
