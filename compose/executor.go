package compose

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/sirupsen/logrus"
)

// Executor handles Docker Compose command execution with merged configurations
type Executor struct {
	project    *types.Project
	workingDir string
	dryRun     bool
}

// NewExecutor creates a new Docker Compose executor
func NewExecutor(project *types.Project, workingDir string, dryRun bool) *Executor {
	return &Executor{
		project:    project,
		workingDir: workingDir,
		dryRun:     dryRun,
	}
}

// writeConfig writes the merged configuration to a temporary file
func (e *Executor) writeConfig() (string, error) {
	logger := logrus.New().WithField("function", "writeConfig")

	// If this is a dry run, just return a placeholder path
	if e.dryRun {
		return filepath.Join(e.workingDir, "docker-compose.merged.yml"), nil
	}

	// Create a temporary file for the merged configuration
	configFile := filepath.Join(e.workingDir, "docker-compose.merged.yml")

	// Marshal the configuration to YAML
	yaml, err := e.project.MarshalYAML()
	if err != nil {
		return "", fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Write the configuration to the file
	err = os.WriteFile(configFile, []byte(yaml), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write configuration file: %w", err)
	}

	logger.Debugf("Wrote merged configuration to %s", configFile)
	return configFile, nil
}

// ExecuteCommand executes a Docker Compose command with the merged configuration
func (e *Executor) ExecuteCommand(cmdName string, args ...string) error {
	logger := logrus.New().WithField("function", "ExecuteCommand")

	// First check if Docker Compose is available
	if err := CheckDockerCompose(); err != nil {
		return fmt.Errorf("docker compose check failed: %w", err)
	}

	// Write the merged configuration to a file
	configFile, err := e.writeConfig()
	if err != nil {
		return err
	}

	// Create the Docker Compose command
	cmd, err := NewDockerComposeCmd()
	if err != nil {
		return fmt.Errorf("failed to create docker compose command: %w", err)
	}

	// Build the command arguments
	cmdArgs := []string{"-f", configFile, cmdName}
	cmdArgs = append(cmdArgs, args...)

	// Configure the command
	cmd.WithArgs(cmdArgs...).WithWorkingDir(e.workingDir)

	// If this is a dry run, just log what would be done
	if e.dryRun {
		logger.Info("Dry run: would execute docker compose " + cmdName)
		logger.Debugf("Command: %s %v", cmd.Executable, cmd.Args)
		return nil
	}

	// Run the command
	output, err := cmd.Run()
	if err != nil {
		return fmt.Errorf("docker compose %s failed: %w\nOutput: %s", cmdName, err, output.Output)
	}

	// For certain commands, we want to print the output
	switch cmdName {
	case "ps", "logs", "config":
		fmt.Print(output.Output)
	}

	return nil
}
