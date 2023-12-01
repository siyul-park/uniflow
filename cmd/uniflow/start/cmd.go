package start

import (
	"context"
	"io/fs"
	"os"
	"os/signal"
	"syscall"

	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/cmd/flag"
	"github.com/siyul-park/uniflow/cmd/resource"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/spf13/cobra"
)

// Config holds the configuration for the uniflow command.
type Config struct {
	Scheme   *scheme.Scheme
	Hook     *hook.Hook
	Database database.Database
	FS       fs.FS
}

// NewCmd creates a new Cobra command for the uniflow application.
func NewCmd(config Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a worker process",
		RunE:  runStartCommand(config),
	}

	cmd.PersistentFlags().StringP(FlagNamespace, flag.ToShorthand(FlagNamespace), "", "Set the worker's namespace.")
	cmd.PersistentFlags().StringP(FlagBoot, flag.ToShorthand(FlagBoot), "", "Set the boot file path for initializing nodes.")

	return cmd
}

func runStartCommand(config Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		ns, err := cmd.Flags().GetString(FlagNamespace)
		if err != nil {
			return err
		}

		boot, err := cmd.Flags().GetString(FlagBoot)
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

func initializeNamespace(ctx context.Context, config Config, ns, boot string) error {
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

func installBootFile(ctx context.Context, config Config, ns, boot string) error {
	b := resource.NewBuilder().Scheme(config.Scheme).Namespace(ns).FS(config.FS).Filename(boot)
	specs, err := b.Build()
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
