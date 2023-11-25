package apply

import (
	"fmt"
	"io/fs"

	"github.com/oklog/ulid/v2"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/cmd/flag"
	"github.com/siyul-park/uniflow/cmd/printer"
	"github.com/siyul-park/uniflow/cmd/resource"
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

var (
	SpecTableColumnDefinitions = []printer.TableColumnDefinition{
		{
			Name:   "id",
			Format: "$.id",
		},
		{
			Name:   "kind",
			Format: "$.kind",
		},
		{
			Name:   "name",
			Format: "$.name",
		},
		{
			Name:   "namespace",
			Format: "$.namespace",
		},
		{
			Name:   "links",
			Format: "$.links",
		},
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

			b := resource.NewBuilder().
				Scheme(sc).
				Namespace(ns).
				FS(fsys).
				Filename(fl)

			specs, err := b.Build()
			if err != nil {
				return err
			}

			for _, spec := range specs {
				if spec.GetID() == (ulid.ULID{}) {
					if spec.GetName() != "" {
						filter := storage.Where[string](scheme.KeyName).EQ(spec.GetName()).And(storage.Where[string](scheme.KeyNamespace).EQ(spec.GetNamespace()))
						if exist, err := st.FindOne(ctx, filter); err != nil {
							return err
						} else if exist != nil {
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
				Limit: lo.ToPtr[int](len(ids)),
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

			tablePrinter, err := printer.NewTable(SpecTableColumnDefinitions)
			if err != nil {
				return err
			}

			table, err := tablePrinter.Print(specs)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprint(cmd.OutOrStdout(), table); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringP(FlagNamespace, flag.ToShorthand(FlagNamespace), "", "Set the namespace. If not set it up, use default namespace. In this case.")
	cmd.PersistentFlags().StringP(FlagFile, flag.ToShorthand(FlagFile), "", "Set the file path that want to be applied.")

	return cmd
}
