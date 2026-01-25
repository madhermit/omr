package state

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

const stateFileName = ".omr-state.toml"

// State represents the current omr state
type State struct {
	Worktree  string `mapstructure:"worktree"`
	Branch    string `mapstructure:"branch"`
	Service   string `mapstructure:"service"`
	Timestamp time.Time
}

// stateFile is the on-disk representation
type stateFile struct {
	Worktree  string `mapstructure:"worktree"`
	Branch    string `mapstructure:"branch"`
	Service   string `mapstructure:"service"`
	Timestamp string `mapstructure:"timestamp"`
}

// Read reads the state file from the given root directory
func Read(root string) (*State, error) {
	statePath := filepath.Join(root, stateFileName)

	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return nil, nil // No state file, not an error
	}

	v := viper.New()
	v.SetConfigFile(statePath)
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var sf stateFile
	if err := v.Unmarshal(&sf); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	state := &State{
		Worktree: sf.Worktree,
		Branch:   sf.Branch,
		Service:  sf.Service,
	}

	// Parse timestamp
	if sf.Timestamp != "" {
		ts, err := time.Parse(time.RFC3339, sf.Timestamp)
		if err == nil {
			state.Timestamp = ts
		}
	}

	return state, nil
}

// Write writes the state file to the given root directory
func Write(root string, state *State) error {
	statePath := filepath.Join(root, stateFileName)

	// Set timestamp
	state.Timestamp = time.Now()

	// Create content
	content := fmt.Sprintf(`# OMR State File - Auto-generated
worktree = %q
branch = %q
service = %q
timestamp = %q
`, state.Worktree, state.Branch, state.Service, state.Timestamp.Format(time.RFC3339))

	if err := os.WriteFile(statePath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// GetStatePath returns the path to the state file
func GetStatePath(root string) string {
	return filepath.Join(root, stateFileName)
}
