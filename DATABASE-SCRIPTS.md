# Database-Based Update Scripts

These scripts use the SQLite database approach for more efficient and flexible data management.

## Scripts Overview

### 1. `update-weekly-db.sh` - Weekly Full Rebuild
**Purpose**: Download full FCC database, rebuild SQLite database from scratch, generate all JSON files.

**When to use**:
- Initial setup (first time)
- Weekly maintenance
- After database corruption
- When you want a clean rebuild

**What it does**:
1. Backs up existing database (if any)
2. Downloads full FCC ULS data
3. Creates fresh SQLite database
4. Generates all JSON files from database
5. Syncs to output directory with rsync

**Run it**:
```bash
./update-weekly-db.sh
```

### 2. `update-daily-db.sh` - Daily Incremental Updates
**Purpose**: Download daily FCC changes, update database, regenerate only changed files.

**When to use**:
- Daily automated updates
- After initial database is created

**What it does**:
1. Checks that database exists
2. Downloads daily FCC changes
3. Updates/inserts changed records in database
4. Regenerates JSON files for changed callsigns
5. Syncs to output directory with rsync

**Run it**:
```bash
./update-daily-db.sh
```

### 3. `regenerate-json-db.sh` - Regenerate JSON from Database
**Purpose**: Regenerate all JSON files from existing database without downloading new data.

**When to use**:
- JSON files got corrupted or deleted
- Changed JSON output format
- Added new fields to output
- Testing changes

**What it does**:
1. Reads from existing database
2. Generates all JSON files
3. Syncs to output directory

**Run it**:
```bash
./regenerate-json-db.sh
```

## Setup Instructions

### First Time Setup

```bash
# 1. Make scripts executable
chmod +x update-weekly-db.sh
chmod +x update-daily-db.sh
chmod +x regenerate-json-db.sh

# 2. Run initial full build
./update-weekly-db.sh

# This creates:
# - hamqrzdb.sqlite (database file)
# - output/ (JSON files directory)
```

### Automated Updates with Cron

```bash
# Edit crontab
crontab -e

# Add these lines:
# Daily updates at 2 AM
0 2 * * * cd /path/to/hamqrzdb && ./update-daily-db.sh >> logs/cron.log 2>&1

# Weekly full rebuild on Sunday at 3 AM
0 3 * * 0 cd /path/to/hamqrzdb && ./update-weekly-db.sh >> logs/cron.log 2>&1
```

## File Structure

```
hamqrzdb/
├── hamqrzdb.sqlite          # SQLite database (source of truth)
├── hamqrzdb.sqlite.backup   # Auto-backup (created during weekly updates)
├── output/                  # Generated JSON files (served by nginx)
├── logs/                    # Log files from updates
│   ├── update-*.log
│   ├── weekly-*.log
│   └── regenerate-*.log
├── process_uls_db.py        # Database processor
├── update-daily-db.sh       # Daily update script
├── update-weekly-db.sh      # Weekly rebuild script
└── regenerate-json-db.sh    # Regenerate JSON script
```

## Advantages of Database Approach

### 1. **Memory Efficient**
- Loads data in small batches
- Never holds entire dataset in RAM
- Works on systems with limited memory

### 2. **Fast Daily Updates**
- Only downloads changed records
- Updates database with upserts
- Regenerates only changed JSON files
- Typical daily update: 1-2 minutes

### 3. **Disaster Recovery**
- Database is the source of truth
- Lost JSON files? Regenerate in ~5-10 minutes
- Automatic backups during weekly rebuilds

### 4. **Flexible**
- Want to change JSON format? Update code, regenerate
- Need statistics? Query the database
- Want to add coordinates? Update schema, regenerate

### 5. **Efficient Storage**
- Database: ~300-500MB (compressed with indexes)
- JSON files: ~1-2GB
- Only need to backup the database

## Database Schema

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
```

## Manual Operations

### Query the Database
```bash
# Open SQLite database
sqlite3 hamqrzdb.sqlite

# Example queries:
sqlite> SELECT COUNT(*) FROM callsigns;
sqlite> SELECT * FROM callsigns WHERE callsign = 'KJ5DJC';
sqlite> SELECT operator_class, COUNT(*) FROM callsigns GROUP BY operator_class;
sqlite> .quit
```

### Generate Single Callsign
```bash
# Regenerate just one callsign
python3 process_uls_db.py --generate --callsign KJ5DJC --db hamqrzdb.sqlite
```

### Check Database Size
```bash
# Database size
du -h hamqrzdb.sqlite

# Output directory size
du -sh output/

# Detailed breakdown
du -sh output/*/
```

## Troubleshooting

### Database Doesn't Exist
```bash
# Run weekly rebuild to create it
./update-weekly-db.sh
```

### Daily Update Fails
```bash
# Check if database exists
ls -lh hamqrzdb.sqlite

# If corrupted, restore from backup
cp hamqrzdb.sqlite.backup hamqrzdb.sqlite

# Or rebuild from scratch
./update-weekly-db.sh
```

### JSON Files Missing
```bash
# Regenerate all JSON files from database
./regenerate-json-db.sh
```

### Out of Disk Space
```bash
# Clean up old logs
find logs/ -name "*.log" -mtime +30 -delete

# Remove old backups
rm hamqrzdb.sqlite.backup

# Check space
df -h .
```

### Database Locked Error
```bash
# Check for running processes
ps aux | grep process_uls_db.py

# Kill if stuck
pkill -f process_uls_db.py

# Wait a moment and retry
sleep 5
./update-daily-db.sh
```

## Performance Tips

### For Systems with Limited Memory
```bash
# The database version already uses minimal memory (~50-100MB)
# If still having issues, ensure you have swap enabled:
free -h
sudo swapon --show
```

### For Faster Processing
```bash
# Use SSD for database and output files
# Database benefits from fast random access
# JSON generation benefits from fast sequential writes
```

### For Large Scale Deployments
```bash
# Keep database on fast storage
# Serve JSON files from separate volume
# Use database for queries, JSON for serving
```

## Comparison: Database vs. Streaming

| Feature | Database Version | Streaming Version |
|---------|-----------------|-------------------|
| Memory Usage | ~50-100MB | ~100-500MB |
| Initial Build | ~15-20 min | ~30-45 min |
| Daily Updates | ~1-2 min | ~10-15 min |
| Regenerate All | ~5-10 min | ~30-45 min |
| Disk Usage | DB + JSON (~2GB total) | JSON only (~1-2GB) |
| Flexibility | High (query, modify) | Low (regenerate all) |
| Best For | Production, automation | One-off processing |

## Logs

All scripts log to `logs/` directory:
- `update-*.log` - Daily updates
- `weekly-*.log` - Weekly rebuilds  
- `regenerate-*.log` - JSON regeneration
- Old logs auto-deleted after 30 days

View logs:
```bash
# Latest daily update
tail -f logs/update-$(date +%Y%m%d).log

# Latest weekly rebuild
tail -f logs/weekly-$(date +%Y%m%d).log

# All recent logs
ls -lt logs/ | head
```
