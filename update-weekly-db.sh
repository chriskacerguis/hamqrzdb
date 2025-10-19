#!/bin/bash

# Weekly full rebuild script for Ham Radio Callsign Lookup (Database Version)
# Run this weekly via cron to rebuild the entire database
# Non-destructive: rebuilds database and regenerates all JSON files

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="${SCRIPT_DIR}/output"
DB_FILE="${SCRIPT_DIR}/hamqrzdb.sqlite"
DB_BACKUP="${SCRIPT_DIR}/hamqrzdb.sqlite.backup"
TEMP_DIR="${SCRIPT_DIR}/output.tmp.$$"
LOG_FILE="${SCRIPT_DIR}/logs/weekly-$(date +%Y%m%d).log"

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

log "Starting weekly full database rebuild process..."

# Backup existing database if it exists
if [ -f "$DB_FILE" ]; then
    log "Backing up existing database..."
    cp "$DB_FILE" "$DB_BACKUP"
    log "Backup created: $DB_BACKUP"
    
    # Remove old database to start fresh
    rm "$DB_FILE"
fi

# Create temporary directory for new JSON files
mkdir -p "$TEMP_DIR"

# Download and process full database
log "Downloading and processing full database..."
if python3 "${SCRIPT_DIR}/process_uls_db.py" --full --db "$DB_FILE" --output "$TEMP_DIR" 2>&1 | tee -a "$LOG_FILE"; then
    log "Successfully created new database and JSON files"
else
    log "ERROR: Failed to process full database"
    
    # Restore backup if it exists
    if [ -f "$DB_BACKUP" ]; then
        log "Restoring database from backup..."
        cp "$DB_BACKUP" "$DB_FILE"
    fi
    
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

log "Weekly rebuild completed successfully"

# Get database stats
if [ -f "$DB_FILE" ]; then
    DB_SIZE=$(du -h "$DB_FILE" | cut -f1)
    log "Database size: $DB_SIZE"
fi

# Clean up old backup (keep most recent)
if [ -f "$DB_BACKUP" ]; then
    BACKUP_AGE=$(find "$DB_BACKUP" -mtime +7 2>/dev/null | wc -l)
    if [ "$BACKUP_AGE" -gt 0 ]; then
        log "Removing old database backup (>7 days old)"
        rm "$DB_BACKUP"
    fi
fi

# Clean up old logs (keep last 30 days)
find "${SCRIPT_DIR}/logs" -name "weekly-*.log" -mtime +30 -delete 2>/dev/null || true
