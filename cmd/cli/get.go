package cli

import (
	"github.com/siyul-park/uniflow/cmd/printer"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/spf13/cobra"
)

// GetConfig represents the configuration for the get command.
type GetConfig struct {
	Scheme   *spec.Scheme
	Database database.Database
}

// NewGetCommand creates a new cobra.Command for the get command.
func NewGetCommand(config GetConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get and display resources in namespace",
		RunE:  runGetCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), "", "Set the resource's namespace. If not set, use all namespace")

	return cmd
}

func runGetCommand(config GetConfig) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()

		namespace, err := cmd.Flags().GetString(flagNamespace)
		if err != nil {
			return err
		}

		st, err := spec.NewStorage(ctx, spec.StorageConfig{
			Scheme:   config.Scheme,
			Database: config.Database,
		})
		if err != nil {
			return err
		}

		var filter *spec.Filter
		if namespace != "" {
			filter = spec.Where[string](spec.KeyNamespace).EQ(namespace)
		}

		specs, err := st.FindMany(ctx, filter)
		if err != nil {
			return err
		}

		return printer.PrintTable(cmd.OutOrStdout(), specs, printer.SpecTableColumnDefinitions)
	}
}
