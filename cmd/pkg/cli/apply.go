package cli

import (
	"github.com/siyul-park/uniflow/cmd/pkg/resource"
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
	FS          afero.Fs
}

// NewApplyCommand creates a new cobra.Command for the apply command.
func NewApplyCommand(config ApplyConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "apply",
		Short:     "Apply resources to the specified namespace",
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{argNodes, argSecrets},
		RunE:      runApplyCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), resourcebase.DefaultNamespace, "Set the resource's namespace. If not set, use the default namespace")
	cmd.PersistentFlags().StringP(flagFilename, toShorthand(flagFilename), "", "Set the file path to be applied")

	return cmd
}

func runApplyCommand(config ApplyConfig) func(cmd *cobra.Command, args []string) error {
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
		writer := resource.NewWriter(cmd.OutOrStdout())

		switch args[0] {
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

			ok, err := config.SpecStore.Load(ctx, specs...)
			if err != nil {
				return err
			}

			var inserts []spec.Spec
			var updates []spec.Spec
			for _, sp := range specs {
				if match := resourcebase.Match(sp, ok...); len(match) > 0 {
					sp.SetID(match[0].GetID())
					updates = append(updates, sp)
				} else {
					inserts = append(inserts, sp)
				}
			}

			if _, err := config.SpecStore.Store(ctx, inserts...); err != nil {
				return err
			}
			if _, err := config.SpecStore.Swap(ctx, updates...); err != nil {
				return err
			}

			return writer.Write(specs)
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

			ok, err := config.SecretStore.Load(ctx, secrets...)
			if err != nil {
				return err
			}

			var inserts []*secret.Secret
			var updates []*secret.Secret
			for _, scrt := range secrets {
				if match := resourcebase.Match(scrt, ok...); len(match) > 0 {
					scrt.SetID(match[0].GetID())
					updates = append(updates, scrt)
				} else {
					inserts = append(inserts, scrt)
				}
			}

			if _, err := config.SecretStore.Store(ctx, inserts...); err != nil {
				return err
			}
			if _, err := config.SecretStore.Swap(ctx, updates...); err != nil {
				return err
			}

			return writer.Write(secrets)
		}

		return nil
	}
}
