package cli

import (
	"github.com/siyul-park/uniflow/cmd/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/chart"
	resourcebase "github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/spf13/cobra"
)

// GetConfig represents the configuration for the get command.
type GetConfig struct {
	ChartStore chart.Store
	SpecStore  spec.Store
	ValueStore value.Store
}

// NewGetCommand creates a new cobra.Command for the get command.
func NewGetCommand(config GetConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "get",
		Short:     "Load resources from the specified namespace",
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{specs, values, charts},
		RunE: runs(map[string]func(cmd *cobra.Command) error{
			specs:  runGetCommand(config.SpecStore, spec.New),
			values: runGetCommand(config.ValueStore, value.New),
			charts: runGetCommand(config.ChartStore, chart.New),
		}),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), resourcebase.DefaultNamespace, "Set the resource's namespace. If not set, use all namespace")

	return cmd
}

func runGetCommand[T resourcebase.Resource](store resourcebase.Store[T], zero func() T, alias ...func(map[string]string)) func(cmd *cobra.Command) error {
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

		rsc := zero()
		rsc.SetNamespace(namespace)

		resources, err := store.Load(ctx, rsc)
		if err != nil {
			return err
		}

		return writer.Write(resources)
	}
}
