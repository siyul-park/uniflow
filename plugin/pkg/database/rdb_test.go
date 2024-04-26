package database

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	_ "github.com/mattn/go-sqlite3"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewRDBNode(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	n := NewRDBNode(db)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestRDBNode_SendAndReceive(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()

	_, err := db.Exec(
		"CREATE TABLE IF NOT EXISTS Foo (" +
			"id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL," +
			"name VARCHAR(255) NOT NULL" +
			")",
	)
	assert.NoError(t, err)

	n := NewRDBNode(db)
	defer n.Close()

	io := port.NewOut()
	io.Link(n.In(node.PortIO))

	proc := process.New()
	defer proc.Close()

	ioWriter := io.Open(proc)

	inPayload := primitive.NewSlice(
		primitive.NewString("INSERT INTO Foo(name) VALUES (?)"),
		primitive.NewSlice(primitive.NewString(faker.UUIDHyphenated())),
	)
	inPck := packet.New(inPayload)

	ioWriter.Write(inPck)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	select {
	case outPck := <-ioWriter.Receive():
		assert.Equal(t, primitive.NewSlice(), outPck.Payload())
	case <-ctx.Done():
		assert.Fail(t, ctx.Err().Error())
	}
}
