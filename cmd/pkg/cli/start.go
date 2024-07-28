package cli

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/siyul-park/uniflow/cmd/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// StartConfig holds the configuration for the start command.
type StartConfig struct {
	Scheme      *scheme.Scheme
	Hook        *hook.Hook
	SpecStore   spec.Store
	SecretStore secret.Store
	FS          afero.Fs
}

// NewStartCommand creates a new cobra.Command for the start command.
func NewStartCommand(config StartConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Starts the workflow engine within the specified namespace",
		RunE:  runStartCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), spec.DefaultNamespace, "Set the namespace for running")
	cmd.PersistentFlags().StringP(flagFilename, toShorthand(flagFilename), "", "Set the file path to be applied")

	return cmd
}

// runStartCommand runs the start command with the given configuration.
func runStartCommand(config StartConfig) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()

		namespace, err := cmd.Flags().GetString(flagNamespace)
		if err != nil {
			return err
		}
		filename, err := cmd.Flags().GetString(flagFilename)
		if err != nil {
			return err
		}

		if filename != "" {
			specs, err := config.SpecStore.Load(ctx, &spec.Meta{Namespace: namespace})
			if err != nil {
				return err
			}
			secrets, err := config.SecretStore.Load(ctx, &secret.Secret{Namespace: namespace})
			if err != nil {
				return err
			}

			if len(specs) == 0 && len(secrets) == 0 {
				file, err := config.FS.Open(filename)
				if err != nil {
					return err
				}
				defer file.Close()

				reader := resource.NewReader(file)

				var raws []types.Value
				if err := reader.Read(&raws); err != nil {
					return err
				}

				for _, raw := range raws {
					var spec spec.Spec
					if err := types.Decoder.Decode(raw, &spec); err == nil {
						specs = append(specs, spec)
						continue
					}

					var secret *secret.Secret
					if err := types.Decoder.Decode(raw, &secret); err == nil {
						secrets = append(secrets, secret)
						continue
					}
				}

				for _, spec := range specs {
					if spec.GetNamespace() == "" {
						spec.SetNamespace(namespace)
					}
				}

				for _, sec := range secrets {
					if sec.GetNamespace() == "" {
						sec.SetNamespace(namespace)
					}
				}

				if _, err = config.SpecStore.Store(ctx, specs...); err != nil {
					return err
				}

				if _, err = config.SecretStore.Store(ctx, secrets...); err != nil {
					return err
				}
			}
		}

		r := runtime.New(runtime.Config{
			Namespace:   namespace,
			Scheme:      config.Scheme,
			Hook:        config.Hook,
			SpecStore:   config.SpecStore,
			SecretStore: config.SecretStore,
		})

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigs
			_ = r.Close()
		}()

		return r.Listen(ctx)
	}
}
