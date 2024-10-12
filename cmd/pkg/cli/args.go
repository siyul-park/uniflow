package cli

import (
	"github.com/spf13/cobra"
)

const (
	specs   = "specs"
	secrets = "secrets"
	charts  = "charts"
)

func runs(runs map[string]func(cmd *cobra.Command) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return nil
		}
		run, ok := runs[args[0]]
		if !ok {
			return nil
		}
		return run(cmd)
	}
}
