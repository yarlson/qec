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

	// Create the first compose file
	file1 := filepath.Join(tmpDir, "docker-compose-1.yml")
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
	err := os.WriteFile(file1, content1, 0644)
	require.NoError(t, err)

	// Create the second compose file
	file2 := filepath.Join(tmpDir, "docker-compose-2.yml")
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
	err = os.MkdirAll(filepath.Join(tmpDir, "app1"), 0755)
	require.NoError(t, err)
	err = os.MkdirAll(filepath.Join(tmpDir, "app1-override"), 0755)
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
	assert.Len(t, merged.Services, 3)

	// Check app1 (merged service)
	app1 := merged.Services["app1"]
	assert.Equal(t, "nginx", app1.Image)
	assert.Equal(t, filepath.Join(tmpDir, "app1-override"), app1.Build.Context)
	assert.Contains(t, app1.Environment, "DEBUG")

	// Check app2 (from first file only)
	app2 := merged.Services["app2"]
	assert.Equal(t, "redis", app2.Image)
	assert.Equal(t, uint32(6379), app2.Ports[0].Target)

	// Check app3 (from second file only)
	app3 := merged.Services["app3"]
	assert.Equal(t, "postgres", app3.Image)
	assert.Equal(t, uint32(5432), app3.Ports[0].Target)
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
