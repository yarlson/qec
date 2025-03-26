package compose

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

// CommandOutput represents the output from a Docker Compose command
type CommandOutput struct {
	ExitCode int    // Exit code from the command
	Output   string // Combined stdout and stderr output
}

// DockerComposeCmd represents a Docker Compose command configuration
type DockerComposeCmd struct {
	Executable string   // Path to docker-compose or docker executable
	IsPlugin   bool     // Whether we're using the docker compose plugin
	Args       []string // Command arguments
	WorkingDir string   // Working directory for the command
}

// NewDockerComposeCmd creates a new Docker Compose command configuration
func NewDockerComposeCmd() (*DockerComposeCmd, error) {
	logger := logrus.New().WithField("function", "NewDockerComposeCmd")

	// Find the docker-compose executable
	path, err := exec.LookPath("docker-compose")
	if err != nil {
		// If docker-compose is not found, try docker compose (newer versions)
		path, err = exec.LookPath("docker")
		if err != nil {
			return nil, fmt.Errorf("neither docker-compose nor docker executable found in PATH")
		}
	}

	// Determine if we're using the plugin
	isPlugin := !strings.HasSuffix(path, "docker-compose")
	if isPlugin {
		logger.Debugf("Using Docker Compose plugin at %s", path)
	} else {
		logger.Debugf("Using standalone Docker Compose at %s", path)
	}

	return &DockerComposeCmd{
		Executable: path,
		IsPlugin:   isPlugin,
		Args:       make([]string, 0),
	}, nil
}

// WithArgs adds arguments to the command
func (cmd *DockerComposeCmd) WithArgs(args ...string) *DockerComposeCmd {
	cmd.Args = append(cmd.Args, args...)
	return cmd
}

// WithWorkingDir sets the working directory for the command
func (cmd *DockerComposeCmd) WithWorkingDir(dir string) *DockerComposeCmd {
	cmd.WorkingDir = dir
	return cmd
}

// Build constructs and returns the final exec.Cmd
func (cmd *DockerComposeCmd) Build() *exec.Cmd {
	logger := logrus.New().WithField("function", "Build")

	// Prepare the command arguments
	var finalArgs []string
	if cmd.IsPlugin {
		// For docker compose plugin, prepend "compose"
		finalArgs = append([]string{"compose"}, cmd.Args...)
	} else {
		finalArgs = cmd.Args
	}

	// Create the command
	command := exec.Command(cmd.Executable, finalArgs...)

	// Set working directory if specified
	if cmd.WorkingDir != "" {
		command.Dir = cmd.WorkingDir
	}

	// Log the command being executed
	logger.Debugf("Executing command: %s %s", cmd.Executable, strings.Join(finalArgs, " "))

	return command
}

// Run executes the Docker Compose command and returns its output
func (cmd *DockerComposeCmd) Run() (*CommandOutput, error) {
	logger := logrus.New().WithField("function", "Run")

	// Build the command
	execCmd := cmd.Build()

	// Create buffers for stdout and stderr
	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	// Run the command
	err := execCmd.Run()

	// Combine stdout and stderr
	var output bytes.Buffer
	_, _ = io.Copy(&output, &stdout)
	_, _ = io.Copy(&output, &stderr)

	// Create the command output
	cmdOutput := &CommandOutput{
		ExitCode: 0,
		Output:   output.String(),
	}

	// Handle error and exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			cmdOutput.ExitCode = exitErr.ExitCode()
			logger.WithFields(logrus.Fields{
				"exit_code": cmdOutput.ExitCode,
				"error":     err,
			}).Debug("Command failed")
			return cmdOutput, fmt.Errorf("command failed with exit code %d: %w", cmdOutput.ExitCode, err)
		}
		logger.WithError(err).Debug("Command failed to execute")
		return cmdOutput, fmt.Errorf("failed to execute command: %w", err)
	}

	logger.WithField("output", cmdOutput.Output).Debug("Command completed successfully")
	return cmdOutput, nil
}

// RunBackground executes the Docker Compose command in the background
func (cmd *DockerComposeCmd) RunBackground() error {
	logger := logrus.New().WithField("function", "RunBackground")

	// Build the command
	execCmd := cmd.Build()

	// Start the command without waiting for it to complete
	if err := execCmd.Start(); err != nil {
		logger.WithError(err).Debug("Failed to start background command")
		return fmt.Errorf("failed to start background command: %w", err)
	}

	logger.Debug("Command started in background")
	return nil
}

// CheckDockerCompose verifies that Docker Compose is installed and accessible
func CheckDockerCompose() error {
	logger := logrus.New().WithField("function", "CheckDockerCompose")

	// First try to find the docker-compose executable
	path, err := exec.LookPath("docker-compose")
	if err != nil {
		// If docker-compose is not found, try docker compose (newer versions)
		path, err = exec.LookPath("docker")
		if err != nil {
			return fmt.Errorf("neither docker-compose nor docker executable found in PATH")
		}
	}

	// Check if we found docker-compose or docker
	if strings.HasSuffix(path, "docker-compose") {
		logger.Debugf("Found docker-compose at %s", path)
		return checkDockerComposeVersion(path)
	} else {
		logger.Debugf("Found docker at %s, checking compose plugin", path)
		return checkDockerComposePlugin(path)
	}
}

// checkDockerComposeVersion checks the version of the docker-compose executable
func checkDockerComposeVersion(path string) error {
	logger := logrus.New().WithField("function", "checkDockerComposeVersion")

	cmd := exec.Command(path, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to check docker-compose version: %w", err)
	}

	version := strings.TrimSpace(string(output))
	logger.Debugf("Docker Compose version: %s", version)

	// TODO: Add version parsing and minimum version check if needed
	return nil
}

// checkDockerComposePlugin checks if the docker compose plugin is available
func checkDockerComposePlugin(path string) error {
	logger := logrus.New().WithField("function", "checkDockerComposePlugin")

	cmd := exec.Command(path, "compose", "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose plugin not found or not working: %w", err)
	}

	version := strings.TrimSpace(string(output))
	logger.Debugf("Docker Compose plugin version: %s", version)

	// TODO: Add version parsing and minimum version check if needed
	return nil
}
