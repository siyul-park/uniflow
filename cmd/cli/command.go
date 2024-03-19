package cli

import (
	"io/fs"
	"os"
	"runtime"
	"runtime/pprof"

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
		PreRunE: func(cmd *cobra.Command, args []string) error {
			cpuprofile, err := cmd.Flags().GetString(flagCPUProfile)
			if err != nil {
				return err
			}

			if cpuprofile != "" {
				f, err := os.Create(cpuprofile)
				if err != nil {
					return err
				}
				defer f.Close()

				if err := pprof.StartCPUProfile(f); err != nil {
					return err
				}
				defer pprof.StopCPUProfile()
			}
			return nil
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			memprofile, err := cmd.Flags().GetString(flagMemProfile)
			if err != nil {
				return err
			}

			if memprofile != "" {
				f, err := os.Create(memprofile)
				if err != nil {
					return err
				}
				defer f.Close()

				runtime.GC()
				if err := pprof.WriteHeapProfile(f); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.PersistentFlags().String(flagCPUProfile, "", "write cpu profile to `file`")
	cmd.PersistentFlags().String(flagMemProfile, "", "write memory profile to `file`")

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
