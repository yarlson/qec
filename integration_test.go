package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite defines the test suite for end-to-end testing
type IntegrationTestSuite struct {
	suite.Suite
	logger *logrus.Entry
	tmpDir string
}

// SetupTest runs before each test
func (suite *IntegrationTestSuite) SetupTest() {
	suite.logger = logrus.New().WithField("test", true)
	suite.tmpDir = suite.T().TempDir()
}

// createTestFiles creates test compose files and directories
func (suite *IntegrationTestSuite) createTestFiles() (string, string) {
	// Create the first compose file in a subdirectory
	folder1 := filepath.Join(suite.tmpDir, "web")
	err := os.MkdirAll(folder1, 0755)
	require.NoError(suite.T(), err)

	file1 := filepath.Join(folder1, "docker-compose.yml")
	content1 := []byte(`services:
  frontend:
    image: nginx:latest
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "80:80"
    environment:
      - NODE_ENV=production
    depends_on:
      - api
  api:
    image: node:16
    build:
      context: ./api
    ports:
      - "3000:3000"
    environment:
      - API_KEY=test123
    volumes:
      - web_data:/data
volumes:
  web_data:`)
	err = os.WriteFile(file1, content1, 0644)
	require.NoError(suite.T(), err)

	// Create necessary directories for first compose file
	err = os.MkdirAll(filepath.Join(folder1, "frontend"), 0755)
	require.NoError(suite.T(), err)
	err = os.WriteFile(filepath.Join(folder1, "frontend", "Dockerfile"), []byte("FROM nginx:latest"), 0644)
	require.NoError(suite.T(), err)
	err = os.MkdirAll(filepath.Join(folder1, "api"), 0755)
	require.NoError(suite.T(), err)

	// Create the second compose file in a different subdirectory
	folder2 := filepath.Join(suite.tmpDir, "db")
	err = os.MkdirAll(folder2, 0755)
	require.NoError(suite.T(), err)

	file2 := filepath.Join(folder2, "docker-compose.yml")
	content2 := []byte(`services:
  api:
    build:
      context: ./api-override
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
    depends_on:
      - postgres
  postgres:
    image: postgres:13
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_PASSWORD=secret
    volumes:
      - db_data:/var/lib/postgresql/data
volumes:
  db_data:`)
	err = os.WriteFile(file2, content2, 0644)
	require.NoError(suite.T(), err)

	// Create necessary directories for second compose file
	err = os.MkdirAll(filepath.Join(folder2, "api-override"), 0755)
	require.NoError(suite.T(), err)

	return file1, file2
}

// TestEndToEndConfig tests the complete configuration processing pipeline
func (suite *IntegrationTestSuite) TestEndToEndConfig() {
	// Create test files
	file1, file2 := suite.createTestFiles()

	// Reset flags and global variables
	resetFlags()
	composeFiles = []string{file1, file2}
	command = "config"
	verbose = true
	dryRun = false

	// Run the program
	err := run()
	require.NoError(suite.T(), err)

	// Verify the merged configuration file exists and contains expected content
	configFile := filepath.Join(filepath.Dir(file1), "docker-compose.merged.yml")
	content, err := os.ReadFile(configFile)
	require.NoError(suite.T(), err)

	contentStr := string(content)

	// Check for prefixed service names
	assert.Contains(suite.T(), contentStr, "web_frontend")
	assert.Contains(suite.T(), contentStr, "web_api")
	assert.Contains(suite.T(), contentStr, "db_api")
	assert.Contains(suite.T(), contentStr, "db_postgres")

	// Check for absolute build contexts
	assert.Contains(suite.T(), contentStr, filepath.Join(filepath.Dir(file1), "frontend"))
	assert.Contains(suite.T(), contentStr, filepath.Join(filepath.Dir(file1), "api"))
	assert.Contains(suite.T(), contentStr, filepath.Join(filepath.Dir(file2), "api-override"))

	// Check for prefixed volume names
	assert.Contains(suite.T(), contentStr, "web_web_data")
	assert.Contains(suite.T(), contentStr, "db_db_data")

	// Check for updated dependencies
	assert.Contains(suite.T(), contentStr, "web_api")
	assert.Contains(suite.T(), contentStr, "db_postgres")
}

// TestEndToEndDryRun tests the dry-run functionality
func (suite *IntegrationTestSuite) TestEndToEndDryRun() {
	// Create test files
	file1, file2 := suite.createTestFiles()

	// Reset flags and global variables
	resetFlags()
	composeFiles = []string{file1, file2}
	command = "up"
	verbose = true
	dryRun = true

	// Run the program
	err := run()
	require.NoError(suite.T(), err)

	// Verify the merged configuration file was not created
	configFile := filepath.Join(filepath.Dir(file1), "docker-compose.merged.yml")
	_, err = os.Stat(configFile)
	assert.True(suite.T(), os.IsNotExist(err))
}

// TestEndToEndPortConflicts tests port conflict resolution
func (suite *IntegrationTestSuite) TestEndToEndPortConflicts() {
	// Create test files with conflicting ports
	folder1 := filepath.Join(suite.tmpDir, "app1")
	folder2 := filepath.Join(suite.tmpDir, "app2")
	err := os.MkdirAll(folder1, 0755)
	require.NoError(suite.T(), err)
	err = os.MkdirAll(folder2, 0755)
	require.NoError(suite.T(), err)

	file1 := filepath.Join(folder1, "docker-compose.yml")
	content1 := []byte(`services:
  web:
    image: nginx
    ports:
      - "80:80"
      - "443:443"`)
	err = os.WriteFile(file1, content1, 0644)
	require.NoError(suite.T(), err)

	file2 := filepath.Join(folder2, "docker-compose.yml")
	content2 := []byte(`services:
  web:
    image: nginx
    ports:
      - "80:80"
      - "443:443"`)
	err = os.WriteFile(file2, content2, 0644)
	require.NoError(suite.T(), err)

	// Reset flags and global variables
	resetFlags()
	composeFiles = []string{file1, file2}
	command = "config"
	verbose = true

	// Run the program
	err = run()
	require.NoError(suite.T(), err)

	// Verify the merged configuration
	configFile := filepath.Join(filepath.Dir(file1), "docker-compose.merged.yml")
	content, err := os.ReadFile(configFile)
	require.NoError(suite.T(), err)

	contentStr := string(content)

	// Check for port mappings in the expected format
	assert.Contains(suite.T(), contentStr, `target: 80`)
	assert.Contains(suite.T(), contentStr, `published: "80"`)
	assert.Contains(suite.T(), contentStr, `target: 443`)
	assert.Contains(suite.T(), contentStr, `published: "443"`)
	assert.Contains(suite.T(), contentStr, `target: 80`)
	assert.Contains(suite.T(), contentStr, `published: "180"`)
	assert.Contains(suite.T(), contentStr, `target: 443`)
	assert.Contains(suite.T(), contentStr, `published: "543"`)
}

// TestEndToEndErrorHandling tests error scenarios
func (suite *IntegrationTestSuite) TestEndToEndErrorHandling() {
	// Test with non-existent file
	resetFlags()
	composeFiles = []string{"nonexistent.yml"}
	command = "up"

	err := run()
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "error loading compose file")

	// Test with invalid YAML
	invalidFile := filepath.Join(suite.tmpDir, "invalid.yml")
	err = os.WriteFile(invalidFile, []byte("invalid: yaml: content"), 0644)
	require.NoError(suite.T(), err)

	resetFlags()
	composeFiles = []string{invalidFile}
	command = "up"

	err = run()
	assert.Error(suite.T(), err)
}

// resetFlags resets all global flags to their default values
func resetFlags() {
	composeFiles = nil
	verbose = false
	dryRun = false
	detach = false
	command = "up"
	showHelp = false
	args = nil
}

// Run the test suite
func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
