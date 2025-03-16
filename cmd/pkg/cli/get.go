package cli

import (
	"github.com/siyul-park/uniflow/cmd/pkg/io"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/spf13/cobra"
)

// GetConfig represents the configuration for the get command.
type GetConfig struct {
	SpecStore  store.Store
	ValueStore store.Store
}

// NewGetCommand creates a new cobra.Command for the get command.
func NewGetCommand(config GetConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "get",
		Short:     "Load resources from the specified namespace",
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{specs, values},
		RunE: runs(map[string]func(cmd *cobra.Command) error{
			specs:  runGetCommand[spec.Spec](config.SpecStore),
			values: runGetCommand[*value.Value](config.ValueStore),
		}),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), resource.DefaultNamespace, "Set the io's namespace. If not set, use all namespace")

	return cmd
}

func runGetCommand[T resource.Resource](store store.Store, alias ...func(map[string]string)) func(cmd *cobra.Command) error {
	flags := map[string]string{
		flagNamespace: flagNamespace,
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

		writer := io.NewWriter(cmd.OutOrStdout())

		cursor, err := store.Find(ctx, map[string]any{resource.KeyNamespace: namespace})
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		var resources []T
		for cursor.Next(ctx) {
			var rsc T
			if err := cursor.Decode(&rsc); err != nil {
				return err
			}
			resources = append(resources, rsc)
		}

		return writer.Write(resources)
	}
}
