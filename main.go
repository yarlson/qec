package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"gihub.com/yarlson/qec/compose"
)

const helpText = `qec - Quantum Entanglement Communicator for Docker Compose

A tool that extends Docker Compose to handle multiple compose files with automatic context path adjustments.

Usage:
  qec [OPTIONS] COMMAND [ARGS...]

Options:
  -f, --file FILE        Path to a docker-compose YAML file (can be specified multiple times)
  -d, --detach          Run containers in the background
  --dry-run             Simulate configuration without making runtime changes
  --verbose             Enable verbose logging
  --command COMMAND     Command to execute (default: "up")

Commands:
  up                    Create and start containers
  down                  Stop and remove containers, networks, and volumes
  ps                    List containers
  logs                  View output from containers
  build                 Build or rebuild services
  pull                  Pull service images
  push                  Push service images
  config               Validate and view the merged configuration

Examples:
  # Run services from multiple compose files:
  qec -f folder1/docker-compose.yml -f folder2/docker-compose.yml up -d

  # View the merged configuration:
  qec -f folder1/docker-compose.yml -f folder2/docker-compose.yml --command config

  # Dry run to see what would happen:
  qec -f folder1/docker-compose.yml -f folder2/docker-compose.yml --dry-run up

For more information, visit: https://github.com/yarlson/qec
`

var (
	composeFiles multiFlag
	verbose      bool
	dryRun       bool
	detach       bool
	command      string
	showHelp     bool
	args         []string
)

// multiFlag is a custom flag type to handle multiple -f options
type multiFlag []string

func (m *multiFlag) String() string {
	return strings.Join(*m, ", ")
}

func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

// run executes the main program logic and returns an error if any
func run() error {
	if showHelp {
		fmt.Print(helpText)
		return nil
	}

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
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging for detailed output")
	flag.BoolVar(&dryRun, "dry-run", false, "Simulate configuration without making runtime changes")
	flag.BoolVar(&detach, "d", false, "Run containers in the background")
	flag.StringVar(&command, "command", "up", "Command to execute (up, down, config, ps, logs, build, pull, push)")
	flag.BoolVar(&showHelp, "help", false, "Show help text")
	flag.BoolVar(&showHelp, "h", false, "Show help text")

	// Set custom usage function
	flag.Usage = func() {
		fmt.Print(helpText)
	}

	flag.Parse()

	// Get any additional arguments after the flags
	args = flag.Args()

	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
