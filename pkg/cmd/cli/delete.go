package cli

import (
	"io/fs"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/cmd/scanner"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/spf13/cobra"
)

// DeleteConfig represents the configuration for the delete command.
type DeleteConfig struct {
	Scheme   *scheme.Scheme
	Database database.Database
	FS       fs.FS
}

// NewDeleteCommand creates a new cobra.Command for the delete command.
func NewDeleteCommand(config DeleteConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete resources to runtime",
		RunE:  runDeleteCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), "", "Set the resource's namespace. If not set, use the default namespace")
	cmd.PersistentFlags().StringP(flagFile, toShorthand(flagFile), "", "Set the file path to be deleted")

	return cmd
}

func runDeleteCommand(config DeleteConfig) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()

		ns, err := cmd.Flags().GetString(flagNamespace)
		if err != nil {
			return err
		}
		file, err := cmd.Flags().GetString(flagFile)
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

		specs, err := scanner.New().
			Scheme(config.Scheme).
			Storage(st).
			Namespace(ns).
			FS(config.FS).
			Filename(file).
			Scan(ctx)
		if err != nil {
			return err
		}

		var filter *storage.Filter
		for _, spec := range specs {
			filter = filter.And(storage.Where[ulid.ULID](scheme.KeyID).EQ(spec.GetID()).And(storage.Where[string](scheme.KeyNamespace).EQ(spec.GetNamespace())))
		}

		if _, err := st.DeleteMany(ctx, filter); err != nil {
			return err
		}
		return nil
	}
}
