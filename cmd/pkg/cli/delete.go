package cli

import (
	"github.com/siyul-park/uniflow/cmd/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/chart"
	resourcebase "github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// DeleteConfig represents the configuration for the delete command.
type DeleteConfig struct {
	SpecStore  spec.Store
	ValueStore value.Store
	ChartStore chart.Store
	FS         afero.Fs
}

// NewDeleteCommand creates a new cobra.Command for the delete command.
func NewDeleteCommand(config DeleteConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "delete",
		Short:     "Delete resources from the specified namespace",
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{specs, values, charts},
		RunE: runs(map[string]func(cmd *cobra.Command) error{
			specs:  runDeleteCommand(config.SpecStore, config.FS, spec.New),
			values: runDeleteCommand(config.ValueStore, config.FS, value.New),
			charts: runDeleteCommand(config.ChartStore, config.FS, chart.New),
		}),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), resourcebase.DefaultNamespace, "Set the resource's namespace. If not set, use the default namespace")
	cmd.PersistentFlags().StringP(flagFilename, toShorthand(flagFilename), "", "Set the file path to be deleted")

	return cmd
}

func runDeleteCommand[T resourcebase.Resource](store resourcebase.Store[T], fs afero.Fs, zero func() T, alias ...func(map[string]string)) func(cmd *cobra.Command) error {
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

		reader := resource.NewReader(file)

		var resources []T
		if err := reader.Read(&resources); err != nil {
			return err
		}
		if len(resources) == 0 {
			resources = append(resources, zero())
		}

		for _, rsc := range resources {
			if rsc.GetNamespace() == "" {
				rsc.SetNamespace(namespace)
			}
		}

		_, err = store.Delete(ctx, resources...)
		return err
	}
}
