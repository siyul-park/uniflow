package cmd

import (
	"io"
	"os"
	"regexp"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/siyul-park/uniflow/pkg/driver"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/testing"
	"github.com/siyul-park/uniflow/pkg/value"
)

// TestConfig holds the configuration for the start command.
type TestConfig struct {
	Namespace   string
	Environment map[string]string
	Runner      *testing.Runner
	Scheme      *scheme.Scheme
	Hook        *hook.Hook
	SpecStore   driver.Store
	ValueStore  driver.Store
	FS          afero.Fs
}

// NewTestCommand creates a new cobra.Command for the start command.
func NewTestCommand(config TestConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run tests for the workflow engine within the specified namespace",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runTestCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), config.Namespace, "Inject the namespace for running the workflow")
	cmd.PersistentFlags().String(flagFromSpecs, "", "Specify the file path containing workflow specifications")
	cmd.PersistentFlags().String(flagFromValues, "", "Specify the file path containing values for the workflow")
	cmd.PersistentFlags().StringToStringP(flagEnvironment, toShorthand(flagEnvironment), config.Environment, "Inject environment variables for the workflow execution")

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
