package cli

import (
	"io"
	"os"
	"regexp"

	"github.com/siyul-park/uniflow/pkg/hook"
	resourcebase "github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/testing"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// TestConfig holds the configuration for the start command.
type TestConfig struct {
	Runner     *testing.Runner
	Scheme     *scheme.Scheme
	Hook       *hook.Hook
	SpecStore  spec.Store
	ValueStore value.Store
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

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), resourcebase.DefaultNamespace, "Set the namespace for running the workflow")
	cmd.PersistentFlags().String(flagFromSpecs, "", "Specify the file path containing workflow specifications")
	cmd.PersistentFlags().String(flagFromValues, "", "Specify the file path containing values for the workflow")
	cmd.PersistentFlags().StringToString(flagEnvironment, nil, "Set environment variables for the workflow execution")

	return cmd
}

// runTestCommand runs the start command with the given configuration.
func runTestCommand(config TestConfig) func(cmd *cobra.Command, args []string) error {
	applySpecs := runApplyCommand(config.SpecStore, config.FS, alias(flagFilename, flagFromSpecs))
	applyValues := runApplyCommand(config.ValueStore, config.FS, alias(flagFilename, flagFromValues))

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

		textReporter := testing.NewTextReporter(cmd.OutOrStdout())
		errorReporter := testing.NewErrorReporter()

		config.Runner.AddReporter(textReporter)
		defer config.Runner.RemoveReporter(textReporter)

		config.Runner.AddReporter(errorReporter)
		defer config.Runner.RemoveReporter(errorReporter)

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
		defer r.Close()

		if err := r.Load(ctx); err != nil {
			return err
		}

		if err := config.Runner.Run(ctx, match); err != nil {
			return err
		}
		return errorReporter.Error()
	}
}
