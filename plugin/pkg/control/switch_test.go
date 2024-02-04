package control

import (
	"context"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestSwitchNodeCodec_Decode(t *testing.T) {
	codec := NewSwitchNodeCodec()

	spec := &SwitchNodeSpec{
		Lang: LangJSONata,
		Match: []Condition{
			{
				When: "$.foo = \"bar\"",
				Port: node.MultiPort(node.PortOut, 0),
			},
		},
	}

	n, err := codec.Decode(spec)
	assert.NoError(t, err)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestNewSwitchNode(t *testing.T) {
	n := NewSwitchNode(LangJSONata)
	assert.NotNil(t, n)

	assert.NoError(t, n.Close())
}

func TestSwitchNode_Add(t *testing.T) {
	t.Run(LangTypescript, func(t *testing.T) {
		n := NewSwitchNode(LangTypescript)
		defer n.Close()

		err := n.Add("$.foo === \"bar\"", node.MultiPort(node.PortOut, 0))
		assert.NoError(t, err)
	})

	t.Run(LangJavascript, func(t *testing.T) {
		n := NewSwitchNode(LangJavascript)
		defer n.Close()

		err := n.Add("$.foo === \"bar\"", node.MultiPort(node.PortOut, 0))
		assert.NoError(t, err)
	})

	t.Run(LangJSONata, func(t *testing.T) {
		n := NewSwitchNode(LangJSONata)
		defer n.Close()

		err := n.Add("$.foo = \"bar\"", node.MultiPort(node.PortOut, 0))
		assert.NoError(t, err)
	})
}

func TestSwitchNode_SendAndReceive(t *testing.T) {
	t.Run(LangTypescript, func(t *testing.T) {
		n := NewSwitchNode(LangTypescript)
		defer n.Close()

		_ = n.Add("$.foo === \"bar\"", node.MultiPort(node.PortOut, 0))

		in := port.New()
		inPort, _ := n.Port(node.PortIn)
		inPort.Link(in)

		out := port.New()
		outPort, _ := n.Port(node.MultiPort(node.PortOut, 0))
		outPort.Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inStream := in.Open(proc)
		outStream := out.Open(proc)

		inPayload := primitive.NewMap(primitive.NewString("foo"), primitive.NewString("bar"))
		inPck := packet.New(inPayload)

		inStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-outStream.Receive():
			assert.Equal(t, inPayload, outPck.Payload())
			outStream.Send(outPck)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		select {
		case backPck := <-inStream.Receive():
			assert.NotNil(t, backPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run(LangJavascript, func(t *testing.T) {
		n := NewSwitchNode(LangJavascript)
		defer n.Close()

		_ = n.Add("$.foo === \"bar\"", node.MultiPort(node.PortOut, 0))

		in := port.New()
		inPort, _ := n.Port(node.PortIn)
		inPort.Link(in)

		out := port.New()
		outPort, _ := n.Port(node.MultiPort(node.PortOut, 0))
		outPort.Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inStream := in.Open(proc)
		outStream := out.Open(proc)

		inPayload := primitive.NewMap(primitive.NewString("foo"), primitive.NewString("bar"))
		inPck := packet.New(inPayload)

		inStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-outStream.Receive():
			assert.Equal(t, inPayload, outPck.Payload())
			outStream.Send(outPck)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		select {
		case backPck := <-inStream.Receive():
			assert.NotNil(t, backPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})

	t.Run(LangJSONata, func(t *testing.T) {
		n := NewSwitchNode(LangJSONata)
		defer n.Close()

		_ = n.Add("$.foo = \"bar\"", node.MultiPort(node.PortOut, 0))

		in := port.New()
		inPort, _ := n.Port(node.PortIn)
		inPort.Link(in)

		out := port.New()
		outPort, _ := n.Port(node.MultiPort(node.PortOut, 0))
		outPort.Link(out)

		proc := process.New()
		defer proc.Exit(nil)

		inStream := in.Open(proc)
		outStream := out.Open(proc)

		inPayload := primitive.NewMap(primitive.NewString("foo"), primitive.NewString("bar"))
		inPck := packet.New(inPayload)

		inStream.Send(inPck)

		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		select {
		case outPck := <-outStream.Receive():
			assert.Equal(t, inPayload, outPck.Payload())
			outStream.Send(outPck)
		case <-ctx.Done():
			assert.Fail(t, ctx.Err().Error())
		}

		select {
		case backPck := <-inStream.Receive():
			assert.NotNil(t, backPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
		}
	})
}

func BenchmarkSwitchNode_SendAndReceive(b *testing.B) {
	b.Run(LangTypescript, func(b *testing.B) {
		n := NewSwitchNode(LangTypescript)
		defer n.Close()

		_ = n.Add("$.foo === \"bar\"", node.MultiPort(node.PortOut, 0))

		in := port.New()
		inPort, _ := n.Port(node.PortIn)
		inPort.Link(in)

		out := port.New()
		outPort, _ := n.Port(node.MultiPort(node.PortOut, 0))
		outPort.Link(out)

		proc := process.New()
		defer proc.Exit(nil)
		defer proc.Stack().Close()

		inStream := in.Open(proc)
		outStream := out.Open(proc)

		inPayload := primitive.NewMap(primitive.NewString("foo"), primitive.NewString("bar"))
		inPck := packet.New(inPayload)

		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				inStream.Send(inPck)
				<-outStream.Receive()
			}
		})
	})

	b.Run(LangJavascript, func(b *testing.B) {
		n := NewSwitchNode(LangJavascript)
		defer n.Close()

		_ = n.Add("$.foo === \"bar\"", node.MultiPort(node.PortOut, 0))

		in := port.New()
		inPort, _ := n.Port(node.PortIn)
		inPort.Link(in)

		out := port.New()
		outPort, _ := n.Port(node.MultiPort(node.PortOut, 0))
		outPort.Link(out)

		proc := process.New()
		defer proc.Exit(nil)
		defer proc.Stack().Close()

		inStream := in.Open(proc)
		outStream := out.Open(proc)

		inPayload := primitive.NewMap(primitive.NewString("foo"), primitive.NewString("bar"))
		inPck := packet.New(inPayload)

		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				inStream.Send(inPck)
				<-outStream.Receive()
			}
		})
	})

	b.Run(LangJSONata, func(b *testing.B) {
		n := NewSwitchNode(LangJSONata)
		defer n.Close()

		_ = n.Add("$.foo = \"bar\"", node.MultiPort(node.PortOut, 0))

		in := port.New()
		inPort, _ := n.Port(node.PortIn)
		inPort.Link(in)

		out := port.New()
		outPort, _ := n.Port(node.MultiPort(node.PortOut, 0))
		outPort.Link(out)

		proc := process.New()
		defer proc.Exit(nil)
		defer proc.Stack().Close()

		inStream := in.Open(proc)
		outStream := out.Open(proc)

		inPayload := primitive.NewMap(primitive.NewString("foo"), primitive.NewString("bar"))
		inPck := packet.New(inPayload)

		b.ResetTimer()

		b.RunParallel(func(p *testing.PB) {
			for p.Next() {
				inStream.Send(inPck)
				<-outStream.Receive()
			}
		})
	})
}
