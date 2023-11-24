package start

import (
	"io"
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

type (
	Config struct {
		Scheme   *scheme.Scheme
		Hook     *hook.Hook
		Database database.Database
		FS       fs.FS
	}
)

func NewCmd(config Config) *cobra.Command {
	sc := config.Scheme
	hk := config.Hook
	db := config.Database
	fsys := config.FS

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a uniflow worker",
		RunE: func(cmd *cobra.Command, _ []string) error {
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
				st, err := storage.New(ctx, storage.Config{
					Scheme:   sc,
					Database: db,
				})
				if err != nil {
					return err
				}

				var filter *storage.Filter
				if ns != "" {
					filter = storage.Where[string](scheme.KeyNamespace).EQ(ns)
				}

				if specs, err := st.FindMany(ctx, filter, &database.FindOptions{
					Limit: lo.ToPtr[int](1),
				}); err != nil {
					return err
				} else if len(specs) == 0 {
					file, err := fsys.Open(boot)
					if err != nil {
						return err
					}
					defer func() { _ = file.Close() }()

					data, err := io.ReadAll(file)
					if err != nil {
						return err
					}

					var raws []map[string]any
					if err := resource.UnmarshalYAMLOrJSON(data, &raws); err != nil {
						var e map[string]any
						if err := resource.UnmarshalYAMLOrJSON(data, &e); err != nil {
							return err
						} else {
							raws = []map[string]any{e}
						}
					}

					codec := resource.NewSpecCodec(resource.SpecCodecOptions{
						Scheme:    sc,
						Namespace: ns,
					})

					var specs []scheme.Spec
					for _, raw := range raws {
						if spec, err := codec.Decode(raw); err != nil {
							return err
						} else {
							specs = append(specs, spec)
						}
					}

					if _, err := st.InsertMany(ctx, specs); err != nil {
						return err
					}
				}
			}

			r, err := runtime.New(ctx, runtime.Config{
				Namespace: ns,
				Scheme:    sc,
				Hooks:     hk,
				Database:  db,
			})
			if err != nil {
				return err
			}

			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigs
				_ = r.Close(ctx)
			}()

			return r.Start(ctx)
		},
	}

	cmd.PersistentFlags().StringP(FlagNamespace, flag.ToShorthand(FlagNamespace), "", "Set the namespace. If not set it up, runs all namespaces. In this case, if namespace is sharing resources exclusively, some nodes may not run normally.")
	cmd.PersistentFlags().StringP(FlagBoot, flag.ToShorthand(FlagBoot), "", "Set the boot file path that must be installed initially if the node does not exist in namespace.")

	return cmd
}
