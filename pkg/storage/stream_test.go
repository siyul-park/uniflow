package storage

import (
	"context"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/database/memdb"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/stretchr/testify/assert"
)

func TestStream_Next(t *testing.T) {
	rawStream := memdb.NewStream()

	stream := NewStream(rawStream)
	defer func() { _ = stream.Close() }()
	event := database.Event{OP: database.EventInsert, DocumentID: primitive.NewBinary(ulid.Make().Bytes())}

	rawStream.Emit(event)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	select {
	case evt, ok := <-stream.Next():
		assert.True(t, ok)
		assert.NotZero(t, evt.NodeID)
	case <-ctx.Done():
		assert.Fail(t, "timeout")
	}
}
