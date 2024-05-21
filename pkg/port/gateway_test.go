package port

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/stretchr/testify/assert"
)

func TestGateway_Write(t *testing.T) {
	r := NewReader()
	defer r.Close()

	g := NewGateway([]*Reader{r}, ForwardHookFunc(func(pcks []*packet.Packet) bool {
		return true
	}))
	defer g.Close()

	pck := packet.New(nil)

	count := g.Write(pck, r)
	assert.Equal(t, 1, count)
}
