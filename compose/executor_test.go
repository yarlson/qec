package compose

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ExecutorTestSuite defines the test suite for Docker Compose executor
type ExecutorTestSuite struct {
	suite.Suite
	logger  *logrus.Entry
	tmpDir  string
	project *types.Project
}

// SetupTest runs before each test
func (suite *ExecutorTestSuite) SetupTest() {
	suite.logger = logrus.New().WithField("test", true)
	suite.tmpDir = suite.T().TempDir()

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

	// Load the compose file
	cf, err := NewComposeFile(composeFile, suite.logger)
	require.NoError(suite.T(), err)
	suite.project = cf.Project
}

// TestNewExecutor tests executor creation
func (suite *ExecutorTestSuite) TestNewExecutor() {
	executor := NewExecutor(suite.project, suite.tmpDir, true, suite.logger)
	assert.NotNil(suite.T(), executor)
	assert.Equal(suite.T(), suite.project, executor.project)
	assert.Equal(suite.T(), suite.tmpDir, executor.workingDir)
	assert.True(suite.T(), executor.dryRun)
}

// TestWriteConfig tests configuration file writing
func (suite *ExecutorTestSuite) TestWriteConfig() {
	executor := NewExecutor(suite.project, suite.tmpDir, false, suite.logger)

	// Write the configuration
	configFile, err := executor.writeConfig()
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), configFile)

	// Verify the file exists and contains the expected content
	content, err := os.ReadFile(configFile)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(content), "hello-world")
}

// TestUpDryRun tests the up command in dry-run mode
func (suite *ExecutorTestSuite) TestUpDryRun() {
	executor := NewExecutor(suite.project, suite.tmpDir, true, suite.logger)

	// Test up command
	err := executor.Up(true)
	assert.NoError(suite.T(), err)

	// Verify the merged config file was not created
	_, err = os.Stat(filepath.Join(suite.tmpDir, "docker-compose.merged.yml"))
	assert.True(suite.T(), os.IsNotExist(err))
}

// TestDownDryRun tests the down command in dry-run mode
func (suite *ExecutorTestSuite) TestDownDryRun() {
	executor := NewExecutor(suite.project, suite.tmpDir, true, suite.logger)

	// Test down command
	err := executor.Down()
	assert.NoError(suite.T(), err)

	// Verify the merged config file was not created
	_, err = os.Stat(filepath.Join(suite.tmpDir, "docker-compose.merged.yml"))
	assert.True(suite.T(), os.IsNotExist(err))
}

// TestConfigDryRun tests the config command in dry-run mode
func (suite *ExecutorTestSuite) TestConfigDryRun() {
	executor := NewExecutor(suite.project, suite.tmpDir, true, suite.logger)

	// Test config command
	err := executor.Config()
	assert.NoError(suite.T(), err)

	// Verify the merged config file was not created
	_, err = os.Stat(filepath.Join(suite.tmpDir, "docker-compose.merged.yml"))
	assert.True(suite.T(), os.IsNotExist(err))
}

// TestUpLive tests the up command with actual execution
func (suite *ExecutorTestSuite) TestUpLive() {
	executor := NewExecutor(suite.project, suite.tmpDir, false, suite.logger)

	// Test up command
	err := executor.Up(true)
	assert.NoError(suite.T(), err)

	// Verify the merged config file was created
	configFile := filepath.Join(suite.tmpDir, "docker-compose.merged.yml")
	_, err = os.Stat(configFile)
	assert.NoError(suite.T(), err)

	// Clean up
	_ = executor.Down()
}

// Run the test suite
func TestExecutorTestSuite(t *testing.T) {
	suite.Run(t, new(ExecutorTestSuite))
}
