# HamQRZDB - Quick Reference

## 🚀 Quick Start

```bash
# Build tools
task build

# Process FCC data
./bin/hamqrzdb-process --full

# Add location data (optional)
./bin/hamqrzdb-process --la-file temp_uls/LA.dat

# Or process both in one command
./bin/hamqrzdb-process --full --la-file temp_uls/LA.dat

# Start API
./bin/hamqrzdb-api
```

## 📦 Build Commands

```bash
task build              # Build all tools
task build:process      # Build process binary
task build:api          # Build API binary
task clean              # Clean build artifacts
```

## 🔄 Data Processing

```bash
# Full database
./bin/hamqrzdb-process --full

# Daily updates  
./bin/hamqrzdb-process --daily

# Single callsign
./bin/hamqrzdb-process --full --callsign KJ5DJC

# Custom database path
./bin/hamqrzdb-process --full --db custom.db

# Local file
./bin/hamqrzdb-process --file /path/to/l_amat.zip
```

## 📍 Location Data

```bash
# Add coordinates and grid squares
./bin/hamqrzdb-process --la-file temp_uls/LA.dat

# Process single callsign
./bin/hamqrzdb-process --la-file temp_uls/LA.dat --callsign KJ5DJC

# Custom database
./bin/hamqrzdb-process --la-file temp_uls/LA.dat --db custom.db

# Combined with full processing
./bin/hamqrzdb-process --full --la-file temp_uls/LA.dat
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
task docker:build       # Build image
task docker:up          # Start services
task docker:down        # Stop services
task docker:logs        # View logs
```

## 💾 Database

```bash
task db:full           # Full download & process
task db:daily          # Daily updates
task db:locations      # Process location data
task db:stats          # Show statistics

# Direct SQLite queries
sqlite3 hamqrzdb.sqlite "SELECT COUNT(*) FROM callsigns;"
sqlite3 hamqrzdb.sqlite "SELECT * FROM callsigns WHERE call = 'KJ5DJC';"
```

## ⚡ Performance

| Task | Python | Go | Speedup |
|------|--------|-----|---------|
| Full DB | 15-20 min | 3-5 min | **4-5x** |
| Daily | 2-3 min | 20-30 sec | **4-6x** |
| Memory | ~500 MB | ~100 MB | **5x less** |

## 📄 File Locations

```
bin/hamqrzdb-process    # Data & location processor binary (~6.8 MB)
bin/hamqrzdb-api        # API server binary (~6.7 MB)
hamqrzdb.sqlite         # Database (~500 MB)
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
