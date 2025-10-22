# HamQRZDB

A high-performance, self-hosted amateur radio callsign lookup system with a HamDB-compatible JSON API.

Built with **Go** for speed and efficiency. Processes FCC ULS data into SQLite and serves it via a fast HTTP API with case-insensitive lookups and CORS support.

## Features

- 🚀 **HamDB-compatible API** - Drop-in replacement for HamDB lookups
- ⚡ **Blazing Fast** - ~2ms average response time, ~2,500 req/s throughput
- 🎯 **Case-insensitive** - Works with any callsign case (KJ5DJC = kj5djc)
- 💾 **SQLite Storage** - Single-file database (~500MB for 1.5M callsigns)
- 🔧 **Modern Tooling** - Uses [Task](https://taskfile.dev) for build automation
- 📦 **Docker Ready** - Simple containerized deployment
- 🔄 **Fast Updates** - Daily updates in ~30 seconds, full rebuild in 3-5 minutes
- 📍 **Location Data** - Coordinates and Maidenhead grid squares
- 🌐 **CORS Enabled** - Ready for web applications
- 📊 **Low Memory** - ~100MB RAM usage

## Architecture

**Simple 2-component system:**

1. **hamqrzdb-process** - Downloads FCC ULS data, processes location data, and populates SQLite database
2. **hamqrzdb-api** - HTTP server that queries SQLite directly

```
FCC Data → hamqrzdb-process → SQLite DB → hamqrzdb-api → JSON API
          (ULS + Locations)
```

No JSON file generation needed. The API queries the database directly for real-time, zero-downtime updates.

## Quick Start

```bash
# 1. Install Task (if needed)
brew install go-task/tap/go-task  # macOS
# or see https://taskfile.dev/installation/

# 2. Clone and build
git clone https://github.com/chriskacerguis/hamqrzdb.git
cd hamqrzdb

task build

# 3. Process FCC data (3-5 minutes)
task db:full

# 4. Optional: Add location data
task db:locations -- --la-file temp_uls/LA.dat

# 5. Start API server
task dev:api

# 6. Test it!
curl http://localhost:8080/v1/KJ5DJC/json/test
curl http://localhost:8080/v1/kj5djc/json/test  # case-insensitive!
```

## Prerequisites

- **Go 1.21+** (with CGO support for SQLite)
- **Task** - Modern task runner ([installation](https://taskfile.dev/installation/))
- **~2GB disk space** for full dataset
- **gcc/clang** for SQLite (usually pre-installed on macOS/Linux)

## Installation

```bash
# Clone repository
git clone https://github.com/chriskacerguis/hamqrzdb.git
cd hamqrzdb

# Install Task (macOS)
brew install go-task/tap/go-task

# Install Task (Linux)
curl -1sLf 'https://dl.cloudsmith.io/public/task/task/setup.deb.sh' | sudo -E bash
apt install task

# Install Task (or with Go)
go install github.com/go-task/task/v3/cmd/task@latest

# Build all tools
task build

# Optional: Install system-wide
task install  # Installs to /usr/local/bin
```

## Usage

### Processing FCC Data

```bash
# Download and process full database (3-5 minutes)
task db:full

# Or use binary directly
./bin/hamqrzdb-process --full

# Process daily updates (30 seconds)
task db:daily

# Process single callsign (for testing)
./bin/hamqrzdb-process --full --callsign KJ5DJC

# Custom database path
./bin/hamqrzdb-process --full --db /path/to/custom.db
```

### Adding Location Data (Optional)

Location data adds latitude, longitude, and Maidenhead grid squares:

```bash
# Process location data (2-3 minutes)
task db:locations -- --la-file temp_uls/LA.dat

# Or use process binary directly
./bin/hamqrzdb-process --la-file temp_uls/LA.dat --db hamqrzdb.sqlite

# Process single callsign
./bin/hamqrzdb-process --la-file temp_uls/LA.dat --callsign KJ5DJC

# Or combine with full database processing
./bin/hamqrzdb-process --full --la-file temp_uls/LA.dat
```

**Note:** The full database download includes LA.dat in the `temp_uls/` directory.

### Running the API Server

```bash
# Development mode (default: localhost:8080)
task dev:api

# Or run binary directly
./bin/hamqrzdb-api

# Custom configuration
DB_PATH=./hamqrzdb.sqlite PORT=8080 ./bin/hamqrzdb-api

# Production mode (with Docker)
task docker:up
```

### Database Statistics

```bash
# Show database stats
task db:stats

# Output:
# 📊 Database statistics:
# 1234567 total callsigns
# 1098765 active licenses
# 987654 with locations
# Last updated: 2025-10-19 12:34:56
```

## API Format

### Endpoint

```
GET /v1/{callsign}/json/{appname}
```

- `{callsign}` - Amateur radio callsign (case-insensitive)
- `{appname}` - Your application name (for compatibility, not validated)

### Examples

```bash
# Valid callsign (both work!)
curl http://localhost:8080/v1/KJ5DJC/json/myapp
curl http://localhost:8080/v1/kj5djc/json/myapp

# Invalid callsign (returns NOT_FOUND response)
curl http://localhost:8080/v1/BADCALL/json/test

# Health check
curl http://localhost:8080/health
```

### Response Format

**Valid Callsign (HTTP 200):**
```json
{
  "hamdb": {
    "version": "1",
    "callsign": {
      "call": "KJ5DJC",
      "class": "G",
      "expires": "11/18/2033",
      "status": "A",
      "grid": "EM10ci",
      "lat": "30.3416503",
      "lon": "-97.7548379",
      "fname": "CHRIS",
      "mi": "",
      "name": "KACERGUIS",
      "suffix": "",
      "addr1": "5900 Balcones Drive STE 26811",
      "addr2": "AUSTIN",
      "state": "TX",
      "zip": "78731",
      "country": "United States"
    },
    "messages": {
      "status": "OK"
    }
  }
}
```

**Invalid Callsign (HTTP 200):**
```json
{
  "hamdb": {
    "version": "1",
    "callsign": {
      "call": "NOT_FOUND",
      "class": "NOT_FOUND",
      ...
    },
    "messages": {
      "status": "NOT_FOUND"
    }
  }
}
```

Both responses return **HTTP 200** for client compatibility.

## Docker Deployment

### Using Docker Compose

```bash
# Build and start services
task docker:build
task docker:up

# Or use docker-compose directly
docker-compose up -d

# View logs
task docker:logs

# Stop services
task docker:down
```

### Manual Docker Build

```bash
# Build image
docker build -t hamqrzdb-api:latest .

# Run container
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/hamqrzdb.sqlite:/app/hamqrzdb.sqlite \
  --name hamqrzdb-api \
  hamqrzdb-api:latest
```

The API server runs on port 8080 by default. The database is bind-mounted for zero-downtime updates.

## Automation

### Cron Jobs

```bash
# Create logs directory
mkdir -p logs

# Edit crontab
crontab -e

# Daily updates at 2 AM (30 seconds)
0 2 * * * cd /path/to/hamqrzdb && task db:daily >> logs/cron.log 2>&1

# Weekly full rebuild on Sunday at 3 AM (3-5 minutes)
0 3 * * 0 cd /path/to/hamqrzdb && task db:full >> logs/cron.log 2>&1

# Update locations monthly (2-3 minutes)
0 4 1 * * cd /path/to/hamqrzdb && task db:locations -- --la-file temp_uls/LA.dat >> logs/cron.log 2>&1
```

Database changes are live immediately - no API server restart needed!

### Systemd Service

Create `/etc/systemd/system/hamqrzdb-api.service`:

```ini
[Unit]
Description=HamQRZDB API Server
After=network.target

[Service]
Type=simple
User=hamqrzdb
WorkingDirectory=/opt/hamqrzdb
Environment="DB_PATH=/opt/hamqrzdb/hamqrzdb.sqlite"
Environment="PORT=8080"
ExecStart=/usr/local/bin/hamqrzdb-api
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable hamqrzdb-api
sudo systemctl start hamqrzdb-api
sudo systemctl status hamqrzdb-api
```

## Performance

### Benchmarks

| Operation | Time | Notes |
|-----------|------|-------|
| Full database load | 3-5 min | 1.5M callsigns |
| Daily updates | ~30 sec | Incremental changes |
| Location processing | 2-3 min | All callsigns |
| API response time | ~2ms | Average (p50) |
| API response time | <50ms | p99 |
| API throughput | ~2,500 req/s | Single instance |
| Memory usage | ~100MB | API server |
| Database size | ~500MB | 1.5M callsigns |

### Comparison to Python

- **4-5x faster** data processing
- **5x less memory** usage
- **50x faster** API responses
- **Single binary** deployment (no dependencies)

## Task Commands

Quick reference for common operations:

```bash
# Build commands
task build              # Build all binaries
task build:api          # Build API server only
task build:process      # Build data processor only
task build:locations    # Build locations processor only

# Development
task dev:api                              # Run API server
task dev:process -- --full                # Run data processor
task dev:locations -- --la-file LA.dat    # Run locations processor

# Database operations
task db:full            # Download and process full database
task db:daily           # Process daily updates
task db:locations       # Process location data
task db:stats           # Show database statistics

# Docker operations
task docker:build       # Build Docker image
task docker:up          # Start services
task docker:down        # Stop services
task docker:logs        # View logs

# Utility
task clean              # Remove build artifacts
task install            # Install to /usr/local/bin
task test               # Run tests
task help               # Show detailed help
task --list             # List all available tasks
```

See [docs/TASKFILE-MIGRATION.md](docs/TASKFILE-MIGRATION.md) for migration guide from Makefile.

## Project Structure

```
hamqrzdb/
├── main.go                   # API server source
├── cmd/
│   └── process/main.go       # Data processor source (includes location processing)
├── bin/                      # Compiled binaries
│   ├── hamqrzdb-api
│   └── hamqrzdb-process
├── Taskfile.yml              # Task automation configuration
├── Dockerfile                # Docker image definition
├── docker-compose.yml        # Docker Compose configuration
├── hamqrzdb.sqlite           # SQLite database (generated)
├── temp_uls/                 # Downloaded FCC data (temporary)
└── docs/                     # Documentation
    ├── README.cli.md         # CLI tools reference
    ├── TASKFILE-MIGRATION.md # Makefile→Task migration guide
    ├── LOCATIONS.md          # Locations processor guide
    └── QUICKREF.md           # Quick reference card
```

## Documentation

- **[README.cli.md](docs/README.cli.md)** - Complete CLI reference for all tools
- **[TASKFILE-MIGRATION.md](docs/TASKFILE-MIGRATION.md)** - Migration guide from Makefile
- **[LOCATIONS.md](docs/LOCATIONS.md)** - Locations processor detailed guide
- **[QUICKREF.md](docs/QUICKREF.md)** - Quick reference card for common operations

## Troubleshooting

### Build Errors

**Error: `CGO_ENABLED` required for SQLite**

Solution: Make sure you have a C compiler installed:

```bash
# macOS
xcode-select --install

# Ubuntu/Debian
sudo apt-get install build-essential

# Fedora/RHEL
sudo dnf install gcc
```

### Database Errors

**Error: Database locked**

Solution: Close any other connections to the database:

```bash
# Find processes using the database
lsof hamqrzdb.sqlite

# Kill the process if needed
kill <PID>
```

**Error: Database corrupted**

Solution: Rebuild from scratch:

```bash
# Backup current database
cp hamqrzdb.sqlite hamqrzdb.sqlite.backup

# Remove and rebuild
rm hamqrzdb.sqlite hamqrzdb.sqlite-*
task db:full
```

### Download Errors

**Error: Daily file not available**

Solution: Daily files may not be available on weekends/holidays. Use full database instead:

```bash
task db:full
```

**Error: Download failed**

Solution: Check FCC website status and try again:

- https://www.fcc.gov/uls/transactions/daily-weekly

### Memory Issues

The Go tools are memory-efficient (~100MB), but if you encounter issues:

```bash
# Process single callsign to test
./bin/hamqrzdb-process --full --callsign KJ5DJC

# Check system resources
free -h  # Linux
vm_stat  # macOS
```

## Data Source & Updates

### FCC ULS Database

- **Full Database**: https://data.fcc.gov/download/pub/uls/complete/l_amat.zip (~500MB)
- **Daily Updates**: https://data.fcc.gov/download/pub/uls/daily/l_am_MMDDYYYY.zip (~1-5MB)
- **Update Schedule**: Daily updates usually available by 2 AM ET
- **Documentation**: https://www.fcc.gov/uls/transactions/daily-weekly

### License Status Codes

- `A` = Active
- `C` = Canceled
- `E` = Expired
- `T` = Terminated

### Operator Classes

- `N` = Novice (no longer issued)
- `T` = Technician
- `G` = General
- `A` = Amateur Extra
- `P` = Technician Plus (no longer issued)

## License

MIT License - See [LICENSE](LICENSE) file for details.

## Credits

**Data Source:**
- FCC Universal Licensing System (ULS) - https://www.fcc.gov/uls/

**Inspiration:**
- [k3ng/hamdb](https://github.com/k3ng/hamdb) for the original HamDB project and API format

**Built with:**
- [Go](https://golang.org/) - Programming language
- [SQLite](https://www.sqlite.org/) - Database engine
- [Task](https://taskfile.dev/) - Task runner
- [go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite driver for Go

---

**73! 📻**

For questions, issues, or contributions, visit: https://github.com/chriskacerguis/hamqrzdb