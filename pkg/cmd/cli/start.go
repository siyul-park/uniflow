package cli

import (
	"context"
	"io/fs"
	"os"
	"os/signal"
	"syscall"

	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/cmd/scanner"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
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
		Short: "Start a worker process",
		RunE:  runStartCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), "", "Set the worker's namespace")
	cmd.PersistentFlags().StringP(flagBoot, toShorthand(flagBoot), "", "Set the boot file path for initializing nodes")

	return cmd
}

func runStartCommand(config StartConfig) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		ns, err := cmd.Flags().GetString(flagNamespace)
		if err != nil {
			return err
		}

		boot, err := cmd.Flags().GetString(flagBoot)
		if err != nil {
			return err
		}

		if boot != "" {
			if err := initializeNamespace(ctx, config, ns, boot); err != nil {
				return err
			}
		}

		r, err := runtime.New(ctx, runtime.Config{
			Namespace: ns,
			Scheme:    config.Scheme,
			Hooks:     config.Hook,
			Database:  config.Database,
		})
		if err != nil {
			return err
		}

		handleSignals(ctx, r)
		return r.Start(ctx)
	}
}

func initializeNamespace(ctx context.Context, config StartConfig, ns, boot string) error {
	st, err := storage.New(ctx, storage.Config{
		Scheme:   config.Scheme,
		Database: config.Database,
	})
	if err != nil {
		return err
	}

	filter := storage.Where[string](scheme.KeyNamespace).EQ(ns)
	specs, err := st.FindMany(ctx, filter, &database.FindOptions{Limit: lo.ToPtr[int](1)})
	if err != nil {
		return err
	}

	if len(specs) == 0 {
		if err := installBootFile(ctx, config, ns, boot); err != nil {
			return err
		}
	}
	return nil
}

func installBootFile(ctx context.Context, config StartConfig, ns, boot string) error {
	specs, err := scanner.New().
		Scheme(config.Scheme).
		Namespace(ns).
		FS(config.FS).
		Filename(boot).
		Scan()
	if err != nil {
		return err
	}

	st, err := storage.New(ctx, storage.Config{Scheme: config.Scheme, Database: config.Database})
	if err != nil {
		return err
	}

	_, err = st.InsertMany(ctx, specs)
	return err
}

func handleSignals(ctx context.Context, r *runtime.Runtime) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		_ = r.Close(ctx)
	}()
}
