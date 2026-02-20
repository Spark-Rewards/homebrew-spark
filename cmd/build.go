package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Spark-Rewards/homebrew-spk/internal/npm"
	"github.com/Spark-Rewards/homebrew-spk/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	buildRecursive bool
	buildPublished bool
)

var knownBuildCommands = map[string]string{
	"AppModel":      "npm run build:all",
	"BusinessModel": "npm run build:all",
	"AppAPI":        "npm run build",
	"BusinessAPI":   "npm run build",
}

type depMapping struct {
	api string
	pkg string
}

var modelToAPI = map[string]depMapping{
	"AppModel":      {api: "AppAPI", pkg: "@spark-rewards/sra-sdk"},
	"BusinessModel": {api: "BusinessAPI", pkg: "@spark-rewards/srw-sdk"},
}

var apiToModel = map[string]string{
	"AppAPI":      "AppModel",
	"BusinessAPI": "BusinessModel",
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build current repo with automatic local dependency linking",
	Long: `Builds the current repo and automatically links locally-built dependencies.

Must be run from inside a repo directory.

Like Amazon's Brazil Build, spk automatically detects when a dependency
(like a Smithy model) is built locally and links it to consuming packages
(like APIs) instead of using published versions.

Dependency chain:
  AppModel      -> AppAPI      (@spark-rewards/sra-sdk)
  BusinessModel -> BusinessAPI (@spark-rewards/srw-sdk)

With --recursive (-r), builds dependencies first, then the current repo.

Examples:
  cd AppAPI && spk build       # build AppAPI (links local AppModel if built)
  cd AppAPI && spk build -r    # build AppModel first, then AppAPI
  spk build --published        # force use of published packages`,
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

		repoName, err := detectCurrentRepo(wsPath, ws)
		if err != nil {
			return fmt.Errorf("must be run from inside a repo directory")
		}

		if buildRecursive {
			return buildRecursively(wsPath, ws, repoName)
		}

		return buildRepo(wsPath, ws, repoName)
	},
}

func detectCurrentRepo(wsPath string, ws *workspace.Workspace) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not get current directory: %w", err)
	}

	for name, repo := range ws.Repos {
		repoDir := filepath.Join(wsPath, repo.Path)
		absRepoDir, _ := filepath.Abs(repoDir)

		if cwd == absRepoDir || isSubdir(absRepoDir, cwd) {
			return name, nil
		}
	}

	return "", fmt.Errorf("not inside a repo directory — specify a repo name or use --all")
}

func isSubdir(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return !filepath.IsAbs(rel) && len(rel) > 0 && rel[0] != '.'
}

func getBuildCommand(name string, repo workspace.RepoDef, repoDir string) string {
	if repo.BuildCommand != "" {
		return repo.BuildCommand
	}

	if cmd, ok := knownBuildCommands[name]; ok {
		return cmd
	}

	if fileExists(filepath.Join(repoDir, "package.json")) {
		return "npm run build"
	}
	if fileExists(filepath.Join(repoDir, "build.gradle")) || fileExists(filepath.Join(repoDir, "build.gradle.kts")) {
		return "./gradlew build"
	}
	if fileExists(filepath.Join(repoDir, "Makefile")) {
		return "make"
	}
	if fileExists(filepath.Join(repoDir, "go.mod")) {
		return "go build ./..."
	}

	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func buildRepo(wsPath string, ws *workspace.Workspace, name string) error {
	repo, ok := ws.Repos[name]
	if !ok {
		return fmt.Errorf("repo '%s' not found in workspace", name)
	}

	repoDir := filepath.Join(wsPath, repo.Path)
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return fmt.Errorf("repo directory %s does not exist", repoDir)
	}

	fmt.Printf("=== Building %s ===\n", name)

	if !buildPublished {
		if err := autoLinkDependencies(wsPath, ws, name); err != nil {
			fmt.Printf("Warning: dependency linking issue: %v\n", err)
		}
	}

	buildCmd := getBuildCommand(name, repo, repoDir)
	if buildCmd == "" {
		fmt.Printf("No build command for '%s' — skipping\n", name)
		return nil
	}

	fmt.Printf("Running: %s\n", buildCmd)
	if err := runShell(repoDir, buildCmd); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	if !buildPublished {
		if err := autoLinkToConsumers(wsPath, ws, name); err != nil {
			fmt.Printf("Note: %v\n", err)
		}
	}

	fmt.Printf("[ok] %s built successfully\n", name)
	return nil
}

func autoLinkDependencies(wsPath string, ws *workspace.Workspace, name string) error {
	modelName, isAPI := apiToModel[name]
	if !isAPI {
		return nil
	}

	modelRepo, exists := ws.Repos[modelName]
	if !exists {
		return nil
	}

	modelDir := filepath.Join(wsPath, modelRepo.Path)
	apiDir := filepath.Join(wsPath, ws.Repos[name].Path)
	mapping := modelToAPI[modelName]

	if !npm.IsBuilt(modelDir) {
		fmt.Printf("Using published %s (local not built)\n", mapping.pkg)
		return nil
	}

	if npm.IsLinked(apiDir, mapping.pkg) {
		fmt.Printf("Using local %s (already linked)\n", modelName)
		return nil
	}

	fmt.Printf("Linking local %s -> %s...\n", modelName, name)
	buildDir := npm.BuildOutputDir(modelDir)

	if err := npm.Link(buildDir); err != nil {
		return fmt.Errorf("npm link in %s failed: %w", modelName, err)
	}

	if err := npm.LinkPackage(apiDir, mapping.pkg); err != nil {
		return fmt.Errorf("npm link %s failed: %w", mapping.pkg, err)
	}

	fmt.Printf("Linked: %s now uses local %s\n", name, modelName)
	return nil
}

func autoLinkToConsumers(wsPath string, ws *workspace.Workspace, name string) error {
	mapping, isModel := modelToAPI[name]
	if !isModel {
		return nil
	}

	apiRepo, exists := ws.Repos[mapping.api]
	if !exists {
		return nil
	}

	apiDir := filepath.Join(wsPath, apiRepo.Path)
	if _, err := os.Stat(apiDir); os.IsNotExist(err) {
		return nil
	}

	modelDir := filepath.Join(wsPath, ws.Repos[name].Path)
	buildDir := npm.BuildOutputDir(modelDir)

	if !npm.IsBuilt(modelDir) {
		return nil
	}

	if npm.IsLinked(apiDir, mapping.pkg) {
		return nil
	}

	fmt.Printf("Auto-linking to consumer %s...\n", mapping.api)

	if err := npm.Link(buildDir); err != nil {
		return fmt.Errorf("npm link failed: %w", err)
	}

	if err := npm.LinkPackage(apiDir, mapping.pkg); err != nil {
		return fmt.Errorf("npm link %s in %s failed: %w", mapping.pkg, mapping.api, err)
	}

	fmt.Printf("Linked: %s now uses local %s\n", mapping.api, name)
	return nil
}

func buildRecursively(wsPath string, ws *workspace.Workspace, target string) error {
	deps := getDependencies(ws, target)
	
	if len(deps) > 0 {
		fmt.Printf("Building dependencies first: %v\n\n", deps)
		for _, dep := range deps {
			repo, exists := ws.Repos[dep]
			if !exists {
				continue
			}

			repoDir := filepath.Join(wsPath, repo.Path)
			if _, err := os.Stat(repoDir); os.IsNotExist(err) {
				fmt.Printf("[skip] %s (not cloned)\n\n", dep)
				continue
			}

			if err := buildRepo(wsPath, ws, dep); err != nil {
				return fmt.Errorf("dependency build failed at '%s': %w", dep, err)
			}
			fmt.Println()
		}
	}

	return buildRepo(wsPath, ws, target)
}

func getDependencies(ws *workspace.Workspace, name string) []string {
	var deps []string
	seen := make(map[string]bool)

	var collect func(n string)
	collect = func(n string) {
		if seen[n] {
			return
		}
		seen[n] = true

		if modelName, isAPI := apiToModel[n]; isAPI {
			if _, exists := ws.Repos[modelName]; exists {
				collect(modelName)
				deps = append(deps, modelName)
			}
		}

		if repo, exists := ws.Repos[n]; exists {
			for _, dep := range repo.Dependencies {
				if _, depExists := ws.Repos[dep]; depExists {
					collect(dep)
					if !contains(deps, dep) {
						deps = append(deps, dep)
					}
				}
			}
		}
	}

	collect(name)

	seen[name] = false
	var result []string
	for _, d := range deps {
		if d != name {
			result = append(result, d)
		}
	}
	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func runShell(dir, command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func init() {
	buildCmd.Flags().BoolVarP(&buildRecursive, "recursive", "r", false, "Build dependencies first")
	buildCmd.Flags().BoolVar(&buildPublished, "published", false, "Force use of published packages (no local linking)")
	rootCmd.AddCommand(buildCmd)
}
