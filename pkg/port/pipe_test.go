package port

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestPipe(t *testing.T) {
	inPort0, outPort0 := Pipe()
	defer inPort0.Close()
	defer outPort0.Close()

	inPort1, outPort1 := NewIn(), NewOut()
	defer inPort1.Close()
	defer outPort1.Close()

	outPort0.Link(inPort1)
	outPort1.Link(inPort0)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := outPort1.Open(proc)
	outReader := inPort1.Open(proc)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	inPck := packet.New(types.NewString(faker.UUIDHyphenated()))
	inWriter.Write(inPck)

	select {
	case outPck := <-outReader.Read():
		assert.Equal(t, inPck.Payload(), outPck.Payload())
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}
