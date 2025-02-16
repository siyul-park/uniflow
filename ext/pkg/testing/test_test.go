package testing

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

// 단순 실행 검증 테스트 (out[0]만 사용)
func TestTestNode_SimpleExecution(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	simpleNode := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		return packet.New(types.NewString("success")), nil
	})
	defer simpleNode.Close()

	n := NewTestNode()
	defer n.Close()

	n.Out(node.PortOut).Link(simpleNode.In(node.PortIn))

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	out := port.NewIn()
	simpleNode.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	outReader := out.Open(proc)

	inPck := packet.New(types.NewString("test"))
	inWriter.Write(inPck)

	select {
	case outPck := <-outReader.Read():
		payload := outPck.Payload()
		value, ok := payload.(types.String)
		assert.True(t, ok)
		assert.Equal(t, "success", value.String())

		outReader.Receive(outPck)
	case <-ctx.Done():
		assert.Fail(t, "timeout")
	}
}

func TestTestNode_SimpleExecutionError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	errorNode := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		return nil, packet.New(types.NewError(errors.New("test error")))
	})
	defer errorNode.Close()

	n := NewTestNode()
	defer n.Close()

	n.Out(node.PortOut).Link(errorNode.In(node.PortIn))

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	out := port.NewIn()
	errorNode.Out(node.PortError).Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	outReader := out.Open(proc)

	inPck := packet.New(types.NewString("test"))
	inWriter.Write(inPck)

	select {
	case outPck := <-outReader.Read():
		payload := outPck.Payload()
		e, ok := payload.(types.Error)
		assert.True(t, ok)
		assert.Equal(t, "test error", e.Error())

		outReader.Receive(outPck)
	case <-ctx.Done():
		assert.Fail(t, "timeout")
	}
}

// 확장 검증 테스트 (out[0]와 out[1] 모두 사용)
func TestTestNode_ExtendedExecution(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	workflowNode := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		return packet.New(types.NewString("result")), nil
	})
	defer workflowNode.Close()

	validationNode := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		payload := inPck.Payload()
		slice, ok := payload.(types.Slice)
		if !ok {
			return nil, packet.New(types.NewError(errors.New("invalid payload type")))
		}

		value, ok := types.Get[string](slice, 0)
		if !ok || value != "result" {
			return nil, packet.New(types.NewError(errors.New("unexpected value")))
		}

		index, ok := types.Get[int](slice, 1)
		if !ok || index != -1 {
			return nil, packet.New(types.NewError(errors.New("unexpected index")))
		}

		return packet.New(types.NewString("validation success")), nil
	})
	defer validationNode.Close()

	n := NewTestNode()
	defer n.Close()

	n.Out(node.PortOut).Link(workflowNode.In(node.PortIn))
	n.Out(node.PortWithIndex(node.PortOut, 1)).Link(validationNode.In(node.PortIn))

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	out := port.NewIn()
	validationNode.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	outReader := out.Open(proc)

	inPck := packet.New(types.NewString("test"))
	inWriter.Write(inPck)

	select {
	case outPck := <-outReader.Read():
		payload := outPck.Payload()
		value, ok := payload.(types.String)
		assert.True(t, ok)
		assert.Equal(t, "validation success", value.String())

		outReader.Receive(outPck)
	case <-ctx.Done():
		assert.Fail(t, "timeout")
	}
}

func TestTestNode_ExtendedExecutionError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	errorNode := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		return nil, packet.New(types.NewError(errors.New("workflow error")))
	})
	defer errorNode.Close()

	validationNode := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		payload := inPck.Payload()
		slice, ok := payload.(types.Slice)
		if !ok {
			return nil, packet.New(types.NewError(errors.New("invalid payload type")))
		}

		value, ok := slice.Get(0).(types.Error)
		if !ok {
			return nil, packet.New(types.NewError(errors.New("expected error value")))
		}
		if value.Error() != "workflow error" {
			return nil, packet.New(types.NewError(errors.New("unexpected error message")))
		}

		index, ok := types.Get[int](slice, 1)
		if !ok || index != -1 {
			return nil, packet.New(types.NewError(errors.New("unexpected index")))
		}

		return packet.New(types.NewString("validation success")), nil
	})
	defer validationNode.Close()

	n := NewTestNode()
	defer n.Close()

	n.Out(node.PortOut).Link(errorNode.In(node.PortIn))
	n.Out(node.PortWithIndex(node.PortOut, 1)).Link(validationNode.In(node.PortIn))

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	out := port.NewIn()
	validationNode.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	outReader := out.Open(proc)

	inPck := packet.New(types.NewString("test"))
	inWriter.Write(inPck)

	select {
	case outPck := <-outReader.Read():
		payload := outPck.Payload()
		value, ok := payload.(types.String)
		assert.True(t, ok)
		assert.Equal(t, "validation success", value.String())

		outReader.Receive(outPck)
	case <-ctx.Done():
		assert.Fail(t, "timeout")
	}
}

func TestTestNode_ExtendedValidationError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	workflowNode := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		return packet.New(types.NewString("wrong result")), nil
	})
	defer workflowNode.Close()

	validationNode := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		payload := inPck.Payload()
		slice, ok := payload.(types.Slice)
		if !ok {
			return nil, packet.New(types.NewError(errors.New("invalid payload type")))
		}

		value, ok := types.Get[string](slice, 0)
		if !ok || value != "wrong result" {
			return nil, packet.New(types.NewError(errors.New("unexpected value")))
		}

		index, ok := types.Get[int](slice, 1)
		if !ok || index != -1 {
			return nil, packet.New(types.NewError(errors.New("unexpected index")))
		}

		return nil, packet.New(types.NewError(errors.New("validation error")))
	})
	defer validationNode.Close()

	n := NewTestNode()
	defer n.Close()

	n.Out(node.PortOut).Link(workflowNode.In(node.PortIn))
	n.Out(node.PortWithIndex(node.PortOut, 1)).Link(validationNode.In(node.PortIn))

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	out := port.NewIn()
	workflowNode.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	outReader := out.Open(proc)

	inPck := packet.New(types.NewString("test"))
	inWriter.Write(inPck)

	select {
	case outPck := <-outReader.Read():
		payload := outPck.Payload()
		value, ok := payload.(types.String)
		assert.True(t, ok)
		assert.Equal(t, "wrong result", value.String())

		outReader.Receive(outPck)
	case <-ctx.Done():
		assert.Fail(t, "timeout")
	}
}

// 추가 테스트 시나리오
func TestTestNode_MultipleFrames(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	workflowNode := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		return packet.New(types.NewSlice(
			types.NewString("frame1"),
			types.NewString("frame2"),
			types.NewString("frame3"),
		)), nil
	})
	defer workflowNode.Close()

	var frameIndex int32 = -1
	validationNode := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		payload := inPck.Payload()
		slice, ok := payload.(types.Slice)
		if !ok {
			return nil, packet.New(types.NewError(errors.New("invalid payload type")))
		}

		value, ok := types.Get[string](slice, 0)
		if !ok {
			return nil, packet.New(types.NewError(errors.New("invalid value type")))
		}
		expectedValue := fmt.Sprintf("frame%d", frameIndex+2)
		if value != expectedValue {
			return nil, packet.New(types.NewError(errors.New("unexpected value")))
		}

		index, ok := types.Get[int](slice, 1)
		if !ok {
			return nil, packet.New(types.NewError(errors.New("invalid index type")))
		}

		expectedIndex := int(frameIndex)
		if index != expectedIndex {
			return nil, packet.New(types.NewError(errors.New("unexpected index")))
		}

		frameIndex++
		return packet.New(types.NewString("frame validated")), nil
	})
	defer validationNode.Close()

	n := NewTestNode()
	defer n.Close()

	n.Out(node.PortOut).Link(workflowNode.In(node.PortIn))
	n.Out(node.PortWithIndex(node.PortOut, 1)).Link(validationNode.In(node.PortIn))

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	out := port.NewIn()
	validationNode.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	outReader := out.Open(proc)

	inPck := packet.New(types.NewString("test"))
	inWriter.Write(inPck)

	frameCount := 0
	for frameCount < 3 {
		select {
		case outPck := <-outReader.Read():
			payload := outPck.Payload()
			value, ok := payload.(types.String)
			assert.True(t, ok)
			assert.Equal(t, "frame validated", value.String())

			outReader.Receive(outPck)
			frameCount++
		case <-ctx.Done():
			assert.Fail(t, "timeout")
			return
		}
	}
}

func TestTestNode_EmptyWorkflow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	emptyNode := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		return packet.New(nil), nil
	})
	defer emptyNode.Close()

	validationNode := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		payload := inPck.Payload()
		slice, ok := payload.(types.Slice)
		if !ok {
			return nil, packet.New(types.NewError(errors.New("invalid payload type")))
		}

		value := slice.Get(0)
		if value != nil {
			return nil, packet.New(types.NewError(errors.New("expected nil value")))
		}

		index, ok := types.Get[int](slice, 1)
		if !ok || index != -1 {
			return nil, packet.New(types.NewError(errors.New("unexpected index")))
		}

		return packet.New(types.NewString("empty validated")), nil
	})
	defer validationNode.Close()

	n := NewTestNode()
	defer n.Close()

	n.Out(node.PortOut).Link(emptyNode.In(node.PortIn))
	n.Out(node.PortWithIndex(node.PortOut, 1)).Link(validationNode.In(node.PortIn))

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	out := port.NewIn()
	validationNode.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	outReader := out.Open(proc)

	inPck := packet.New(types.NewString("test"))
	inWriter.Write(inPck)

	select {
	case outPck := <-outReader.Read():
		payload := outPck.Payload()
		value, ok := payload.(types.String)
		assert.True(t, ok)
		assert.Equal(t, "empty validated", value.String())

		outReader.Receive(outPck)
	case <-ctx.Done():
		assert.Fail(t, "timeout")
	}
}

func TestTestNode_ConcurrentExecution(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	workflowNode := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		return packet.New(types.NewString("concurrent")), nil
	})
	defer workflowNode.Close()

	validationNode := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		payload := inPck.Payload()
		slice, ok := payload.(types.Slice)
		if !ok {
			return nil, packet.New(types.NewError(errors.New("invalid payload type")))
		}

		value, ok := types.Get[string](slice, 0)
		if !ok || value != "concurrent" {
			return nil, packet.New(types.NewError(errors.New("unexpected value")))
		}

		index, ok := types.Get[int](slice, 1)
		if !ok || index != -1 {
			return nil, packet.New(types.NewError(errors.New("unexpected index")))
		}

		return packet.New(types.NewString("concurrent validated")), nil
	})
	defer validationNode.Close()

	n := NewTestNode()
	defer n.Close()

	n.Out(node.PortOut).Link(workflowNode.In(node.PortIn))
	n.Out(node.PortWithIndex(node.PortOut, 1)).Link(validationNode.In(node.PortIn))

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			in := port.NewOut()
			in.Link(n.In(node.PortIn))

			out := port.NewIn()
			validationNode.Out(node.PortOut).Link(out)

			proc := process.New()
			defer proc.Exit(nil)

			inWriter := in.Open(proc)
			outReader := out.Open(proc)

			inPck := packet.New(types.NewString("test"))
			inWriter.Write(inPck)

			select {
			case outPck := <-outReader.Read():
				payload := outPck.Payload()
				value, ok := payload.(types.String)
				assert.True(t, ok)
				assert.Equal(t, "concurrent validated", value.String())

				outReader.Receive(outPck)
			case <-ctx.Done():
				assert.Fail(t, "timeout")
			}
		}()
	}

	wg.Wait()
}

func TestTestNode_HelloWorldWorkflow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// hello_world step node
	snippetNode1 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		return packet.New(types.NewString("Hello, World!\n")), nil
	})
	defer snippetNode1.Close()

	printNode1 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		payload := inPck.Payload()
		str, ok := payload.(types.String)
		if !ok {
			return nil, packet.New(types.NewError(errors.New("invalid payload type")))
		}
		return packet.New(str), nil
	})
	defer printNode1.Close()

	// good_bye step node
	snippetNode2 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		return packet.New(types.NewString("Good, Bye!\n")), nil
	})
	defer snippetNode2.Close()

	printNode2 := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		payload := inPck.Payload()
		str, ok := payload.(types.String)
		if !ok {
			return nil, packet.New(types.NewError(errors.New("invalid payload type")))
		}
		return packet.New(str), nil
	})
	defer printNode2.Close()

	snippetNode1.Out(node.PortOut).Link(printNode1.In(node.PortIn))
	snippetNode2.Out(node.PortOut).Link(printNode2.In(node.PortIn))

	collectorNode := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		payload := inPck.Payload()
		if str, ok := payload.(types.String); ok {
			if str.String() == "Hello, World!\n" {
				return packet.New(types.NewString("Hello, World!\n")), nil
			} else if str.String() == "Good, Bye!\n" {
				return packet.New(types.NewString("Good, Bye!\n")), nil
			}
		}
		return nil, packet.New(types.NewError(errors.New("unexpected input")))
	})
	defer collectorNode.Close()

	var outputIndex int32 = -1
	validationNode := node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
		payload := inPck.Payload()
		slice, ok := payload.(types.Slice)
		if !ok {
			return nil, packet.New(types.NewError(errors.New("invalid payload type")))
		}

		value, ok := types.Get[string](slice, 0)
		if !ok {
			return nil, packet.New(types.NewError(errors.New("invalid value type")))
		}

		index, ok := types.Get[int](slice, 1)
		if !ok {
			return nil, packet.New(types.NewError(errors.New("invalid index type")))
		}

		expectedValue := ""
		expectedIndex := int(outputIndex)

		switch outputIndex {
		case -1:
			expectedValue = "Hello, World!\n"
		case 0:
			expectedValue = "Good, Bye!\n"
		default:
			return nil, packet.New(types.NewError(errors.New("unexpected output index")))
		}

		if value != expectedValue {
			return nil, packet.New(types.NewError(fmt.Errorf("unexpected value: got %q, want %q", value, expectedValue)))
		}
		if index != expectedIndex {
			return nil, packet.New(types.NewError(fmt.Errorf("unexpected index: got %d, want %d", index, expectedIndex)))
		}

		atomic.AddInt32(&outputIndex, 1)
		return packet.New(types.NewString("validation success")), nil
	})
	defer validationNode.Close()

	n := NewTestNode()
	defer n.Close()

	n.Out(node.PortOut).Link(snippetNode1.In(node.PortIn))
	printNode1.Out(node.PortOut).Link(collectorNode.In(node.PortIn))
	n.Out(node.PortOut).Link(snippetNode2.In(node.PortIn))
	printNode2.Out(node.PortOut).Link(collectorNode.In(node.PortIn))
	n.Out(node.PortWithIndex(node.PortOut, 1)).Link(validationNode.In(node.PortIn))

	in := port.NewOut()
	in.Link(n.In(node.PortIn))

	out := port.NewIn()
	validationNode.Out(node.PortOut).Link(out)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	outReader := out.Open(proc)

	inPck := packet.New(types.NewString("start"))
	inWriter.Write(inPck)

	for i := 0; i < 2; i++ {
		select {
		case outPck := <-outReader.Read():
			payload := outPck.Payload()
			value, ok := payload.(types.String)
			assert.True(t, ok)
			assert.Equal(t, "validation success", value.String())

			outReader.Receive(outPck)
		case <-ctx.Done():
			assert.Fail(t, "timeout")
			return
		}
	}
}
