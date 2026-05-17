package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigValidateService(t *testing.T) {
	cfg := &Config{
		Services: map[string]Service{
			"api":      {Dir: "api", Procs: []string{"rails"}},
			"frontend": {Dir: "app", Procs: []string{"app"}},
		},
	}

	if err := cfg.ValidateService("api"); err != nil {
		t.Errorf("expected no error for valid service, got: %v", err)
	}

	if err := cfg.ValidateService("nonexistent"); err == nil {
		t.Error("expected error for nonexistent service")
	}
}

func TestConfigServiceNames(t *testing.T) {
	cfg := &Config{
		Services: map[string]Service{
			"api":      {Dir: "api"},
			"frontend": {Dir: "app"},
			"backend":  {Dir: "backend"},
		},
	}

	names := cfg.ServiceNames()

	if len(names) != 3 {
		t.Errorf("expected 3 services, got %d", len(names))
	}

	// Names should be sorted
	expected := []string{"api", "backend", "frontend"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("expected %q at index %d, got %q", expected[i], i, name)
		}
	}
}

func TestConfigValidateRoot(t *testing.T) {
	// Test empty root
	cfg := &Config{Root: ""}
	if err := cfg.ValidateRoot(); err == nil {
		t.Error("expected error for empty root")
	}

	// Test nonexistent root
	cfg = &Config{Root: "/nonexistent/path/12345"}
	if err := cfg.ValidateRoot(); err == nil {
		t.Error("expected error for nonexistent root")
	}

	// Test valid root
	tmpDir, _ := os.MkdirTemp("", "config-test")
	defer os.RemoveAll(tmpDir)

	cfg = &Config{Root: tmpDir}
	if err := cfg.ValidateRoot(); err != nil {
		t.Errorf("expected no error for valid root, got: %v", err)
	}
}

func TestLoadFromFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configContent := `
root = "/tmp/test-root"

[services.api]
dir = "api"
procs = ["rails", "worker"]
detect = "config/application.rb"

[services.frontend]
dir = "app"
procs = ["app"]
`
	configPath := filepath.Join(tmpDir, ".omr.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Root != "/tmp/test-root" {
		t.Errorf("expected root '/tmp/test-root', got %q", cfg.Root)
	}

	if len(cfg.Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(cfg.Services))
	}

	api, ok := cfg.Services["api"]
	if !ok {
		t.Fatal("expected api service")
	}
	if api.Dir != "api" {
		t.Errorf("expected api dir 'api', got %q", api.Dir)
	}
	if len(api.Procs) != 2 {
		t.Errorf("expected 2 procs for api, got %d", len(api.Procs))
	}
	if api.Detect != "config/application.rb" {
		t.Errorf("expected detect 'config/application.rb', got %q", api.Detect)
	}
}

func TestValidateRoot_Port(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "config-test")
	defer os.RemoveAll(tmpDir)

	cfg := &Config{
		Root: tmpDir,
		Services: map[string]Service{
			"api": {Dir: "api", Port: 70000},
		},
	}
	if err := cfg.ValidateRoot(); err == nil {
		t.Error("expected error for out-of-range port")
	}

	cfg.Services["api"] = Service{Dir: "api", Port: 0}
	if err := cfg.ValidateRoot(); err != nil {
		t.Errorf("port=0 should be allowed (unset sentinel), got: %v", err)
	}

	cfg.Services["api"] = Service{Dir: "api", Port: 3000}
	if err := cfg.ValidateRoot(); err != nil {
		t.Errorf("valid port should pass, got: %v", err)
	}
}

func TestValidateRoot_DependsOn(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "config-test")
	defer os.RemoveAll(tmpDir)

	// Unknown dep
	cfg := &Config{
		Root: tmpDir,
		Services: map[string]Service{
			"web": {Dir: "web", DependsOn: []string{"api"}},
		},
	}
	if err := cfg.ValidateRoot(); err == nil {
		t.Error("expected error for unknown depends_on target")
	}

	// Valid linear chain
	cfg = &Config{
		Root: tmpDir,
		Services: map[string]Service{
			"api": {Dir: "api"},
			"web": {Dir: "web", DependsOn: []string{"api"}},
		},
	}
	if err := cfg.ValidateRoot(); err != nil {
		t.Errorf("valid chain should pass, got: %v", err)
	}

	// Two-node cycle
	cfg = &Config{
		Root: tmpDir,
		Services: map[string]Service{
			"a": {Dir: "a", DependsOn: []string{"b"}},
			"b": {Dir: "b", DependsOn: []string{"a"}},
		},
	}
	if err := cfg.ValidateRoot(); err == nil {
		t.Error("expected error for two-node cycle")
	}

	// Three-node cycle
	cfg = &Config{
		Root: tmpDir,
		Services: map[string]Service{
			"a": {Dir: "a", DependsOn: []string{"b"}},
			"b": {Dir: "b", DependsOn: []string{"c"}},
			"c": {Dir: "c", DependsOn: []string{"a"}},
		},
	}
	if err := cfg.ValidateRoot(); err == nil {
		t.Error("expected error for three-node cycle")
	}

	// Self-loop
	cfg = &Config{
		Root: tmpDir,
		Services: map[string]Service{
			"a": {Dir: "a", DependsOn: []string{"a"}},
		},
	}
	if err := cfg.ValidateRoot(); err == nil {
		t.Error("expected error for self-loop")
	}
}

func TestLoadNewFields(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configContent := `
root = "/tmp/test-root"

[services.api]
dir = "current"
procs = ["api"]
port = 9002

[services.web]
dir = "current"
procs = ["web"]
port = 3002
depends_on = ["api"]
`
	configPath := filepath.Join(tmpDir, ".omr.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	api := cfg.Services["api"]
	if api.Port != 9002 {
		t.Errorf("api.Port: expected 9002, got %d", api.Port)
	}

	web := cfg.Services["web"]
	if len(web.DependsOn) != 1 || web.DependsOn[0] != "api" {
		t.Errorf("web.DependsOn: expected [api], got %v", web.DependsOn)
	}
}

func TestSetConfigFile(t *testing.T) {
	// Reset state
	configFile = ""

	SetConfigFile("/custom/path.toml")
	if configFile != "/custom/path.toml" {
		t.Errorf("expected '/custom/path.toml', got %q", configFile)
	}

	// Reset for other tests
	configFile = ""
}
