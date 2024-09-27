# Overmind Manager/Restarter

`omr` is a bash script that simplifies the process of restarting [Overmind](https://github.com/DarthSim/overmind) processes in a Rails or Nuxt.js project. It's specifically designed to support a Git worktree-style workflow, allowing easy management of multiple branches in separate directories. The script automatically detects the project type and restarts the appropriate processes, making it easier to manage your development environment across different branches.

## Features

- Supports Git worktree-style workflow for managing multiple branches
- Uses symlinks to make directory changes persistent
- Automatically detects Rails or Nuxt.js projects
- Restarts appropriate Overmind processes
- Supports custom process specification
- Quiet mode for suppressing non-error output
- Color-coded output for better readability

## Prerequisites

- Unix-like operating system (Linux, macOS, etc.)
- Git
- [Overmind](https://github.com/DarthSim/overmind) installed and configured for your project

## Installation

1. Download the `omr` script:

   ```sh
   curl -o omr https://raw.githubusercontent.com/yourusername/omr-script/main/omr
   ```

2. Make the script executable:

   ```sh
   chmod +x omr
   ```

3. Move the script to a directory in your PATH:

   ```sh
   sudo mv omr /usr/local/bin/
   ```

   Note: You may need to use `sudo` depending on the permissions of your `/usr/local/bin/` directory.

## Usage

```sh
omr [options] [process]
```

### Options

- `-h, --help`: Show help message and exit
- `-q, --quiet`: Suppress non-error output

### Process

- `rails`: Restart Rails and worker processes
- `app`: Restart the app process (for Nuxt.js)

If not specified, the process will be auto-detected based on the repository structure.

### Examples

1. Auto-detect and restart with verbose output (default):

   ```sh
   omr
   ```

2. Restart Rails with quiet output:

   ```sh
   omr -q rails
   ```

3. Restart Nuxt.js app:

   ```sh
   omr app
   ```

## Environment Variables

- `FIRST_ROOT_DIR`: The root directory for the project (required)
- `FIRST_API_DIR`: The API directory name (default: api)
- `FIRST_NUXT_DIR`: The Nuxt directory name (default: app)

## Git Worktree Workflow and Symlink Usage

This script is designed to work seamlessly with a Git worktree setup. It allows you to easily switch between different branches of your project, each in its own directory, and restart the appropriate Overmind processes for that branch. This is particularly useful for projects where you need to maintain multiple versions or feature branches simultaneously.

Key points about how `omr` works with Git worktrees:

1. The script uses symlinks to make directory changes persistent. This means that when you switch between branches, the appropriate directories are automatically linked to your current working branch.

2. The `FIRST_ROOT_DIR` environment variable should point to a directory that contains symlinks to your different branch directories. The script updates these symlinks as you switch branches.

3. When you run `omr`, it creates or updates symlinks in `FIRST_ROOT_DIR` to point to the current Git worktree. This ensures that Overmind always runs processes for your current branch.

To use `omr` with Git worktrees:

1. Set up your Git worktrees for different branches.
2. Navigate to the worktree directory for the branch you want to work on.
3. Run `omr` to update the symlinks and restart the Overmind processes for that specific branch.

The script will automatically detect the current branch and project type, update the necessary symlinks, and restart the appropriate processes.

## Troubleshooting

If you encounter any issues running the script, ensure that:

1. The script has execute permissions (`chmod +x omr`).
2. The script is in a directory listed in your PATH.
3. All required environment variables are set correctly.
4. You're running the script from within a Git repository.
5. Your user has permissions to create and modify symlinks in the `FIRST_ROOT_DIR`.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This script is open-source software licensed under the MIT license.
