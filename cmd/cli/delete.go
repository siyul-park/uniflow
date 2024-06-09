package cli

import (
	"io/fs"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/cmd/scanner"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/spf13/cobra"
)

// DeleteConfig represents the configuration for the delete command.
type DeleteConfig struct {
	Scheme   *spec.Scheme
	Database database.Database
	FS       fs.FS
}

// NewDeleteCommand creates a new cobra.Command for the delete command.
func NewDeleteCommand(config DeleteConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete resources in namespace",
		RunE:  runDeleteCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), "", "Set the resource's namespace. If not set, use the default namespace")
	cmd.PersistentFlags().StringP(flagFilename, toShorthand(flagFilename), "", "Set the file path to be deleted")

	return cmd
}

func runDeleteCommand(config DeleteConfig) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()

		namespace, err := cmd.Flags().GetString(flagNamespace)
		if err != nil {
			return err
		}
		filename, err := cmd.Flags().GetString(flagFilename)
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

		specs, err := scanner.New().
			Scheme(config.Scheme).
			Storage(st).
			Namespace(namespace).
			FS(config.FS).
			Filename(filename).
			Scan(ctx)
		if err != nil {
			return err
		}

		var filter *spec.Filter
		for _, v := range specs {
			filter = filter.And(spec.Where[uuid.UUID](spec.KeyID).EQ(v.GetID()).
				And(spec.Where[string](spec.KeyNamespace).EQ(v.GetNamespace())))
		}

		_, err = st.DeleteMany(ctx, filter)
		return err
	}
}
