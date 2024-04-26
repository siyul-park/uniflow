package database

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewRDBNode(t *testing.T) {
	db, _ := sqlx.Connect("sqlite3", ":memory:")
	defer db.Close()

	n := NewRDBNode(db)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestRDBNode_SendAndReceive(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	db, _ := sqlx.Connect("sqlite3", "file::memory:?cache=shared")
	defer db.Close()

	n := NewRDBNode(db)
	defer n.Close()

	_, err := db.ExecContext(ctx,
		"CREATE TABLE Foo ("+
			"id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,"+
			"name VARCHAR(255) NOT NULL"+
			")",
	)
	assert.NoError(t, err)

	io := port.NewOut()
	io.Link(n.In(node.PortIO))

	proc := process.New()
	defer proc.Close()

	ioWriter := io.Open(proc)

	var inPayload primitive.Value
	inPayload = primitive.NewSlice(
		primitive.NewString("INSERT INTO Foo(name) VALUES (?)"),
		primitive.NewSlice(primitive.NewString(faker.UUIDHyphenated())),
	)
	inPck := packet.New(inPayload)

	ioWriter.Write(inPck)

	select {
	case outPck := <-ioWriter.Receive():
		assert.Equal(t, primitive.NewSlice(), outPck.Payload())
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}

	inPayload = primitive.NewString("SELECT * FROM Foo")
	inPck = packet.New(inPayload)

	ioWriter.Write(inPck)

	select {
	case outPck := <-ioWriter.Receive():
		outPayload, ok := outPck.Payload().(*primitive.Slice)
		assert.True(t, ok)
		assert.Equal(t, 1, outPayload.Len())
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}
