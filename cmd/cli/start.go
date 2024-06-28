package cli

import (
	"io/fs"
	"os"
	"os/signal"
	"syscall"

	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/cmd/scanner"
	"github.com/siyul-park/uniflow/database"
	"github.com/siyul-park/uniflow/hook"
	"github.com/siyul-park/uniflow/runtime"
	"github.com/siyul-park/uniflow/scheme"
	"github.com/siyul-park/uniflow/spec"
	"github.com/siyul-park/uniflow/store"
	"github.com/spf13/cobra"
)

// StartConfig holds the configuration for the uniflow command.
type StartConfig struct {
	Scheme   *scheme.Scheme
	Hook     *hook.Hook
	Database database.Database
	FS       fs.FS
}

// NewStartCommand creates a new Cobra command for the uniflow application.
func NewStartCommand(config StartConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a application",
		RunE:  runStartCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), "", "Set the namespace for running")
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

		if filename != "" {
			st, err := store.New(ctx, store.Config{
				Scheme:   config.Scheme,
				Database: config.Database,
			})
			if err != nil {
				return err
			}

			filter := store.Where[string](spec.KeyNamespace).EQ(namespace)
			specs, err := st.FindMany(ctx, filter, &database.FindOptions{Limit: lo.ToPtr[int](1)})
			if err != nil {
				return err
			}
			if len(specs) != 0 {
				return nil
			}

			specs, err = scanner.New().
				Scheme(config.Scheme).
				Store(st).
				Namespace(namespace).
				FS(config.FS).
				Filename(filename).
				Scan(ctx)
			if err != nil {
				return err
			}
			if _, err = st.InsertMany(ctx, specs); err != nil {
				return err
			}
		}

		r, err := runtime.New(ctx, runtime.Config{
			Namespace: namespace,
			Scheme:    config.Scheme,
			Hook:      config.Hook,
			Database:  config.Database,
		})
		if err != nil {
			return err
		}

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigs
			_ = r.Close()
		}()

		return r.Start(ctx)
	}
}
