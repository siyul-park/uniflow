package process

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcess_Data(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	assert.NotNil(t, proc.Data())
}

func TestProcess_AddExitHook(t *testing.T) {
	proc := New()

	count := 0
	h := ExitHookFunc(func(err error) {
		count++
	})
	proc.AddExitHook(h)

	proc.Exit(nil)
	assert.Equal(t, 1, count)
}

func BenchmarkNewProcess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		proc := New()
		proc.Exit(nil)
	}
}
