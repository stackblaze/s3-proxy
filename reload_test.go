package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReloadableHandler_ConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	
	configContent := `- host: test.localhost
  awsKey: test-key
  awsSecret: test-secret
  awsRegion: us-east-1
  awsBucket: test-bucket
`
	
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	
	handler, err := NewReloadableHandler(configFile)
	if err != nil {
		t.Fatalf("NewReloadableHandler() error = %v", err)
	}
	if handler == nil {
		t.Fatal("NewReloadableHandler() returned nil handler")
	}
}

func TestReloadableHandler_InvalidConfig(t *testing.T) {
	// Create a temporary invalid config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	
	configContent := `invalid yaml content
`
	
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	
	_, err = NewReloadableHandler(configFile)
	if err == nil {
		t.Error("NewReloadableHandler() should return error for invalid config")
	}
}

func TestReloadableHandler_MissingConfig(t *testing.T) {
	_, err := NewReloadableHandler("/nonexistent/config.yaml")
	if err == nil {
		t.Error("NewReloadableHandler() should return error for missing config file")
	}
}

