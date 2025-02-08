package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/siyul-park/uniflow/pkg/hook"
	resourcebase "github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// StartConfig holds the configuration for the start command.
type StartConfig struct {
	Scheme     *scheme.Scheme
	Hook       *hook.Hook
	SpecStore  spec.Store
	ValueStore value.Store
	FS         afero.Fs
}

// loadSpecs loads specs from a file and stores them in the spec store.
func loadSpecs(ctx context.Context, filename string, store spec.Store, fs afero.Fs, env map[string]string) error {
	file, err := fs.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open specs file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read specs file: %v", err)
	}

	var unstructuredSpecs []*spec.Unstructured
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &unstructuredSpecs); err != nil {
			return fmt.Errorf("failed to parse YAML specs: %v", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &unstructuredSpecs); err != nil {
			return fmt.Errorf("failed to parse JSON specs: %v", err)
		}
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}

	for _, unstructured := range unstructuredSpecs {
		if _, err := store.Store(ctx, unstructured); err != nil {
			return fmt.Errorf("failed to store spec: %v", err)
		}
	}

	return nil
}

// loadValues loads values from a file and stores them in the value store.
func loadValues(ctx context.Context, filename string, store value.Store, fs afero.Fs) error {
	file, err := fs.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open values file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read values file: %v", err)
	}

	var values []value.Value
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &values); err != nil {
			return fmt.Errorf("failed to parse YAML values: %v", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &values); err != nil {
			return fmt.Errorf("failed to parse JSON values: %v", err)
		}
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}

	for _, v := range values {
		if _, err := store.Store(ctx, &v); err != nil {
			return fmt.Errorf("failed to store value: %v", err)
		}
	}

	return nil
}

// NewStartCommand creates a new cobra.Command for the start command.
func NewStartCommand(config StartConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Starts the workflow engine within the specified namespace",
		RunE:  runStartCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), resourcebase.DefaultNamespace, "Set the namespace for running the workflow")
	cmd.PersistentFlags().String(flagFromSpecs, "", "Specify the file path containing workflow specifications")
	cmd.PersistentFlags().String(flagFromValues, "", "Specify the file path containing values for the workflow")
	cmd.PersistentFlags().Bool(flagDebug, false, "Enable debug mode for detailed output during execution")
	cmd.PersistentFlags().StringToString(flagEnvironment, nil, "Set environment variables for the workflow execution")

	return cmd
}

// runStartCommand runs the start command with the given configuration.
func runStartCommand(config StartConfig) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()

		// By default, silence usage unless we have a setup error
		cmd.SilenceUsage = true

		namespace, err := cmd.Flags().GetString(flagNamespace)
		if err != nil {
			cmd.SilenceUsage = false
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

		// Load specs from file if specified
		if specsFile, _ := cmd.Flags().GetString(flagFromSpecs); specsFile != "" {
			if err := loadSpecs(ctx, specsFile, config.SpecStore, config.FS, environment); err != nil {
				return err
			}
		}

		// Load values from file if specified
		if valuesFile, _ := cmd.Flags().GetString(flagFromValues); valuesFile != "" {
			if err := loadValues(ctx, valuesFile, config.ValueStore, config.FS); err != nil {
				return err
			}
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
				r.Close()
			}()

			go func() {
				<-sigs
				d.Kill()
			}()

			if err := r.Watch(ctx); err != nil {
				return err
			}
			r.Load(ctx)
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
		r.Load(ctx)
		return r.Reconcile(ctx)
	}
}
