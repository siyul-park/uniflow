package cli

import (
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/cmd/pkg/printer"
	"github.com/siyul-park/uniflow/cmd/pkg/scanner"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// ApplyConfig represents the configuration for the apply command.
type ApplyConfig struct {
	Scheme *scheme.Scheme
	Store  *spec.Store
	FS     afero.Fs
}

// NewApplyCommand creates a new cobra.Command for the apply command.
func NewApplyCommand(config ApplyConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply node specifications to the specified namespace",
		RunE:  runApplyCommand(config),
	}

	cmd.PersistentFlags().StringP(flagNamespace, toShorthand(flagNamespace), spec.DefaultNamespace, "Set the resource's namespace. If not set, use the default namespace")
	cmd.PersistentFlags().StringP(flagFilename, toShorthand(flagFilename), "", "Set the file path to be applied")

	return cmd
}

func runApplyCommand(config ApplyConfig) func(cmd *cobra.Command, args []string) error {
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

		specs, err := scanner.New().
			Scheme(config.Scheme).
			Store(config.Store).
			Namespace(namespace).
			FS(config.FS).
			Filename(filename).
			Scan(ctx)
		if err != nil {
			return err
		}

		origins, err := config.Store.Load(ctx, specs...)
		if err != nil {
			return err
		}

		exists := make(map[uuid.UUID]struct{}, len(origins))
		for _, spec := range origins {
			exists[spec.GetID()] = struct{}{}
		}

		var inserts []spec.Spec
		var updates []spec.Spec
		for _, spec := range specs {
			if _, ok := exists[spec.GetID()]; ok {
				updates = append(updates, spec)
			} else {
				inserts = append(inserts, spec)
			}
		}

		if _, err := config.Store.Store(ctx, inserts...); err != nil {
			return err
		}
		if _, err := config.Store.Swap(ctx, updates...); err != nil {
			return err
		}

		return printer.PrintTable(cmd.OutOrStdout(), specs, printer.SpecTableColumnDefinitions)
	}
}
