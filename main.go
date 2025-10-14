package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	// Add this
)

type Config struct {
	Name    string `yaml:"name"`
	Entry   string `yaml:"entry"`
	Scripts string `yaml:"scripts"`
	Version string `yaml:"version"`
}

type ScriptInfo struct {
	Path string
}

const runtimeTemplate = `package main

import (
	_ "embed"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

{{range .Scripts}}
//go:embed {{.Path}}
var script_{{sanitize .Path}} []byte
{{end}}

var scripts = map[string][]byte{
{{range .Scripts}}	"{{.Path}}": script_{{sanitize .Path}},
{{end}}
}

func main() {
	// Get cache directory
	cacheDir, err := getCacheDir("{{.Name}}", "{{.Version}}")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting cache directory: %v\n", err)
		os.Exit(1)
	}
	
	// Extract scripts if needed
	if needsExtraction(cacheDir, scripts) {
		if err := extractScripts(cacheDir, scripts); err != nil {
			fmt.Fprintf(os.Stderr, "Error extracting scripts: %v\n", err)
			os.Exit(1)
		}
	}

	// Execute main.sh
	mainScript := filepath.Join(cacheDir, "{{.MainScript}}")
	cmd := exec.Command(mainScript, os.Args[1:]...)
	cmd.Dir = cacheDir

	// Set up environment
	cmd.Env = append(os.Environ(),
		"SHPACK_SCRIPT_DIR="+cacheDir,
		"SHPACK_VERSION={{.Version}}",
	)
	
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Run and get exit code
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "Error executing script: %v\n", err)
		os.Exit(1)
	}
}

func getCacheDir(name, version string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	h := sha256.New()
	h.Write([]byte(version))
	versionHash := fmt.Sprintf("%x", h.Sum(nil))[:8]
	
	cacheDir := filepath.Join(homeDir, ".cache", name, versionHash)
	return cacheDir, nil
}

func needsExtraction(cacheDir string, scripts map[string][]byte) bool {
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return true
	}
	
	for relPath := range scripts {
		scriptPath := filepath.Join(cacheDir, relPath)
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			return true
		}
	}
	
	return false
}

func extractScripts(cacheDir string, scripts map[string][]byte) error {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}
	
	for relPath, content := range scripts {
		path := filepath.Join(cacheDir, relPath)
		
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		
		if err := os.WriteFile(path, content, 0755); err != nil {
			return err
		}
	}
	
	return nil
}
`

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	srcDir := "."
	if len(os.Args) == 3 {
		srcDir = os.Args[2]
	}

	switch os.Args[1] {
	case "build":
		if err := buildCommand(srcDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "make":
		if err := makeCommand(srcDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "init":
		if err := initCommand(srcDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "version":
		fmt.Println("shpack version 1.0.2")
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`shpack - Shell Script Bundler

Usage:
  shpack build [source_dir]  Build executable from scripts
  shpack make <folder>       Quick build from existing script folder
  shpack init [source_dir]   Initialize a new project
  shpack version             Show version
  
Examples:
  shpack build              # Build from current directory
  shpack make ./myscripts   # Quick build from folder (auto-setup)
  shpack init ./newproject  # Initialize new project`)
}

func buildCommand(srcDir string) error {
	// Parse flags
	configFile := "shpack.yaml"
	outputPath := ""

	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	if srcDir != "." {
		absPath, err := filepath.Abs(srcDir)
		if err != nil {
			return fmt.Errorf("invalid source directory: %w", err)
		}
		if err := os.Chdir(absPath); err != nil {
			return fmt.Errorf("failed to change to source directory: %w", err)
		}
		defer os.Chdir(originalDir)
		fmt.Printf("Building from: %s\n", absPath)
	}

	// Load config
	config, err := loadConfig(configFile)
	if err != nil {
		return err
	}

	// Discover scripts
	scripts, err := discoverScripts(config)
	if err != nil {
		return err
	}

	// Validate main script exists
	mainScriptPath := config.Entry
	if mainScriptPath == "" {
		mainScriptPath = "scripts/main.sh"
	}

	found := slices.Contains(scripts, mainScriptPath)
	if !found {
		return fmt.Errorf("entry script not found: %s", mainScriptPath)
	}

	// Set default output path
	if outputPath == "" {
		outputPath = filepath.Join("build", config.Name)
	} else if !filepath.IsAbs(outputPath) {
		// If relative output path, resolve from original directory
		outputPath = filepath.Join(originalDir, outputPath)
	}

	// Create build directory
	buildDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}

	// Create temporary build directory for generated code
	tmpBuildDir, err := os.MkdirTemp("", "shpack-build-*")
	if err != nil {
		return fmt.Errorf("failed to create temp build directory: %w", err)
	}
	defer os.RemoveAll(tmpBuildDir)

	// Copy scripts to temp build directory
	baseDir := config.Scripts

	for _, script := range scripts {
		content, err := os.ReadFile(script)
		if err != nil {
			return fmt.Errorf("failed to read script %s: %w", script, err)
		}

		// Strip the base directory from the path
		relPath, err := filepath.Rel(baseDir, script)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", script, err)
		}

		destPath := filepath.Join(tmpBuildDir, relPath)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create subdirectory for %s: %w", script, err)
		}

		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return fmt.Errorf("failed to copy script %s: %w", script, err)
		}
	}

	// Generate main.go
	mainGoPath := filepath.Join(tmpBuildDir, "main.go")
	if err := generateMainGo(mainGoPath, config, scripts); err != nil {
		return fmt.Errorf("failed to generate main.go: %w", err)
	}

	// Initialize go module
	if err := initGoModule(tmpBuildDir, config.Name); err != nil {
		return fmt.Errorf("failed to initialize go module: %w", err)
	}

	// Build executable
	fmt.Printf("Building %s...\n", config.Name)
	if err := buildExecutable(tmpBuildDir, outputPath); err != nil {
		return fmt.Errorf("failed to build executable: %w", err)
	}

	fmt.Printf("✓ Built successfully: %s\n", outputPath)
	return nil
}

func initCommand(srcDir string) error {

	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", srcDir, err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	if srcDir != "." {
		absPath, err := filepath.Abs(srcDir)
		if err != nil {
			return fmt.Errorf("invalid source directory: %w", err)
		}
		if err := os.Chdir(absPath); err != nil {
			return fmt.Errorf("failed to change to source directory: %w", err)
		}
		defer os.Chdir(originalDir)
		fmt.Printf("Building from: %s\n", absPath)
	}

	// Create directory structure
	dirs := []string{"scripts", "build"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create sample config
	configContent := `name: mytool
entry: scripts/main.sh
scripts: scripts
version: 1.0.0
`
	if err := os.WriteFile("shpack.yaml", []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create shpack.yaml: %w", err)
	}

	// Create sample main.sh
	mainShContent := `#!/bin/bash
# Main entry point script
`
	if err := os.WriteFile("scripts/main.sh", []byte(mainShContent), 0755); err != nil {
		return fmt.Errorf("failed to create main.sh: %w", err)
	}

	fmt.Println("✓ Initialized shpack project")
	return nil
}

func makeCommand(srcDir string) error {

	// Validate source directory exists
	absPath, err := filepath.Abs(srcDir)
	if err != nil {
		return fmt.Errorf("invalid source directory: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("folder does not exist: %s", absPath)
	}

	// Get folder name for tool name
	toolName := filepath.Base(absPath)

	// Create temporary build directory
	tmpDir, err := os.MkdirTemp("", "shpack-make-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	fmt.Printf("Making %s from: %s\n", toolName, absPath)

	// Create scripts subdirectory in temp
	scriptsDir := filepath.Join(tmpDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		return fmt.Errorf("failed to create scripts directory: %w", err)
	}

	// Copy all .sh files from source to temp/scripts
	var mainScriptFound bool
	err = filepath.WalkDir(absPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		matched, _ := filepath.Match("*.sh", filepath.Base(path))
		if !matched {
			return nil
		}

		// Get relative path from source
		relPath, err := filepath.Rel(absPath, path)
		if err != nil {
			return err
		}

		// Check if this is main.sh
		if filepath.Base(path) == "main.sh" {
			mainScriptFound = true
		}

		// Copy to temp/scripts/
		destPath := filepath.Join(scriptsDir, relPath)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(destPath, content, 0755)
	})
	if err != nil {
		return fmt.Errorf("failed to copy scripts: %w", err)
	}

	if !mainScriptFound {
		return fmt.Errorf("no main.sh found in %s", absPath)
	}

	// Generate random version for fresh build
	randomVersion := randomVersion()

	// Create shpack.yaml in temp directory
	configContent := fmt.Sprintf(`name: %s
entry: scripts/main.sh
scripts: scripts
version: %s
`, toolName, randomVersion)

	if err := os.WriteFile(filepath.Join(tmpDir, "shpack.yaml"), []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	// Build from temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		return fmt.Errorf("failed to change to temp directory: %w", err)
	}
	defer os.Chdir(originalDir)

	// Load config and build
	config, err := loadConfig("shpack.yaml")
	if err != nil {
		return err
	}

	scripts, err := discoverScripts(config)
	if err != nil {
		return err
	}

	// Output to original directory
	outputPath := filepath.Join(originalDir, toolName)

	// Create build directory for compilation
	buildTmpDir, err := os.MkdirTemp("", "shpack-build-*")
	if err != nil {
		return fmt.Errorf("failed to create build temp directory: %w", err)
	}
	defer os.RemoveAll(buildTmpDir)

	// Copy scripts with stripped paths
	baseDir := config.Scripts
	for _, script := range scripts {
		content, err := os.ReadFile(script)
		if err != nil {
			return fmt.Errorf("failed to read script %s: %w", script, err)
		}

		relPath, err := filepath.Rel(baseDir, script)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", script, err)
		}

		destPath := filepath.Join(buildTmpDir, relPath)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create subdirectory: %w", err)
		}

		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return fmt.Errorf("failed to copy script: %w", err)
		}
	}

	// Generate main.go
	mainGoPath := filepath.Join(buildTmpDir, "main.go")
	if err := generateMainGo(mainGoPath, config, scripts); err != nil {
		return fmt.Errorf("failed to generate main.go: %w", err)
	}

	// Initialize go module and build
	if err := initGoModule(buildTmpDir, config.Name); err != nil {
		return fmt.Errorf("failed to initialize go module: %w", err)
	}

	fmt.Printf("Building %s...\n", config.Name)
	if err := buildExecutable(buildTmpDir, outputPath); err != nil {
		return fmt.Errorf("failed to build executable: %w", err)
	}

	fmt.Printf("✓ Built successfully: %s\n", outputPath)
	fmt.Printf("  (version: %s - fresh build, no cache)\n", randomVersion)

	return nil
}
