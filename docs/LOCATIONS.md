# HamQRZDB Locations Processor

Go CLI tool to process FCC ULS location data (LA.dat) and add coordinates and grid squares to callsigns.

## Overview

The **`hamqrzdb-locations`** tool processes location data from the FCC's LA.dat file and updates the database with:
- **Latitude** and **Longitude** (decimal degrees)
- **Maidenhead Grid Square** (6-character, e.g., EM10ci)

This is an **optional** processor that should be run **after** the main data load to add geographic information to callsigns.

## Quick Start

```bash
# Build the tool
make build

# Process location data from LA.dat
./bin/hamqrzdb-locations --la-file path/to/LA.dat

# Process and regenerate JSON files
./bin/hamqrzdb-locations --la-file path/to/LA.dat --regenerate

# Process specific callsign only
./bin/hamqrzdb-locations --la-file path/to/LA.dat --callsign KJ5DJC
```

## Usage

### Command Line Options

```
./bin/hamqrzdb-locations [options]

Options:
  --la-file string      Path to LA.dat file (required)
  --callsign string     Process only this specific callsign
  --db string           Path to SQLite database (default: hamqrzdb.sqlite)
  --regenerate          Regenerate JSON files after updating locations
```

### Examples

**Process all location data:**
```bash
./bin/hamqrzdb-locations --la-file temp_uls/LA.dat
```

**Process with custom database:**
```bash
./bin/hamqrzdb-locations --la-file temp_uls/LA.dat --db /var/lib/hamqrzdb.sqlite
```

**Process single callsign:**
```bash
./bin/hamqrzdb-locations --la-file temp_uls/LA.dat --callsign KJ5DJC
```

**Process and regenerate JSON files:**
```bash
# Note: The tool will show you the command to run
./bin/hamqrzdb-locations --la-file temp_uls/LA.dat --regenerate

# Then run the suggested command:
./bin/hamqrzdb-process --generate
```

## Workflow

### 1. Download FCC Data

The LA.dat file is included in the full FCC database download:

```bash
# Download full database
wget https://data.fcc.gov/download/pub/uls/complete/l_amat.zip

# Extract
unzip l_amat.zip -d temp_uls/

# You'll find LA.dat inside
ls temp_uls/LA.dat
```

### 2. Process Main Data First

**Always process the main callsign data before location data:**

```bash
# Process main data (HD, EN, AM files)
./bin/hamqrzdb-process --full

# OR use the downloaded files directly
./bin/hamqrzdb-process --file l_amat.zip
```

### 3. Add Location Data

```bash
# Add location data to existing database
./bin/hamqrzdb-locations --la-file temp_uls/LA.dat
```

### 4. (Optional) Regenerate JSON Files

If you're serving static JSON files via nginx:

```bash
# Regenerate JSON files with updated location data
./bin/hamqrzdb-process --generate
```

**Note:** If you're using the Go API server, you don't need to regenerate JSON files. The API reads directly from the database!

## What It Does

### Input: LA.dat File

The LA.dat file contains location records with fields like:
- Latitude (degrees, minutes, seconds, direction)
- Longitude (degrees, minutes, seconds, direction)
- Location type, site status, etc.

Example LA.dat record:
```
LA|12345|1|KJ5DJC|...|30|20|30|N|97|45|17|W|...
```

### Processing Steps

1. **Parse coordinates** - Converts degrees/minutes/seconds to decimal
2. **Calculate grid square** - Computes 6-character Maidenhead grid
3. **Update database** - Adds lat/lon/grid to callsign record
4. **Validate data** - Skips invalid coordinates (0,0)

### Output: Updated Database

The tool updates three fields in the `callsigns` table:
- `latitude` (REAL) - Decimal degrees (-90 to 90)
- `longitude` (REAL) - Decimal degrees (-180 to 180)
- `grid_square` (TEXT) - 6-character Maidenhead grid (e.g., "EM10ci")

## Maidenhead Grid Square

The tool calculates standard 6-character Maidenhead locator grid squares:

**Format:** `AABBCC`
- **AA** - Field (20° longitude × 10° latitude)
- **BB** - Square (2° longitude × 1° latitude)  
- **CC** - Subsquare (5' longitude × 2.5' latitude)

**Example:** `EM10ci`
- **EM** - Field covering Texas/Oklahoma region
- **10** - Square in central Texas
- **ci** - Subsquare for Austin, TX area

## Performance

Processing ~1M location records:
- **Time**: 2-3 minutes
- **Memory**: ~50 MB
- **Updates**: ~800K callsigns (those with location data)
- **Batch size**: 1,000 records per transaction

**Comparison with Python:**
| Task | Python | Go | Improvement |
|------|--------|-----|-------------|
| Process 1M records | 8-10 min | 2-3 min | **3-4x faster** |
| Memory usage | ~200 MB | ~50 MB | **4x less** |

## Integration

### With hamqrzdb-process

Process everything in one go:

```bash
# 1. Download and extract
wget https://data.fcc.gov/download/pub/uls/complete/l_amat.zip
unzip l_amat.zip -d temp_uls/

# 2. Process main data
./bin/hamqrzdb-process --file l_amat.zip

# 3. Add location data
./bin/hamqrzdb-locations --la-file temp_uls/LA.dat
```

### With Docker

Add location processing to your Docker workflow:

```bash
# In your update script
docker exec hamqrzdb-api /app/hamqrzdb-locations \
  --la-file /data/temp_uls/LA.dat \
  --db /data/hamqrzdb.sqlite
```

### Automation

**Daily Location Updates:**

```bash
#!/bin/bash
# daily-locations-update.sh

DATE=$(date +%m%d%Y)
URL="https://data.fcc.gov/download/pub/uls/daily/l_am_${DATE}.zip"
TEMP_DIR="/tmp/uls-daily-$$"

# Download and extract
mkdir -p $TEMP_DIR
wget -O $TEMP_DIR/daily.zip $URL
unzip -q $TEMP_DIR/daily.zip -d $TEMP_DIR

# Process locations if LA.dat exists
if [ -f $TEMP_DIR/LA.dat ]; then
    /usr/local/bin/hamqrzdb-locations \
        --la-file $TEMP_DIR/LA.dat \
        --db /var/lib/hamqrzdb/hamqrzdb.sqlite
fi

# Cleanup
rm -rf $TEMP_DIR
```

**Cron job:**
```bash
# Daily at 3 AM (after main data update)
0 3 * * * /usr/local/bin/daily-locations-update.sh
```

## Troubleshooting

### File Not Found

```bash
# Error: file not found: temp_uls/LA.dat

# Solution: Check file exists
ls -l temp_uls/LA.dat

# Make sure you extracted the ZIP first
unzip l_amat.zip -d temp_uls/
```

### No Callsigns Updated

```bash
# Processed 0 LA records, updated 0 callsigns

# Cause: Database doesn't have callsign records yet
# Solution: Run hamqrzdb-process first
./bin/hamqrzdb-process --full
```

### Invalid Coordinates

The tool automatically skips:
- Coordinates at (0, 0)
- Missing or malformed data
- Records with invalid degrees/minutes/seconds

### Database Locked

```bash
# Error: database is locked

# Solution: Close other connections
# WAL mode (default) should prevent this
```

## Database Queries

After processing, you can query location data:

```bash
# Find callsigns with location data
sqlite3 hamqrzdb.sqlite "
  SELECT callsign, latitude, longitude, grid_square 
  FROM callsigns 
  WHERE latitude IS NOT NULL 
  LIMIT 10;
"

# Count callsigns with locations
sqlite3 hamqrzdb.sqlite "
  SELECT COUNT(*) 
  FROM callsigns 
  WHERE latitude IS NOT NULL;
"

# Find callsigns in a specific grid square
sqlite3 hamqrzdb.sqlite "
  SELECT callsign, first_name, last_name, city, state
  FROM callsigns 
  WHERE grid_square LIKE 'EM10%';
"

# Find callsigns near coordinates (Austin, TX)
sqlite3 hamqrzdb.sqlite "
  SELECT callsign, grid_square,
         ROUND(latitude, 4) as lat,
         ROUND(longitude, 4) as lon
  FROM callsigns
  WHERE latitude BETWEEN 30.0 AND 30.5
    AND longitude BETWEEN -98.0 AND -97.5
  ORDER BY callsign
  LIMIT 20;
"
```

## API Response

When using the Go API server, location data is included in responses:

```json
{
  "hamdb": {
    "version": "1",
    "callsign": {
      "call": "KJ5DJC",
      "class": "G",
      "status": "A",
      "grid": "EM10ci",
      "lat": "30.3416503",
      "lon": "-97.7548379",
      "fname": "CHRIS",
      "name": "KACERGUIS",
      ...
    },
    "messages": {
      "status": "OK"
    }
  }
}
```

## Best Practices

1. ✅ **Process main data first** - LA.dat references existing callsigns
2. ✅ **Run after full updates** - Weekly or monthly
3. ✅ **Skip for daily updates** - Location data rarely changes
4. ✅ **Use with Go API** - No need to regenerate JSON files
5. ✅ **Validate coordinates** - Tool automatically skips invalid data
6. ✅ **Batch processing** - 1,000 records per transaction for speed

## See Also

- **README.cli.md** - Main CLI documentation
- **COMPARISON.md** - Python vs Go performance comparison
- **DEPLOY.md** - Production deployment guide
- **QUICKREF.md** - Quick reference card

## Credits

- **Author**: Chris Kacerguis (KJ5DJC)
- **Data Source**: FCC Universal Licensing System (ULS)
- **Grid Square Algorithm**: Maidenhead Locator System
- **License**: MIT

73! 📻
