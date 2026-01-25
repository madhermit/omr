# OMR - Overmind Restart

OMR manages git worktree symlinks and restarts overmind processes for seamless branch switching in development.

## Installation

```bash
# Build from source
go build -o omr .

# Or install directly
go install github.com/madhermit/omr@latest
```

## Quick Start

1. Generate a config file:
   ```bash
   omr init > .omr.toml
   ```

2. Edit `.omr.toml` to match your project structure

3. Use omr to switch branches and restart services:
   ```bash
   omr restart api          # Restart the api service
   omr restart --all        # Restart all services
   omr switch main          # Switch to main worktree
   omr status               # Show current status
   ```

## Commands

```
omr [global-flags] <command> [args]

Commands:
  status              Show current service status
  restart [services]  Restart services (auto-detects if none specified)
  switch <worktree>   Switch worktree and restart all services
  init                Generate example config file
  version             Show version info

Global Flags:
  -h, --help           Help
  -q, --quiet          Suppress output
  -c, --config FILE    Config file path

Restart Flags:
  -a, --all            Restart all services
```

## Configuration

OMR looks for configuration in these locations (in order):

1. `--config` flag
2. `.omr.toml` in current directory
3. `.omr.toml` in git repository root
4. `~/.omr.toml`
5. `~/.config/omr/config.toml`

### Example Config

```toml
# Root directory where symlinks are managed
root = "/path/to/active"

[services.api]
dir = "api"                         # Symlink name in root
procs = ["rails", "worker"]         # Overmind process names
detect = "config/application.rb"    # Auto-detection file (optional)

[services.frontend]
dir = "app"
procs = ["app"]
detect = "nuxt.config.ts"
```

### Environment Variables

- `OMR_ROOT` - Override the root directory
- `OMR_CONFIG` - Set config file path
- `FIRST_ROOT_DIR` - Legacy alias for `OMR_ROOT` (deprecated, shows warning)

## How It Works

1. **Symlink Management**: OMR creates/updates symlinks in your root directory pointing to git worktrees
2. **Process Restart**: After updating symlinks, OMR restarts the configured overmind processes
3. **Auto-detection**: When no service is specified, OMR detects the service type based on marker files

## Examples

```bash
# Show current status of all services
omr status

# Restart api service (updates symlink, restarts rails and worker)
omr restart api

# Restart all configured services
omr restart --all

# Switch all services to the main branch worktree
omr switch main

# Auto-detect and restart (based on current directory)
omr restart

# Quiet mode (suppress output)
omr -q restart api

# Use custom config file
omr -c ~/myconfig.toml status
```

## Requirements

- Go 1.21+ (for building)
- [Overmind](https://github.com/DarthSim/overmind) process manager
- Git with worktree support

## License

MIT
