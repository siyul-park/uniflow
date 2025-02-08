package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/siyul-park/uniflow/ext/pkg/control"
	"github.com/siyul-park/uniflow/ext/pkg/language"
	"github.com/siyul-park/uniflow/ext/pkg/language/javascript"
	"github.com/siyul-park/uniflow/ext/pkg/language/text"
	"github.com/siyul-park/uniflow/pkg/hook"
	resourcebase "github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/testing"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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

	cmd.Flags().String(flagNamespace, resourcebase.DefaultNamespace, "Namespace to run tests in")
	cmd.Flags().StringToString(flagEnvironment, nil, "Environment variables for test execution")
	cmd.Flags().String(flagFromSpecs, "", "Load specs from file")
	cmd.Flags().String(flagFromValues, "", "Load values from file")
	cmd.Flags().Bool(flagDebug, false, "Enable debug logging")

	return cmd
}

func runTestCommand(config TestConfig) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		// By default, silence usage unless we have a setup error
		cmd.SilenceUsage = true

		debug, _ := cmd.Flags().GetBool(flagDebug)
		if !debug {
			// If not in debug mode, discard logs
			log.SetOutput(io.Discard)
		} else {
			// In debug mode, output to stderr with file and line number
			log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
			log.SetOutput(os.Stderr)
		}

		namespace, err := cmd.Flags().GetString(flagNamespace)
		if err != nil {
			return fmt.Errorf("failed to get namespace: %v", err)
		}
		if debug {
			log.Printf("Using namespace: %s", namespace)
		}

		environment, err := cmd.Flags().GetStringToString(flagEnvironment)
		if err != nil {
			return fmt.Errorf("failed to get environment: %v", err)
		}
		if debug {
			log.Printf("Environment variables: %v", environment)
		}

		match := func(string) bool { return true }
		if len(args) > 0 {
			exp, err := regexp.Compile(args[0])
			if err != nil {
				return fmt.Errorf("failed to compile regex: %v", err)
			}
			match = func(name string) bool {
				return exp.Match([]byte(name))
			}
		}

		// Initialize language module
		module := language.NewModule()
		module.Store(text.Language, text.NewCompiler())
		module.Store(javascript.Language, javascript.NewCompiler())
		if debug {
			log.Printf("Initialized language module with: %s, %s", text.Language, javascript.Language)
		}

		// Register test node
		if err := testing.AddToScheme().AddToScheme(config.Scheme); err != nil {
			return fmt.Errorf("failed to register test node: %v", err)
		}
		if debug {
			log.Printf("Registered test node")
		}

		// Register control nodes with text language for conditions and assertions
		textRegister := control.AddToScheme(module, text.Language)
		if err := textRegister.AddToScheme(config.Scheme); err != nil {
			return fmt.Errorf("failed to register control nodes with text language: %v", err)
		}
		if debug {
			log.Printf("Registered control nodes with text language")
		}

		// Register control nodes with JavaScript language for snippets
		jsRegister := control.AddToScheme(module, javascript.Language)
		if err := jsRegister.AddToScheme(config.Scheme); err != nil {
			return fmt.Errorf("failed to register control nodes with JavaScript language: %v", err)
		}
		if debug {
			log.Printf("Registered control nodes with JavaScript language")
		}

		reporter := testing.NewPrettyReporter(os.Stderr)
		config.Runner.AddReporter(reporter)
		defer config.Runner.RemoveReporter(reporter)

		// Process specs from file
		if specsFile, _ := cmd.Flags().GetString(flagFromSpecs); specsFile != "" {
			if debug {
				log.Printf("Loading specs from file: %s", specsFile)
			}

			file, err := config.FS.Open(specsFile)
			if err != nil {
				return fmt.Errorf("failed to open specs file: %v", err)
			}
			defer file.Close()

			data, err := io.ReadAll(file)
			if err != nil {
				return fmt.Errorf("failed to read specs file: %v", err)
			}

			var unstructuredSpecs []*spec.Unstructured
			ext := strings.ToLower(filepath.Ext(specsFile))
			switch ext {
			case ".yaml", ".yml":
				if err := yaml.Unmarshal(data, &unstructuredSpecs); err != nil {
					return fmt.Errorf("failed to parse YAML specs: %v", err)
				}
				if debug {
					log.Printf("Parsed %d specs from YAML", len(unstructuredSpecs))
				}
			case ".json":
				if err := json.Unmarshal(data, &unstructuredSpecs); err != nil {
					return fmt.Errorf("failed to parse JSON specs: %v", err)
				}
				if debug {
					log.Printf("Parsed %d specs from JSON", len(unstructuredSpecs))
				}
			default:
				return fmt.Errorf("unsupported file format: %s", ext)
			}

			for _, unstructured := range unstructuredSpecs {
				if debug {
					log.Printf("Processing spec: kind=%s, name=%s, namespace=%s",
						unstructured.GetKind(), unstructured.GetName(), unstructured.GetNamespace())
				}

				if _, err := config.SpecStore.Store(ctx, unstructured); err != nil {
					return fmt.Errorf("failed to store spec: %v", err)
				}

				// Register the test suite if it's a test node
				if unstructured.GetKind() == testing.KindTest {
					if debug {
						log.Printf("Compiling test node: %s", unstructured.GetName())
					}

					node, err := config.Scheme.Compile(unstructured)
					if err != nil {
						if debug {
							log.Printf("Failed to compile test node: %+v", err)
						}
						return fmt.Errorf("failed to compile test node: %v", err)
					}
					if suite, ok := node.(testing.Suite); ok {
						config.Runner.Register(unstructured.GetName(), suite)
						if debug {
							log.Printf("Successfully registered test suite: %s", unstructured.GetName())
						}
					} else {
						return fmt.Errorf("node is not a test suite: %v", node)
					}
				}
			}
		}

		// Load values from file if specified
		if valuesFile, _ := cmd.Flags().GetString(flagFromValues); valuesFile != "" {
			file, err := config.FS.Open(valuesFile)
			if err != nil {
				return fmt.Errorf("failed to open values file: %v", err)
			}
			defer file.Close()

			data, err := io.ReadAll(file)
			if err != nil {
				return fmt.Errorf("failed to read values file: %v", err)
			}

			var values []value.Value
			ext := strings.ToLower(filepath.Ext(valuesFile))
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
				if _, err := config.ValueStore.Store(ctx, &v); err != nil {
					return fmt.Errorf("failed to store value: %v", err)
				}
			}
		}

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

		return config.Runner.Run(ctx, match)
	}
}
