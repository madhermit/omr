package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/fatih/color"
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

	// Config search paths
	v.AddConfigPath(".")
	if gitRoot, err := findGitRoot(); err == nil {
		v.AddConfigPath(gitRoot)
	}
	if home, err := os.UserHomeDir(); err == nil {
		v.AddConfigPath(home)
		v.AddConfigPath(filepath.Join(home, ".config", "omr"))
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

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

func handleLegacyEnvVars(v *viper.Viper) {
	if legacyRoot := os.Getenv("FIRST_ROOT_DIR"); legacyRoot != "" {
		if os.Getenv("OMR_ROOT") == "" {
			yellow := color.New(color.FgYellow).SprintFunc()
			fmt.Fprintf(os.Stderr, "%s FIRST_ROOT_DIR is deprecated, use OMR_ROOT instead\n", yellow("Warning:"))
			v.Set("root", legacyRoot)
		}
	}

	if envConfig := os.Getenv("OMR_CONFIG"); envConfig != "" && configFile == "" {
		configFile = envConfig
		v.SetConfigFile(envConfig)
	}
}

func findGitRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not in a git repository")
		}
		dir = parent
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
	return nil
}
