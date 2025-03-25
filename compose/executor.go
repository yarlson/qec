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
	logger     *logrus.Entry
	project    *types.Project
	workingDir string
	dryRun     bool
}

// NewExecutor creates a new Docker Compose executor
func NewExecutor(project *types.Project, workingDir string, dryRun bool, logger *logrus.Entry) *Executor {
	return &Executor{
		logger:     logger.WithField("component", "executor"),
		project:    project,
		workingDir: workingDir,
		dryRun:     dryRun,
	}
}

// writeConfig writes the merged configuration to a temporary file
func (e *Executor) writeConfig() (string, error) {
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

	e.logger.Debugf("Wrote merged configuration to %s", configFile)
	return configFile, nil
}

// Up starts the services defined in the configuration
func (e *Executor) Up(detach bool) error {
	// First check if Docker Compose is available
	if err := CheckDockerCompose(e.logger); err != nil {
		return fmt.Errorf("docker compose check failed: %w", err)
	}

	// Write the merged configuration to a file
	configFile, err := e.writeConfig()
	if err != nil {
		return err
	}

	// Create the Docker Compose command
	cmd, err := NewDockerComposeCmd(e.logger)
	if err != nil {
		return fmt.Errorf("failed to create docker compose command: %w", err)
	}

	// Build the command arguments
	args := []string{"-f", configFile, "up", "--remove-orphans"}
	if detach {
		args = append(args, "-d")
	}

	// Configure the command
	cmd.WithArgs(args...).WithWorkingDir(e.workingDir)

	// If this is a dry run, just log what would be done
	if e.dryRun {
		e.logger.Info("Dry run: would execute docker compose up")
		e.logger.Debugf("Command: %s %v", cmd.Executable, cmd.Args)
		return nil
	}

	// Run the command
	output, err := cmd.Run(e.logger)
	if err != nil {
		return fmt.Errorf("docker compose up failed: %w\nOutput: %s", err, output.Output)
	}

	return nil
}

// Down stops and removes the services defined in the configuration
func (e *Executor) Down() error {
	// First check if Docker Compose is available
	if err := CheckDockerCompose(e.logger); err != nil {
		return fmt.Errorf("docker compose check failed: %w", err)
	}

	// Write the merged configuration to a file
	configFile, err := e.writeConfig()
	if err != nil {
		return err
	}

	// Create the Docker Compose command
	cmd, err := NewDockerComposeCmd(e.logger)
	if err != nil {
		return fmt.Errorf("failed to create docker compose command: %w", err)
	}

	// Configure the command
	cmd.WithArgs("-f", configFile, "down", "--remove-orphans").WithWorkingDir(e.workingDir)

	// If this is a dry run, just log what would be done
	if e.dryRun {
		e.logger.Info("Dry run: would execute docker compose down")
		e.logger.Debugf("Command: %s %v", cmd.Executable, cmd.Args)
		return nil
	}

	// Run the command
	output, err := cmd.Run(e.logger)
	if err != nil {
		return fmt.Errorf("docker compose down failed: %w\nOutput: %s", err, output.Output)
	}

	return nil
}

// Config validates and shows the merged configuration
func (e *Executor) Config() error {
	// First check if Docker Compose is available
	if err := CheckDockerCompose(e.logger); err != nil {
		return fmt.Errorf("docker compose check failed: %w", err)
	}

	// Write the merged configuration to a file
	configFile, err := e.writeConfig()
	if err != nil {
		return err
	}

	// Create the Docker Compose command
	cmd, err := NewDockerComposeCmd(e.logger)
	if err != nil {
		return fmt.Errorf("failed to create docker compose command: %w", err)
	}

	// Configure the command
	cmd.WithArgs("-f", configFile, "config").WithWorkingDir(e.workingDir)

	// If this is a dry run, just log what would be done
	if e.dryRun {
		e.logger.Info("Dry run: would execute docker compose config")
		e.logger.Debugf("Command: %s %v", cmd.Executable, cmd.Args)
		return nil
	}

	// Run the command
	output, err := cmd.Run(e.logger)
	if err != nil {
		return fmt.Errorf("docker compose config failed: %w\nOutput: %s", err, output.Output)
	}

	// Print the configuration
	fmt.Println(output.Output)
	return nil
}
