#!/bin/bash

# Daily update script for Ham Radio Callsign Lookup
# Run this daily via cron to keep the database up-to-date

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="${SCRIPT_DIR}/output"
S3_BUCKET="s3://your-bucket-name"
LOG_FILE="${SCRIPT_DIR}/logs/update-$(date +%Y%m%d).log"

# Create logs directory if it doesn't exist
mkdir -p "${SCRIPT_DIR}/logs"

# Log function
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log "Starting daily update process..."

# Run the daily update
log "Downloading and processing daily updates..."
if python3 "${SCRIPT_DIR}/process_uls.py" --daily --output "$OUTPUT_DIR" 2>&1 | tee -a "$LOG_FILE"; then
    log "Successfully processed daily updates"
else
    log "ERROR: Failed to process daily updates"
    exit 1
fi

# Sync to S3
log "Syncing to S3..."
if aws s3 sync "$OUTPUT_DIR/" "$S3_BUCKET/" --delete --only-show-errors 2>&1 | tee -a "$LOG_FILE"; then
    log "Successfully synced to S3"
else
    log "ERROR: Failed to sync to S3"
    exit 1
fi

# Create CloudFront invalidation (optional, only if needed)
# DISTRIBUTION_ID="YOUR_DISTRIBUTION_ID"
# log "Creating CloudFront invalidation..."
# aws cloudfront create-invalidation --distribution-id "$DISTRIBUTION_ID" --paths "/*" | tee -a "$LOG_FILE"

log "Daily update completed successfully"

# Clean up old logs (keep last 30 days)
find "${SCRIPT_DIR}/logs" -name "update-*.log" -mtime +30 -delete

exit 0
