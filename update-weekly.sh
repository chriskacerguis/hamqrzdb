#!/bin/bash

# Weekly full rebuild script for Ham Radio Callsign Lookup
# Run this weekly via cron to rebuild the entire database

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="${SCRIPT_DIR}/output"
S3_BUCKET="s3://your-bucket-name"
LOG_FILE="${SCRIPT_DIR}/logs/weekly-$(date +%Y%m%d).log"

# Create logs directory if it doesn't exist
mkdir -p "${SCRIPT_DIR}/logs"

# Log function
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log "Starting weekly full rebuild process..."

# Remove old output directory
log "Cleaning old output directory..."
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# Run the full rebuild
log "Downloading and processing full database..."
if python3 "${SCRIPT_DIR}/process_uls.py" --full --output "$OUTPUT_DIR" 2>&1 | tee -a "$LOG_FILE"; then
    log "Successfully processed full database"
else
    log "ERROR: Failed to process full database"
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

# Create CloudFront invalidation (optional)
# DISTRIBUTION_ID="YOUR_DISTRIBUTION_ID"
# log "Creating CloudFront invalidation..."
# aws cloudfront create-invalidation --distribution-id "$DISTRIBUTION_ID" --paths "/*" | tee -a "$LOG_FILE"

log "Weekly full rebuild completed successfully"

# Clean up old logs (keep last 90 days for weekly logs)
find "${SCRIPT_DIR}/logs" -name "weekly-*.log" -mtime +90 -delete

exit 0
