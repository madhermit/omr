# OMR - Overmind Worktree Manager

OMR manages git worktree symlinks and restarts [Overmind](https://github.com/DarthSim/overmind) processes for seamless branch switching in development.

## Installation

**With mise** (recommended):

```toml
# Add to your project's mise.toml
[tools]
"ubi:DarthSim/overmind" = "latest"
"ubi:madhermit/omr" = "latest"
```

```bash
mise install
```

To upgrade: `mise upgrade ubi:madhermit/omr` (run `mise cache clear` first if it doesn't see the new version).

**Or download directly**:

```bash
curl -fsSL https://raw.githubusercontent.com/madhermit/omr/main/install.sh | bash
```

**Or with Go**:

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
  restart [services]  Restart overmind processes (no symlink changes)
  switch [branch]     Switch symlinks to branch worktree and restart
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

## Shell Completions

```bash
omr completion fish > ~/.config/fish/completions/omr.fish
omr completion bash >> ~/.bashrc
omr completion zsh >> ~/.zshrc
```

## Configuration

OMR looks for `.omr.toml` by walking up from the current directory. When multiple `.omr.toml` files exist (e.g., inside a worktree and at the project root), the outermost one is preferred.

### Example: Multi-Repo (per-service symlinks)

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

### Example: Monorepo (shared root symlink)

When multiple services share the same `dir`, OMR manages one symlink and restarts all services on switch. Auto-detection falls back to all services, so `--all` is not required:

```toml
[services.api]
dir = "current"
procs = ["api", "worker"]
detect = "config/application.rb"

[services.web]
dir = "current"
procs = ["web"]
detect = "nuxt.config.ts"
```

### Config Options

| Option       | Description                                                               |
| ------------ | ------------------------------------------------------------------------- |
| `dir`        | Symlink path, relative to the config file's directory                     |
| `procs`      | Overmind process names to restart (from your Procfile)                    |
| `detect`     | File to look for when auto-detecting service (optional)                   |
| `port`       | TCP port for HTTP readiness probing after restart (optional)              |
| `depends_on` | Services this one must wait for during a multi-service restart (optional) |
| `root`       | Override the root directory (defaults to config file's directory)         |

### Environment Variables

- `OMR_ROOT` - Override the root directory
- `OMR_CONFIG` - Set config file path

### Readiness probing

When `port` is set, OMR waits for the restarted service's HTTP layer to respond
before moving on. The probe is generic — TCP-connect to `localhost:<port>` (tries
both IPv4 and IPv6 so dev servers that bind only `::1`, like Nuxt by default,
work without extra config), then HTTP `HEAD /`. Any HTTP response (200, 404, 405, …)
counts as ready, so no `/health` endpoint is required.

```toml
[services.api]
port = 9002
```

OMR polls until any HTTP response comes back. There's a small theoretical
window where a dying old process could respond before overmind kills it, but
in interactive dev use the worst case is a one-refresh 502 — far better than
the polling-race hangs a "wait for the port to drop" gate introduces.

### Sequencing

When multiple services restart in one invocation, `depends_on` orders them into
waves. OMR waits for each wave's `port` probes to succeed before starting the
next — this fixes the common 502 race where a fast frontend boots before a slow
backend and serves a request before the backend can answer.

```toml
[services.api]
dir = "current"
port = 9002

[services.web]
dir = "current"
port = 3002
depends_on = ["api"]
```

`depends_on` is only honored among services *being restarted in this invocation*.
If `api` isn't being restarted, `web` assumes it's already up and starts
immediately.

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

# Restart api processes (no symlink change)
omr restart api

# Restart all processes
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
