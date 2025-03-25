package compose

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewComposeFile(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create a test compose file
	testFile := filepath.Join(tmpDir, "docker-compose.yml")
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
	require.NoError(t, err)

	// Create necessary directories
	err = os.MkdirAll(filepath.Join(tmpDir, "app"), 0755)
	require.NoError(t, err)

	// Create a logger for testing
	logger := logrus.New().WithField("test", true)

	// Test loading the compose file
	cf, err := NewComposeFile(testFile, logger)
	require.NoError(t, err)
	assert.Equal(t, testFile, cf.Path)
	assert.Equal(t, tmpDir, cf.BaseDir)
	assert.NotNil(t, cf.Project)

	// Verify the loaded service configuration
	require.Contains(t, cf.Project.Services, "app")
	service := cf.Project.Services["app"]
	assert.Equal(t, "nginx", service.Image)
	assert.NotNil(t, service.Build)
	assert.Equal(t, filepath.Join(tmpDir, "app"), service.Build.Context)
}

func TestMergeComposeFiles(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create the first compose file in a subdirectory
	folder1 := filepath.Join(tmpDir, "folder1")
	err := os.MkdirAll(folder1, 0755)
	require.NoError(t, err)
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
	require.NoError(t, err)

	// Create the second compose file in a different subdirectory
	folder2 := filepath.Join(tmpDir, "folder2")
	err = os.MkdirAll(folder2, 0755)
	require.NoError(t, err)
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
	require.NoError(t, err)

	// Create necessary directories
	err = os.MkdirAll(filepath.Join(folder1, "app1"), 0755)
	require.NoError(t, err)
	err = os.MkdirAll(filepath.Join(folder2, "app1-override"), 0755)
	require.NoError(t, err)

	// Create a logger for testing
	logger := logrus.New().WithField("test", true)

	// Load and merge the compose files
	cf1, err := NewComposeFile(file1, logger)
	require.NoError(t, err)
	cf2, err := NewComposeFile(file2, logger)
	require.NoError(t, err)

	merged, err := MergeComposeFiles([]*ComposeFile{cf1, cf2})
	require.NoError(t, err)

	// Verify merged configuration
	assert.Len(t, merged.Services, 4)

	// Check app1 from folder1
	folder1App1 := merged.Services["folder1_app1"]
	assert.Equal(t, "nginx", folder1App1.Image)
	assert.Equal(t, filepath.Join(folder1, "app1"), folder1App1.Build.Context)

	// Check app2 from folder1
	folder1App2 := merged.Services["folder1_app2"]
	assert.Equal(t, "redis", folder1App2.Image)
	assert.Equal(t, uint32(6379), folder1App2.Ports[0].Target)

	// Check app1 from folder2
	folder2App1 := merged.Services["folder2_app1"]
	assert.Equal(t, filepath.Join(folder2, "app1-override"), folder2App1.Build.Context)
	assert.Contains(t, folder2App1.Environment, "DEBUG")

	// Check app3 from folder2
	folder2App3 := merged.Services["folder2_app3"]
	assert.Equal(t, "postgres", folder2App3.Image)
	assert.Equal(t, uint32(5432), folder2App3.Ports[0].Target)
}

func TestAdjustBuildContexts(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create a test compose file with relative and absolute build contexts
	testFile := filepath.Join(tmpDir, "docker-compose.yml")
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
	require.NoError(t, err)

	// Create necessary directory
	err = os.MkdirAll(filepath.Join(tmpDir, "app1"), 0755)
	require.NoError(t, err)

	// Create a logger for testing
	logger := logrus.New().WithField("test", true)

	// Load the compose file
	cf, err := NewComposeFile(testFile, logger)
	require.NoError(t, err)

	// Adjust build contexts
	err = cf.adjustBuildContexts()
	require.NoError(t, err)

	// Verify the adjusted build contexts
	app1 := cf.Project.Services["app1"]
	assert.Equal(t, filepath.Join(tmpDir, "app1"), app1.Build.Context)

	app2 := cf.Project.Services["app2"]
	assert.Equal(t, "/absolute/path", app2.Build.Context)
}

func TestPrefixResourceNames(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create a test compose file with various resources
	testFile := filepath.Join(tmpDir, "docker-compose.yml")
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
	require.NoError(t, err)

	// Create necessary directories
	err = os.MkdirAll(filepath.Join(tmpDir, "secrets"), 0755)
	require.NoError(t, err)

	// Create a logger for testing
	logger := logrus.New().WithField("test", true)

	// Load the compose file
	cf, err := NewComposeFile(testFile, logger)
	require.NoError(t, err)

	// Test prefixing with a sample prefix
	prefix := "test"
	err = cf.prefixResourceNames(prefix)
	require.NoError(t, err)

	// Verify service names are prefixed
	assert.Contains(t, cf.Project.Services, prefix+"_app")
	assert.Contains(t, cf.Project.Services, prefix+"_db")
	assert.Contains(t, cf.Project.Services, prefix+"_redis")

	// Verify volume names are prefixed
	assert.Contains(t, cf.Project.Volumes, prefix+"_data")

	// Verify config names are prefixed
	assert.Contains(t, cf.Project.Configs, prefix+"_app_config")

	// Verify secret names are prefixed
	assert.Contains(t, cf.Project.Secrets, prefix+"_db_password")

	// Verify dependencies are updated
	appService := cf.Project.Services[prefix+"_app"]
	assert.Contains(t, appService.DependsOn, prefix+"_db")

	// Verify links are updated
	assert.Contains(t, appService.Links, prefix+"_redis:redis")
}

func TestMergeComposeFilesWithPrefixing(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create the first compose file
	file1 := filepath.Join(tmpDir, "folder1", "docker-compose.yml")
	err := os.MkdirAll(filepath.Dir(file1), 0755)
	require.NoError(t, err)
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
	require.NoError(t, err)

	// Create the second compose file
	file2 := filepath.Join(tmpDir, "folder2", "docker-compose.yml")
	err = os.MkdirAll(filepath.Dir(file2), 0755)
	require.NoError(t, err)
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
	require.NoError(t, err)

	// Create a logger for testing
	logger := logrus.New().WithField("test", true)

	// Load and merge the compose files
	cf1, err := NewComposeFile(file1, logger)
	require.NoError(t, err)
	cf2, err := NewComposeFile(file2, logger)
	require.NoError(t, err)

	merged, err := MergeComposeFiles([]*ComposeFile{cf1, cf2})
	require.NoError(t, err)

	// Verify that services from both files are present with correct prefixes
	assert.Contains(t, merged.Services, "folder1_app")
	assert.Contains(t, merged.Services, "folder1_db")
	assert.Contains(t, merged.Services, "folder2_app")
	assert.Contains(t, merged.Services, "folder2_db")

	// Verify that volumes from both files are present with correct prefixes
	assert.Contains(t, merged.Volumes, "folder1_data")
	assert.Contains(t, merged.Volumes, "folder2_data")

	// Verify that dependencies are correctly updated
	folder1App := merged.Services["folder1_app"]
	assert.Contains(t, folder1App.DependsOn, "folder1_db")

	folder2App := merged.Services["folder2_app"]
	assert.Contains(t, folder2App.DependsOn, "folder2_db")
}
