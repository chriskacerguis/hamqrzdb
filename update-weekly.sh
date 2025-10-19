#!/bin/bash

# Weekly full rebuild script for Ham Radio Callsign Lookup
# Run this weekly via cron to rebuild the entire database
# Non-destructive: processes to temp directory then rsyncs changes

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="${SCRIPT_DIR}/output"
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

log "Starting weekly full rebuild process..."

# Create temporary directory
mkdir -p "$TEMP_DIR"

# Run the full rebuild to temp directory
log "Downloading and processing full database..."
if python3 "${SCRIPT_DIR}/process_uls.py" --full --output "$TEMP_DIR" 2>&1 | tee -a "$LOG_FILE"; then
    log "Successfully processed full database"
else
    log "ERROR: Failed to process full database"
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

log "Weekly rebuild completed successfully"

# Clean up old logs (keep last 30 days)
find "${SCRIPT_DIR}/logs" -name "weekly-*.log" -mtime +30 -delete 2>/dev/null || true

log "Weekly full rebuild completed successfully"

# Clean up old logs (keep last 90 days for weekly logs)
find "${SCRIPT_DIR}/logs" -name "weekly-*.log" -mtime +90 -delete

exit 0
