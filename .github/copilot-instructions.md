# GitHub Copilot Instructions for HamQRZDB

## Project Overview

This is a Ham Radio Callsign Lookup System that downloads FCC ULS (Universal Licensing System) data and serves it as a JSON API compatible with the HamDB format. The system is designed to be self-hosted using Docker and nginx.

## Architecture

- **Data Processing**: Python script (`process_uls.py`) downloads and processes FCC data
- **Storage**: Static JSON files in nested directory structure (`/K/J/5/KJ5DJC.json`)
- **Serving**: nginx in Docker container with URL rewriting
- **Updates**: Shell scripts with rsync for zero-downtime updates

## Code Style & Conventions

### Python
- Use Python 3.7+ compatible syntax
- Follow PEP 8 style guidelines
- Use type hints where appropriate
- Prefer streaming/generator patterns for large datasets to minimize memory usage
- Use `csv.DictReader` for parsing CSV/DAT files
- Log important operations for debugging

### Shell Scripts (Bash)
- Use `#!/bin/bash` shebang
- Always use `set -e` for error handling
- Use proper quoting: `"$variable"` instead of `$variable`
- Create cleanup functions with `trap cleanup EXIT`
- Log all major operations with timestamps
- Use rsync for file synchronization (not mv/cp for production updates)

### nginx Configuration
- Use location blocks with regex for URL matching
- Set proper CORS headers (`Access-Control-Allow-Origin: *`)
- Enable gzip compression for JSON
- Handle OPTIONS requests for CORS preflight
- Use `try_files` with fallback to 404.json

### Docker
- Use bind mounts for data (not COPY in Dockerfile)
- Keep images minimal (alpine-based)
- Don't include data in the image
- Document port mappings clearly

## API Format

### URL Pattern
```
/v1/{callsign}/json/{appname}
```

### Response Format
Always return HTTP 200 with this structure:

**Valid Callsign:**
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

**Invalid Callsign (NOT_FOUND):**
```json
{
  "hamdb": {
    "version": "1",
    "callsign": {
      "call": "NOT_FOUND",
      "class": "NOT_FOUND",
      // ... all fields set to "NOT_FOUND"
    },
    "messages": {
      "status": "NOT_FOUND"
    }
  }
}
```

## File Structure

### Data Files
```
output/
├── {first_char}/
│   └── {second_char}/
│       └── {third_char}/
│           └── {CALLSIGN}.json
```

### Key Files
- `process_uls.py` - Main data processing script
- `update-daily.sh` - Daily incremental updates
- `update-weekly.sh` - Weekly full rebuild
- `Dockerfile` - nginx container configuration
- `nginx.conf` - nginx URL rewriting and error handling
- `404.json` - NOT_FOUND response template
- `docker-compose.yml` - Docker deployment configuration

## Important Constraints

### Data Updates Must Be Non-Destructive
- Always process to a temporary directory first
- Use rsync to sync changes to live directory
- Never delete the output directory while the service is running
- Ensure zero downtime during updates

### FCC Data Sources
- Full database: `https://data.fcc.gov/download/pub/uls/complete/l_amat.zip`
- Daily updates: `https://data.fcc.gov/download/pub/uls/daily/l_am_{{DATE}}.zip`
- Date format: `MMDDYYYY` (e.g., `01152025`)

### Required FCC Data Files
- `HD.dat` - Header data (callsign, operator class, status)
- `EN.dat` - Entity data (name, address)
- `AM.dat` - Amateur radio specific data

## When Making Changes

### Adding New Features
1. Maintain backward compatibility with HamDB API format
2. Ensure changes work with Docker bind mounts
3. Update documentation in DOCKER.md and README.md
4. Test with both daily and weekly update scripts

### Modifying Data Processing
1. Test with sample data first
2. Ensure memory-efficient processing (streaming)
3. Handle missing/malformed data gracefully
4. Log all data transformation steps

### Changing nginx Configuration
1. Test configuration with `nginx -t` before deployment
2. Ensure CORS headers are present
3. Maintain NOT_FOUND JSON response for 404/403 errors
4. Preserve URL rewriting for HamDB compatibility

### Updating Shell Scripts
1. Always use temporary directories with unique names (`$$` for PID)
2. Implement proper cleanup with trap
3. Use rsync for production updates
4. Log all steps with timestamps
5. Handle errors gracefully (set -e)

## Deployment Notes

- This is a Docker-only deployment (AWS/CloudFront references are legacy)
- Service runs on port 8080 by default
- Consider using a CDN (Cloudflare) in front for production
- Updates via bind mounts are instant (no container restart needed)

## Testing

### Local Testing
```bash
# Start service
docker-compose up -d

# Test valid callsign
curl http://localhost:8080/v1/KJ5DJC/json/test

# Test invalid callsign
curl http://localhost:8080/v1/BADCALL/json/test

# Check health
curl http://localhost:8080/health
```

### Data Processing Testing
```bash
# Process single callsign
./process_uls.py --full --callsign KJ5DJC

# Process daily updates
./process_uls.py --daily

# Process full database
./process_uls.py --full
```

## Performance Considerations

- Database contains ~1M records, generates ~1-2GB of JSON files
- Processing takes a few minutes
- Use streaming/generators to minimize memory usage
- nginx serves static files efficiently
- Docker image is only ~10MB (data is bind-mounted)

## Credits

- Data source: FCC Universal Licensing System (ULS)
- Inspired by: [k3ng/hamdb](https://github.com/k3ng/hamdb)
- License: MIT
