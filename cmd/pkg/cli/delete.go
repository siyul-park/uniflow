package cli

import (
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/cmd/pkg/io"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// DeleteConfig represents the configuration for the delete command.
type DeleteConfig struct {
	SpecStore  store.Store
	ValueStore store.Store
	FS         afero.Fs
}

// NewDeleteCommand creates a new cobra.Command for the delete command.
func NewDeleteCommand(config DeleteConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "delete",
		Short:     "Delete resources from the specified namespace",
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{specs, values},
		RunE: runs(map[string]func(cmd *cobra.Command) error{
			specs:  runDeleteCommand[spec.Spec](config.SpecStore, config.FS),
			values: runDeleteCommand[*value.Value](config.ValueStore, config.FS),
		}),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), resource.DefaultNamespace, "Set the io's namespace. If not set, use the default namespace")
	cmd.PersistentFlags().StringP(flagFilename, toShorthand(flagFilename), "", "Set the file path to be deleted")

	return cmd
}

func runDeleteCommand[T resource.Resource](store store.Store, fs afero.Fs, alias ...func(map[string]string)) func(cmd *cobra.Command) error {
	flags := map[string]string{
		flagNamespace: flagNamespace,
		flagFilename:  flagFilename,
	}
	for _, init := range alias {
		init(flags)
	}

	return func(cmd *cobra.Command) error {
		ctx := cmd.Context()

		namespace, err := cmd.Flags().GetString(flags[flagNamespace])
		if err != nil {
			return err
		}
		filename, err := cmd.Flags().GetString(flags[flagFilename])
		if err != nil {
			return err
		}
		if filename == "" {
			return nil
		}

		file, err := fs.Open(filename)
		if err != nil {
			return err
		}
		defer file.Close()

		reader := io.NewReader(file)

		var resources []T
		if err := reader.Read(&resources); err != nil {
			return err
		}

		filters := make([]any, 0, len(resources))
		for _, rsc := range resources {
			filter := map[string]any{}
			if rsc.GetID() != uuid.Nil {
				filter[resource.KeyID] = rsc.GetID()
			}
			if rsc.GetName() != "" {
				filter[resource.KeyName] = rsc.GetName()
			}
			filters = append(filters, filter)
		}

		_, err = store.Delete(ctx, map[string]any{
			"$and": []any{
				map[string]any{resource.KeyNamespace: namespace},
				map[string]any{"$or": filters},
			},
		})
		return err
	}
}
