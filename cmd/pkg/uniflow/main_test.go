package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	if os.Getenv("_TEST_IS_MAIN") != "" {
		main()
		os.Exit(0)
	}

	os.Setenv("_TEST_IS_MAIN", "1")
	os.Exit(m.Run())
}

func TestCommand(t *testing.T) {
	exe, err := os.Executable()
	assert.NoError(t, err)

	tests := []struct {
		args []string
	}{
		{
			args: []string{"start", "-h"},
		},
		{
			args: []string{"test", "-h"},
		},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			cmd := exec.Command(exe, tt.args...)
			cmd.Env = os.Environ()
			cmd.Dir = t.TempDir()

			err := cmd.Run()
			assert.NoError(t, err)
		})
	}
}
