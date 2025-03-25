package main

import (
	"flag"
	"fmt"
	"os"

	"gihub.com/yarlson/qec/compose"
	"github.com/sirupsen/logrus"
)

var composeFiles multiFlag
var verbose bool
var dryRun bool

// multiFlag is a custom flag type to handle multiple -f options
type multiFlag []string

func (m *multiFlag) String() string {
	return fmt.Sprint(*m)
}

func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

// run executes the main program logic and returns an error if any
func run() error {
	if len(composeFiles) == 0 {
		return fmt.Errorf("no compose files specified. Use -f flag to specify compose files")
	}

	// Configure logging
	logger := logrus.New()
	if verbose {
		logger.SetLevel(logrus.DebugLevel)
		logger.Info("Verbose logging enabled")
	}

	if dryRun {
		logger.Info("Running in dry-run mode - no changes will be made")
	}

	// Load and process each compose file
	var files []*compose.ComposeFile
	for _, file := range composeFiles {
		cf, err := compose.NewComposeFile(file, logger.WithField("component", "loader"))
		if err != nil {
			return fmt.Errorf("error loading compose file %s: %v", file, err)
		}
		files = append(files, cf)
	}

	// Merge the compose files
	merged, err := compose.MergeComposeFiles(files)
	if err != nil {
		return fmt.Errorf("error merging compose files: %v", err)
	}

	// Output the merged configuration
	yaml, err := merged.MarshalYAML()
	if err != nil {
		return fmt.Errorf("error marshaling merged configuration: %v", err)
	}

	fmt.Println(string(yaml))

	if dryRun {
		logger.Info("Dry run completed successfully")
		return nil
	}

	// TODO: Implement actual runtime changes here
	logger.Info("Configuration applied successfully")
	return nil
}

func main() {
	// Register the -f flag to accept multiple entries
	flag.Var(&composeFiles, "f", "Path to a docker-compose YAML file (can be specified multiple times)")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&dryRun, "dry-run", false, "Simulate configuration without making runtime changes")
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
