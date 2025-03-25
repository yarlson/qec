package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"gihub.com/yarlson/qec/compose"
)

var (
	composeFiles multiFlag
	verbose      bool
	dryRun       bool
	detach       bool
	command      string
	args         []string
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

	// Add command-specific arguments
	if command == "up" {
		args = append([]string{"--remove-orphans"}, args...)
		if detach {
			args = append(args, "-d")
		}
	} else if command == "down" {
		args = append([]string{"--remove-orphans"}, args...)
	}

	// Execute the command
	if err := executor.ExecuteCommand(command, args...); err != nil {
		return fmt.Errorf("error executing %s command: %v", command, err)
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
	flag.StringVar(&command, "command", "up", "Command to execute (up, down, config, ps, logs, build, pull, push)")
	flag.Parse()

	// Get any additional arguments after the flags
	args = flag.Args()

	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
