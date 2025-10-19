# Ham Radio Callsign Lookup System

A high-performance system for creating and serving static JSON files for amateur radio callsign lookups using FCC ULS data.

## Architecture

1. **Data Processing Script** (`process_uls.py`) - Downloads and processes FCC ULS data into static JSON files
2. **nginx** - Serves static files with URL rewriting and NOT_FOUND handling
3. **Docker** - Containerized deployment with bind mounts for easy updates

> [!TIP]
> The service is lightweight, but consider using Cloudflare or another CDN in front of it for production deployments.

## Quick Start

```bash
# Process the database
./process_uls.py --full

# Start the server
docker-compose up -d

# Test it
curl http://localhost:8080/v1/KJ5DJC/json/test
```

See [DOCKER.md](DOCKER.md) for complete deployment guide.

## Setup

### Prerequisites

- Python 3.7+
- Docker and Docker Compose

### Installation

```bash
# Clone the repository
git clone https://github.com/chriskacerguis/hamqrzdb.git
cd hamqrzdb

# Make the script executable
chmod +x process_uls.py
chmod +x update-daily.sh
chmod +x update-weekly.sh
```

## Usage

### Download and Process Full Database

```bash
./process_uls.py --full
```

This will:
1. Download the complete ULS amateur radio database (~500MB)
2. Extract and process HD.dat, EN.dat, and AM.dat files
3. Generate JSON files in the `output/` directory with structure: `output/K/J/5/KJ5DJC.json`

### Download and Process Daily Updates

```bash
./process_uls.py --daily
```

This downloads only today's incremental changes and processes them.

### Process a Specific Callsign

```bash
./process_uls.py --full --callsign KJ5DJC
```

Downloads the full database but only generates a JSON file for the specified callsign.

### Process a Local File

```bash
./process_uls.py --file /path/to/l_amat.zip
```

Processes a ZIP file you've already downloaded.

### Custom Output Directory

```bash
./process_uls.py --full --output /path/to/output
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

## API Endpoints

### HamDB Compatible Format

```
/v1/{callsign}/json/{appname}
```

**Examples:**
```bash
# Valid callsign
curl http://localhost:8080/v1/KJ5DJC/json/myapp
curl https://lookup.kj5djc.com/v1/KJ5DJC/json/hamdb

# Invalid callsign (returns NOT_FOUND)
curl http://localhost:8080/v1/BADCALL/json/test
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

## API URL Format

The API follows the HamDB-compatible format:

```
http://your-domain.com/v1/{callsign}/json/{appname}
```

**Examples:**
```
http://localhost:8080/v1/KJ5DJC/json/myapp
https://lookup.kj5djc.com/v1/KJ5DJC/json/hamdb
```

The `{appname}` parameter is required for compatibility but is not used by the API.

## Automation
https://lookup.kj5djc.com/v1/KJ5DJC/json/hamdb
```

## Automation

With Docker bind mounts, updates are instant and don't require container restarts:

```bash
# Add to crontab
crontab -e

# Daily updates at 2 AM
0 2 * * * cd /path/to/hamqrzdb && ./update-daily.sh >> logs/cron.log 2>&1

# Weekly full rebuild on Sunday at 3 AM
0 3 * * 0 cd /path/to/hamqrzdb && ./update-weekly.sh >> logs/cron.log 2>&1
```

Changes are live immediately - no container restart needed!

### Included Scripts

- `update-daily.sh` - Downloads daily changes and updates data
- `update-weekly.sh` - Full database rebuild

## Performance Notes

- **Processing Speed**: Processes ~1M records in a few minutes
- **Disk Space**: Full database generates ~1-2GB of JSON files
- **Memory**: Minimal memory usage with streaming CSV parsing
- **Docker Image**: Only ~10MB (nginx + config, data is bind-mounted)
- **Updates**: Instant with bind mounts (no container restart required)

## Documentation

- **[DOCKER.md](DOCKER.md)** - Complete Docker deployment guide
- **[docs/QUICKSTART.md](docs/QUICKSTART.md)** - Quick start guide
- **[docs/STRUCTURE.md](docs/STRUCTURE.md)** - Project structure details

## Important Notes

### Missing Coordinates

The current script doesn't include latitude/longitude data because that information is in additional files (LA.dat - Location/Antenna). To add coordinate support:

1. Modify the script to also parse LA.dat
2. Match location records to callsigns
3. Calculate grid squares from coordinates

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

### Download Fails

Check the FCC website for changes to URLs or file formats:
- https://www.fcc.gov/uls/transactions/daily-weekly

### Daily File Not Available

Daily files may not be available on weekends or holidays. Use `--full` or provide a specific `--file`.

### Out of Memory

If processing fails due to memory, consider processing in batches or increasing system memory.

## License

MIT License - Feel free to use and modify for your needs.

## Credits

**Data Source:**
- FCC Universal Licensing System (ULS) - https://www.fcc.gov/uls/

**Inspiration:**
- Special thanks to [k3ng/hamdb](https://github.com/k3ng/hamdb) for the original HamDB project and API format inspiration
