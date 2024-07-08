package cli

import (
	"io/fs"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/cmd/pkg/printer"
	"github.com/siyul-park/uniflow/cmd/pkg/scanner"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/spf13/cobra"
)

// ApplyConfig represents the configuration for the apply command.
type ApplyConfig struct {
	Scheme *scheme.Scheme
	Store  *store.Store
	FS     fs.FS
}

// NewApplyCommand creates a new cobra.Command for the apply command.
func NewApplyCommand(config ApplyConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply node specifications to the specified namespace",
		RunE:  runApplyCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), spec.DefaultNamespace, "Set the resource's namespace. If not set, use the default namespace")
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

		var ids []uuid.UUID
		for _, spec := range specs {
			ids = append(ids, spec.GetID())
		}

		origins, err := config.Store.FindMany(ctx, store.Where[uuid.UUID](spec.KeyID).IN(ids...), &database.FindOptions{
			Limit: lo.ToPtr[int](len(ids)),
		})
		if err != nil {
			return err
		}

		exists := make(map[uuid.UUID]struct{}, len(origins))
		for _, spec := range origins {
			exists[spec.GetID()] = struct{}{}
		}

		var inserted []spec.Spec
		var updated []spec.Spec
		for _, spec := range specs {
			if _, ok := exists[spec.GetID()]; ok {
				updated = append(updated, spec)
			} else {
				inserted = append(inserted, spec)
			}
		}

		if _, err := config.Store.InsertMany(ctx, inserted); err != nil {
			return err
		}
		if _, err := config.Store.UpdateMany(ctx, updated); err != nil {
			return err
		}

		return printer.PrintTable(cmd.OutOrStdout(), specs, printer.SpecTableColumnDefinitions)
	}
}
