package cli

import (
	"fmt"
	"io/fs"
	"log"
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

	var cpuprof *os.File

	cmd := &cobra.Command{
		Use:  "uniflow",
		Long: "Low-Code Engine for Backend Workflows",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cpuprofile, err := cmd.Flags().GetString(flagCPUProfile)
			if err != nil {
				return err
			}

			if cpuprofile != "" {
				fmt.Printf("Using cpu profile: %s\n", cpuprofile)

				cpuprof, err = os.Create(cpuprofile)
				if err != nil {
					return err
				}

				if err := pprof.StartCPUProfile(cpuprof); err != nil {
					return err
				}
			}
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			memprofile, err := cmd.Flags().GetString(flagMemProfile)
			if err != nil {
				return err
			}

			if cpuprof != nil {
				defer func() {
					cpuprof = nil
				}()

				pprof.StopCPUProfile()
				if err := cpuprof.Close(); err != nil {
					return err
				}
			}

			if memprofile != "" {
				log.Printf("Using mem profile: %s\n", memprofile)

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
	cmd.AddCommand(NewExecCommand(ExecConfig{
		Scheme:   sc,
		Hook:     hk,
		Database: db,
		FS:       fsys,
	}))

	return cmd
}
