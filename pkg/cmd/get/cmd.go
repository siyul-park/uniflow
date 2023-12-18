package get

import (
	"github.com/siyul-park/uniflow/pkg/cmd/flag"
	"github.com/siyul-park/uniflow/pkg/cmd/printer"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/spf13/cobra"
)

// Config represents the configuration for the get command.
type Config struct {
	Scheme   *scheme.Scheme
	Database database.Database
}

const (
	flagNamespace = "namespace"
)

// NewCmd creates a new cobra.Command for the get command.
func NewCmd(config Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get and display applied resources",
		RunE:  runGetCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, flag.ToShorthand(flagNamespace), "", "Set the resource's namespace. If not set, use all namespace")

	return cmd
}

func runGetCommand(config Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()

		ns, err := cmd.Flags().GetString(flagNamespace)
		if err != nil {
			return err
		}

		st, err := storage.New(ctx, storage.Config{
			Scheme:   config.Scheme,
			Database: config.Database,
		})
		if err != nil {
			return err
		}

		filter := createNamespaceFilter(ns)
		specs, err := st.FindMany(ctx, filter)
		if err != nil {
			return err
		}

		return printer.PrintTable(cmd.OutOrStdout(), specs, printer.SpecTableColumnDefinitions)
	}
}

func createNamespaceFilter(ns string) *storage.Filter {
	if ns == "" {
		return nil
	}
	return storage.Where[string](scheme.KeyNamespace).EQ(ns)
}
