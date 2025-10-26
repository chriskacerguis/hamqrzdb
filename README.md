# HamQRZDB

A high-performance, self-hosted amateur radio callsign lookup system with a HamDB-compatible JSON API, built with **Go** for speed and efficiency. It processes FCC ULS data into SQLite and serves it via a fast HTTP API with case-insensitive lookups and CORS support.

## Quick Start

### Docker Compose (Recommended)

The container automatically creates an empty database on first run. You then populate it with FCC data.

```bash
# 1. Create docker-compose.yml
cat > docker-compose.yml << 'EOF'
services:
  api:
    image: ghcr.io/chriskacerguis/hamqrzdb:latest
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - hamqrzdb_data:/data
    environment:
      - DB_PATH=/data/hamqrzdb.sqlite
      - PORT=8080
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s

volumes:
  hamqrzdb_data:
EOF

# 2. Start the container (creates empty database)
docker compose up -d

# 3. Populate the database with FCC data (3-5 minutes, one-time, includes location data)
docker compose exec api /app/hamqrzdb-import-us --full --db /data/hamqrzdb.sqlite

# 4. Test the API
curl http://localhost:8080/v1/kj5djc/json/test
```

**That's it!** The database is persistent across container restarts. Location data (latitude, longitude, and grid squares) is automatically processed if LA.dat is included in the FCC download.

### Docker Compose Commands

#### Database Management

```bash
# Populate database with FCC data (first time, 3-5 minutes)
docker compose exec api /app/hamqrzdb-import-us --full --db /data/hamqrzdb.sqlite

# Daily updates (30 seconds)
docker compose exec api /app/hamqrzdb-import-us --daily --db /data/hamqrzdb.sqlite

# Import UK amateur radio data (Ofcom)
docker compose exec api /app/hamqrzdb-import-uk --db /data/hamqrzdb.sqlite
```

#### Database Inspection

```bash
# Check total callsign count
docker compose exec api sqlite3 /data/hamqrzdb.sqlite "SELECT COUNT(*) FROM callsigns"

# Check active licenses
docker compose exec api sqlite3 /data/hamqrzdb.sqlite "SELECT COUNT(*) FROM callsigns WHERE license_status = 'A'"

# Look up a specific callsign
docker compose exec api sqlite3 /data/hamqrzdb.sqlite "SELECT callsign, first_name, last_name, city, state FROM callsigns WHERE callsign = 'KJ5DJC'"

# Check UK callsigns
docker compose exec api sqlite3 /data/hamqrzdb.sqlite "SELECT COUNT(*) FROM callsigns WHERE radio_service_code = 'UK'"

# Interactive SQLite shell
docker compose exec api sqlite3 /data/hamqrzdb.sqlite
```

#### Container Management

```bash
# View logs
docker compose logs -f api

# Check container health
docker compose ps

# Restart the API (keeps database)
docker compose restart api

# Stop all services
docker compose down

# Stop and remove volumes (deletes database!)
docker compose down -v
```

#### Troubleshooting

```bash
# Open shell inside container
docker compose exec api sh

# Check files in data directory
docker compose exec api ls -lh /data/

# View database schema
docker compose exec api sqlite3 /data/hamqrzdb.sqlite ".schema callsigns"

# Check database integrity
docker compose exec api sqlite3 /data/hamqrzdb.sqlite "PRAGMA integrity_check"
```

### Updating the Database

```bash
# Daily updates (30 seconds)
docker compose exec api /app/hamqrzdb-import-us --daily --db /data/hamqrzdb.sqlite

# Full rebuild (includes location data)
docker compose exec api /app/hamqrzdb-import-us --full --db /data/hamqrzdb.sqlite
```