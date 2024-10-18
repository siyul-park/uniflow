package cli

import (
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/siyul-park/uniflow/pkg/agent"
	"github.com/siyul-park/uniflow/pkg/chart"
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
	ChartStore  chart.Store
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
	cmd.PersistentFlags().String(flagFromSpecs, "", "Specify the file path containing specs")
	cmd.PersistentFlags().String(flagFromSecrets, "", "Specify the file path containing secrets")
	cmd.PersistentFlags().String(flagFromCharts, "", "Specify the file path containing charts")
	cmd.PersistentFlags().Bool(flagDebug, false, "Enable debug mode")

	return cmd
}

// runStartCommand runs the start command with the given configuration.
func runStartCommand(config StartConfig) func(cmd *cobra.Command, args []string) error {
	applySpecs := runApplyCommand(config.SpecStore, config.FS, alias(flagFilename, flagFromSpecs))
	applySecrets := runApplyCommand(config.SecretStore, config.FS, alias(flagFilename, flagFromSecrets))
	applyCharts := runApplyCommand(config.ChartStore, config.FS, alias(flagFilename, flagFromCharts))

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

		if err := applySpecs(cmd); err != nil {
			return err
		}
		if err := applySecrets(cmd); err != nil {
			return err
		}
		if err := applyCharts(cmd); err != nil {
			return err
		}

		h := config.Hook
		if h == nil {
			h = hook.New()
		}

		r := runtime.New(runtime.Config{
			Namespace:   namespace,
			Scheme:      config.Scheme,
			Hook:        h,
			SpecStore:   config.SpecStore,
			SecretStore: config.SecretStore,
			ChartStore:  config.ChartStore,
		})

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		if enableDebug {
			a := agent.New()

			h.AddLoadHook(a)
			h.AddUnloadHook(a)
			h.AddLinkHook(a)
			h.AddUnlinkHook(a)

			d := NewDebugger(
				a,
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
