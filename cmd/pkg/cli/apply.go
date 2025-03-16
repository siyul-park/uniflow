package cli

import (
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/cmd/pkg/io"
	"github.com/siyul-park/uniflow/pkg/meta"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/siyul-park/uniflow/pkg/value"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// ApplyConfig represents the configuration for the apply command.
type ApplyConfig struct {
	SpecStore  store.Store
	ValueStore store.Store
	FS         afero.Fs
}

// NewApplyCommand creates a new cobra.Command for the apply command.
func NewApplyCommand(config ApplyConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "apply",
		Short:     "Apply resources to the specified namespace",
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{specs, values},
		RunE: runs(map[string]func(cmd *cobra.Command) error{
			specs:  runApplyCommand[spec.Spec](config.SpecStore, config.FS),
			values: runApplyCommand[*value.Value](config.ValueStore, config.FS),
		}),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), meta.DefaultNamespace, "Set the io's namespace. If not set, use the default namespace")
	cmd.PersistentFlags().StringP(flagFilename, toShorthand(flagFilename), "", "Set the file path to be applied")

	return cmd
}

func runApplyCommand[T meta.Meta](st store.Store, fs afero.Fs, alias ...func(map[string]string)) func(cmd *cobra.Command) error {
	flags := map[string]string{
		flagNamespace: flagNamespace,
		flagFilename:  flagFilename,
	}
	for _, init := range alias {
		init(flags)
	}

	return func(cmd *cobra.Command) error {
		ctx := cmd.Context()

		namespace, err := cmd.Flags().GetString(flags[flagNamespace])
		if err != nil {
			return err
		}
		filename, err := cmd.Flags().GetString(flags[flagFilename])
		if err != nil {
			return err
		}
		if filename == "" {
			return nil
		}

		file, err := fs.Open(filename)
		if err != nil {
			return err
		}

		defer file.Close()

		reader := io.NewReader(file)
		writer := io.NewWriter(cmd.OutOrStdout())

		var metas []T
		if err := reader.Read(&metas); err != nil {
			return err
		}

		if len(metas) == 0 {
			return nil
		}

		for _, m := range metas {
			if m.GetNamespace() == "" {
				m.SetNamespace(namespace)
			}

			filter := map[string]any{}
			if m.GetID() != uuid.Nil {
				filter[meta.KeyID] = m.GetID()
			}
			if m.GetName() != "" {
				filter[meta.KeyName] = m.GetName()
			}

			cursor, err := st.Find(ctx, filter, store.FindOptions{Limit: 1})
			if err != nil {
				return err
			}

			ok := cursor.Next(ctx)
			_ = cursor.Close(ctx)

			if ok {
				_, err := st.Update(ctx, filter, map[string]any{"$set": m})
				if err != nil {
					return err
				}
			} else {
				if m.GetID() == uuid.Nil {
					m.SetID(uuid.Must(uuid.NewV7()))
				}

				err := st.Insert(ctx, []any{m})
				if err != nil {
					return err
				}
			}
		}

		return writer.Write(metas)
	}
}
