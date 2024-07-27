package cli

import (
	"github.com/siyul-park/uniflow/cmd/pkg/printer"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/spf13/cobra"
)

// GetConfig represents the configuration for the get command.
type GetConfig struct {
	SpecStore spec.Store
}

// NewGetCommand creates a new cobra.Command for the get command.
func NewGetCommand(config GetConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get node specifications from the specified namespace",
		RunE:  runGetCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), spec.DefaultNamespace, "Set the resource's namespace. If not set, use all namespace")

	return cmd
}

func runGetCommand(config GetConfig) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()

		namespace, err := cmd.Flags().GetString(flagNamespace)
		if err != nil {
			return err
		}

		specs, err := config.SpecStore.Load(ctx, &spec.Meta{Namespace: namespace})
		if err != nil {
			return err
		}

		return printer.PrintTable(cmd.OutOrStdout(), specs, printer.SpecTableColumnDefinitions)
	}
}
