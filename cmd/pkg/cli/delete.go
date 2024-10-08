package cli

import (
	"github.com/siyul-park/uniflow/cmd/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/chart"
	resourcebase "github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// DeleteConfig represents the configuration for the delete command.
type DeleteConfig struct {
	ChartStore  chart.Store
	SpecStore   spec.Store
	SecretStore secret.Store
	FS          afero.Fs
}

// NewDeleteCommand creates a new cobra.Command for the delete command.
func NewDeleteCommand(config DeleteConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "delete",
		Short:     "Delete resources from the specified namespace",
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{argCharts, argNodes, argSecrets},
		RunE:      runDeleteCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), resourcebase.DefaultNamespace, "Set the resource's namespace. If not set, use the default namespace")
	cmd.PersistentFlags().StringP(flagFilename, toShorthand(flagFilename), "", "Set the file path to be deleted")

	return cmd
}

func runDeleteCommand(config DeleteConfig) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		namespace, err := cmd.Flags().GetString(flagNamespace)
		if err != nil {
			return err
		}
		filename, err := cmd.Flags().GetString(flagFilename)
		if err != nil {
			return err
		}

		file, err := config.FS.Open(filename)
		if err != nil {
			return err
		}
		defer file.Close()

		reader := resource.NewReader(file)

		switch args[0] {
		case argCharts:
			var charts []*chart.Chart
			if err := reader.Read(&charts); err != nil {
				return err
			}

			for _, chrt := range charts {
				if chrt.GetNamespace() == "" {
					chrt.SetNamespace(namespace)
				}
			}

			_, err := config.ChartStore.Delete(ctx, charts...)
			return err
		case argNodes:
			var specs []spec.Spec
			if err := reader.Read(&specs); err != nil {
				return err
			}

			for _, sp := range specs {
				if sp.GetNamespace() == "" {
					sp.SetNamespace(namespace)
				}
			}

			_, err := config.SpecStore.Delete(ctx, specs...)
			return err
		case argSecrets:
			var secrets []*secret.Secret
			if err := reader.Read(&secrets); err != nil {
				return err
			}

			for _, scrt := range secrets {
				if scrt.GetNamespace() == "" {
					scrt.SetNamespace(namespace)
				}
			}

			_, err := config.SecretStore.Delete(ctx, secrets...)
			return err
		}
		return nil
	}
}
