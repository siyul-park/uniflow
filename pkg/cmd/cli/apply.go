package cli

import (
	"context"
	"io/fs"

	"github.com/oklog/ulid/v2"
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
		Short: "Apply a configuration to a resource",
		RunE:  runApplyCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), "", "Set the resource's namespace. If not set, use the default namespace")
	cmd.PersistentFlags().StringP(flagFile, toShorthand(flagFile), "", "Set the file path to be applied")

	return cmd
}

func runApplyCommand(config ApplyConfig) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()

		ns, err := cmd.Flags().GetString(flagNamespace)
		if err != nil {
			return err
		}
		fl, err := cmd.Flags().GetString(flagFile)
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
			Namespace(ns).
			FS(config.FS).
			Filename(fl).
			Scan()
		if err != nil {
			return err
		}

		if err := updateSpecIDs(ctx, st, specs); err != nil {
			return err
		}

		if err := applySpecs(ctx, st, specs); err != nil {
			return err
		}

		if err := printer.PrintTable(cmd.OutOrStdout(), specs, printer.SpecTableColumnDefinitions); err != nil {
			return err
		}

		return nil
	}
}

func updateSpecIDs(ctx context.Context, st *storage.Storage, specs []scheme.Spec) error {
	for _, spec := range specs {
		if spec.GetID() == (ulid.ULID{}) {
			if spec.GetName() != "" {
				filter := storage.Where[string](scheme.KeyName).EQ(spec.GetName()).And(storage.Where[string](scheme.KeyNamespace).EQ(spec.GetNamespace()))
				if exist, err := st.FindOne(ctx, filter); err != nil {
					return err
				} else if exist != nil {
					spec.SetID(exist.GetID())
				}
			}
		}

		if spec.GetID() == (ulid.ULID{}) {
			spec.SetID(ulid.Make())
		}
	}
	return nil
}

func applySpecs(ctx context.Context, st *storage.Storage, specs []scheme.Spec) error {
	var ids []ulid.ULID
	for _, spec := range specs {
		ids = append(ids, spec.GetID())
	}

	exists, err := st.FindMany(ctx, storage.Where[ulid.ULID](scheme.KeyID).IN(ids...), &database.FindOptions{
		Limit: lo.ToPtr[int](len(ids)),
	})
	if err != nil {
		return err
	}
	existsIds := make(map[ulid.ULID]struct{}, len(exists))
	for _, spec := range exists {
		existsIds[spec.GetID()] = struct{}{}
	}

	var inserted []scheme.Spec
	var updated []scheme.Spec
	for _, spec := range specs {
		if _, ok := existsIds[spec.GetID()]; ok {
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
	return nil
}
