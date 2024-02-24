package cli

import (
	"io/fs"

	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/spf13/cobra"
)

// Config holds the configuration parameters for the root command.
type Config struct {
	Scheme   *scheme.Scheme
	Hook     *hook.Hook
	Database database.Database
	FS       fs.FS
}

// NewCommand creates the root cobra command for the 'uniflow' CLI.
func NewCommand(config Config) *cobra.Command {
	sc := config.Scheme
	hk := config.Hook
	db := config.Database
	fsys := config.FS

	cmd := &cobra.Command{
		Use:  "uniflow",
		Long: "Low-Code Engine for Backend Workflows",
	}

	cmd.AddCommand(NewApplyCommand(ApplyConfig{
		Scheme:   sc,
		Database: db,
		FS:       fsys,
	}))
	cmd.AddCommand(NewDeleteCommand(DeleteConfig{
		Scheme:   sc,
		Database: db,
		FS:       fsys,
	}))
	cmd.AddCommand(NewGetCommand(GetConfig{
		Scheme:   sc,
		Database: db,
	}))
	cmd.AddCommand(NewStartCommand(StartConfig{
		Scheme:   sc,
		Hook:     hk,
		Database: db,
		FS:       fsys,
	}))

	return cmd
}
