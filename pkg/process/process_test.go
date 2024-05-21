package process

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	assert.NotNil(t, proc)
}

func TestProcess_Data(t *testing.T) {
	proc := New()
	defer proc.Exit(nil)

	assert.NotNil(t, proc.Data())
}

func BenchmarkNewProcess(b *testing.B) {
	for i := 0; i < b.N; i++ {
		proc := New()
		proc.Exit(nil)
	}
}
