package cmd

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestCommand_Execute(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	fs := afero.NewMemMapFs()

	cpuprofile := "cpu.prof"
	memprofile := "mem.prof"

	output := new(bytes.Buffer)

	cmd := NewCommand(Config{
		FS: fs,
	})
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetContext(ctx)

	cmd.AddCommand(&cobra.Command{
		Use:       "dummy",
		ValidArgs: []string{specs, values},
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	})

	cmd.SetArgs([]string{"dummy", fmt.Sprintf("--%s", flagCPUProfile), cpuprofile, fmt.Sprintf("--%s", flagMemProfile), memprofile})

	err := cmd.Execute()
	require.NoError(t, err)
}
