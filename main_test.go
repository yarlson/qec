package main

import (
	"os"
	"path/filepath"
	"testing"

	"gihub.com/yarlson/qec/compose"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComposeFileHandling(t *testing.T) {
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
	cf, err := compose.NewComposeFile(testFile, logger)
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

func TestMultipleComposeFiles(t *testing.T) {
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
	cf1, err := compose.NewComposeFile(file1, logger)
	require.NoError(t, err)
	cf2, err := compose.NewComposeFile(file2, logger)
	require.NoError(t, err)

	merged, err := compose.MergeComposeFiles([]*compose.ComposeFile{cf1, cf2})
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
