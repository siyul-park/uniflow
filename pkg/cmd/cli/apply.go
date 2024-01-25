package cli

import (
	"io/fs"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/cmd/printer"
	"github.com/siyul-park/uniflow/pkg/cmd/scanner"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/spf13/cobra"
)

// ApplyConfig represents the configuration for the apply command.
type ApplyConfig struct {
	Scheme   *scheme.Scheme
	Database database.Database
	FS       fs.FS
}

// NewApplyCommand creates a new cobra.Command for the apply command.
func NewApplyCommand(config ApplyConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply resources in namespace",
		RunE:  runApplyCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), "", "Set the resource's namespace. If not set, use the default namespace")
	cmd.PersistentFlags().StringP(flagFilename, toShorthand(flagFilename), "", "Set the file path to be applied")

	return cmd
}

func runApplyCommand(config ApplyConfig) func(cmd *cobra.Command, args []string) error {
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
			Namespace(namespace).
			FS(config.FS).
			Filename(filename).
			Scan(ctx)
		if err != nil {
			return err
		}

		var ids []uuid.UUID
		for _, spec := range specs {
			ids = append(ids, spec.GetID())
		}

		origins, err := st.FindMany(ctx, storage.Where[uuid.UUID](scheme.KeyID).IN(ids...), &database.FindOptions{
			Limit: lo.ToPtr[int](len(ids)),
		})
		if err != nil {
			return err
		}

		exists := make(map[uuid.UUID]struct{}, len(origins))
		for _, spec := range origins {
			exists[spec.GetID()] = struct{}{}
		}

		var inserted []scheme.Spec
		var updated []scheme.Spec
		for _, spec := range specs {
			if _, ok := exists[spec.GetID()]; ok {
				updated = append(updated, spec)
			} else {
				inserted = append(inserted, spec)
			}
		}

		if _, err := st.InsertMany(ctx, inserted); err != nil {
			return err
		}
		if _, err := st.UpdateMany(ctx, updated); err != nil {
			return err
		}

		return printer.PrintTable(cmd.OutOrStdout(), specs, printer.SpecTableColumnDefinitions)
	}
}
