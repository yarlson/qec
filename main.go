package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gihub.com/yarlson/qec/compose"
	"github.com/sirupsen/logrus"
)

var (
	composeFiles multiFlag
	verbose      bool
	dryRun       bool
	detach       bool
	command      string
)

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

	// Create an executor with the merged configuration
	workingDir := filepath.Dir(composeFiles[0])
	executor := compose.NewExecutor(merged, workingDir, dryRun, logger.WithField("component", "executor"))

	// Execute the requested command
	switch command {
	case "up":
		if err := executor.Up(detach); err != nil {
			return fmt.Errorf("error executing up command: %v", err)
		}
	case "down":
		if err := executor.Down(); err != nil {
			return fmt.Errorf("error executing down command: %v", err)
		}
	case "config":
		if err := executor.Config(); err != nil {
			return fmt.Errorf("error executing config command: %v", err)
		}
	default:
		return fmt.Errorf("unknown command: %s", command)
	}

	logger.Info("Command executed successfully")
	return nil
}

func main() {
	// Register flags
	flag.Var(&composeFiles, "f", "Path to a docker-compose YAML file (can be specified multiple times)")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&dryRun, "dry-run", false, "Simulate configuration without making runtime changes")
	flag.BoolVar(&detach, "d", false, "Run containers in the background")
	flag.StringVar(&command, "command", "up", "Command to execute (up, down, or config)")
	flag.Parse()

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
