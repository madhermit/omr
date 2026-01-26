# OMR - Overmind Restart

OMR manages git worktree symlinks and restarts overmind processes for seamless branch switching in development.

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/madhermit/omr/main/install.sh | bash
```

Or with Go:
```bash
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
   omr status               # Show current status
   omr switch               # Switch detected service to current worktree
   omr switch --all main    # Switch all services to main worktree
   omr restart api          # Restart the api service
   ```

## Commands

```
omr [global-flags] <command> [args]

Commands:
  status              Show current service status
  restart [services]  Restart services (auto-detects if none specified)
  switch [branch]     Switch worktree and restart services
  init                Generate example config file
  version             Show version info

Global Flags:
  -h, --help           Help
  -q, --quiet          Suppress output
  -c, --config FILE    Config file path

Restart Flags:
  -a, --all            Restart all services

Switch Flags:
  -a, --all            Switch all services
```

## Configuration

OMR looks for `.omr.toml` in the current directory and parent directories.

### Example Config

```toml
[services.api]
dir = "first-api/current"           # Symlink path (relative to config file)
procs = ["rails", "worker"]         # Overmind process names to restart
detect = "config/application.rb"    # Auto-detection file (optional)

[services.frontend]
dir = "first-nuxt/current"
procs = ["app"]
detect = "nuxt.config.ts"
```

### Config Options

| Option | Description |
|--------|-------------|
| `dir` | Symlink path, relative to the config file's directory |
| `procs` | Overmind process names to restart (from your Procfile) |
| `detect` | File to look for when auto-detecting service (optional) |
| `root` | Override the root directory (defaults to config file's directory) |

### Environment Variables

- `OMR_ROOT` - Override the root directory
- `OMR_CONFIG` - Set config file path

## How It Works

1. **Symlink Management**: OMR creates/updates symlinks pointing to git worktrees
2. **Process Restart**: After updating symlinks, OMR restarts the configured overmind processes
3. **Auto-detection**: When you run `omr switch` from inside a worktree, OMR uses the `detect` files to identify which service you're in and switches just that symlink

## Examples

```bash
# Show current status of all services
omr status

# cd into a worktree, then switch that service's symlink to point here
cd first-api/feature-branch
omr switch

# Switch all services to the main branch worktree
omr switch --all main

# Restart api service (updates symlink, restarts rails and worker)
omr restart api

# Restart all configured services
omr restart --all
```

## Requirements

- [Overmind](https://github.com/DarthSim/overmind) process manager
- Git with worktree support

## Development

```bash
mise install       # install Go 1.25
mise run test      # run tests
mise run build     # build binary
mise run check     # fmt + lint + test
mise run release   # build all platform binaries
```

## License

MIT
