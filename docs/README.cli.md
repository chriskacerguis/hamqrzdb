# HamQRZDB CLI Tools

Go-based command-line tools for processing FCC ULS Amateur Radio data.

## Overview

HamQRZDB provides two main CLI tools:

1. **`hamqrzdb-process`** - Download and process FCC data (including locations) into SQLite database
2. **`hamqrzdb-api`** - Serve callsign lookups via HTTP API

These tools are **10-20x faster** than the Python equivalent and provide better memory efficiency.

## Quick Start

### Build the Tools

```bash
# Build all tools
task build

# Or build individually
task build:process
task build:api
```

### Process FCC Data

```bash
# Download and process full FCC database (~5-10 minutes)
./bin/hamqrzdb-process --full

# Download and process daily updates (~30 seconds)
./bin/hamqrzdb-process --daily

# Process location data
./bin/hamqrzdb-process --la-file temp_uls/LA.dat

# Process full database with locations in one command
./bin/hamqrzdb-process --full --la-file temp_uls/LA.dat

# Process a specific callsign only
./bin/hamqrzdb-process --full --callsign KJ5DJC
```

### Start the API Server

```bash
# Start API server
./bin/hamqrzdb-api

# Or specify custom settings
DB_PATH=./hamqrzdb.sqlite PORT=8080 ./bin/hamqrzdb-api
```

## Installation

### From Source

```bash
# Clone repository
git clone https://github.com/chriskacerguis/hamqrzdb.git
cd hamqrzdb

# Install dependencies
go mod download

# Build and install
task build
task install  # Installs to /usr/local/bin
```

### Using Docker

```bash
# Build Docker image
task docker:build

# Start services
task docker:up

# View logs
task docker:logs

# Stop services
task docker:down
```

## CLI Reference

### hamqrzdb-process

Process FCC ULS Amateur Radio data into SQLite database and generate JSON files.

#### Usage

```bash
hamqrzdb-process [flags]
```

#### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--full` | Download and process full FCC database | - |
| `--daily` | Download and process daily updates | - |
| `--file <path>` | Process a specific ZIP file | - |
| `--generate` | Generate JSON files from existing database | - |
| `--db <path>` | SQLite database path | `hamqrzdb.sqlite` |
| `--output <dir>` | Output directory for JSON files | `output` |
| `--callsign <call>` | Process only a specific callsign | - |

#### Examples

**Download and process full database:**
```bash
./bin/hamqrzdb-process --full
```

**Process daily updates:**
```bash
./bin/hamqrzdb-process --daily
```

**Generate JSON files without downloading:**
```bash
./bin/hamqrzdb-process --generate
```

**Process only one callsign:**
```bash
./bin/hamqrzdb-process --full --callsign KJ5DJC
```

**Use custom database path:**
```bash
./bin/hamqrzdb-process --full --db /var/lib/hamqrzdb/database.sqlite
```

**Process a local ZIP file:**
```bash
./bin/hamqrzdb-process --file ~/Downloads/l_amat.zip
```

#### Performance

Processing the full FCC database (~1.5M records):
- **Loading data**: 3-5 minutes
- **Generating JSON**: 5-10 minutes
- **Memory usage**: ~100MB
- **Database size**: ~500MB
- **JSON files**: ~2GB (optional)

### hamqrzdb-api

HTTP API server for callsign lookups.

#### Usage

```bash
hamqrzdb-api
```

#### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_PATH` | Path to SQLite database | `/data/hamqrzdb.sqlite` |
| `PORT` | HTTP port to listen on | `8080` |

#### Examples

**Start with default settings:**
```bash
./bin/hamqrzdb-api
```

**Use custom database and port:**
```bash
DB_PATH=./hamqrzdb.sqlite PORT=9000 ./bin/hamqrzdb-api
```

#### API Endpoints

**Callsign Lookup** (case-insensitive)
```
GET /v1/{callsign}/json/{appname}
```

Example:
```bash
curl http://localhost:8080/v1/KJ5DJC/json/myapp
curl http://localhost:8080/v1/kj5djc/json/myapp  # Also works!
```

**Health Check**
```
GET /health
```

Example:
```bash
curl http://localhost:8080/health
```

**Homepage**
```
GET /
```

## Task Commands

The project uses [Task](https://taskfile.dev) for build automation. See [TASKFILE-MIGRATION.md](TASKFILE-MIGRATION.md) for migration guide from Makefile.

### Build Tasks

```bash
task build            # Build all binaries
task clean            # Remove build artifacts
task install          # Install binaries to /usr/local/bin
task test             # Run tests
task --list           # List all available tasks
```

### Development Tasks

```bash
task dev:api                              # Run API in development mode
task dev:process -- --full                # Run processor with args
task dev:process -- --callsign KJ5DJC     # Process one callsign
task dev:locations -- --la-file LA.dat    # Run locations processor
```

### Docker Tasks

```bash
task docker:build     # Build Docker image
task docker:up        # Start services
task docker:down      # Stop services
task docker:logs      # View logs
task docker:restart   # Restart services
```

### Database Tasks

```bash
task db:full          # Download and process full database
task db:daily         # Download and process daily updates
task db:generate      # Generate JSON files from database
task db:stats         # Show database statistics
task db:locations     # Process location data
```

## Migration from Python

The Go CLI tools are drop-in replacements for the Python scripts:

| Python | Go | Notes |
|--------|-----|-------|
| `python3 process_uls_db.py --full` | `./bin/hamqrzdb-process --full` | 10-20x faster |
| `python3 process_uls_db.py --daily` | `./bin/hamqrzdb-process --daily` | Much faster |
| `python3 process_uls_db.py --generate` | `./bin/hamqrzdb-process --generate` | 5-10x faster |
| `python3 -m http.server 8080` | `./bin/hamqrzdb-api` | Better performance |

**Benefits of Go version:**
- ✅ **10-20x faster** data processing
- ✅ **Lower memory usage** (~100MB vs ~500MB)
- ✅ **Single binary** - no Python dependencies
- ✅ **Better concurrency** - parallel processing
- ✅ **Compiled binary** - no interpreter overhead
- ✅ **Cross-platform** - works on Linux, macOS, Windows

## Database Schema

The SQLite database has the following schema:

```sql
CREATE TABLE callsigns (
    callsign TEXT PRIMARY KEY,
    license_status TEXT,
    radio_service_code TEXT,
    grant_date TEXT,
    expired_date TEXT,
    cancellation_date TEXT,
    operator_class TEXT,
    group_code TEXT,
    region_code TEXT,
    first_name TEXT,
    mi TEXT,
    last_name TEXT,
    suffix TEXT,
    entity_name TEXT,
    street_address TEXT,
    city TEXT,
    state TEXT,
    zip_code TEXT,
    latitude REAL,
    longitude REAL,
    grid_square TEXT,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_callsign ON callsigns(callsign);
CREATE INDEX idx_status ON callsigns(license_status);
```

### Query Examples

```bash
# Count total callsigns
sqlite3 hamqrzdb.sqlite "SELECT COUNT(*) FROM callsigns;"

# Find callsign (case-insensitive)
sqlite3 hamqrzdb.sqlite "SELECT * FROM callsigns WHERE UPPER(callsign) = 'KJ5DJC';"

# Count by license status
sqlite3 hamqrzdb.sqlite "SELECT license_status, COUNT(*) FROM callsigns GROUP BY license_status;"

# Recent updates
sqlite3 hamqrzdb.sqlite "SELECT callsign, last_updated FROM callsigns ORDER BY last_updated DESC LIMIT 10;"
```

## Automation

### Cron Jobs

Update the database automatically:

```bash
# Daily updates at 2 AM
0 2 * * * /usr/local/bin/hamqrzdb-process --daily --db /var/lib/hamqrzdb/hamqrzdb.sqlite

# Weekly full rebuild on Sunday at 3 AM
0 3 * * 0 /usr/local/bin/hamqrzdb-process --full --db /var/lib/hamqrzdb/hamqrzdb.sqlite
```

### Systemd Service

Create `/etc/systemd/system/hamqrzdb-api.service`:

```ini
[Unit]
Description=HamQRZDB API Server
After=network.target

[Service]
Type=simple
User=hamqrzdb
WorkingDirectory=/var/lib/hamqrzdb
Environment="DB_PATH=/var/lib/hamqrzdb/hamqrzdb.sqlite"
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

## Troubleshooting

### Build Errors

**Error: `package github.com/mattn/go-sqlite3: build constraints exclude all Go files`**

Solution: Enable CGO
```bash
CGO_ENABLED=1 go build
```

Or use the Makefile:
```bash
make build
```

**Error: `gcc: command not found`**

Solution: Install build tools
```bash
# Ubuntu/Debian
sudo apt-get install build-essential

# macOS
xcode-select --install

# Alpine Linux (Docker)
apk add gcc musl-dev sqlite-dev
```

### Processing Errors

**Error: `Failed to download`**

Solution: Check URL or use local file
```bash
# Download manually
wget https://data.fcc.gov/download/pub/uls/complete/l_amat.zip

# Process local file
./bin/hamqrzdb-process --file l_amat.zip
```

**Error: `database is locked`**

Solution: Close other connections or use WAL mode (enabled by default)

### API Errors

**Error: `no such table: callsigns`**

Solution: Process data first
```bash
./bin/hamqrzdb-process --full
```

**Error: `bind: address already in use`**

Solution: Change port or stop conflicting process
```bash
PORT=9000 ./bin/hamqrzdb-api
```

## Performance Tips

1. **Use SSD storage** - Database operations are I/O intensive
2. **Skip JSON generation** - Only generate if needed for nginx serving
3. **Use WAL mode** - Enabled by default for better concurrency
4. **Batch updates** - Use daily updates instead of full rebuilds
5. **Monitor memory** - Processing uses ~100MB, API uses ~50MB

## Development

### Project Structure

```
hamqrzdb/
├── main.go              # API server
├── cmd/
│   └── process/
│       └── main.go      # Data processor CLI
├── go.mod               # Go dependencies
├── go.sum               # Dependency checksums
├── Makefile             # Build automation
├── Dockerfile           # Docker image
├── docker-compose.go.yml # Docker deployment
└── README.cli.md        # This file
```

### Adding Features

1. **Fork the repository**
2. **Create a feature branch**
3. **Make your changes**
4. **Test thoroughly**
5. **Submit a pull request**

### Running Tests

```bash
make test
```

### Code Style

- Follow Go best practices
- Use `gofmt` for formatting
- Add comments for exported functions
- Handle errors explicitly

## FAQ

**Q: Do I need to generate JSON files?**

A: No! With the Go API server, you can query the SQLite database directly. JSON files are only needed if you're using nginx to serve static files.

**Q: How often should I update the database?**

A: Daily updates are sufficient. The FCC releases daily change files. Use `--daily` flag.

**Q: Can I run this on a Raspberry Pi?**

A: Yes! Build with `GOARCH=arm64` or `GOARCH=arm`:
```bash
GOARCH=arm64 CGO_ENABLED=1 go build -o hamqrzdb-process cmd/process/main.go
```

**Q: Is this faster than the Python version?**

A: Yes, 10-20x faster for data processing and much lower memory usage.

**Q: Can I use this commercially?**

A: Yes, MIT license. See LICENSE file.

## Support

- **GitHub Issues**: https://github.com/chriskacerguis/hamqrzdb/issues
- **QRZ**: https://www.qrz.com/db/KJ5DJC
- **Documentation**: See README.md and DEPLOY.md

## Credits

- **Data Source**: [FCC Universal Licensing System (ULS)](https://www.fcc.gov/uls)
- **Inspired By**: [k3ng/hamdb](https://github.com/k3ng/hamdb)
- **Author**: Chris Kacerguis (KJ5DJC)
- **License**: MIT

73! 📻
