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
	Dir    string   `mapstructure:"dir"`
	Procs  []string `mapstructure:"procs"`
	Detect string   `mapstructure:"detect"`
}

// Config represents the application configuration
type Config struct {
	Root     string             `mapstructure:"root"`
	Services map[string]Service `mapstructure:"services"`
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

// parentDirs returns directories to search for config
func parentDirs() []string {
	var dirs []string
	dir, err := os.Getwd()
	if err != nil {
		return dirs
	}

	for {
		dirs = append(dirs, dir)
		// Stop at config file or project root (Procfile.dev)
		if _, err := os.Stat(filepath.Join(dir, ".omr.toml")); err == nil {
			break
		}
		if _, err := os.Stat(filepath.Join(dir, "Procfile.dev")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return dirs
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
	}
	return nil
}
