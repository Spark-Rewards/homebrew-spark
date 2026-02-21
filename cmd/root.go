package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:     "spark-cli",
	Short:   "spark-cli — multi-repo workspace CLI",
	Version: Version,
	Long: `spark-cli manages multi-repo workspaces with shared environment and smart builds.
`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("spark-cli %s (%s %s)\n", Version, Commit, Date))
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// No "help" subcommand — use -h/--help only
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
}
