package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/viper"
)

// Service represents a configured service
type Service struct {
	Dir       string   `mapstructure:"dir"`
	Procs     []string `mapstructure:"procs"`
	Detect    string   `mapstructure:"detect"`
	Port      int      `mapstructure:"port"`
	DependsOn []string `mapstructure:"depends_on"`
}

// Config represents the application configuration
type Config struct {
	Root     string             `mapstructure:"root"`
	Services map[string]Service `mapstructure:"services"`
}

// HasSingleDir returns true if all services share the same dir
func (c *Config) HasSingleDir() bool {
	var dir string
	for _, svc := range c.Services {
		if dir == "" {
			dir = svc.Dir
		} else if svc.Dir != dir {
			return false
		}
	}
	return true
}

// ProcsForDir returns all process names from services that share the given dir
func (c *Config) ProcsForDir(dir string) []string {
	var procs []string
	for _, svc := range c.Services {
		if svc.Dir == dir {
			procs = append(procs, svc.Procs...)
		}
	}
	return procs
}

var configFile string

// SetConfigFile sets the config file path (from --config flag)
func SetConfigFile(file string) {
	configFile = file
}

// Load loads the configuration from various sources
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName(".omr")
	v.SetConfigType("toml")

	// Walk up from cwd looking for .omr.toml
	for _, dir := range parentDirs() {
		v.AddConfigPath(dir)
	}

	if configFile != "" {
		v.SetConfigFile(configFile)
	}

	v.SetEnvPrefix("OMR")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	handleLegacyEnvVars(v)

	// Default root to config file's directory
	if v.GetString("root") == "" && v.ConfigFileUsed() != "" {
		v.Set("root", filepath.Dir(v.ConfigFileUsed()))
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

// parentDirs returns the outermost directory containing .omr.toml.
// Walks up from cwd to find all .omr.toml files, returns the outermost one
// so the project root is preferred over nested worktree copies.
func parentDirs() []string {
	var outermost string
	dir, err := os.Getwd()
	if err != nil {
		return nil
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".omr.toml")); err == nil {
			outermost = dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	if outermost != "" {
		return []string{outermost}
	}
	return nil
}

func handleLegacyEnvVars(v *viper.Viper) {
	if envConfig := os.Getenv("OMR_CONFIG"); envConfig != "" && configFile == "" {
		configFile = envConfig
		v.SetConfigFile(envConfig)
	}
}

// ValidateService checks if a service name is valid
func (c *Config) ValidateService(name string) error {
	if _, ok := c.Services[name]; !ok {
		return fmt.Errorf("unknown service: %s", name)
	}
	return nil
}

// ServiceNames returns all configured service names (sorted)
func (c *Config) ServiceNames() []string {
	names := make([]string, 0, len(c.Services))
	for name := range c.Services {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ValidateRoot checks if the root directory is configured and exists
func (c *Config) ValidateRoot() error {
	if c.Root == "" {
		return fmt.Errorf("root directory not configured (set 'root' in config or OMR_ROOT env var)")
	}
	if _, err := os.Stat(c.Root); os.IsNotExist(err) {
		return fmt.Errorf("root directory does not exist: %s", c.Root)
	}
	for name, svc := range c.Services {
		if svc.Dir == "" {
			return fmt.Errorf("service %q is missing required 'dir' field", name)
		}
		if svc.Port < 0 || svc.Port > 65535 {
			return fmt.Errorf("service %q has invalid port %d (must be 1-65535, or 0 for unset)", name, svc.Port)
		}
	}
	return c.validateDeps()
}

// validateDeps checks that depends_on references known services and contains no cycles.
// Uses Kahn's algorithm — any node with non-zero in-degree at the end is in a cycle.
func (c *Config) validateDeps() error {
	inDegree := map[string]int{}
	for name := range c.Services {
		inDegree[name] = 0
	}
	for name, svc := range c.Services {
		for _, dep := range svc.DependsOn {
			if _, ok := c.Services[dep]; !ok {
				return fmt.Errorf("service %q depends on unknown service %q", name, dep)
			}
			inDegree[name]++
		}
	}

	queue := []string{}
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}
	visited := 0
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		visited++
		for other, svc := range c.Services {
			for _, dep := range svc.DependsOn {
				if dep == name {
					inDegree[other]--
					if inDegree[other] == 0 {
						queue = append(queue, other)
					}
				}
			}
		}
	}
	if visited != len(c.Services) {
		return fmt.Errorf("depends_on cycle detected among services")
	}
	return nil
}
