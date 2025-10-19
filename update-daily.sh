#!/bin/bash

# Daily update script for Ham Radio Callsign Lookup
# Run this daily via cron to keep the database up-to-date
# Non-destructive: processes to temp directory then rsyncs changes

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="${SCRIPT_DIR}/output"
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

log "Starting daily update process..."

# Create temporary directory
mkdir -p "$TEMP_DIR"

# Copy existing data to temp directory (preserve current state)
log "Copying existing data to temporary directory..."
if [ -d "$OUTPUT_DIR" ]; then
    rsync -a "$OUTPUT_DIR/" "$TEMP_DIR/" 2>&1 | tee -a "$LOG_FILE"
fi

# Run the daily update to temp directory
log "Downloading and processing daily updates..."
if python3 "${SCRIPT_DIR}/process_uls.py" --daily --output "$TEMP_DIR" 2>&1 | tee -a "$LOG_FILE"; then
    log "Successfully processed daily updates"
else
    log "ERROR: Failed to process daily updates"
    exit 1
fi

# Rsync changes to output directory (only updates changed files)
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

exit 0
