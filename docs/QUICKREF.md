# HamQRZDB - Quick Reference

## 🚀 Quick Start

```bash
# Build tools
./build.sh

# Process FCC data
./bin/hamqrzdb-process --full

# Add location data (optional)
./bin/hamqrzdb-locations --la-file temp_uls/LA.dat

# Start API
./bin/hamqrzdb-api
```

## 📦 Build Commands

```bash
./build.sh              # Easy build script
make build              # Build both tools
make clean              # Clean build artifacts
make install            # Install to /usr/local/bin
```

## 🔄 Data Processing

```bash
# Full database
./bin/hamqrzdb-process --full

# Daily updates  
./bin/hamqrzdb-process --daily

# Generate JSON files
./bin/hamqrzdb-process --generate

# Single callsign
./bin/hamqrzdb-process --full --callsign KJ5DJC

# Custom paths
./bin/hamqrzdb-process --full --db custom.db --output /var/www/data

# Local file
./bin/hamqrzdb-process --file /path/to/l_amat.zip
```

## 📍 Location Data

```bash
# Add coordinates and grid squares
./bin/hamqrzdb-locations --la-file temp_uls/LA.dat

# Process single callsign
./bin/hamqrzdb-locations --la-file temp_uls/LA.dat --callsign KJ5DJC

# Custom database
./bin/hamqrzdb-locations --la-file temp_uls/LA.dat --db custom.db

# Note about regeneration
./bin/hamqrzdb-locations --la-file temp_uls/LA.dat --regenerate
```

## 🌐 API Server

```bash
# Start API
./bin/hamqrzdb-api

# Custom config
DB_PATH=./hamqrzdb.sqlite PORT=9000 ./bin/hamqrzdb-api

# Test endpoints
curl http://localhost:8080/v1/KJ5DJC/json/test
curl http://localhost:8080/v1/kj5djc/json/test  # Case-insensitive!
curl http://localhost:8080/health
```

## 🐳 Docker

```bash
make docker-build       # Build image
make docker-run         # Start services
make docker-stop        # Stop services
make docker-logs        # View logs
```

## 💾 Database

```bash
make db-full           # Full download & process
make db-daily          # Daily updates
make db-generate       # Generate JSON files
make db-stats          # Show statistics

# Direct SQLite queries
sqlite3 hamqrzdb.sqlite "SELECT COUNT(*) FROM callsigns;"
sqlite3 hamqrzdb.sqlite "SELECT * FROM callsigns WHERE UPPER(callsign) = 'KJ5DJC';"
```

## ⚡ Performance

| Task | Python | Go | Speedup |
|------|--------|-----|---------|
| Full DB | 15-20 min | 3-5 min | **4-5x** |
| Daily | 2-3 min | 20-30 sec | **4-6x** |
| JSON | 25-30 min | 5-10 min | **3-5x** |
| Memory | ~500 MB | ~100 MB | **5x less** |

## 📄 File Locations

```
bin/hamqrzdb-process    # Data processor binary
bin/hamqrzdb-locations  # Locations processor binary
bin/hamqrzdb-api        # API server binary
hamqrzdb.sqlite         # Database (~500 MB)
output/                 # JSON files (~2 GB, optional)
```

## 🔧 Automation

### Cron Job
```bash
# Daily updates at 2 AM
0 2 * * * /usr/local/bin/hamqrzdb-process --daily --db /var/lib/hamqrzdb.sqlite
```

### Systemd
```bash
# /etc/systemd/system/hamqrzdb-api.service
[Service]
ExecStart=/usr/local/bin/hamqrzdb-api
Environment="DB_PATH=/var/lib/hamqrzdb.sqlite"
Environment="PORT=8080"
```

## 🔍 Troubleshooting

```bash
# Check database
ls -lh hamqrzdb.sqlite

# View table count
sqlite3 hamqrzdb.sqlite "SELECT COUNT(*) FROM callsigns;"

# Check binary
./bin/hamqrzdb-process --help
./bin/hamqrzdb-api --help  # Not yet implemented

# Check API
curl http://localhost:8080/health

# View logs (Docker)
docker-compose -f docker-compose.go.yml logs -f api
```

## 📚 Documentation

- **QUICKREF.md** - This quick reference guide
- **CLI-SUMMARY.md** - Complete summary of CLI tools
- **README.cli.md** - Full CLI documentation
- **LOCATIONS.md** - Locations processor guide
- **COMPARISON.md** - Python vs Go benchmarks
- **DEPLOY.md** - Production deployment guide
- **README.md** - General project overview

## 🆘 Common Issues

**Build fails with "gcc not found":**
```bash
# macOS
xcode-select --install

# Ubuntu/Debian
sudo apt-get install build-essential

# Alpine
apk add gcc musl-dev sqlite-dev
```

**Database locked:**
```bash
# Close other connections
# WAL mode is enabled by default
```

**API returns NOT_FOUND:**
```bash
# Make sure database exists and has data
./bin/hamqrzdb-process --full
```

**Port already in use:**
```bash
# Use different port
PORT=9000 ./bin/hamqrzdb-api
```

## 🎯 Best Practices

1. ✅ Use Go CLI for production
2. ✅ Run daily updates via cron
3. ✅ Skip JSON generation (use API directly)
4. ✅ Use Docker for deployment
5. ✅ Monitor with health checks
6. ✅ Use WAL mode (default)
7. ✅ Keep backups of database

## 📊 Database Schema

```sql
CREATE TABLE callsigns (
    callsign TEXT PRIMARY KEY,
    license_status TEXT,
    operator_class TEXT,
    first_name TEXT,
    last_name TEXT,
    -- ... more fields
    last_updated TIMESTAMP
);
```

## 🔗 URLs

- **GitHub**: https://github.com/chriskacerguis/hamqrzdb
- **QRZ**: https://www.qrz.com/db/KJ5DJC
- **FCC ULS**: https://www.fcc.gov/uls

## 📞 Support

- Open issues on GitHub
- See documentation files
- Contact via QRZ

---
73! 📻 - KJ5DJC
