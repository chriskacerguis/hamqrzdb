# Ham Radio Callsign Lookup System

A high-performance, self-hosted system for serving amateur radio callsign data via a HamDB-compatible JSON API. Uses SQLite for efficient data storage and nginx for lightning-fast static file serving.

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
- [API Format](#api-format)
- [Docker Deployment](#docker-deployment)
- [Automation](#automation)
- [Performance Notes](#performance-notes)
- [Project Files](#project-files)
- [Troubleshooting](#troubleshooting)
- [License](#license)
- [Credits](#credits)

## Features

- 🚀 **HamDB-compatible API** - Drop-in replacement for HamDB lookups
- 💾 **SQLite database** - Efficient storage with ~50-100MB RAM usage
- 📦 **Docker deployment** - Simple setup with stock nginx:alpine
- 🔄 **Incremental updates** - Daily updates without full rebuilds
- 📍 **Optional location data** - Coordinates and Maidenhead grid squares
- ⚡ **Zero-downtime updates** - Changes are instant with bind mounts
- 🌐 **CORS enabled** - Ready for web applications

## Architecture

1. **Data Processing** (`process_uls_db.py`) - Downloads FCC ULS data, loads into SQLite, generates JSON files
2. **Location Processing** (`process_uls_locations.py`) - Optional: Adds coordinates and grid squares
3. **nginx** (Docker) - Serves static JSON files with URL rewriting
4. **SQLite Database** - Single-file database (~500MB for full dataset)
5. **Static JSON Files** - Nested directory structure (~1-2GB for 1.5M callsigns)

> [!TIP]
> Consider using Cloudflare or another CDN in front for production deployments.

## Quick Start

```bash
# 1. Process the full FCC database
python3 process_uls_db.py --full

# 2. Optional: Add location data (coordinates, grid squares)
python3 process_uls_locations.py --la-file temp_uls/LA.dat --regenerate

# 3. Start the nginx server
docker-compose up -d

# 4. Test it
curl http://localhost/v1/KJ5DJC/json/test
```

See [DOCKER.md](DOCKER.md) for complete deployment guide.

## Prerequisites

- Python 3.7+
- Docker and Docker Compose
- ~2GB disk space for full dataset

## Installation

```bash
# Clone the repository
git clone https://github.com/chriskacerguis/hamqrzdb.git
cd hamqrzdb

# Make scripts executable
chmod +x *.sh
chmod +x process_uls_db.py
chmod +x process_uls_locations.py
```

## Usage

### Initial Database Setup

**Full database load** (processes all 1.5M callsigns):

```bash
python3 process_uls_db.py --full
```

This will:
1. Download the complete ULS amateur radio database (~500MB ZIP)
2. Extract HD.dat, EN.dat, and AM.dat files
3. Load data into SQLite database (`hamqrzdb.sqlite`)
4. Generate 1.5M JSON files in nested directory structure
5. Takes ~10-20 minutes depending on disk speed

**Single callsign** (for testing):

```bash
python3 process_uls_db.py --full --callsign KJ5DJC
```

### Add Location Data (Optional)

Location data adds latitude, longitude, and Maidenhead grid squares:

```bash
# Download full database if not already done
python3 process_uls_db.py --full

# Process location data and regenerate JSON files
python3 process_uls_locations.py --la-file temp_uls/LA.dat --regenerate
```

**Note:** The full database download includes LA.dat in `temp_uls/` directory.

### Generate JSON Files from Database

If you already have the database loaded and just need to regenerate JSON files:

```bash
# Generate all JSON files
python3 process_uls_db.py --generate

# Generate single callsign
python3 process_uls_db.py --generate --callsign KJ5DJC
```

### Daily Updates

Update with daily changes from FCC (incremental):

```bash
# Using database scripts (recommended)
./update-daily-db.sh

# Manual
python3 process_uls_db.py --daily
```

Daily updates:
- Download only today's changes (~1-5MB)
- Upsert changes into database
- Regenerate only affected JSON files
- Much faster than full rebuild

### Weekly Full Rebuild

Rebuild the entire database from scratch:

```bash
# Using database script (recommended)
./update-weekly-db.sh

# Manual
rm hamqrzdb.sqlite
python3 process_uls_db.py --full
```

## File Structure

The script creates a nested directory structure to avoid too many files in one directory:

```
output/
├── K/
│   └── J/
│       └── 5/
│           └── KJ5DJC.json
├── W/
│   └── 1/
│       └── A/
│           └── W1AW.json
...
```

## JSON Output Format

Each JSON file follows this structure:

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

## API Format

### Endpoint

```
GET /v1/{callsign}/json/{appname}
```

- `{callsign}` - Amateur radio callsign (e.g., KJ5DJC, W1AW)
- `{appname}` - Your application name (required for compatibility, not used)

### Examples

```bash
# Valid callsign
curl http://localhost/v1/KJ5DJC/json/myapp
curl https://lookup.kj5djc.com/v1/KJ5DJC/json/hamdb

# Invalid callsign (returns NOT_FOUND response)
curl http://localhost/v1/BADCALL/json/test

# Health check
curl http://localhost/health
```

### Response Format

**Valid Callsign:**
```json
{
  "hamdb": {
    "version": "1",
    "callsign": {
      "call": "KJ5DJC",
      "class": "G",
      ...
    },
    "messages": {
      "status": "OK"
    }
  }
}
```

**Invalid Callsign (NOT_FOUND):**
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

Both return **HTTP 200** for client compatibility.

### Features

- **Always returns HTTP 200** - Even for invalid callsigns (client compatibility)
- **CORS enabled** - `Access-Control-Allow-Origin: *`
- **Compressed responses** - gzip enabled for JSON
- **Caching headers** - 24-hour cache for callsign data
- **Sub-10ms response time** - nginx static file serving

## Docker Deployment

Start the nginx server to serve the JSON files:

```bash
# Start service
docker-compose up -d

# View logs
docker-compose logs -f

# Stop service
docker-compose down
```

The service runs on port 80 by default. Edit `docker-compose.yml` to change the port.

## Automation

With Docker bind mounts, updates are instant and don't require container restarts:

```bash
# Create logs directory
mkdir -p logs

# Add to crontab
crontab -e

# Daily updates at 2 AM (using database approach)
0 2 * * * cd /path/to/hamqrzdb && ./update-daily-db.sh >> logs/cron.log 2>&1

# Weekly full rebuild on Sunday at 3 AM (using database approach)
0 3 * * 0 cd /path/to/hamqrzdb && ./update-weekly-db.sh >> logs/cron.log 2>&1
```

Changes are live immediately - no container restart needed!

### Included Scripts

**Database scripts (recommended):**
- `update-daily-db.sh` - Daily incremental updates with database upserts
- `update-weekly-db.sh` - Weekly full rebuild with automatic backup
- `regenerate-json-db.sh` - Regenerate all JSON files from existing database

**Legacy scripts:**
- `update-daily.sh` - Daily updates (direct file processing, deprecated)
- `update-weekly.sh` - Weekly rebuild (direct file processing, deprecated)

## Performance Notes

- **Database Size**: ~500MB SQLite file for 1.5M callsigns
- **JSON Files**: ~1-2GB total (nested directory structure)
- **Memory Usage**: 50-100MB RAM (SQLite batch processing)
- **Initial Load**: ~10-20 minutes for full database + JSON generation
- **Daily Updates**: 1-5 minutes (incremental changes only)
- **Docker Image**: Only ~10MB (nginx:alpine, data is bind-mounted)
- **Response Time**: <10ms with nginx static file serving
- **Updates**: Instant with bind mounts (no container restart required)

## Project Files

**Main Scripts:**
- `process_uls_db.py` - Main database processor (load, update, generate JSON)
- `process_uls_locations.py` - Optional location data processor (lat/lon, grid squares)
- `process_uls_streaming.py` - Memory-optimized streaming processor (legacy)
- `process_uls_enhanced.py` - In-memory processor (deprecated)

**Update Scripts:**
- `update-daily-db.sh` - Daily incremental updates with database
- `update-weekly-db.sh` - Weekly full rebuild with database backup
- `regenerate-json-db.sh` - Regenerate JSON from existing database

**Configuration:**
- `docker-compose.yml` - Docker service configuration
- `nginx.conf` - nginx URL rewriting and CORS configuration
- `404.json` - NOT_FOUND response template

**Documentation:**
- `DOCKER.md` - Complete Docker deployment guide
- `DATABASE-SCRIPTS.md` - Database script documentation
- `MEMORY-OPTIMIZATION.md` - Handling large datasets
- `.github/copilot-instructions.md` - GitHub Copilot guidelines

## Important Notes

### Location Data

Location data (latitude, longitude, grid squares) is **optional** and processed separately:

```bash
# Add location data after initial setup
python3 process_uls_locations.py --la-file temp_uls/LA.dat --regenerate
```

Not all callsigns have location data in the FCC database. The location processor:
- Parses LA.dat (Location/Antenna) records
- Calculates Maidenhead grid squares from coordinates
- Updates the database and regenerates affected JSON files

### License Status Codes

- `A` = Active
- `C` = Canceled  
- `E` = Expired
- `T` = Terminated

### Operator Classes

- `N` = Novice
- `T` = Technician
- `G` = General
- `A` = Amateur Extra
- `P` = Technician Plus

## Troubleshooting

### Out of Memory

The database version (`process_uls_db.py`) uses batch processing and should handle large datasets with 50-100MB RAM. If you still encounter issues:

```bash
# Use streaming version (very low memory)
python3 process_uls_streaming.py --full
```

See [MEMORY-OPTIMIZATION.md](MEMORY-OPTIMIZATION.md) for details.

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
