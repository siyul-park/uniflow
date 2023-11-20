package process

import (
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	proc := New()
	defer proc.Close()

	assert.NotNil(t, proc)
}

func TestProcess_ID(t *testing.T) {
	proc := New()
	defer proc.Close()

	assert.NotEqual(t, ulid.ULID{}, proc.ID())
}

func TestProcess_Stack(t *testing.T) {
	proc := New()
	defer proc.Close()

	assert.NotNil(t, proc.Stack())
}

func TestProcess_Close(t *testing.T) {
	proc := New()

	select {
	case <-proc.Done():
		assert.Fail(t, "proc.Done() is not empty.")
	default:
	}

	proc.Close()

	select {
	case <-proc.Done():
	default:
		assert.Fail(t, "proc.Done() is empty.")
	}
}
