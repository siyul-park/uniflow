package cli

import (
	"github.com/siyul-park/uniflow/cmd/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/spf13/cobra"
)

// GetConfig represents the configuration for the get command.
type GetConfig struct {
	SpecStore   spec.Store
	SecretStore secret.Store
}

// NewGetCommand creates a new cobra.Command for the get command.
func NewGetCommand(config GetConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "get",
		Short:     "Get resources from the specified namespace",
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{argNodes, argSecrets},
		RunE:      runGetCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), spec.DefaultNamespace, "Set the resource's namespace. If not set, use all namespace")

	return cmd
}

func runGetCommand(config GetConfig) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		namespace, err := cmd.Flags().GetString(flagNamespace)
		if err != nil {
			return err
		}

		writer := resource.NewWriter(cmd.OutOrStdout())

		switch args[0] {
		case argNodes:
			specs, err := config.SpecStore.Load(ctx, &spec.Meta{
				Namespace: namespace,
			})
			if err != nil {
				return err
			}

			return writer.Write(specs)
		case argSecrets:
			secrets, err := config.SecretStore.Load(ctx, &secret.Secret{
				Namespace: namespace,
			})
			if err != nil {
				return err
			}

			return writer.Write(secrets)
		}

		return nil
	}
}
