package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BrianBFarias/homebrew-spark/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	createAWSProfile string
	createAWSRegion  string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create resources (workspace)",
}

var createWorkspaceCmd = &cobra.Command{
	Use:   "workspace [path]",
	Short: "Create a new spark workspace",
	Long: `Creates a new workspace directory with a .spark/workspace.json manifest.
If the directory doesn't exist, it will be created.

Examples:
  spark create workspace .                     # current dir
  spark create workspace ./my-project          # relative path
  spark create workspace ~/Projects/my-app     # absolute path`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath := args[0]

		absPath, err := filepath.Abs(targetPath)
		if err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}

		// Create directory if it doesn't exist
		if err := os.MkdirAll(absPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Check if workspace already exists
		manifestPath := workspace.ManifestPath(absPath)
		if _, err := os.Stat(manifestPath); err == nil {
			return fmt.Errorf("workspace already exists at %s", absPath)
		}

		name := filepath.Base(absPath)

		ws, err := workspace.Create(absPath, name, createAWSProfile, createAWSRegion)
		if err != nil {
			return err
		}

		fmt.Printf("Workspace '%s' created at %s\n", ws.Name, absPath)
		fmt.Printf("  AWS Profile: %s\n", orDefault(ws.AWSProfile, "(not set)"))
		fmt.Printf("  AWS Region:  %s\n", orDefault(ws.AWSRegion, "(not set)"))
		fmt.Println("\nNext steps:")
		fmt.Println("  cd", absPath)
		fmt.Println("  spark use <org/repo>")
		return nil
	},
}

func orDefault(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

func init() {
	createWorkspaceCmd.Flags().StringVar(&createAWSProfile, "aws-profile", "", "AWS SSO profile name")
	createWorkspaceCmd.Flags().StringVar(&createAWSRegion, "aws-region", "", "Default AWS region")
	createCmd.AddCommand(createWorkspaceCmd)
	rootCmd.AddCommand(createCmd)
}
