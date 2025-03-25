package compose

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

// DockerComposeCmd represents a Docker Compose command configuration
type DockerComposeCmd struct {
	Executable string   // Path to docker-compose or docker executable
	IsPlugin   bool     // Whether we're using the docker compose plugin
	Args       []string // Command arguments
	WorkingDir string   // Working directory for the command
}

// NewDockerComposeCmd creates a new Docker Compose command configuration
func NewDockerComposeCmd(logger *logrus.Entry) (*DockerComposeCmd, error) {
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
func (cmd *DockerComposeCmd) Build(logger *logrus.Entry) *exec.Cmd {
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

// CheckDockerCompose verifies that Docker Compose is installed and accessible
func CheckDockerCompose(logger *logrus.Entry) error {
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
		return checkDockerComposeVersion(path, logger)
	} else {
		logger.Debugf("Found docker at %s, checking compose plugin", path)
		return checkDockerComposePlugin(path, logger)
	}
}

// checkDockerComposeVersion checks the version of the docker-compose executable
func checkDockerComposeVersion(path string, logger *logrus.Entry) error {
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
func checkDockerComposePlugin(path string, logger *logrus.Entry) error {
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
