package main

import (
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func loadConfig(configFile string) (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Use defaults
		return &Config{
			Name:    "mytool",
			Entry:   "scripts/main.sh",
			Scripts: "scripts",
			Version: "1.0.0",
		}, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.Name == "" {
		config.Name = "mytool"
	}
	if config.Entry == "" {
		config.Entry = "scripts/main.sh"
	}
	if config.Scripts == "" {
		config.Scripts = "scripts"
	}
	if config.Version == "" {
		config.Version = "1.0.0"
	}

	return &config, nil
}

func discoverScripts(config *Config) ([]string, error) {

	scriptMap := make(map[string]bool)

	base := config.Scripts

	filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		matched, _ := filepath.Match("*.sh", filepath.Base(path))

		if !d.IsDir() && matched {
			scriptMap[path] = true
		}
		return nil
	})

	// Ensure entry script is included
	if config.Entry != "" {
		scriptMap[config.Entry] = true
	}

	scripts := make([]string, 0, len(scriptMap))
	for script := range scriptMap {
		scripts = append(scripts, script)
	}

	if len(scripts) == 0 {
		return nil, fmt.Errorf("no scripts found")
	}

	//fmt.Println(scripts)

	return scripts, nil
}

func generateMainGo(outputPath string, config *Config, scripts []string) error {
	funcMap := template.FuncMap{
		"sanitize": func(s string) string {
			s = strings.ReplaceAll(s, "/", "_")
			s = strings.ReplaceAll(s, ".", "_")
			s = strings.ReplaceAll(s, "-", "_")
			return s
		},
		"base": filepath.Base,
	}

	tmpl, err := template.New("runtime").Funcs(funcMap).Parse(runtimeTemplate)
	if err != nil {
		return err
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Convert scripts to ScriptInfo with stripped paths
	baseDir := config.Scripts
	scriptInfos := make([]ScriptInfo, len(scripts))
	for i, script := range scripts {
		relPath, _ := filepath.Rel(baseDir, script)
		scriptInfos[i] = ScriptInfo{Path: relPath}
	}

	// Also strip from MainScript
	mainScriptRel, _ := filepath.Rel(baseDir, config.Entry)

	data := map[string]interface{}{
		"Name":       config.Name,
		"Version":    config.Version,
		"MainScript": mainScriptRel, // e.g. "main.sh" instead of "scripts/main.sh"
		"Scripts":    scriptInfos,
	}

	return tmpl.Execute(f, data)
}

func initGoModule(buildDir, moduleName string) error {
	cmd := exec.Command("go", "mod", "init", moduleName)
	cmd.Dir = buildDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod init failed: %w\n%s", err, output)
	}
	return nil
}

func buildExecutable(buildDir, outputPath string) error {
	absOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return err
	}

	cmd := exec.Command("go", "build", "-o", absOutputPath)
	cmd.Dir = buildDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	return nil
}

func randomVersion() string {
	now := time.Now()
	year, day := now.Year(), now.YearDay()
	secondOfDay := now.Hour()*3600 + now.Minute()*60 + now.Second()
	return fmt.Sprintf("%d.%03d.%05d", year, day, secondOfDay)
}
