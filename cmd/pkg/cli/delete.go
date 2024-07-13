package cli

import (
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/cmd/pkg/scanner"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// DeleteConfig represents the configuration for the delete command.
type DeleteConfig struct {
	Scheme *scheme.Scheme
	Store  *store.Store
	FS     afero.Fs
}

// NewDeleteCommand creates a new cobra.Command for the delete command.
func NewDeleteCommand(config DeleteConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete node specifications from the specified namespace",
		RunE:  runDeleteCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), spec.DefaultNamespace, "Set the resource's namespace. If not set, use the default namespace")
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

		specs, err := scanner.New().
			Scheme(config.Scheme).
			Store(config.Store).
			Namespace(namespace).
			FS(config.FS).
			Filename(filename).
			Scan(ctx)
		if err != nil {
			return err
		}

		var filter *store.Filter
		for _, v := range specs {
			filter = filter.And(store.Where[uuid.UUID](spec.KeyID).Equal(v.GetID()).
				And(store.Where[string](spec.KeyNamespace).Equal(v.GetNamespace())))
		}

		_, err = config.Store.DeleteMany(ctx, filter)
		return err
	}
}
