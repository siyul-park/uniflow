package packet

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReader_AddHook(t *testing.T) {
	w := NewWriter()
	defer w.Close()

	r := NewReader()
	defer r.Close()

	count := 0
	r.AddInboundHook(HookFunc(func(_ *Packet) {
		count += 1
	}))
	r.AddOutboundHook(HookFunc(func(_ *Packet) {
		count += 1
	}))

	w.Link(r)

	out := New(nil)

	w.Write(out)

	in := <-r.Read()
	require.Equal(t, 1, count)

	r.Receive(in)
	require.Equal(t, 2, count)

	back, ok := <-w.Receive()
	require.True(t, ok)
	require.Equal(t, in, back)
}

func TestReader_Receive(t *testing.T) {
	w := NewWriter()
	defer w.Close()

	r := NewReader()
	defer r.Close()

	w.Link(r)

	out1 := New(nil)
	out2 := New(nil)

	w.Write(out1)
	w.Write(out2)

	in1 := <-r.Read()
	in2 := <-r.Read()

	ok := r.Receive(in1)
	require.True(t, ok)

	ok = r.Receive(in2)
	require.True(t, ok)

	back1, ok := <-w.Receive()
	require.True(t, ok)
	require.Equal(t, in1, back1)

	back2, ok := <-w.Receive()
	require.True(t, ok)
	require.Equal(t, in2, back2)
}

func BenchmarkReader_Receive(b *testing.B) {
	w := NewWriter()
	defer w.Close()

	r := NewReader()
	defer r.Close()

	w.Link(r)

	b.RunParallel(func(p *testing.PB) {
		out := New(nil)

		for p.Next() {
			b.StopTimer()

			count := w.Write(out)
			require.Equal(b, 1, count)

			in, ok := <-r.Read()
			require.True(b, ok)

			b.StartTimer()

			ok = r.Receive(in)
			require.True(b, ok)

			_, ok = <-w.Receive()
			require.True(b, ok)
		}
	})
}
