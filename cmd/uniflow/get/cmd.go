package get

import (
	"fmt"

	"github.com/siyul-park/uniflow/cmd/flag"
	"github.com/siyul-park/uniflow/cmd/printer"
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

// NewCmd creates a new cobra.Command for the get command.
func NewCmd(config Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get and display applied resources",
		RunE:  runGetCommand(config),
	}

	cmd.PersistentFlags().StringP(FlagNamespace, flag.ToShorthand(FlagNamespace), "", "Set the resource's namespace. If not set, use all namespace")

	return cmd
}

func runGetCommand(config Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()

		ns, err := cmd.Flags().GetString(FlagNamespace)
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

		return printSpecTable(cmd, specs)
	}
}

func createNamespaceFilter(ns string) *storage.Filter {
	if ns == "" {
		return nil
	}
	return storage.Where[string](scheme.KeyNamespace).EQ(ns)
}

func printSpecTable(cmd *cobra.Command, specs []scheme.Spec) error {
	tablePrinter, err := printer.NewTable(printer.SpecTableColumnDefinitions)
	if err != nil {
		return err
	}

	table, err := tablePrinter.Print(specs)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprint(cmd.OutOrStdout(), table); err != nil {
		return err
	}

	return nil
}
