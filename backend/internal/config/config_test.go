package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Create a minimal config file
	tmpFile := "/tmp/test_config.yaml"
	content := `
server:
  port: 9090
docker:
  host: tcp://localhost:2375
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}
	defer os.Remove(tmpFile)

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}

	if cfg.Docker.Host != "tcp://localhost:2375" {
		t.Errorf("expected docker host tcp://localhost:2375, got %s", cfg.Docker.Host)
	}

	// Check defaults
	if cfg.App.Name != "DockerSphere" {
		t.Errorf("expected app name DockerSphere, got %s", cfg.App.Name)
	}

	if cfg.Task.WorkerPoolSize != 4 {
		t.Errorf("expected worker pool size 4, got %d", cfg.Task.WorkerPoolSize)
	}
}

func TestLoad_EnvironmentOverride(t *testing.T) {
	tmpFile := "/tmp/test_config_env.yaml"
	content := `
server:
  port: 8080
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}
	defer os.Remove(tmpFile)

	// Set environment variable
	os.Setenv("DOCKERSPHERE_SERVER_PORT", "9999")
	defer os.Unsetenv("DOCKERSPHERE_SERVER_PORT")

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}

	if cfg.Server.Port != 9999 {
		t.Errorf("expected port 9999 from env, got %d", cfg.Server.Port)
	}
}
