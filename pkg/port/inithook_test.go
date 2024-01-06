package port

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/stretchr/testify/assert"
)

func TestInitOnceHook(t *testing.T) {
	count := 0

	h := &InitOnceHook{
		Hook: InitHookFunc(func(_ *process.Process) {
			count += 1
		}),
	}

	proc := process.New()
	defer proc.Exit(nil)

	h.Init(proc)
	assert.Equal(t, 1, count)

	h.Init(proc)
	assert.Equal(t, 1, count)
}
