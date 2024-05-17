package process

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/stretchr/testify/assert"
)

func TestStack_Has(t *testing.T) {
	s := newStack()
	defer s.Close()

	pck1 := packet.New(nil)
	pck2 := packet.New(nil)

	assert.False(t, s.Has(pck1, pck2))
	assert.False(t, s.Has(nil, pck1))
	assert.False(t, s.Has(nil, pck2))

	s.Add(pck1, pck2)
	assert.True(t, s.Has(pck1, pck2))
	assert.True(t, s.Has(nil, pck1))
	assert.True(t, s.Has(nil, pck2))
}

func TestStack_Add(t *testing.T) {
	s := newStack()
	defer s.Close()

	pck1 := packet.New(nil)
	pck2 := packet.New(nil)

	s.Add(nil, pck1)
	assert.True(t, s.Has(nil, pck1))

	s.Add(nil, pck2)
	assert.True(t, s.Has(nil, pck2))

	s.Add(pck1, pck2)
	assert.True(t, s.Has(pck1, pck2))
}

func TestStack_Stems(t *testing.T) {
	s := newStack()
	defer s.Close()

	pck1 := packet.New(nil)
	pck2 := packet.New(nil)

	s.Add(pck1, pck2)

	assert.Len(t, s.Stems(pck2), 1)
	assert.Len(t, s.Stems(pck1), 0)
}

func TestStack_Unwind(t *testing.T) {
	s := newStack()
	defer s.Close()

	pck1 := packet.New(nil)
	pck2 := packet.New(nil)
	pck3 := packet.New(nil)

	s.Add(pck1, pck2)
	s.Add(pck2, pck3)

	s.Unwind(pck3, pck2)
	assert.False(t, s.Has(pck2, pck3))
	assert.False(t, s.Has(nil, pck2))
	assert.False(t, s.Has(nil, pck3))

	s.Unwind(pck3, pck1)
	assert.False(t, s.Has(nil, pck1))
}

func TestStack_Clear(t *testing.T) {
	s := newStack()
	defer s.Close()

	pck1 := packet.New(nil)
	pck2 := packet.New(nil)
	pck3 := packet.New(nil)

	s.Add(pck1, pck2)
	s.Add(pck1, pck3)

	s.Clear(pck3)
	assert.True(t, s.Has(nil, pck1))
	assert.False(t, s.Has(nil, pck3))

	s.Clear(pck2)
	assert.False(t, s.Has(nil, pck1))
	assert.False(t, s.Has(nil, pck2))
}

func TestStack_Cost(t *testing.T) {
	s := newStack()
	defer s.Close()

	pck1 := packet.New(nil)
	pck2 := packet.New(nil)
	pck3 := packet.New(nil)

	s.Add(pck1, pck2)
	s.Add(pck1, pck3)

	cost := s.Cost(pck3, pck3)
	assert.Equal(t, 0, cost)

	cost = s.Cost(pck2, pck2)
	assert.Equal(t, 0, cost)

	cost = s.Cost(nil, pck1)
	assert.Equal(t, 1, cost)

	cost = s.Cost(pck1, pck2)
	assert.Equal(t, 1, cost)

	cost = s.Cost(pck1, pck3)
	assert.Equal(t, 1, cost)

	cost = s.Cost(pck2, pck3)
	assert.Equal(t, math.MaxInt, cost)
}

func TestStack_Near(t *testing.T) {
	s := newStack()
	defer s.Close()

	pck1 := packet.New(nil)
	pck2 := packet.New(nil)
	pck3 := packet.New(nil)

	s.Add(pck1, pck2)
	s.Add(pck1, pck3)

	near, cost := s.Near(pck3, []*packet.Packet{nil, pck1, pck2})
	assert.Equal(t, pck1, near)
	assert.Equal(t, 1, cost)
}

func TestStack_Done(t *testing.T) {
	s := newStack()
	defer s.Close()

	pck1 := packet.New(nil)
	pck2 := packet.New(nil)
	pck3 := packet.New(nil)

	s.Add(pck1, pck2)
	s.Add(pck1, pck3)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s.Unwind(pck3, pck1)

	select {
	case <-s.Done(pck3):
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}

	s.Unwind(pck2, pck2)

	select {
	case <-s.Done(nil):
	case <-ctx.Done():
		assert.NoError(t, ctx.Err())
	}
}
