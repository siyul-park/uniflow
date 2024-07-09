package cli

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/cmd/pkg/scanner"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// StartConfig holds the configuration for the uniflow command.
type StartConfig struct {
	Scheme *scheme.Scheme
	Hook   *hook.Hook
	Store  *store.Store
	FS     afero.Fs
}

// NewStartCommand creates a new Cobra command for the uniflow application.
func NewStartCommand(config StartConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Starts the workflow engine within the specified namespace",
		RunE:  runStartCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), spec.DefaultNamespace, "Set the namespace for running")
	cmd.PersistentFlags().StringP(flagFilename, toShorthand(flagFilename), "", "Set the boot file path for initializing nodes")

	return cmd
}

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

		if err := config.Store.Index(ctx); err != nil {
			return err
		}

		if filename != "" {
			filter := store.Where[string](spec.KeyNamespace).EQ(namespace)
			specs, err := config.Store.FindMany(ctx, filter, &database.FindOptions{Limit: lo.ToPtr[int](1)})
			if err != nil {
				return err
			}
			if len(specs) != 0 {
				return nil
			}

			specs, err = scanner.New().
				Scheme(config.Scheme).
				Store(config.Store).
				Namespace(namespace).
				FS(config.FS).
				Filename(filename).
				Scan(ctx)
			if err != nil {
				return err
			}

			if _, err = config.Store.InsertMany(ctx, specs); err != nil {
				return err
			}
		}

		r := runtime.New(runtime.Config{
			Namespace: namespace,
			Scheme:    config.Scheme,
			Hook:      config.Hook,
			Store:     config.Store,
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
