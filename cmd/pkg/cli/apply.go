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

// ApplyConfig represents the configuration for the apply command.
type ApplyConfig struct {
	SpecStore   spec.Store
	SecretStore secret.Store
	ChartStore  chart.Store
	FS          afero.Fs
}

// NewApplyCommand creates a new cobra.Command for the apply command.
func NewApplyCommand(config ApplyConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "apply",
		Short:     "Apply resources to the specified namespace",
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{specs, secrets, charts},
		RunE: runs(map[string]func(cmd *cobra.Command) error{
			specs:   runApplyCommand(config.SpecStore, config.FS),
			secrets: runApplyCommand(config.SecretStore, config.FS),
			charts:  runApplyCommand(config.ChartStore, config.FS),
		}),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), resourcebase.DefaultNamespace, "Set the resource's namespace. If not set, use the default namespace")
	cmd.PersistentFlags().StringP(flagFilename, toShorthand(flagFilename), "", "Set the file path to be applied")

	return cmd
}

func runApplyCommand[T resourcebase.Resource](store resourcebase.Store[T], fs afero.Fs, alias ...func(map[string]string)) func(cmd *cobra.Command) error {
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
		writer := resource.NewWriter(cmd.OutOrStdout())

		var resources []T
		if err := reader.Read(&resources); err != nil {
			return err
		}
		if len(resources) == 0 {
			return nil
		}

		for _, rsc := range resources {
			if rsc.GetNamespace() == "" {
				rsc.SetNamespace(namespace)
			}
		}

		loads, err := store.Load(ctx, resources...)
		if err != nil {
			return err
		}

		var inserts []T
		var updates []T
		for _, rsc := range resources {
			exists := false
			for _, r := range loads {
				if len(resourcebase.Match(r, rsc)) > 0 {
					rsc.SetID(r.GetID())
					updates = append(updates, rsc)
					exists = true
					break
				}
			}
			if !exists {
				inserts = append(inserts, rsc)
			}
		}

		if _, err := store.Store(ctx, inserts...); err != nil {
			return err
		}
		if _, err := store.Swap(ctx, updates...); err != nil {
			return err
		}

		return writer.Write(resources)
	}
}
