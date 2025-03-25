package compose

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MergeTestSuite defines the test suite for merge functionality
type MergeTestSuite struct {
	suite.Suite
	logger *logrus.Entry
	tmpDir string
}

// SetupTest runs before each test
func (suite *MergeTestSuite) SetupTest() {
	suite.logger = logrus.New().WithField("test", true)
	suite.tmpDir = suite.T().TempDir()
}

// TestNewComposeFile tests loading a single compose file
func (suite *MergeTestSuite) TestNewComposeFile() {
	// Create a test compose file
	testFile := filepath.Join(suite.tmpDir, "docker-compose.yml")
	content := []byte(`
version: '3'
services:
  app:
    image: nginx
    build:
      context: ./app
    ports:
      - "80:80"
`)
	err := os.WriteFile(testFile, content, 0644)
	require.NoError(suite.T(), err)

	// Create necessary directories
	err = os.MkdirAll(filepath.Join(suite.tmpDir, "app"), 0755)
	require.NoError(suite.T(), err)

	// Test loading the compose file
	cf, err := NewComposeFile(testFile, suite.logger)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), testFile, cf.Path)
	assert.Equal(suite.T(), suite.tmpDir, cf.BaseDir)
	assert.NotNil(suite.T(), cf.Project)

	// Verify the loaded service configuration
	require.Contains(suite.T(), cf.Project.Services, "app")
	service := cf.Project.Services["app"]
	assert.Equal(suite.T(), "nginx", service.Image)
	assert.NotNil(suite.T(), service.Build)
	assert.Equal(suite.T(), filepath.Join(suite.tmpDir, "app"), service.Build.Context)
}

// TestMergeComposeFiles tests merging multiple compose files
func (suite *MergeTestSuite) TestMergeComposeFiles() {
	// Create the first compose file in a subdirectory
	folder1 := filepath.Join(suite.tmpDir, "folder1")
	err := os.MkdirAll(folder1, 0755)
	require.NoError(suite.T(), err)
	file1 := filepath.Join(folder1, "docker-compose.yml")
	content1 := []byte(`
version: '3'
services:
  app1:
    image: nginx
    build:
      context: ./app1
    ports:
      - "80:80"
  app2:
    image: redis
    ports:
      - "6379:6379"
`)
	err = os.WriteFile(file1, content1, 0644)
	require.NoError(suite.T(), err)

	// Create the second compose file in a different subdirectory
	folder2 := filepath.Join(suite.tmpDir, "folder2")
	err = os.MkdirAll(folder2, 0755)
	require.NoError(suite.T(), err)
	file2 := filepath.Join(folder2, "docker-compose.yml")
	content2 := []byte(`
version: '3'
services:
  app1:
    build:
      context: ./app1-override
    environment:
      - DEBUG=true
  app3:
    image: postgres
    ports:
      - "5432:5432"
`)
	err = os.WriteFile(file2, content2, 0644)
	require.NoError(suite.T(), err)

	// Create necessary directories
	err = os.MkdirAll(filepath.Join(folder1, "app1"), 0755)
	require.NoError(suite.T(), err)
	err = os.MkdirAll(filepath.Join(folder2, "app1-override"), 0755)
	require.NoError(suite.T(), err)

	// Load and merge the compose files
	cf1, err := NewComposeFile(file1, suite.logger)
	require.NoError(suite.T(), err)
	cf2, err := NewComposeFile(file2, suite.logger)
	require.NoError(suite.T(), err)

	merged, err := MergeComposeFiles([]*ComposeFile{cf1, cf2})
	require.NoError(suite.T(), err)

	// Verify merged configuration
	assert.Len(suite.T(), merged.Services, 4)

	// Check app1 from folder1
	folder1App1 := merged.Services["folder1_app1"]
	assert.Equal(suite.T(), "nginx", folder1App1.Image)
	assert.Equal(suite.T(), filepath.Join(folder1, "app1"), folder1App1.Build.Context)

	// Check app2 from folder1
	folder1App2 := merged.Services["folder1_app2"]
	assert.Equal(suite.T(), "redis", folder1App2.Image)
	assert.Equal(suite.T(), uint32(6379), folder1App2.Ports[0].Target)

	// Check app1 from folder2
	folder2App1 := merged.Services["folder2_app1"]
	assert.Equal(suite.T(), filepath.Join(folder2, "app1-override"), folder2App1.Build.Context)
	assert.Contains(suite.T(), folder2App1.Environment, "DEBUG")

	// Check app3 from folder2
	folder2App3 := merged.Services["folder2_app3"]
	assert.Equal(suite.T(), "postgres", folder2App3.Image)
	assert.Equal(suite.T(), uint32(5432), folder2App3.Ports[0].Target)
}

// TestAdjustBuildContexts tests the build context adjustment functionality
func (suite *MergeTestSuite) TestAdjustBuildContexts() {
	// Create a test compose file with relative and absolute build contexts
	testFile := filepath.Join(suite.tmpDir, "docker-compose.yml")
	content := []byte(`
version: '3'
services:
  app1:
    build:
      context: ./app1
  app2:
    build:
      context: /absolute/path
`)
	err := os.WriteFile(testFile, content, 0644)
	require.NoError(suite.T(), err)

	// Create necessary directory
	err = os.MkdirAll(filepath.Join(suite.tmpDir, "app1"), 0755)
	require.NoError(suite.T(), err)

	// Load the compose file
	cf, err := NewComposeFile(testFile, suite.logger)
	require.NoError(suite.T(), err)

	// Adjust build contexts
	err = cf.adjustBuildContexts()
	require.NoError(suite.T(), err)

	// Verify the adjusted build contexts
	app1 := cf.Project.Services["app1"]
	assert.Equal(suite.T(), filepath.Join(suite.tmpDir, "app1"), app1.Build.Context)

	app2 := cf.Project.Services["app2"]
	assert.Equal(suite.T(), "/absolute/path", app2.Build.Context)
}

// TestPrefixResourceNames tests the resource name prefixing functionality
func (suite *MergeTestSuite) TestPrefixResourceNames() {
	// Create a test compose file with various resources
	testFile := filepath.Join(suite.tmpDir, "docker-compose.yml")
	content := []byte(`
version: '3'
services:
  app:
    image: nginx
    depends_on:
      db:
        condition: service_started
      redis:
        condition: service_started
        restart: true
    links:
      - redis:redis
  db:
    image: postgres
  redis:
    image: redis
volumes:
  data:
    driver: local
configs:
  app_config:
    file: ./config.json
secrets:
  db_password:
    file: ./secrets/db_password.txt
`)
	err := os.WriteFile(testFile, content, 0644)
	require.NoError(suite.T(), err)

	// Create necessary directories
	err = os.MkdirAll(filepath.Join(suite.tmpDir, "secrets"), 0755)
	require.NoError(suite.T(), err)

	// Load the compose file
	cf, err := NewComposeFile(testFile, suite.logger)
	require.NoError(suite.T(), err)

	// Test prefixing with a sample prefix
	prefix := "test"
	err = cf.prefixResourceNames(prefix)
	require.NoError(suite.T(), err)

	// Verify service names are prefixed
	assert.Contains(suite.T(), cf.Project.Services, prefix+"_app")
	assert.Contains(suite.T(), cf.Project.Services, prefix+"_db")
	assert.Contains(suite.T(), cf.Project.Services, prefix+"_redis")

	// Verify volume names are prefixed
	assert.Contains(suite.T(), cf.Project.Volumes, prefix+"_data")

	// Verify config names are prefixed
	assert.Contains(suite.T(), cf.Project.Configs, prefix+"_app_config")

	// Verify secret names are prefixed
	assert.Contains(suite.T(), cf.Project.Secrets, prefix+"_db_password")

	// Verify dependencies are updated
	appService := cf.Project.Services[prefix+"_app"]
	assert.Contains(suite.T(), appService.DependsOn, prefix+"_db")

	// Verify links are updated
	assert.Contains(suite.T(), appService.Links, prefix+"_redis:redis")
}

// TestMergeComposeFilesWithPrefixing tests merging compose files with resource name prefixing
func (suite *MergeTestSuite) TestMergeComposeFilesWithPrefixing() {
	// Create the first compose file
	file1 := filepath.Join(suite.tmpDir, "folder1", "docker-compose.yml")
	err := os.MkdirAll(filepath.Dir(file1), 0755)
	require.NoError(suite.T(), err)
	content1 := []byte(`
version: '3'
services:
  app:
    image: nginx
    depends_on:
      db:
        condition: service_started
  db:
    image: postgres
volumes:
  data:
    driver: local
`)
	err = os.WriteFile(file1, content1, 0644)
	require.NoError(suite.T(), err)

	// Create the second compose file
	file2 := filepath.Join(suite.tmpDir, "folder2", "docker-compose.yml")
	err = os.MkdirAll(filepath.Dir(file2), 0755)
	require.NoError(suite.T(), err)
	content2 := []byte(`
version: '3'
services:
  app:
    image: nginx
    depends_on:
      db:
        condition: service_started
  db:
    image: postgres
volumes:
  data:
    driver: local
`)
	err = os.WriteFile(file2, content2, 0644)
	require.NoError(suite.T(), err)

	// Load and merge the compose files
	cf1, err := NewComposeFile(file1, suite.logger)
	require.NoError(suite.T(), err)
	cf2, err := NewComposeFile(file2, suite.logger)
	require.NoError(suite.T(), err)

	merged, err := MergeComposeFiles([]*ComposeFile{cf1, cf2})
	require.NoError(suite.T(), err)

	// Verify that services from both files are present with correct prefixes
	assert.Contains(suite.T(), merged.Services, "folder1_app")
	assert.Contains(suite.T(), merged.Services, "folder1_db")
	assert.Contains(suite.T(), merged.Services, "folder2_app")
	assert.Contains(suite.T(), merged.Services, "folder2_db")

	// Verify that volumes from both files are present with correct prefixes
	assert.Contains(suite.T(), merged.Volumes, "folder1_data")
	assert.Contains(suite.T(), merged.Volumes, "folder2_data")

	// Verify that dependencies are correctly updated
	folder1App := merged.Services["folder1_app"]
	assert.Contains(suite.T(), folder1App.DependsOn, "folder1_db")

	folder2App := merged.Services["folder2_app"]
	assert.Contains(suite.T(), folder2App.DependsOn, "folder2_db")
}

// TestMergeComposeFilesWithPortConflicts tests merging compose files with port conflict resolution
func (suite *MergeTestSuite) TestMergeComposeFilesWithPortConflicts() {
	// Create the first compose file
	file1 := filepath.Join(suite.tmpDir, "folder1", "docker-compose.yml")
	err := os.MkdirAll(filepath.Dir(file1), 0755)
	require.NoError(suite.T(), err)
	content1 := []byte(`
version: '3'
services:
  web:
    image: nginx
    ports:
      - "80:80"
      - "443:443"
  redis:
    image: redis
    ports:
      - "6379:6379"
`)
	err = os.WriteFile(file1, content1, 0644)
	require.NoError(suite.T(), err)

	// Create the second compose file with conflicting ports
	file2 := filepath.Join(suite.tmpDir, "folder2", "docker-compose.yml")
	err = os.MkdirAll(filepath.Dir(file2), 0755)
	require.NoError(suite.T(), err)
	content2 := []byte(`
version: '3'
services:
  web:
    image: nginx
    ports:
      - "80:80"
      - "443:443"
  postgres:
    image: postgres
    ports:
      - "5432:5432"
`)
	err = os.WriteFile(file2, content2, 0644)
	require.NoError(suite.T(), err)

	// Load and merge the compose files
	cf1, err := NewComposeFile(file1, suite.logger)
	require.NoError(suite.T(), err)
	cf2, err := NewComposeFile(file2, suite.logger)
	require.NoError(suite.T(), err)

	merged, err := MergeComposeFiles([]*ComposeFile{cf1, cf2})
	require.NoError(suite.T(), err)

	// Verify that services from both files are present with correct prefixes
	assert.Contains(suite.T(), merged.Services, "folder1_web")
	assert.Contains(suite.T(), merged.Services, "folder1_redis")
	assert.Contains(suite.T(), merged.Services, "folder2_web")
	assert.Contains(suite.T(), merged.Services, "folder2_postgres")

	// Verify that port conflicts are resolved
	folder1Web := merged.Services["folder1_web"]
	folder2Web := merged.Services["folder2_web"]

	// First file's services should keep their original ports
	assert.Equal(suite.T(), "80", folder1Web.Ports[0].Published)
	assert.Equal(suite.T(), "443", folder1Web.Ports[1].Published)

	// Second file's services should have their ports adjusted
	assert.Equal(suite.T(), "180", folder2Web.Ports[0].Published)
	assert.Equal(suite.T(), "543", folder2Web.Ports[1].Published)

	// Non-conflicting ports should remain unchanged
	folder1Redis := merged.Services["folder1_redis"]
	folder2Postgres := merged.Services["folder2_postgres"]
	assert.Equal(suite.T(), "6379", folder1Redis.Ports[0].Published)
	assert.Equal(suite.T(), "5432", folder2Postgres.Ports[0].Published)
}

// Run the test suite
func TestMergeTestSuite(t *testing.T) {
	suite.Run(t, new(MergeTestSuite))
}
