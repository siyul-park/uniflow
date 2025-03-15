package store

import (
	"context"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/types"
)

type Cursor interface {
	Next(ctx context.Context) bool
	Decode(val any) error
	Close(ctx context.Context) error
}

type cursor struct {
	docs []types.Map
}

var _ Cursor = (*cursor)(nil)

func newCursor(docs []types.Map) *cursor {
	return &cursor{docs: append([]types.Map{nil}, docs...)}
}

func (c *cursor) Next(_ context.Context) bool {
	if len(c.docs) <= 1 {
		return false
	}
	c.docs = c.docs[1:]
	return true
}

func (c *cursor) Decode(val any) error {
	if len(c.docs) == 0 {
		return errors.WithStack(encoding.ErrUnsupportedType)
	}
	return types.Unmarshal(c.docs[0], val)
}

func (c *cursor) Close(_ context.Context) error {
	c.docs = nil
	return nil
}
