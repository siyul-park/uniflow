package cli

import (
	"io"
	"os"
	"regexp"

	"github.com/siyul-park/uniflow/hook"
	"github.com/siyul-park/uniflow/meta"
	"github.com/siyul-park/uniflow/runtime"
	"github.com/siyul-park/uniflow/scheme"
	"github.com/siyul-park/uniflow/spec"
	"github.com/siyul-park/uniflow/store"
	"github.com/siyul-park/uniflow/testing"
	"github.com/siyul-park/uniflow/value"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// TestConfig holds the configuration for the start command.
type TestConfig struct {
	Runner     *testing.Runner
	Scheme     *scheme.Scheme
	Hook       *hook.Hook
	SpecStore  store.Store
	ValueStore store.Store
	FS         afero.Fs
}

// NewTestCommand creates a new cobra.Command for the start command.
func NewTestCommand(config TestConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run tests for the workflow engine within the specified namespace",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runTestCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), meta.DefaultNamespace, "Set the namespace for running the workflow")
	cmd.PersistentFlags().String(flagFromSpecs, "", "Specify the file path containing workflow specifications")
	cmd.PersistentFlags().String(flagFromValues, "", "Specify the file path containing values for the workflow")
	cmd.PersistentFlags().StringToString(flagEnvironment, nil, "Set environment variables for the workflow execution")

	return cmd
}

// runTestCommand runs the start command with the given configuration.
func runTestCommand(config TestConfig) func(cmd *cobra.Command, args []string) error {
	applySpecs := runApplyCommand[spec.Spec](config.SpecStore, config.FS, alias(flagFilename, flagFromSpecs))
	applyValues := runApplyCommand[*value.Value](config.ValueStore, config.FS, alias(flagFilename, flagFromValues))

	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		namespace, err := cmd.Flags().GetString(flagNamespace)
		if err != nil {
			return err
		}
		environment, err := cmd.Flags().GetStringToString(flagEnvironment)
		if err != nil {
			return err
		}

		match := func(string) bool { return true }
		if len(args) > 0 {
			exp, err := regexp.Compile(args[0])
			if err != nil {
				return err
			}
			match = func(name string) bool {
				return exp.Match([]byte(name))
			}
		}

		reporter := testing.NewTextReporter(cmd.OutOrStdout())

		config.Runner.AddReporter(reporter)
		defer config.Runner.RemoveReporter(reporter)

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

		if err := r.Load(ctx, nil); err != nil {
			return err
		}

		return config.Runner.Run(ctx, match)
	}
}
