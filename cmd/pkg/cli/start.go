package cli

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/siyul-park/uniflow/cmd/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/debug"
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

		var debugger *debug.Debugger
		if enableDebug {
			debugger = debug.NewDebugger()
		}

		r := runtime.New(runtime.Config{
			Namespace:   namespace,
			Scheme:      config.Scheme,
			Hook:        config.Hook,
			SpecStore:   config.SpecStore,
			SecretStore: config.SecretStore,
			Debugger:    debugger,
		})

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigs
			r.Close()
		}()

		if enableDebug {
			d := NewDebugger(debugger)

			go func() {
				d.Wait()
				r.Close()
			}()

			go func() {
				<-sigs
				d.Kill()
			}()

			go r.Listen(ctx)
			return d.Run()
		}

		return r.Listen(ctx)
	}
}
