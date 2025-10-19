# HamQRZDB# HamQRZDB



A high-performance, self-hosted amateur radio callsign lookup system with a HamDB-compatible JSON API.A high-performance, self-hosted amateur radio callsign lookup system with a HamDB-compatible JSON API.



Built with **Go** for speed and efficiency. Processes FCC ULS data into SQLite and serves it via a fast HTTP API with case-insensitive lookups and CORS support.Built with **Go** for speed and efficiency. Processes FCC ULS data into SQLite and serves it via a fast HTTP API with case-insensitive lookups and CORS support.



## Features## Table of Contents



- 🚀 **HamDB-compatible API** - Drop-in replacement for HamDB lookups- [Features](#features)

- ⚡ **Blazing Fast** - ~2ms average response time, ~2,500 req/s throughput- [Architecture](#architecture)

- 🎯 **Case-insensitive** - Works with any callsign case (KJ5DJC = kj5djc)- [Quick Start](#quick-start)

- 💾 **SQLite Storage** - Single-file database (~500MB for 1.5M callsigns)- [Prerequisites](#prerequisites)

- 🔧 **Modern Tooling** - Uses [Task](https://taskfile.dev) for build automation- [Installation](#installation)

- 📦 **Docker Ready** - Simple containerized deployment- [Usage](#usage)

- 🔄 **Fast Updates** - Daily updates in ~30 seconds, full rebuild in 3-5 minutes- [API Format](#api-format)

- 📍 **Location Data** - Coordinates and Maidenhead grid squares- [Docker Deployment](#docker-deployment)

- 🌐 **CORS Enabled** - Ready for web applications- [Automation](#automation)

- 📊 **Low Memory** - ~100MB RAM usage- [Performance](#performance)

- [Task Commands](#task-commands)

## Architecture- [Troubleshooting](#troubleshooting)

- [License](#license)

**Simple 3-component system:**- [Credits](#credits)



1. **hamqrzdb-process** - Downloads FCC ULS data and populates SQLite database## Features

2. **hamqrzdb-locations** - Adds coordinates and Maidenhead grid squares (optional)

3. **hamqrzdb-api** - HTTP server that queries SQLite directly- 🚀 **HamDB-compatible API** - Drop-in replacement for HamDB lookups

- ⚡ **High Performance** - Go-based tools are 4-5x faster than Python equivalents

```- 💾 **SQLite database** - Efficient storage with ~50-100MB RAM usage

FCC Data → hamqrzdb-process → SQLite DB → hamqrzdb-api → JSON API- � **Modern Tooling** - Uses [Task](https://taskfile.dev) for build automation

                                    ↑- �📦 **Docker deployment** - Simple setup with containerized deployment

                          hamqrzdb-locations (optional)- 🔄 **Incremental updates** - Daily updates without full rebuilds

```- 📍 **Location data** - Coordinates and Maidenhead grid squares

- ⚡ **Zero-downtime updates** - Changes are instant with bind mounts

No JSON file generation needed. The API queries the database directly for real-time, zero-downtime updates.- 🌐 **CORS enabled** - Ready for web applications

- 🎯 **Case-insensitive lookups** - Works with any callsign case

## Quick Start

## Architecture

```bash

# 1. Install Task (if needed)### Go-based Tools (Recommended)

brew install go-task/tap/go-task  # macOS

# or see https://taskfile.dev/installation/1. **hamqrzdb-process** - Go CLI for downloading FCC ULS data and populating SQLite

2. **hamqrzdb-locations** - Go CLI for adding coordinates and Maidenhead grid squares

# 2. Clone and build3. **hamqrzdb-api** - Go HTTP server with case-insensitive lookups and CORS support

git clone https://github.com/chriskacerguis/hamqrzdb.git4. **SQLite Database** - Single-file database (~500MB for full dataset)

cd hamqrzdb5. **Task** - Modern build tool for automation

task build

### Legacy Python Tools (Still supported)

# 3. Process FCC data (3-5 minutes)

task db:full1. **process_uls_db.py** - Python script for data processing (slower but works)

2. **process_uls_locations.py** - Python script for location processing

# 4. Optional: Add location data3. **nginx** (Docker) - Serves static JSON files with URL rewriting

task db:locations -- --la-file temp_uls/LA.dat

> [!TIP]

# 5. Start API server> The Go tools are **4-5x faster** and use **5x less memory** than Python equivalents. They're recommended for production use.

task dev:api

> [!NOTE]

# 6. Test it!> Consider using Cloudflare or another CDN in front for production deployments.

curl http://localhost:8080/v1/KJ5DJC/json/test

curl http://localhost:8080/v1/kj5djc/json/test  # case-insensitive!## Quick Start

```

### Using Go Tools (Recommended)

## Prerequisites

```bash

- **Go 1.21+** (with CGO support for SQLite)# 1. Build the tools

- **Task** - Modern task runner ([installation](https://taskfile.dev/installation/))task build

- **~2GB disk space** for full dataset

- **gcc/clang** for SQLite (usually pre-installed on macOS/Linux)# 2. Process the full FCC database (3-5 minutes)

task db:full

## Installation

# 3. Optional: Add location data

```bashtask db:locations -- --la-file temp_uls/LA.dat

# Clone repository

git clone https://github.com/chriskacerguis/hamqrzdb.git# 4. Start the API server

cd hamqrzdbtask dev:api



# Install Task (macOS)# 5. Test it (case-insensitive!)

brew install go-task/tap/go-taskcurl http://localhost:8080/v1/KJ5DJC/json/test

curl http://localhost:8080/v1/kj5djc/json/test

# Install Task (Linux)```

sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d

### Using Python Tools (Legacy)

# Install Task (or with Go)

go install github.com/go-task/task/v3/cmd/task@latest```bash

# 1. Process the full FCC database (15-20 minutes)

# Build all toolspython3 process_uls_db.py --full

task build

# 2. Optional: Add location data

# Optional: Install system-widepython3 process_uls_locations.py --la-file temp_uls/LA.dat --regenerate

task install  # Installs to /usr/local/bin

```# 3. Start the nginx server

docker-compose up -d

## Usage

# 4. Test it

### Processing FCC Datacurl http://localhost/v1/KJ5DJC/json/test

```

```bash

# Download and process full database (3-5 minutes)See [README.cli.md](docs/README.cli.md) for complete Go CLI reference, or [DOCKER.md](DOCKER.md) for nginx deployment guide.

task db:full

## Prerequisites

# Or use binary directly

./bin/hamqrzdb-process --full### For Go Tools (Recommended)

- Go 1.21+ (with CGO support for SQLite)

# Process daily updates (30 seconds)- [Task](https://taskfile.dev) - `brew install go-task/tap/go-task`

task db:daily- ~2GB disk space for full dataset



# Process single callsign (for testing)### For Python Tools (Legacy)

./bin/hamqrzdb-process --full --callsign KJ5DJC- Python 3.7+

- Docker and Docker Compose (optional)

# Custom database path- ~2GB disk space for full dataset

./bin/hamqrzdb-process --full --db /path/to/custom.db

```## Installation



### Adding Location Data (Optional)```bash

# Clone the repository

Location data adds latitude, longitude, and Maidenhead grid squares:git clone https://github.com/chriskacerguis/hamqrzdb.git

cd hamqrzdb

```bash

# Process location data (2-3 minutes)# Install Task (if not already installed)

task db:locations -- --la-file temp_uls/LA.datbrew install go-task/tap/go-task  # macOS

# or see https://taskfile.dev/installation/

# Or use binary directly

./bin/hamqrzdb-locations --la-file temp_uls/LA.dat --db hamqrzdb.sqlite# Install Go dependencies

task deps

# Process single callsign

./bin/hamqrzdb-locations --la-file temp_uls/LA.dat --callsign KJ5DJC# Build all tools

```task build



**Note:** The full database download includes LA.dat in the `temp_uls/` directory.# Optional: Install system-wide

task install  # Installs to /usr/local/bin

### Running the API Server```



```bash## Usage

# Development mode (default: localhost:8080)

task dev:api### Initial Database Setup



# Or run binary directly#### Using Go Tools (Fast - 3-5 minutes)

./bin/hamqrzdb-api

**Full database load** (processes all 1.5M callsigns):

# Custom configuration

DB_PATH=./hamqrzdb.sqlite PORT=8080 ./bin/hamqrzdb-api```bash

task db:full

# Production mode (with Docker)```

task docker:up

```**Or using the binary directly:**



### Database Statistics```bash

./bin/hamqrzdb-process --full

```bash```

# Show database stats

task db:stats**Single callsign** (for testing):



# Output:```bash

# 📊 Database statistics:./bin/hamqrzdb-process --full --callsign KJ5DJC

# 1234567 total callsigns```

# 1098765 active licenses

# 987654 with locations#### Using Python Tools (Slower - 15-20 minutes)

# Last updated: 2025-10-19 12:34:56

``````bash

python3 process_uls_db.py --full

## API Format```



### Endpoint**What happens:**

1. Downloads the complete ULS amateur radio database (~500MB ZIP)

```2. Extracts HD.dat, EN.dat, and AM.dat files

GET /v1/{callsign}/json/{appname}3. Loads data into SQLite database (`hamqrzdb.sqlite`)

```4. Optionally generates 1.5M JSON files in nested directory structure



- `{callsign}` - Amateur radio callsign (case-insensitive)### Add Location Data (Optional)

- `{appname}` - Your application name (for compatibility, not validated)

Location data adds latitude, longitude, and Maidenhead grid squares:

### Examples

#### Using Go Tools (Fast - 2-3 minutes)

```bash

# Valid callsign (both work!)```bash

curl http://localhost:8080/v1/KJ5DJC/json/myapp# Process location data from LA.dat

curl http://localhost:8080/v1/kj5djc/json/myapptask db:locations -- --la-file temp_uls/LA.dat



# Invalid callsign (returns NOT_FOUND response)# Or use the binary directly

curl http://localhost:8080/v1/BADCALL/json/test./bin/hamqrzdb-locations --la-file temp_uls/LA.dat --db hamqrzdb.sqlite

```

# Health check

curl http://localhost:8080/health#### Using Python Tools (Slower - 8-10 minutes)

```

```bash

### Response Format# Download full database if not already done

python3 process_uls_db.py --full

**Valid Callsign (HTTP 200):**

```json# Process location data and regenerate JSON files

{python3 process_uls_locations.py --la-file temp_uls/LA.dat --regenerate

  "hamdb": {```

    "version": "1",

    "callsign": {**Note:** The full database download includes LA.dat in `temp_uls/` directory.

      "call": "KJ5DJC",

      "class": "G",### Generate JSON Files from Database

      "expires": "11/18/2033",

      "status": "A",If you already have the database loaded and just need to regenerate JSON files:

      "grid": "EM10ci",

      "lat": "30.3416503",#### Using Go Tools

      "lon": "-97.7548379",

      "fname": "CHRIS",```bash

      "mi": "",task db:generate

      "name": "KACERGUIS",

      "suffix": "",# Or use the binary directly

      "addr1": "5900 Balcones Drive STE 26811",./bin/hamqrzdb-process --generate --db hamqrzdb.sqlite --output output

      "addr2": "AUSTIN",```

      "state": "TX",

      "zip": "78731",#### Using Python Tools

      "country": "United States"

    },```bash

    "messages": {# Generate all JSON files

      "status": "OK"python3 process_uls_db.py --generate

    }

  }# Generate single callsign

}python3 process_uls_db.py --generate --callsign KJ5DJC

``````



**Invalid Callsign (HTTP 200):**> [!NOTE]

```json> JSON file generation is optional when using the Go API server, which queries SQLite directly.

{

  "hamdb": {### Daily Updates

    "version": "1",

    "callsign": {Update with daily changes from FCC (incremental):

      "call": "NOT_FOUND",

      "class": "NOT_FOUND",#### Using Go Tools (Fast - 30 seconds)

      ...

    },```bash

    "messages": {task db:daily

      "status": "NOT_FOUND"

    }# Or use the binary directly

  }./bin/hamqrzdb-process --daily --db hamqrzdb.sqlite

}```

```

#### Using Python Tools (Slower - 2-3 minutes)

Both responses return **HTTP 200** for client compatibility.

```bash

## Docker Deployment# Using database scripts (recommended)

./update-daily-db.sh

### Using Docker Compose

# Or manually

```bashpython3 process_uls_db.py --daily

# Build and start services```

task docker:build

task docker:up**Daily updates:**

- Download only today's changes (~1-5MB)

# Or use docker-compose directly- Upsert changes into database

docker-compose up -d- Optionally regenerate affected JSON files

- Much faster than full rebuild

# View logs

task docker:logs### Weekly Full Rebuild



# Stop servicesRebuild the entire database from scratch:

task docker:down

``````bash

# Using database script (recommended)

### Manual Docker Build./update-weekly-db.sh



```bash# Manual

# Build imagerm hamqrzdb.sqlite

docker build -t hamqrzdb-api:latest .python3 process_uls_db.py --full

```

# Run container

docker run -d \## File Structure

  -p 8080:8080 \

  -v $(pwd)/hamqrzdb.sqlite:/app/hamqrzdb.sqlite \The script creates a nested directory structure to avoid too many files in one directory:

  --name hamqrzdb-api \

  hamqrzdb-api:latest```

```output/

├── K/

The API server runs on port 8080 by default. The database is bind-mounted for zero-downtime updates.│   └── J/

│       └── 5/

## Automation│           └── KJ5DJC.json

├── W/

### Cron Jobs│   └── 1/

│       └── A/

```bash│           └── W1AW.json

# Create logs directory...

mkdir -p logs```



# Edit crontab## JSON Output Format

crontab -e

Each JSON file follows this structure:

# Daily updates at 2 AM (30 seconds)

0 2 * * * cd /path/to/hamqrzdb && task db:daily >> logs/cron.log 2>&1```json

{

# Weekly full rebuild on Sunday at 3 AM (3-5 minutes)  "hamdb": {

0 3 * * 0 cd /path/to/hamqrzdb && task db:full >> logs/cron.log 2>&1    "version": "1",

    "callsign": {

# Update locations monthly (2-3 minutes)      "call": "KJ5DJC",

0 4 1 * * cd /path/to/hamqrzdb && task db:locations -- --la-file temp_uls/LA.dat >> logs/cron.log 2>&1      "class": "G",

```      "expires": "11/18/2033",

      "status": "A",

Database changes are live immediately - no API server restart needed!      "grid": "EM10ci",

      "lat": "30.3416503",

### Systemd Service      "lon": "-97.7548379",

      "fname": "CHRIS",

Create `/etc/systemd/system/hamqrzdb-api.service`:      "mi": "",

      "name": "KACERGUIS",

```ini      "suffix": "",

[Unit]      "addr1": "5900 Balcones Drive STE 26811",

Description=HamQRZDB API Server      "addr2": "AUSTIN",

After=network.target      "state": "TX",

      "zip": "78731",

[Service]      "country": "United States"

Type=simple    },

User=hamqrzdb    "messages": {

WorkingDirectory=/opt/hamqrzdb      "status": "OK"

Environment="DB_PATH=/opt/hamqrzdb/hamqrzdb.sqlite"    }

Environment="PORT=8080"  }

ExecStart=/usr/local/bin/hamqrzdb-api}

Restart=always```

RestartSec=10

## API Format

[Install]

WantedBy=multi-user.target### Endpoint

```

```

Enable and start:GET /v1/{callsign}/json/{appname}

```

```bash

sudo systemctl enable hamqrzdb-api- `{callsign}` - Amateur radio callsign (e.g., KJ5DJC, W1AW)

sudo systemctl start hamqrzdb-api- `{appname}` - Your application name (required for compatibility, not used)

sudo systemctl status hamqrzdb-api

```### Examples



## Performance```bash

# Valid callsign

### Benchmarkscurl http://localhost/v1/KJ5DJC/json/myapp

curl https://lookup.kj5djc.com/v1/KJ5DJC/json/hamdb

| Operation | Time | Notes |

|-----------|------|-------|# Invalid callsign (returns NOT_FOUND response)

| Full database load | 3-5 min | 1.5M callsigns |curl http://localhost/v1/BADCALL/json/test

| Daily updates | ~30 sec | Incremental changes |

| Location processing | 2-3 min | All callsigns |# Health check

| API response time | ~2ms | Average (p50) |curl http://localhost/health

| API response time | <50ms | p99 |```

| API throughput | ~2,500 req/s | Single instance |

| Memory usage | ~100MB | API server |### Response Format

| Database size | ~500MB | 1.5M callsigns |

**Valid Callsign:**

### Comparison to Python```json

{

- **4-5x faster** data processing  "hamdb": {

- **5x less memory** usage    "version": "1",

- **50x faster** API responses    "callsign": {

- **Single binary** deployment (no dependencies)      "call": "KJ5DJC",

      "class": "G",

## Task Commands      ...

    },

Quick reference for common operations:    "messages": {

      "status": "OK"

```bash    }

# Build commands  }

task build              # Build all binaries}

task build:api          # Build API server only```

task build:process      # Build data processor only

task build:locations    # Build locations processor only**Invalid Callsign (NOT_FOUND):**

```json

# Development{

task dev:api                              # Run API server  "hamdb": {

task dev:process -- --full                # Run data processor    "version": "1",

task dev:locations -- --la-file LA.dat    # Run locations processor    "callsign": {

      "call": "NOT_FOUND",

# Database operations      "class": "NOT_FOUND",

task db:full            # Download and process full database      ...

task db:daily           # Process daily updates    },

task db:locations       # Process location data    "messages": {

task db:stats           # Show database statistics      "status": "NOT_FOUND"

    }

# Docker operations  }

task docker:build       # Build Docker image}

task docker:up          # Start services```

task docker:down        # Stop services

task docker:logs        # View logsBoth return **HTTP 200** for client compatibility.



# Utility### Features

task clean              # Remove build artifacts

task install            # Install to /usr/local/bin- **Always returns HTTP 200** - Even for invalid callsigns (client compatibility)

task test               # Run tests- **Case-insensitive** - Works with any callsign case (Go API only)

task help               # Show detailed help- **CORS enabled** - `Access-Control-Allow-Origin: *`

task --list             # List all available tasks- **Compressed responses** - gzip enabled for JSON

```- **Fast responses** - <2ms with Go API, <10ms with nginx



See [docs/TASKFILE-MIGRATION.md](docs/TASKFILE-MIGRATION.md) for migration guide from Makefile.## Running the API Server



## Project Structure### Using Go API Server (Recommended)



```The Go API server queries SQLite directly - no JSON files needed!

hamqrzdb/

├── main.go                   # API server source```bash

├── cmd/# Start the API server (development)

│   ├── process/main.go       # Data processor sourcetask dev:api

│   └── locations/main.go     # Locations processor source

├── bin/                      # Compiled binaries# Or run the binary directly

│   ├── hamqrzdb-api./bin/hamqrzdb-api

│   ├── hamqrzdb-process

│   └── hamqrzdb-locations# Custom configuration

├── Taskfile.yml              # Task automation configurationDB_PATH=./hamqrzdb.sqlite PORT=8080 ./bin/hamqrzdb-api

├── Dockerfile                # Docker image definition```

├── docker-compose.yml        # Docker Compose configuration

├── hamqrzdb.sqlite           # SQLite database (generated)**Benefits:**

├── temp_uls/                 # Downloaded FCC data (temporary)- ⚡ **50x faster** than Python http.server (~2,500 req/s vs ~50 req/s)

└── docs/                     # Documentation- 🎯 **Case-insensitive** lookups (KJ5DJC = kj5djc)

    ├── README.cli.md         # CLI tools reference- 💾 **No JSON files** needed (queries SQLite directly)

    ├── TASKFILE-MIGRATION.md # Makefile→Task migration guide- 🔄 **Real-time updates** (database changes are instant)

    ├── LOCATIONS.md          # Locations processor guide- 🌐 **Built-in CORS** support

    └── QUICKREF.md           # Quick reference card

```### Using nginx (Static Files)



## DocumentationIf you prefer serving static JSON files with nginx:



- **[README.cli.md](docs/README.cli.md)** - Complete CLI reference for all tools## Docker Deployment

- **[TASKFILE-MIGRATION.md](docs/TASKFILE-MIGRATION.md)** - Migration guide from Makefile

- **[LOCATIONS.md](docs/LOCATIONS.md)** - Locations processor detailed guideStart the nginx server to serve the JSON files:

- **[QUICKREF.md](docs/QUICKREF.md)** - Quick reference card for common operations

```bash

## Troubleshooting# Start service

docker-compose up -d

### Build Errors

# View logs

**Error: `CGO_ENABLED` required for SQLite**docker-compose logs -f



Solution: Make sure you have a C compiler installed:# Stop service

docker-compose down

```bash```

# macOS

xcode-select --installThe service runs on port 80 by default. Edit `docker-compose.yml` to change the port.



# Ubuntu/Debian## Automation

sudo apt-get install build-essential

### Using Task (Go Tools)

# Fedora/RHEL

sudo dnf install gcc```bash

```# Create logs directory

mkdir -p logs

### Database Errors

# Add to crontab

**Error: Database locked**crontab -e



Solution: Close any other connections to the database:# Daily updates at 2 AM (fast - 30 seconds)

0 2 * * * cd /path/to/hamqrzdb && task db:daily >> logs/cron.log 2>&1

```bash

# Find processes using the database# Weekly full rebuild on Sunday at 3 AM (fast - 3-5 minutes)

lsof hamqrzdb.sqlite0 3 * * 0 cd /path/to/hamqrzdb && task db:full >> logs/cron.log 2>&1

```

# Kill the process if needed

kill <PID>### Using Shell Scripts (Legacy)

```

With Docker bind mounts, updates are instant and don't require container restarts:

**Error: Database corrupted**

```bash

Solution: Rebuild from scratch:# Daily updates at 2 AM (slower - 2-3 minutes)

0 2 * * * cd /path/to/hamqrzdb && ./update-daily-db.sh >> logs/cron.log 2>&1

```bash

# Backup current database# Weekly full rebuild on Sunday at 3 AM (slower - 15-20 minutes)

cp hamqrzdb.sqlite hamqrzdb.sqlite.backup0 3 * * 0 cd /path/to/hamqrzdb && ./update-weekly-db.sh >> logs/cron.log 2>&1

```

# Remove and rebuild

rm hamqrzdb.sqlite hamqrzdb.sqlite-*Changes are live immediately - no container restart needed!

task db:full

```### Included Scripts



### Download Errors**Database scripts (recommended):**

- `update-daily-db.sh` - Daily incremental updates with database upserts

**Error: Daily file not available**- `update-weekly-db.sh` - Weekly full rebuild with automatic backup

- `regenerate-json-db.sh` - Regenerate all JSON files from existing database

Solution: Daily files may not be available on weekends/holidays. Use full database instead:

**Legacy scripts:**

```bash- `update-daily.sh` - Daily updates (direct file processing, deprecated)

task db:full- `update-weekly.sh` - Weekly rebuild (direct file processing, deprecated)

```

## Performance Notes

**Error: Download failed**

### Go Tools (Recommended)

Solution: Check FCC website status and try again:

- **Processing Speed**: 4-5x faster than Python (3-5 min vs 15-20 min full load)

- https://www.fcc.gov/uls/transactions/daily-weekly- **Memory Usage**: 5x less than Python (~100MB vs ~500MB)

- **API Response Time**: ~2ms average (<50ms p99)

### Memory Issues- **API Throughput**: ~2,500 requests/second

- **Daily Updates**: ~30 seconds (incremental changes only)

The Go tools are memory-efficient (~100MB), but if you encounter issues:- **Binary Size**: ~17MB total for all three tools



```bash### General Metrics

# Process single callsign to test

./bin/hamqrzdb-process --full --callsign KJ5DJC- **Database Size**: ~500MB SQLite file for 1.5M callsigns

- **JSON Files**: ~1-2GB total (optional, for nginx deployment)

# Check system resources- **Docker Image**: Only ~10MB (nginx:alpine, data is bind-mounted)

free -h  # Linux- **Updates**: Instant with Go API (database changes are live immediately)

vm_stat  # macOS

```## Project Files



## Data Source & Updates**Go Tools (Recommended):**

- `main.go` - Go API server with case-insensitive lookups

### FCC ULS Database- `cmd/process/main.go` - Go data processor (FCC downloads and SQLite)

- `cmd/locations/main.go` - Go locations processor (coordinates and grids)

- **Full Database**: https://data.fcc.gov/download/pub/uls/complete/l_amat.zip (~500MB)- `Taskfile.yml` - Task automation configuration

- **Daily Updates**: https://data.fcc.gov/download/pub/uls/daily/l_am_MMDDYYYY.zip (~1-5MB)

- **Update Schedule**: Daily updates usually available by 2 AM ET**Python Scripts (Legacy):**

- **Documentation**: https://www.fcc.gov/uls/transactions/daily-weekly- `process_uls_db.py` - Main database processor (load, update, generate JSON)

- `process_uls_locations.py` - Optional location data processor (lat/lon, grid squares)

### License Status Codes

**Update Scripts:**

- `A` = Active- `update-daily-db.sh` - Daily incremental updates with database

- `C` = Canceled- `update-weekly-db.sh` - Weekly full rebuild with database backup

- `E` = Expired- `regenerate-json-db.sh` - Regenerate JSON from existing database

- `T` = Terminated

**Configuration:**

### Operator Classes- `docker-compose.yml` - Docker service configuration

- `nginx.conf` - nginx URL rewriting and CORS configuration

- `N` = Novice (no longer issued)- `404.json` - NOT_FOUND response template

- `T` = Technician

- `G` = General**Documentation:**

- `A` = Amateur Extra- `docs/README.cli.md` - Go CLI tools reference

- `P` = Technician Plus (no longer issued)- `docs/TASKFILE-MIGRATION.md` - Migration guide from Makefile to Task

- `docs/LOCATIONS.md` - Locations processor guide

## License- `docs/QUICKREF.md` - Quick reference card

- `DOCKER.md` - Complete Docker deployment guide

MIT License - See [LICENSE](LICENSE) file for details.- `.github/copilot-instructions.md` - GitHub Copilot guidelines



## Credits## Important Notes



**Data Source:**### Location Data

- FCC Universal Licensing System (ULS) - https://www.fcc.gov/uls/

Location data (latitude, longitude, grid squares) is **optional** and processed separately:

**Inspiration:**

- [k3ng/hamdb](https://github.com/k3ng/hamdb) for the original HamDB project and API format```bash

# Add location data after initial setup

**Built with:**python3 process_uls_locations.py --la-file temp_uls/LA.dat --regenerate

- [Go](https://golang.org/) - Programming language```

- [SQLite](https://www.sqlite.org/) - Database engine

- [Task](https://taskfile.dev/) - Task runnerNot all callsigns have location data in the FCC database. The location processor:

- [go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite driver for Go- Parses LA.dat (Location/Antenna) records

- Calculates Maidenhead grid squares from coordinates

---- Updates the database and regenerates affected JSON files



**73! 📻**### License Status Codes



For questions, issues, or contributions, visit: https://github.com/chriskacerguis/hamqrzdb- `A` = Active

- `C` = Canceled  
- `E` = Expired
- `T` = Terminated

### Operator Classes

- `N` = Novice
- `T` = Technician
- `G` = General
- `A` = Amateur Extra
- `P` = Technician Plus

## Task Commands Reference

Quick reference for common Task commands. See [docs/TASKFILE-MIGRATION.md](docs/TASKFILE-MIGRATION.md) for complete guide.

```bash
# List all available tasks
task --list

# Build commands
task build              # Build all binaries
task build:api          # Build API server only
task build:process      # Build data processor only
task build:locations    # Build locations processor only

# Development commands
task dev:api                              # Run API server
task dev:process -- --full                # Run data processor
task dev:locations -- --la-file LA.dat    # Run locations processor

# Database commands
task db:full            # Download and process full database
task db:daily           # Process daily updates
task db:locations       # Process location data
task db:stats           # Show database statistics

# Docker commands
task docker:build       # Build Docker image
task docker:up          # Start services
task docker:down        # Stop services
task docker:logs        # View logs

# Utility commands
task clean              # Remove build artifacts
task install            # Install to /usr/local/bin
task test               # Run tests
task help               # Show detailed help
```

## Troubleshooting

### Out of Memory

**Go tools** use efficient batch processing (~100MB RAM). If using Python and encountering memory issues:

```bash
# Use the Go tools instead (recommended)
task db:full

# Or use Python streaming version (very low memory)
python3 process_uls_streaming.py --full
```

### Download Fails

Check the FCC website for changes to URLs or file formats:
- https://www.fcc.gov/uls/transactions/daily-weekly

### Daily File Not Available

Daily files may not be available on weekends or holidays. Run `--full` or wait for next daily update.

### nginx Not Serving Files

Check that:
1. JSON files exist in `output/` directory
2. `404.json` exists in `output/404.json`
3. nginx.conf is mounted correctly
4. File permissions allow reading

```bash
# Check files
ls -la output/K/J/5/KJ5DJC.json
ls -la output/404.json

# Check nginx config
docker-compose exec nginx nginx -t

# Restart nginx
docker-compose restart
```

### Database Corruption

If the database becomes corrupted:

```bash
# Backup current database
cp hamqrzdb.sqlite hamqrzdb.sqlite.backup

# Rebuild from scratch
rm hamqrzdb.sqlite
python3 process_uls_db.py --full
```

## License

MIT License - Feel free to use and modify for your needs.

## Credits

**Data Source:**
- FCC Universal Licensing System (ULS) - https://www.fcc.gov/uls/

**Inspiration:**
- Special thanks to [k3ng/hamdb](https://github.com/k3ng/hamdb) for the original HamDB project and API format inspiration
