package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Spark-Rewards/homebrew-spk/internal/workspace"
	"github.com/spf13/cobra"
)

var testWatch bool

var knownTestCommands = map[string]string{
	"AppAPI":      "npm test",
	"BusinessAPI": "npm test",
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run tests for current repo",
	Long: `Runs the test command for the current repo.

Must be run from inside a repo directory.

Examples:
  cd AppAPI && spk test        # run tests
  spk test --watch             # run tests in watch mode`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		wsPath, err := workspace.Find()
		if err != nil {
			return err
		}

		ws, err := workspace.Load(wsPath)
		if err != nil {
			return err
		}

		repoName, err := detectCurrentRepoForTest(wsPath, ws)
		if err != nil {
			return fmt.Errorf("must be run from inside a repo directory")
		}

		return testRepo(wsPath, ws, repoName)
	},
}

func detectCurrentRepoForTest(wsPath string, ws *workspace.Workspace) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not get current directory: %w", err)
	}

	for name, repo := range ws.Repos {
		repoDir := filepath.Join(wsPath, repo.Path)
		absRepoDir, _ := filepath.Abs(repoDir)

		if cwd == absRepoDir || isSubdirTest(absRepoDir, cwd) {
			return name, nil
		}
	}

	return "", fmt.Errorf("not inside a repo directory — specify a repo name or use --all")
}

func isSubdirTest(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return !filepath.IsAbs(rel) && len(rel) > 0 && rel[0] != '.'
}

func getTestCommand(name string, repo workspace.RepoDef, repoDir string) string {
	if repo.TestCommand != "" {
		return repo.TestCommand
	}

	if cmd, ok := knownTestCommands[name]; ok {
		if testWatch {
			return cmd + ":watch"
		}
		return cmd
	}

	if fileExists(filepath.Join(repoDir, "package.json")) {
		if testWatch {
			return "npm run test:watch"
		}
		return "npm test"
	}

	if fileExists(filepath.Join(repoDir, "build.gradle")) || fileExists(filepath.Join(repoDir, "build.gradle.kts")) {
		return "./gradlew test"
	}

	if fileExists(filepath.Join(repoDir, "go.mod")) {
		return "go test ./..."
	}

	return ""
}

func testRepo(wsPath string, ws *workspace.Workspace, name string) error {
	repo, ok := ws.Repos[name]
	if !ok {
		return fmt.Errorf("repo '%s' not found in workspace", name)
	}

	repoDir := filepath.Join(wsPath, repo.Path)
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return fmt.Errorf("repo directory %s does not exist", repoDir)
	}

	testCmd := getTestCommand(name, repo, repoDir)
	if testCmd == "" {
		fmt.Printf("No test command for '%s' — skipping\n", name)
		return nil
	}

	fmt.Printf("Testing %s: %s\n", name, testCmd)
	return runShell(repoDir, testCmd)
}

func init() {
	testCmd.Flags().BoolVarP(&testWatch, "watch", "w", false, "Run tests in watch mode")
	rootCmd.AddCommand(testCmd)
}
