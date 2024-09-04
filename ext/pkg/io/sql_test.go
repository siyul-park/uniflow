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
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestSQLNodeCodec_Decode(t *testing.T) {
	codec := NewSQLNodeCodec()

	spec := &SQLNodeSpec{
		Driver: "sqlite3",
		Source: ":memory:",
	}

	n, err := codec.Compile(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestNewSQLNode(t *testing.T) {
	db, _ := sqlx.Connect("sqlite3", ":memory:")
	defer db.Close()

	n := NewSQLNode(db)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestSQLNode_Isolation(t *testing.T) {
	db, _ := sqlx.Connect("sqlite3", ":memory:")
	defer db.Close()

	n := NewSQLNode(db)
	defer n.Close()

	isolation := sql.LevelSerializable
	n.SetIsolation(isolation)
	assert.Equal(t, isolation, n.Isolation())
}

func TestSQLNode_SendAndReceive(t *testing.T) {
	t.Run("RawSQL", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		db, _ := sqlx.Connect("sqlite3", "file::memory:?cache=shared")
		defer db.Close()

		n := NewSQLNode(db)
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

		var inPayload types.Value
		inPayload = types.NewSlice(
			types.NewString("INSERT INTO Foo(name) VALUES (?)"),
			types.NewSlice(types.NewString(faker.UUIDHyphenated())),
		)
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-inWriter.Receive():
			assert.Equal(t, types.NewSlice(), outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		inPayload = types.NewString("SELECT * FROM Foo")
		inPck = packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-inWriter.Receive():
			outPayload, ok := outPck.Payload().(types.Slice)
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

		n := NewSQLNode(db)
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

		var inPayload types.Value
		inPayload = types.NewSlice(
			types.NewString("INSERT INTO Foo(name) VALUES (:name)"),
			types.NewMap(
				types.NewString("name"),
				types.NewString(faker.UUIDHyphenated()),
			),
		)
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-inWriter.Receive():
			assert.Equal(t, types.NewSlice(), outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		inPayload = types.NewString("SELECT * FROM Foo")
		inPck = packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-inWriter.Receive():
			outPayload, ok := outPck.Payload().(types.Slice)
			assert.True(t, ok)
			assert.Equal(t, 1, outPayload.Len())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})
}

func BenchmarkSQLNode_SendAndReceive(b *testing.B) {
	db, _ := sqlx.Connect("sqlite3", "file::memory:?cache=shared")
	defer db.Close()

	_, _ = db.Exec(
		"CREATE TABLE Foo (" +
			"id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL," +
			"name VARCHAR(255) NOT NULL" +
			")",
	)

	n := NewSQLNode(db)
	defer n.Close()

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)

	inPayload := types.NewSlice(
		types.NewString("INSERT INTO Foo(name) VALUES (?)"),
		types.NewSlice(types.NewString(faker.UUIDHyphenated())),
	)
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		<-inWriter.Receive()
	}
}
