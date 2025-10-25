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
docker compose exec api /app/hamqrzdb-process --full --db /data/hamqrzdb.sqlite

# 4. Test the API
curl http://localhost:8080/v1/kj5djc/json/test
```

**That's it!** The database is persistent across container restarts. Location data (latitude, longitude, and grid squares) is automatically processed if LA.dat is included in the FCC download.

### Updating the Database

```bash
# Daily updates (30 seconds)
docker compose exec api /app/hamqrzdb-process --daily --db /data/hamqrzdb.sqlite

# Full rebuild (includes location data)
docker compose exec api /app/hamqrzdb-process --full --db /data/hamqrzdb.sqlite
```