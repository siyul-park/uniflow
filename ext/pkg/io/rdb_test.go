package io

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
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

func TestRDBNode_Isolation(t *testing.T) {
	db, _ := sqlx.Connect("sqlite3", ":memory:")
	defer db.Close()

	n := NewRDBNode(db)
	defer n.Close()

	isolation := sql.LevelSerializable
	n.SetIsolation(isolation)
	assert.Equal(t, isolation, n.Isolation())
}

func TestRDBNode_SendAndReceive(t *testing.T) {
	t.Run("RawSQL", func(t *testing.T) {
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

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)

		var inPayload object.Object
		inPayload = object.NewSlice(
			object.NewString("INSERT INTO Foo(name) VALUES (?)"),
			object.NewSlice(object.NewString(faker.UUIDHyphenated())),
		)
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-inWriter.Receive():
			assert.Equal(t, object.NewSlice(), outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		inPayload = object.NewString("SELECT * FROM Foo")
		inPck = packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-inWriter.Receive():
			outPayload, ok := outPck.Payload().(object.Slice)
			assert.True(t, ok)
			assert.Equal(t, 1, outPayload.Len())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("NamedSQL", func(t *testing.T) {
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

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		proc := process.New()
		defer proc.Exit(nil)

		inWriter := in.Open(proc)

		var inPayload object.Object
		inPayload = object.NewSlice(
			object.NewString("INSERT INTO Foo(name) VALUES (:name)"),
			object.NewMap(
				object.NewString("name"),
				object.NewString(faker.UUIDHyphenated()),
			),
		)
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-inWriter.Receive():
			assert.Equal(t, object.NewSlice(), outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		inPayload = object.NewString("SELECT * FROM Foo")
		inPck = packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-inWriter.Receive():
			outPayload, ok := outPck.Payload().(object.Slice)
			assert.True(t, ok)
			assert.Equal(t, 1, outPayload.Len())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})
}

func TestRDBNodeCodec_Decode(t *testing.T) {
	codec := NewRDBNodeCodec()

	spec := &RDBNodeSpec{
		Driver: "sqlite3",
		Source: ":memory:",
	}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func BenchmarkRDBNode_SendAndReceive(b *testing.B) {
	db, _ := sqlx.Connect("sqlite3", "file::memory:?cache=shared")
	defer db.Close()

	_, _ = db.Exec(
		"CREATE TABLE Foo (" +
			"id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL," +
			"name VARCHAR(255) NOT NULL" +
			")",
	)

	n := NewRDBNode(db)
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := object.NewSlice(
		object.NewString("INSERT INTO Foo(name) VALUES (?)"),
		object.NewSlice(object.NewString(faker.UUIDHyphenated())),
	)
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		<-inWriter.Receive()
	}
}
