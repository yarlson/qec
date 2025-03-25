package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateAndResolve(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-compose.yml")
	err := os.WriteFile(tmpFile, []byte("version: '3'\nservices:\n  app:\n    image: busybox"), 0644)
	if err != nil {
		t.Fatalf("Failed to write temporary file: %v", err)
	}

	// Test valid file
	info, err := validateAndResolve(tmpFile)
	if err != nil {
		t.Errorf("Expected valid file, got error: %v", err)
	}

	if info.BaseDir != tmpDir {
		t.Errorf("Expected base dir %s, got %s", tmpDir, info.BaseDir)
	}

	// Test non-existent file
	_, err = validateAndResolve("non-existent-file.yml")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestLoadYAML(t *testing.T) {
	// Create a temporary test file with valid YAML
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-compose.yml")
	yamlContent := `
version: '3'
services:
  app:
    image: busybox
    command: ["echo", "hello world"]
`
	err := os.WriteFile(tmpFile, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write temporary YAML file: %v", err)
	}

	// Test loading valid YAML
	content, err := loadYAML(tmpFile)
	if err != nil {
		t.Errorf("Expected YAML to load, got error: %v", err)
	}

	// Verify content
	if content["version"] != "3" {
		t.Errorf("Expected version '3', got %v", content["version"])
	}

	services, ok := content["services"].(map[string]interface{})
	if !ok {
		t.Error("Expected services to be a map")
		return
	}

	app, ok := services["app"].(map[string]interface{})
	if !ok {
		t.Error("Expected app service to be a map")
		return
	}

	if app["image"] != "busybox" {
		t.Errorf("Expected image 'busybox', got %v", app["image"])
	}

	// Test loading invalid YAML
	err = os.WriteFile(tmpFile, []byte("invalid: yaml: ]["), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid YAML file: %v", err)
	}

	_, err = loadYAML(tmpFile)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}
