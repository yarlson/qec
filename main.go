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

// multiFlag is a custom flag type to handle multiple -f options
type multiFlag []string

func (m *multiFlag) String() string {
	return fmt.Sprint(*m)
}

func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

func main() {
	// Register the -f flag to accept multiple entries
	flag.Var(&composeFiles, "f", "Path to a docker-compose YAML file (can be specified multiple times)")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.Parse()

	if len(composeFiles) == 0 {
		fmt.Println("No compose files specified. Use -f flag to specify compose files.")
		os.Exit(1)
	}

	// Configure logging
	logger := logrus.New()
	if verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	// Load and process each compose file
	var files []*compose.ComposeFile
	for _, file := range composeFiles {
		cf, err := compose.NewComposeFile(file, logger.WithField("component", "loader"))
		if err != nil {
			logger.Fatalf("Error loading compose file %s: %v", file, err)
		}
		files = append(files, cf)
	}

	// Merge the compose files
	merged, err := compose.MergeComposeFiles(files)
	if err != nil {
		logger.Fatalf("Error merging compose files: %v", err)
	}

	// Output the merged configuration
	yaml, err := merged.MarshalYAML()
	if err != nil {
		logger.Fatalf("Error marshaling merged configuration: %v", err)
	}

	fmt.Println(string(yaml))
}
