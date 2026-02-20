package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "spark",
	Short: "Spark — workspace CLI for multi-repo development",
	Long: `Spark is a workspace-oriented CLI that keeps multiple repositories
in sync, manages AWS credentials at the workspace level, and provides
dependency-aware builds across projects.

Get started:
  spark create workspace ./my-project
  cd my-project
  spark use org/repo-name
  spark sync`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
