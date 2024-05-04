package datastore

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewSQLNode(t *testing.T) {
	n, err := NewSQLNode("", "")
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func TestSQLNode_SetArguments(t *testing.T) {
	n, _ := NewSQLNode("", "")
	defer n.Close()

	err := n.SetArguments("[\"foo\", \"bar\"]")
	assert.NoError(t, err)
}

func TestSQLNode_SendAndReceive(t *testing.T) {
	t.Run("Query", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n, _ := NewSQLNode("SELECT * FROM Foo", "")
		defer n.Close()

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		proc := process.New()
		defer proc.Close()

		inWriter := in.Open(proc)

		inPayload := primitive.NewMap(
			primitive.NewString("name"),
			primitive.NewString(faker.UUIDHyphenated()),
		)
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-inWriter.Receive():
			assert.Equal(t, primitive.NewString("SELECT * FROM Foo"), outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})

	t.Run("Arguments", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		n, _ := NewSQLNode("SELECT * FROM Foo WHERE id = :name", "")
		defer n.Close()

		err := n.SetArguments("$")
		assert.NoError(t, err)

		in := port.NewOut()
		in.Link(n.In(node.PortIn))

		proc := process.New()
		defer proc.Close()

		inWriter := in.Open(proc)

		inPayload := primitive.NewMap(
			primitive.NewString("name"),
			primitive.NewString(faker.UUIDHyphenated()),
		)
		inPck := packet.New(inPayload)

		inWriter.Write(inPck)

		select {
		case outPck := <-inWriter.Receive():
			expect := primitive.NewSlice(
				primitive.NewString("SELECT * FROM Foo WHERE id = :name"),
				inPayload,
			)
			assert.Equal(t, expect, outPck.Payload())
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}
	})
}

func TestSQLNodeCodec_Decode(t *testing.T) {
	codec := NewSQLNodeCodec()

	spec := &SQLNodeSpec{
		Query:     "SELECT * FROM Foo",
		Arguments: "null",
	}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.NoError(t, n.Close())
}

func BenchmarkSQLNode_SendAndReceive(b *testing.B) {
	n, _ := NewSQLNode("SELECT * FROM Foo WHERE id = :name", "")
	defer n.Close()

	_ = n.SetArguments("$")

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	proc := process.New()
	defer proc.Close()

	inWriter := in.Open(proc)

	inPayload := primitive.NewMap(
		primitive.NewString("name"),
		primitive.NewString(faker.UUIDHyphenated()),
	)
	inPck := packet.New(inPayload)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inWriter.Write(inPck)
		<-inWriter.Receive()
	}
}
