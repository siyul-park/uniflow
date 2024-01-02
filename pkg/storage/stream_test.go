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
	dbStream := memdb.NewStream()

	stream := NewStream(dbStream)
	defer stream.Close()

	id := ulid.Make()
	event := database.Event{OP: database.EventInsert, DocumentID: primitive.NewBinary(id.Bytes())}
	dbStream.Emit(event)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case evt, ok := <-stream.Next():
		assert.True(t, ok)
		assert.Equal(t, id, evt.NodeID)
	case <-ctx.Done():
		assert.Fail(t, "timeout")
	}
}
