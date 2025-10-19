# HamQRZDB - Go API Deployment Guide

## Overview

This guide covers deploying the Go API version of HamQRZDB, which queries SQLite directly instead of serving static JSON files. This approach provides:

- ✅ **Case-insensitive lookups** (both `/v1/KJ5DJC/json/app` and `/v1/kj5djc/json/app` work)
- ✅ **Smaller disk footprint** (~500MB database vs ~2GB of JSON files)
- ✅ **Faster queries** (direct database access with connection pooling)
- ✅ **Better performance** (Go's compiled binary is very fast)
- ✅ **Simpler deployment** (no need to regenerate 1.5M JSON files)

## Quick Start

### 1. Build the Database

First, process the FCC data into SQLite:

```bash
# Download and process full FCC database
python3 process_uls_db.py --full

# Optional: Add location data (lat/lon/grid squares)
python3 process_uls_locations.py --full
```

This creates `hamqrzdb.sqlite` (~500MB) with all callsign data.

### 2. Build and Run the Go API

Using Docker Compose (recommended):

```bash
# Build and start the API
docker-compose -f docker-compose.go.yml up -d

# View logs
docker-compose -f docker-compose.go.yml logs -f api

# Test the API
curl http://localhost:8080/v1/KJ5DJC/json/test
curl http://localhost:8080/v1/kj5djc/json/test  # Case-insensitive!
```

Or build locally:

```bash
# Install Go 1.21+ from https://golang.org/dl/

# Download dependencies
go mod download

# Build the binary
CGO_ENABLED=1 go build -o hamqrzdb-api main.go

# Run the API
./hamqrzdb-api
```

## Production Deployment with SSL

For production with HTTPS:

### 1. Update Domain Name

Edit `nginx-proxy.conf` and replace `lookup.kj5djc.com` with your domain.

### 2. Initial SSL Setup

```bash
# Start services without SSL first
docker-compose -f docker-compose.go.yml up -d

# Get SSL certificate
docker-compose -f docker-compose.go.yml run --rm certbot certonly \
  --webroot --webroot-path /var/www/certbot \
  -d lookup.yourdomain.com \
  --email your-email@example.com \
  --agree-tos \
  --no-eff-email

# Restart nginx with SSL enabled
docker-compose -f docker-compose.go.yml restart nginx
```

### 3. Test Production

```bash
# Test HTTPS
curl https://lookup.yourdomain.com/v1/KJ5DJC/json/test

# Test case-insensitivity
curl https://lookup.yourdomain.com/v1/kj5djc/json/test

# Check health
curl https://lookup.yourdomain.com/health
```

## API Endpoints

### Callsign Lookup
```
GET /v1/{callsign}/json/{appname}
```

**Case-insensitive** - both work:
- `https://lookup.kj5djc.com/v1/KJ5DJC/json/myapp`
- `https://lookup.kj5djc.com/v1/kj5djc/json/myapp`

**Response (200 OK):**
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
      // ... more fields
    },
    "messages": {
      "status": "OK"
    }
  }
}
```

**Not Found Response (200 OK):**
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

### Health Check
```
GET /health
```

Returns `200 OK` with `{"status": "healthy"}` if the API and database are working.

### Homepage
```
GET /
```

Serves the beautiful API documentation page with live demo.

## Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │ HTTPS
       ▼
┌─────────────┐
│   Nginx     │ ← SSL termination, rate limiting, gzip
│  (port 443) │
└──────┬──────┘
       │ HTTP
       ▼
┌─────────────┐
│  Go API     │ ← Fast HTTP server with connection pooling
│  (port 8080)│
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   SQLite    │ ← Single database file (~500MB)
│  Database   │    1,575,334+ callsigns
└─────────────┘
```

## Performance

- **Database size**: ~500MB (vs ~2GB of JSON files)
- **Query time**: < 5ms average (direct database access)
- **Memory usage**: ~50MB (Go binary + connection pool)
- **Concurrent requests**: 25 max connections, 5 idle
- **Rate limit**: 10 req/s per IP (burst up to 20)

## Configuration

### Environment Variables

The Go API accepts these environment variables:

- `DB_PATH` - Path to SQLite database (default: `/data/hamqrzdb.sqlite`)
- `PORT` - HTTP port to listen on (default: `8080`)

### Docker Compose Configuration

Edit `docker-compose.go.yml` to customize:

```yaml
environment:
  - DB_PATH=/data/hamqrzdb.sqlite
  - PORT=8080
ports:
  - "8080:8080"  # Change external port here
volumes:
  - ./hamqrzdb.sqlite:/data/hamqrzdb.sqlite:ro  # Read-only mount
```

## Updating Data

### Daily Updates (Incremental)

```bash
# Update database with daily changes
python3 process_uls_db.py --daily

# No need to restart API - it reads directly from database!
# Changes are available immediately on next query
```

### Weekly Full Rebuild

```bash
# Full rebuild from FCC data
python3 process_uls_db.py --full

# Optional: Update location data
python3 process_uls_locations.py --full

# Still no restart needed!
```

**Note**: Unlike the static file approach, you don't need to regenerate JSON files or restart the API. The database is the single source of truth.

## Monitoring

### Check Logs

```bash
# API logs
docker-compose -f docker-compose.go.yml logs -f api

# Nginx logs
docker-compose -f docker-compose.go.yml logs -f nginx

# All services
docker-compose -f docker-compose.go.yml logs -f
```

### Health Check

```bash
# Check API health
curl http://localhost:8080/health

# Check through nginx
curl https://lookup.yourdomain.com/health
```

### Database Stats

```bash
# Connect to database
sqlite3 hamqrzdb.sqlite

# Count records
SELECT COUNT(*) FROM callsigns;

# Recent updates
SELECT callsign, expires, status 
FROM callsigns 
ORDER BY rowid DESC 
LIMIT 10;
```

## Troubleshooting

### API Won't Start

```bash
# Check database exists
ls -lh hamqrzdb.sqlite

# Check database is readable
sqlite3 hamqrzdb.sqlite "SELECT COUNT(*) FROM callsigns;"

# Check container logs
docker-compose -f docker-compose.go.yml logs api

# Check database permissions
chmod 644 hamqrzdb.sqlite
```

### Case-Insensitive Lookup Not Working

The Go API handles this automatically with `UPPER(callsign) = UPPER(?)` in SQL queries. If it's not working:

```bash
# Test database directly
sqlite3 hamqrzdb.sqlite "SELECT * FROM callsigns WHERE UPPER(callsign) = 'KJ5DJC';"

# Check API logs for errors
docker-compose -f docker-compose.go.yml logs api | grep -i error
```

### Slow Queries

```bash
# Check database has index on callsign
sqlite3 hamqrzdb.sqlite "PRAGMA index_list('callsigns');"

# Should see idx_callsigns_callsign

# If missing, create it:
sqlite3 hamqrzdb.sqlite "CREATE INDEX idx_callsigns_callsign ON callsigns(callsign);"
```

### SSL Certificate Issues

```bash
# Check certificate exists
ls -l certbot/conf/live/lookup.yourdomain.com/

# Renew certificate manually
docker-compose -f docker-compose.go.yml run --rm certbot renew

# Check nginx config
docker-compose -f docker-compose.go.yml exec nginx nginx -t
```

## Migration from Static Files

If you're migrating from the nginx static file approach:

1. **Stop old services**:
   ```bash
   docker-compose down
   ```

2. **Build database** (if not already done):
   ```bash
   python3 process_uls_db.py --full
   ```

3. **Start Go API**:
   ```bash
   docker-compose -f docker-compose.go.yml up -d
   ```

4. **Test both cases**:
   ```bash
   curl https://lookup.kj5djc.com/v1/KJ5DJC/json/test
   curl https://lookup.kj5djc.com/v1/kj5djc/json/test
   ```

5. **Optional: Delete old JSON files** (if you want to save space):
   ```bash
   # Backup first!
   tar -czf output-backup.tar.gz output/
   
   # Delete JSON files (keep index.html and 404.json)
   find output -name "*.json" ! -name "404.json" -delete
   ```

## Development

### Local Development

```bash
# Install dependencies
go mod download

# Run locally (no Docker)
DB_PATH=./hamqrzdb.sqlite PORT=8080 go run main.go

# Build binary
CGO_ENABLED=1 go build -o hamqrzdb-api main.go

# Run binary
./hamqrzdb-api
```

### Testing

```bash
# Test valid callsign
curl http://localhost:8080/v1/KJ5DJC/json/test | jq .

# Test invalid callsign
curl http://localhost:8080/v1/INVALID/json/test | jq .

# Test case-insensitivity
curl http://localhost:8080/v1/kj5djc/json/test | jq .
curl http://localhost:8080/v1/KJ5DJC/json/test | jq .

# Test health endpoint
curl http://localhost:8080/health

# Test homepage
curl http://localhost:8080/
```

## Credits

- **Data Source**: [FCC Universal Licensing System (ULS)](https://www.fcc.gov/uls)
- **Inspired By**: [k3ng/hamdb](https://github.com/k3ng/hamdb)
- **Author**: Chris Kacerguis (KJ5DJC)
- **License**: MIT

## Support

For issues or questions:
- GitHub: [chriskacerguis/hamqrzdb](https://github.com/chriskacerguis/hamqrzdb)
- QRZ: [KJ5DJC](https://www.qrz.com/db/KJ5DJC)

73! 📻
