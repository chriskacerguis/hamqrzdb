#!/bin/bash

# Regenerate JSON files from existing database
# Useful if JSON files are lost or corrupted, or after schema changes

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="${SCRIPT_DIR}/output"
DB_FILE="${SCRIPT_DIR}/hamqrzdb.sqlite"
TEMP_DIR="${SCRIPT_DIR}/output.tmp.$$"
LOG_FILE="${SCRIPT_DIR}/logs/regenerate-$(date +%Y%m%d-%H%M%S).log"

# Create logs directory if it doesn't exist
mkdir -p "${SCRIPT_DIR}/logs"

# Cleanup function
cleanup() {
    if [ -d "$TEMP_DIR" ]; then
        log "Cleaning up temporary directory..."
        rm -rf "$TEMP_DIR"
    fi
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Log function
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log "Starting JSON file regeneration from database..."

# Check if database exists
if [ ! -f "$DB_FILE" ]; then
    log "ERROR: Database file not found: $DB_FILE"
    log "Please run update-weekly-db.sh first to create the database"
    exit 1
fi

# Get database info
DB_SIZE=$(du -h "$DB_FILE" | cut -f1)
log "Database size: $DB_SIZE"

# Create temporary directory for regenerated files
mkdir -p "$TEMP_DIR"

# Regenerate all JSON files from database
log "Regenerating JSON files from database..."
if python3 "${SCRIPT_DIR}/process_uls_db.py" --generate --db "$DB_FILE" --output "$TEMP_DIR" 2>&1 | tee -a "$LOG_FILE"; then
    log "Successfully regenerated JSON files"
else
    log "ERROR: Failed to regenerate JSON files"
    exit 1
fi

# Sync changes to output directory
log "Syncing changes to output directory..."
mkdir -p "$OUTPUT_DIR"
if rsync -av --delete "$TEMP_DIR/" "$OUTPUT_DIR/" 2>&1 | tee -a "$LOG_FILE"; then
    log "Successfully synced changes"
else
    log "ERROR: Failed to sync changes"
    exit 1
fi

# Get output directory size
OUTPUT_SIZE=$(du -sh "$OUTPUT_DIR" | cut -f1)
log "Output directory size: $OUTPUT_SIZE"

log "Regeneration completed successfully"

# Clean up old logs (keep last 30 days)
find "${SCRIPT_DIR}/logs" -name "regenerate-*.log" -mtime +30 -delete 2>/dev/null || true
