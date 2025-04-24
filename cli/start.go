package cli

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/siyul-park/uniflow/hook"
	"github.com/siyul-park/uniflow/meta"
	"github.com/siyul-park/uniflow/runtime"
	"github.com/siyul-park/uniflow/scheme"
	"github.com/siyul-park/uniflow/spec"
	"github.com/siyul-park/uniflow/store"
	"github.com/siyul-park/uniflow/value"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// StartConfig holds the configuration for the start command.
type StartConfig struct {
	Scheme     *scheme.Scheme
	Hook       *hook.Hook
	SpecStore  store.Store
	ValueStore store.Store
	FS         afero.Fs
}

// NewStartCommand creates a new cobra.Command for the start command.
func NewStartCommand(config StartConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Starts the workflow engine within the specified namespace",
		RunE:  runStartCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), meta.DefaultNamespace, "Set the namespace for running the workflow")
	cmd.PersistentFlags().String(flagFromSpecs, "", "Specify the file path containing workflow specifications")
	cmd.PersistentFlags().String(flagFromValues, "", "Specify the file path containing values for the workflow")
	cmd.PersistentFlags().Bool(flagDebug, false, "Enable debug mode for detailed output during execution")
	cmd.PersistentFlags().StringToString(flagEnvironment, nil, "Set environment variables for the workflow execution")

	return cmd
}

// runStartCommand runs the start command with the given configuration.
func runStartCommand(config StartConfig) func(cmd *cobra.Command, args []string) error {
	applySpecs := runApplyCommand[spec.Spec](config.SpecStore, config.FS, alias(flagFilename, flagFromSpecs))
	applyValues := runApplyCommand[*value.Value](config.ValueStore, config.FS, alias(flagFilename, flagFromValues))

	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()

		namespace, err := cmd.Flags().GetString(flagNamespace)
		if err != nil {
			return err
		}
		enableDebug, err := cmd.Flags().GetBool(flagDebug)
		if err != nil {
			return err
		}
		environment, err := cmd.Flags().GetStringToString(flagEnvironment)
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		if out == os.Stdout {
			out = nil
		}

		cmd.SetOut(io.Discard)

		if err := applySpecs(cmd); err != nil {
			return err
		}
		if err := applyValues(cmd); err != nil {
			return err
		}

		cmd.SetOut(out)

		h := config.Hook
		if h == nil {
			h = hook.New()
		}

		r := runtime.New(runtime.Config{
			Namespace:   namespace,
			Environment: environment,
			Scheme:      config.Scheme,
			Hook:        h,
			SpecStore:   config.SpecStore,
			ValueStore:  config.ValueStore,
		})
		defer r.Close(ctx)

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		if enableDebug {
			a := runtime.NewAgent()

			h.AddLoadHook(a)
			h.AddUnloadHook(a)

			d := NewDebugger(
				a,
				tea.WithContext(ctx),
				tea.WithInput(cmd.InOrStdin()),
				tea.WithOutput(cmd.OutOrStdout()),
			)

			go func() {
				d.Wait()
				r.Close(ctx)
			}()

			go func() {
				<-sigs
				d.Kill()
			}()

			if err := r.Watch(ctx); err != nil {
				return err
			}
			_ = r.Load(ctx, nil)
			go r.Reconcile(ctx)
			return d.Run()
		}

		go func() {
			<-sigs
			r.Close(ctx)
		}()

		if err := r.Watch(ctx); err != nil {
			return err
		}
		_ = r.Load(ctx, nil)
		return r.Reconcile(ctx)
	}
}
