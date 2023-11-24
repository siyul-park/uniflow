package apply

import (
	"io"
	"io/fs"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/cmd/flag"
	"github.com/siyul-park/uniflow/cmd/resource"
	"github.com/siyul-park/uniflow/internal/util"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
	"github.com/spf13/cobra"
)

type (
	Config struct {
		Scheme   *scheme.Scheme
		Database database.Database
		FS       fs.FS
	}
)

func NewCmd(config Config) *cobra.Command {
	sc := config.Scheme
	db := config.Database
	fsys := config.FS

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply a configuration to a resource by file name",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			ns, err := cmd.Flags().GetString(FlagNamespace)
			if err != nil {
				return err
			}
			fl, err := cmd.Flags().GetString(FlagFile)
			if err != nil {
				return err
			}

			st, err := storage.New(ctx, storage.Config{
				Scheme:   sc,
				Database: db,
			})
			if err != nil {
				return err
			}

			file, err := fsys.Open(fl)
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

			for _, spec := range specs {
				if spec.GetID() == (ulid.ULID{}) {
					if spec.GetName() != "" {
						exist, err := st.FindOne(ctx, storage.Where[string](scheme.KeyName).EQ(spec.GetName()).And(storage.Where[string](scheme.KeyNamespace).EQ(spec.GetNamespace())))
						if err != nil {
							return err
						}
						if exist != nil {
							spec.SetID(exist.GetID())
						}
					} else {
						spec.SetID(ulid.Make())
					}
				}
			}

			var ids []ulid.ULID
			for _, spec := range specs {
				ids = append(ids, spec.GetID())
			}

			exists, err := st.FindMany(ctx, storage.Where[ulid.ULID](scheme.KeyID).IN(ids...), &database.FindOptions{
				Limit: util.Ptr[int](len(ids)),
			})
			if err != nil {
				return err
			}
			existsIds := make(map[ulid.ULID]struct{}, len(exists))
			for _, spec := range exists {
				existsIds[spec.GetID()] = struct{}{}
			}

			var inserted []scheme.Spec
			var updated []scheme.Spec
			for _, spec := range specs {
				if _, ok := existsIds[spec.GetID()]; ok {
					updated = append(updated, spec)
				} else {
					inserted = append(inserted, spec)
				}
			}

			if _, err := st.InsertMany(ctx, inserted); err != nil {
				return err
			}
			if _, err := st.UpdateMany(ctx, updated); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringP(FlagNamespace, flag.ToShorthand(FlagNamespace), "", "uniflow namespace")
	cmd.PersistentFlags().StringP(FlagFile, flag.ToShorthand(FlagFile), "", "configuration file name")

	return cmd
}
