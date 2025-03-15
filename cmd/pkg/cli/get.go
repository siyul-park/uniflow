package cli

import (
	"github.com/siyul-park/uniflow/cmd/pkg/resource"
	resourcebase "github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/store"
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
			specs:  runGetCommand(config.SpecStore),
			values: runGetCommand(config.ValueStore),
		}),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), resourcebase.DefaultNamespace, "Set the resource's namespace. If not set, use all namespace")

	return cmd
}

func runGetCommand(store store.Store, alias ...func(map[string]string)) func(cmd *cobra.Command) error {
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

		writer := resource.NewWriter(cmd.OutOrStdout())

		cursor, err := store.Find(ctx, map[string]any{resourcebase.KeyNamespace: namespace})
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		var resources []*resourcebase.Unstructured
		for cursor.Next(ctx) {
			rsc := &resourcebase.Unstructured{}
			if err := cursor.Decode(rsc); err != nil {
				return err
			}
			resources = append(resources, rsc)
		}

		return writer.Write(resources)
	}
}
