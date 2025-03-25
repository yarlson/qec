package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var composeFiles multiFlag

// multiFlag is a custom flag type to handle multiple -f options
type multiFlag []string

func (m *multiFlag) String() string {
	return fmt.Sprint(*m)
}

func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

// FileInfo holds information about a compose file
type FileInfo struct {
	AbsPath string
	BaseDir string
	Content map[string]interface{}
}

// validateAndResolve checks if a file exists and determines its absolute path and base directory
func validateAndResolve(filePath string) (*FileInfo, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %v", filePath, err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", absPath)
	}

	baseDir := filepath.Dir(absPath)
	return &FileInfo{
		AbsPath: absPath,
		BaseDir: baseDir,
	}, nil
}

// loadYAML loads and parses a YAML file into a generic map structure
func loadYAML(filePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file %s: %v", filePath, err)
	}

	var content map[string]interface{}
	if err := yaml.Unmarshal(data, &content); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML content from %s: %v", filePath, err)
	}

	return content, nil
}

func main() {
	// Register the -f flag to accept multiple entries
	flag.Var(&composeFiles, "f", "Path to a docker-compose YAML file (can be specified multiple times)")
	flag.Parse()

	if len(composeFiles) == 0 {
		fmt.Println("No compose files specified. Use -f flag to specify compose files.")
		os.Exit(1)
	}

	// Process each file: validate, get base dir, and load YAML
	for _, file := range composeFiles {
		info, err := validateAndResolve(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		content, err := loadYAML(info.AbsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading YAML: %v\n", err)
			os.Exit(1)
		}

		info.Content = content
		fmt.Printf("Processing File: %s\nBase Directory: %s\nYAML Content: %+v\n\n",
			info.AbsPath, info.BaseDir, info.Content)
	}
}
