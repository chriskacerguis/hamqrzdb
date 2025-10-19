#!/bin/bash

# Daily update script for Ham Radio Callsign Lookup (Database Version)
# Run this daily via cron to keep the database up-to-date
# Non-destructive: updates database, regenerates only changed files

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="${SCRIPT_DIR}/output"
DB_FILE="${SCRIPT_DIR}/hamqrzdb.sqlite"
TEMP_DIR="${SCRIPT_DIR}/output.tmp.$$"
LOG_FILE="${SCRIPT_DIR}/logs/update-$(date +%Y%m%d).log"

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

log "Starting daily database update process..."

# Check if database exists
if [ ! -f "$DB_FILE" ]; then
    log "ERROR: Database file not found: $DB_FILE"
    log "Please run update-weekly-db.sh first to create the initial database"
    exit 1
fi

# Create temporary directory for regenerated files
mkdir -p "$TEMP_DIR"

# Download and process daily updates into database
log "Downloading and processing daily updates into database..."
if python3 "${SCRIPT_DIR}/process_uls_db.py" --daily --db "$DB_FILE" --output "$TEMP_DIR" 2>&1 | tee -a "$LOG_FILE"; then
    log "Successfully updated database with daily changes"
else
    log "ERROR: Failed to process daily updates"
    exit 1
fi

# Sync changes to output directory (only updates changed files)
log "Syncing changes to output directory..."
mkdir -p "$OUTPUT_DIR"
if rsync -av --delete "$TEMP_DIR/" "$OUTPUT_DIR/" 2>&1 | tee -a "$LOG_FILE"; then
    log "Successfully synced changes"
else
    log "ERROR: Failed to sync changes"
    exit 1
fi

log "Daily update completed successfully"

# Clean up old logs (keep last 30 days)
find "${SCRIPT_DIR}/logs" -name "update-*.log" -mtime +30 -delete 2>/dev/null || true
