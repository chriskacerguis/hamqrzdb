#!/bin/bash
set -e

# Database path from environment or default
DB_PATH="${DB_PATH:-/data/hamqrzdb.sqlite}"

# Check if database exists
if [ ! -f "$DB_PATH" ]; then
    echo "� Database file not found at $DB_PATH"
    echo "� Creating empty database with schema..."
    
    # Create empty database with schema
    sqlite3 "$DB_PATH" << 'EOF'
CREATE TABLE IF NOT EXISTS callsigns (
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

CREATE INDEX IF NOT EXISTS idx_callsign ON callsigns(callsign);
CREATE INDEX IF NOT EXISTS idx_status ON callsigns(license_status);
EOF
    
    echo "✅ Empty database created!"
    echo "📥 To populate with FCC data, run:"
    echo "   docker compose exec api /app/hamqrzdb-import-us --full --db $DB_PATH"
else
    # Count records to verify database
    RECORD_COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM callsigns;" 2>/dev/null || echo "0")
    if [ "$RECORD_COUNT" -eq "0" ]; then
        echo "⚠️  Database exists but is empty (0 callsigns)"
        echo "📥 To populate with FCC data, run:"
        echo "   docker compose exec api /app/hamqrzdb-import-us --full --db $DB_PATH"
    else
        echo "📊 Database found with $RECORD_COUNT callsigns"
    fi
fi

# Start the API server
echo "🚀 Starting API server on port ${PORT:-8080}..."
exec /app/hamqrzdb-api
