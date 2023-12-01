package main

import (
	"io/fs"

	"github.com/siyul-park/uniflow/cmd/uniflow/apply"
	"github.com/siyul-park/uniflow/cmd/uniflow/start"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/hook"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/spf13/cobra"
)

// Config holds the configuration parameters for the main command.

type Config struct {
	Scheme   *scheme.Scheme
	Hook     *hook.Hook
	Database database.Database
	FS       fs.FS
}

// NewCmd creates the root cobra command for the 'uniflow' CLI.
func NewCmd(config Config) *cobra.Command {
	sc := config.Scheme
	hk := config.Hook
	db := config.Database
	fsys := config.FS

	cmd := &cobra.Command{
		Use:  "uniflow",
		Long: "Create your uniflow and integrate it anywhere!",
	}

	cmd.AddCommand(start.NewCmd(start.Config{
		Scheme:   sc,
		Hook:     hk,
		Database: db,
		FS:       fsys,
	}))
	cmd.AddCommand(apply.NewCmd(apply.Config{
		Scheme:   sc,
		Database: db,
		FS:       fsys,
	}))

	return cmd
}
