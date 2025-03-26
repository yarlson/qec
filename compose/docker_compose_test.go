package compose

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// DockerComposeTestSuite defines the test suite for Docker Compose integration
type DockerComposeTestSuite struct {
	suite.Suite
	tmpDir string
}

// SetupTest runs before each test
func (suite *DockerComposeTestSuite) SetupTest() {
	suite.tmpDir = suite.T().TempDir()
}

// TestNewDockerComposeCmd tests the creation of a new Docker Compose command
func (suite *DockerComposeTestSuite) TestNewDockerComposeCmd() {
	cmd, err := NewDockerComposeCmd()
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), cmd.Executable)
	assert.NotNil(suite.T(), cmd.Args)
	assert.Empty(suite.T(), cmd.Args)
}

// TestDockerComposeCmdWithArgs tests adding arguments to the command
func (suite *DockerComposeTestSuite) TestDockerComposeCmdWithArgs() {
	cmd, err := NewDockerComposeCmd()
	require.NoError(suite.T(), err)

	// Add arguments
	cmd.WithArgs("up", "-d", "--build")
	assert.Equal(suite.T(), []string{"up", "-d", "--build"}, cmd.Args)

	// Add more arguments
	cmd.WithArgs("--remove-orphans")
	assert.Equal(suite.T(), []string{"up", "-d", "--build", "--remove-orphans"}, cmd.Args)
}

// TestDockerComposeCmdWithWorkingDir tests setting the working directory
func (suite *DockerComposeTestSuite) TestDockerComposeCmdWithWorkingDir() {
	cmd, err := NewDockerComposeCmd()
	require.NoError(suite.T(), err)

	// Set working directory
	workingDir := "/test/dir"
	cmd.WithWorkingDir(workingDir)
	assert.Equal(suite.T(), workingDir, cmd.WorkingDir)
}

// TestDockerComposeCmdBuild tests building the final command
func (suite *DockerComposeTestSuite) TestDockerComposeCmdBuild() {
	cmd, err := NewDockerComposeCmd()
	require.NoError(suite.T(), err)

	// Configure the command
	cmd.WithArgs("up", "-d")
	cmd.WithWorkingDir("/test/dir")

	// Build the command
	execCmd := cmd.Build()

	// Verify the command
	assert.Equal(suite.T(), cmd.Executable, execCmd.Path)
	assert.Equal(suite.T(), "/test/dir", execCmd.Dir)

	// Verify arguments based on whether we're using the plugin
	if cmd.IsPlugin {
		assert.Equal(suite.T(), []string{"compose", "up", "-d"}, execCmd.Args[1:])
	} else {
		assert.Equal(suite.T(), []string{"up", "-d"}, execCmd.Args[1:])
	}
}

// TestDockerComposeCmdRun tests command execution
func (suite *DockerComposeTestSuite) TestDockerComposeCmdRun() {
	cmd, err := NewDockerComposeCmd()
	require.NoError(suite.T(), err)

	// Test successful command (version)
	cmd.Args = []string{"version"}
	output, err := cmd.Run()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, output.ExitCode)
	assert.NotEmpty(suite.T(), output.Output)
	assert.Contains(suite.T(), strings.ToLower(output.Output), "version")

	// Test failed command
	cmd.Args = []string{"non-existent-command"}
	output, err = cmd.Run()
	assert.Error(suite.T(), err)
	assert.NotEqual(suite.T(), 0, output.ExitCode)
	assert.NotEmpty(suite.T(), output.Output)
}

// TestDockerComposeCmdRunBackground tests background command execution
func (suite *DockerComposeTestSuite) TestDockerComposeCmdRunBackground() {
	// Create a test compose file
	composeFile := filepath.Join(suite.tmpDir, "docker-compose.yml")
	content := []byte(`
version: '3'
services:
  test:
    image: hello-world
`)
	err := os.WriteFile(composeFile, content, 0644)
	require.NoError(suite.T(), err)

	// Create and configure the command
	cmd, err := NewDockerComposeCmd()
	require.NoError(suite.T(), err)

	cmd.WithWorkingDir(suite.tmpDir)
	cmd.WithArgs("-f", "docker-compose.yml", "up", "-d")

	// Run in background
	err = cmd.RunBackground()
	assert.NoError(suite.T(), err)

	// Test with invalid working directory
	cmd.WithWorkingDir("/nonexistent")
	err = cmd.RunBackground()
	assert.Error(suite.T(), err)
}

// TestCheckDockerCompose tests the Docker Compose detection functionality
func (suite *DockerComposeTestSuite) TestCheckDockerCompose() {
	// Test with real Docker Compose installation
	err := CheckDockerCompose()
	require.NoError(suite.T(), err, "Docker Compose should be available in the test environment")

	// Test with missing Docker Compose by temporarily modifying PATH
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	// Create a temporary directory with no docker-compose
	tmpPath := filepath.Join(suite.tmpDir, "bin")
	err = os.MkdirAll(tmpPath, 0755)
	require.NoError(suite.T(), err)

	// Set PATH to only include our empty directory
	os.Setenv("PATH", tmpPath)

	// Now check should fail
	err = CheckDockerCompose()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "neither docker-compose nor docker executable found in PATH")
}

// TestCheckDockerComposeVersion tests version checking functionality
func (suite *DockerComposeTestSuite) TestCheckDockerComposeVersion() {
	// Find the real docker-compose executable
	path, err := exec.LookPath("docker-compose")
	if err != nil {
		// Try docker compose plugin
		path, err = exec.LookPath("docker")
		require.NoError(suite.T(), err)
	}

	// Test version check with real executable
	err = checkDockerComposeVersion(path)
	require.NoError(suite.T(), err)

	// Test with non-existent executable
	err = checkDockerComposeVersion("/nonexistent/docker-compose")
	assert.Error(suite.T(), err)
}

// TestCheckDockerComposePlugin tests Docker Compose plugin detection
func (suite *DockerComposeTestSuite) TestCheckDockerComposePlugin() {
	// Find the real docker executable
	path, err := exec.LookPath("docker")
	require.NoError(suite.T(), err)

	// Test plugin check with real docker
	err = checkDockerComposePlugin(path)
	require.NoError(suite.T(), err)

	// Test with non-existent executable
	err = checkDockerComposePlugin("/nonexistent/docker")
	assert.Error(suite.T(), err)
}

// Run the test suite
func TestDockerComposeTestSuite(t *testing.T) {
	suite.Run(t, new(DockerComposeTestSuite))
}
