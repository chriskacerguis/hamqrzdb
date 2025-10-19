# Migration to Task (Taskfile.yml)

## What Changed

The project has migrated from `Makefile` and `build.sh` to [Task](https://taskfile.dev), a modern task runner and build tool.

### Why Task?

- **Simpler Syntax**: YAML instead of Makefile's complex syntax
- **Cross-Platform**: Works identically on macOS, Linux, and Windows
- **Built-in Features**: Variables, dependencies, file watching, and more
- **Better Feedback**: Cleaner output and progress indicators
- **Incremental Builds**: Automatic source/generate tracking

## Installation

Install Task if you haven't already:

```bash
# macOS
brew install go-task/tap/go-task

# Linux
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d

# Or with Go
go install github.com/go-task/task/v3/cmd/task@latest
```

## Command Migration Guide

### Build Commands

| Old Command | New Command | Notes |
|------------|-------------|-------|
| `make` or `make build` | `task` or `task build` | Build all binaries |
| `make build-api` | `task build:api` | Build API server |
| `make build-process` | `task build:process` | Build data processor |
| `make build-locations` | `task build:locations` | Build locations processor |
| `./build.sh` | `task build` | Replaced by task |

### Development Commands

| Old Command | New Command | Notes |
|------------|-------------|-------|
| `make dev-api` | `task dev:api` | Run API in dev mode |
| `make dev-process` | `task dev:process` | Run processor in dev mode |
| `make dev-locations` | `task dev:locations` | Run locations in dev mode |

### Database Commands

| Old Command | New Command | Notes |
|------------|-------------|-------|
| `make db-full` | `task db:full` | Download and process full database |
| `make db-daily` | `task db:daily` | Process daily updates |
| `make db-generate` | `task db:generate` | Generate JSON from database |
| `make db-stats` | `task db:stats` | Show database statistics |
| N/A | `task db:locations` | Process location data |

### Docker Commands

| Old Command | New Command | Notes |
|------------|-------------|-------|
| `make docker-build` | `task docker:build` | Build Docker image |
| `make docker-run` | `task docker:up` | Start services |
| N/A | `task docker:down` | Stop services |
| N/A | `task docker:logs` | View service logs |
| N/A | `task docker:restart` | Restart services |

### Utility Commands

| Old Command | New Command | Notes |
|------------|-------------|-------|
| `make clean` | `task clean` | Remove build artifacts |
| `make install` | `task install` | Install to /usr/local/bin |
| N/A | `task uninstall` | Remove from /usr/local/bin |
| N/A | `task deps` | Download Go dependencies |
| N/A | `task tidy` | Tidy Go modules |
| N/A | `task test` | Run tests |
| N/A | `task fmt` | Format Go code |
| N/A | `task vet` | Run go vet |
| N/A | `task lint` | Run linter |
| N/A | `task check` | Run all checks |

### Help Commands

| Old Command | New Command | Notes |
|------------|-------------|-------|
| `make help` | `task help` | Show detailed help |
| N/A | `task --list` | List all available tasks |
| N/A | `task version` | Show version info |
| N/A | `task info` | Show binary sizes |

## New Features

### Task Colons vs Makefile Hyphens

Task uses colons (`:`) to namespace tasks, while Makefile used hyphens (`-`):

```bash
# Makefile style
make dev-api

# Task style  
task dev:api
```

### Passing Arguments

Task makes it easier to pass arguments to commands:

```bash
# Run processor with arguments
task dev:process -- --full --callsign KJ5DJC

# Run locations with arguments
task dev:locations -- --la-file LA.dat --regenerate
```

### Smart Rebuilds

Task automatically tracks source files and only rebuilds when needed:

```bash
$ task build:api
task: [build:api] echo "🔨 Building hamqrzdb-api..."
🔨 Building hamqrzdb-api...
task: [build:api] CGO_ENABLED=1 go build -ldflags="-s -w" -o bin/hamqrzdb-api main.go
✓ Built bin/hamqrzdb-api

$ task build:api
task: Task "build:api" is up to date
```

### Parallel Builds

Task automatically runs independent tasks in parallel:

```bash
$ task build
task: [build:api] echo "🔨 Building hamqrzdb-api..."
task: [build:process] echo "🔨 Building hamqrzdb-process..."
task: [build:locations] echo "🔨 Building hamqrzdb-locations..."
```

### Better Output

Task provides cleaner, more informative output with emojis and colors:

```
✅ Build complete!
📊 Binary sizes
🚀 Starting API server...
📡 Downloading and processing...
🧪 Running tests...
```

## Common Workflows

### First Time Setup

```bash
# Install dependencies
task deps

# Build everything
task build

# Optionally install to system
task install
```

### Development Workflow

```bash
# Download and process FCC data
task dev:process -- --full

# Add location data
task dev:locations -- --la-file temp_uls/LA.dat

# Start API server
task dev:api

# In another terminal, test API
task test:api
```

### Production Build

```bash
# Clean and rebuild
task clean
task build

# Check database stats
task db:stats

# Build Docker image
task docker:build

# Start services
task docker:up
```

### Daily Updates

```bash
# Process daily FCC updates
task db:daily

# Or use in cron:
# 0 2 * * * cd /path/to/hamqrzdb && task db:daily
```

### Development Checks

```bash
# Format, vet, and test
task check

# Or individually
task fmt
task vet
task test
```

## Tips

1. **Tab Completion**: Task supports shell completion. See [Task Shell Completion](https://taskfile.dev/installation/#setup-completions).

2. **Watch Mode**: Task can watch files and auto-rebuild (requires `--watch` flag):
   ```bash
   task --watch build:api
   ```

3. **Dry Run**: See what would run without executing:
   ```bash
   task --dry build
   ```

4. **Force Run**: Force a task to run even if up-to-date:
   ```bash
   task --force build:api
   ```

5. **Verbose Mode**: See all commands being executed:
   ```bash
   task --verbose build
   ```

6. **Task Graph**: Visualize task dependencies:
   ```bash
   task --list-all
   ```

## Taskfile.yml Structure

The Taskfile is located at the root of the project and contains:

- **vars**: Global variables (BIN_DIR, binary names, etc.)
- **tasks**: All available tasks organized by namespace:
  - `build:*` - Build tasks
  - `dev:*` - Development tasks
  - `docker:*` - Docker tasks
  - `db:*` - Database tasks
  - `test:*` - Testing tasks
  - Top-level tasks (clean, install, help, etc.)

## Troubleshooting

### Task Not Found

```bash
# Make sure Task is installed
which task

# Install if missing
brew install go-task/tap/go-task
```

### Task Version

```bash
# Check Task version (requires v3)
task --version

# Upgrade if needed
brew upgrade go-task/tap/go-task
```

### SQLite Build Errors

Task automatically sets `CGO_ENABLED=1` for SQLite support. If you get build errors:

```bash
# Make sure you have a C compiler
gcc --version  # or clang --version

# On macOS, install Xcode Command Line Tools
xcode-select --install
```

## Resources

- [Task Documentation](https://taskfile.dev)
- [Task GitHub Repository](https://github.com/go-task/task)
- [Task Installation Guide](https://taskfile.dev/installation/)
- [Taskfile Schema](https://taskfile.dev/api/)

## Backward Compatibility

The old `Makefile` and `build.sh` have been removed. If you need them for reference:

```bash
git show HEAD~1:Makefile
git show HEAD~1:build.sh
```

All functionality has been preserved and enhanced in the Taskfile.yml.

---

**73! 📻**
