package cli

import (
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/siyul-park/uniflow/cmd/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/agent"
	"github.com/siyul-park/uniflow/pkg/hook"
	resourcebase "github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
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

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), resourcebase.DefaultNamespace, "Set the namespace for running")
	cmd.PersistentFlags().String(flagFromNodes, "", "Specify the file path containing node specs")
	cmd.PersistentFlags().String(flagFromSecrets, "", "Specify the file path containing secrets")
	cmd.PersistentFlags().Bool(flagDebug, false, "Enable debug mode")

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

		fromNodes, err := cmd.Flags().GetString(flagFromNodes)
		if err != nil {
			return err
		}

		fromSecrets, err := cmd.Flags().GetString(flagFromSecrets)
		if err != nil {
			return err
		}

		enableDebug, err := cmd.Flags().GetBool(flagDebug)
		if err != nil {
			return err
		}

		if fromNodes != "" {
			specs, err := config.SpecStore.Load(ctx, &spec.Meta{Namespace: namespace})
			if err != nil {
				return err
			}

			if len(specs) == 0 {
				file, err := config.FS.Open(fromNodes)
				if err != nil {
					return err
				}
				defer file.Close()

				reader := resource.NewReader(file)
				if err := reader.Read(&specs); err != nil {
					return err
				}

				for _, spec := range specs {
					if spec.GetNamespace() == "" {
						spec.SetNamespace(namespace)
					}
				}

				if _, err = config.SpecStore.Store(ctx, specs...); err != nil {
					return err
				}
			}
		}

		if fromSecrets != "" {
			secrets, err := config.SecretStore.Load(ctx, &secret.Secret{Namespace: namespace})
			if err != nil {
				return err
			}

			if len(secrets) == 0 {
				file, err := config.FS.Open(fromSecrets)
				if err != nil {
					return err
				}
				defer file.Close()

				reader := resource.NewReader(file)
				if err := reader.Read(&secrets); err != nil {
					return err
				}

				for _, sec := range secrets {
					if sec.GetNamespace() == "" {
						sec.SetNamespace(namespace)
					}
				}

				if _, err := config.SecretStore.Store(ctx, secrets...); err != nil {
					return err
				}
			}
		}

		h := config.Hook
		if h == nil {
			h = hook.New()
		}

		var agt *agent.Agent
		if enableDebug {
			agt = agent.New()
			h.AddLoadHook(agt)
			h.AddUnloadHook(agt)
		}

		r := runtime.New(runtime.Config{
			Namespace:   namespace,
			Scheme:      config.Scheme,
			Hook:        h,
			SpecStore:   config.SpecStore,
			SecretStore: config.SecretStore,
		})

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		if enableDebug {
			d := NewDebugger(
				agt,
				tea.WithContext(ctx),
				tea.WithInput(cmd.InOrStdin()),
				tea.WithOutput(cmd.OutOrStdout()),
			)

			go func() {
				d.Wait()
				r.Close()
			}()

			go func() {
				<-sigs
				d.Kill()
			}()

			if err := r.Watch(ctx); err != nil {
				return err
			}
			if err := r.Load(ctx); err != nil {
				return err
			}
			go r.Reconcile(ctx)
			return d.Run()
		}

		go func() {
			<-sigs
			r.Close()
		}()

		if err := r.Watch(ctx); err != nil {
			return err
		}
		if err := r.Load(ctx); err != nil {
			return err
		}
		return r.Reconcile(ctx)
	}
}
