package controllx

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestNewSwitchNode(t *testing.T) {
	n := NewSwitchNode(SwitchNodeConfig{})
	assert.NotNil(t, n)
	assert.NotZero(t, n.ID())

	_ = n.Close()
}

func TestSwitchNode_Send(t *testing.T) {
	n := NewSwitchNode(SwitchNodeConfig{})
	defer func() { _ = n.Close() }()

	in := port.New()
	inPort, _ := n.Port(node.PortIn)
	inPort.Link(in)

	err := n.Add("$.a", "out[0]")
	assert.NoError(t, err)
	err = n.Add("$.b", "out[1]")
	assert.NoError(t, err)
	err = n.Add("$.a = $.b", "out[2]")
	assert.NoError(t, err)
	err = n.Add("true", "out[3]")
	assert.NoError(t, err)

	testCases := []struct {
		when   any
		expect string
	}{
		{
			when: map[string]bool{
				"a": true,
			},
			expect: "out[0]",
		},
		{
			when: map[string]bool{
				"b": true,
			},
			expect: "out[1]",
		},
		{
			when: map[string]int{
				"a": 0,
				"b": 0,
			},
			expect: "out[2]",
		},
		{
			when: map[string]any{
				"a": 0,
				"b": false,
			},
			expect: "out[3]",
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.when), func(t *testing.T) {
			out := port.New()
			defer out.Close()
			outPort, _ := n.Port(tc.expect)
			outPort.Link(out)

			proc := process.New()
			defer proc.Exit(nil)

			inStream := in.Open(proc)
			outStream := out.Open(proc)

			inPayload, err := primitive.MarshalText(tc.when)
			assert.NoError(t, err)

			inStream.Send(packet.New(inPayload))

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			select {
			case outPck := <-outStream.Receive():
				assert.Equal(t, tc.when, outPck.Payload().Interface())
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}
		})
	}
}

func BenchmarkSwitchNode_Send(b *testing.B) {
	n := NewSwitchNode(SwitchNodeConfig{})
	defer func() { _ = n.Close() }()

	in := port.New()
	inPort, _ := n.Port(node.PortIn)
	inPort.Link(in)

	n.Add("$.a", "out[0]")
	n.Add("$.b", "out[1]")
	n.Add("$.a = $.b", "out[2]")
	n.Add("true", "out[3]")

	out := port.New()
	defer out.Close()
	outPort, _ := n.Port(port.SetIndex(node.PortOut, 0))
	outPort.Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	inStream := in.Open(proc)
	outStream := out.Open(proc)

	inPayload, _ := primitive.MarshalText(map[string]bool{
		"a": true,
	})
	
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		inStream.Send(packet.New(inPayload))
		<-outStream.Receive()
	}
}
