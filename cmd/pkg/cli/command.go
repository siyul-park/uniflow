package cli

import (
	"runtime"
	"runtime/pprof"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// Config is a structure to hold the configuration for the CLI command.
type Config struct {
	Use   string
	Short string
	FS    afero.Fs
}

// NewCommand creates the root cobra command for the 'uniflow' CLI.
func NewCommand(config Config) *cobra.Command {
	var cpuprof afero.File

	cmd := &cobra.Command{
		Use:   config.Use,
		Short: config.Short,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cpuprofile, err := cmd.Flags().GetString(flagCPUProfile)
			if err != nil {
				return err
			}

			if cpuprofile != "" {
				cpuprof, err = config.FS.Create(cpuprofile)
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
				f, err := config.FS.Create(memprofile)
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

	return cmd
}
