# Deployment Guide - Go API

This guide will help you deploy the new Go API to your production server at `lookup.kj5djc.com`.

## What Changed

**Before**: nginx serving 1.5M static JSON files (~2GB)
- ❌ Case-sensitive URLs only
- ❌ Must regenerate all JSON files for updates
- ❌ Large disk footprint

**After**: Go API querying SQLite directly (~500MB)
- ✅ Case-insensitive URLs (both `/v1/KJ5DJC/json/app` and `/v1/kj5djc/json/app` work!)
- ✅ No JSON files needed - database is source of truth
- ✅ Smaller disk footprint
- ✅ Faster queries
- ✅ Simpler updates (just update database, no restart needed)

## Step-by-Step Deployment

### 1. On Your Production Server

SSH into your server:
```bash
ssh root@157.230.133.121
cd /root/hamqrzdb
```

### 2. Build the Database (if not already done)

```bash
# Full rebuild
python3 process_uls_db.py --full

# This creates hamqrzdb.sqlite (~500MB)
# Takes a few minutes to process 1.5M records
```

### 3. Pull the New Code

```bash
# Pull the latest code with Go API
git pull origin main

# You should now have:
# - main.go (Go API server)
# - go.mod and go.sum (Go dependencies)
# - Dockerfile (Go API container)
# - docker-compose.go.yml (deployment config)
# - nginx-proxy.conf (SSL proxy)
```

### 4. Build the Docker Image

```bash
# Build the Go API container
docker build -t hamqrzdb-api:latest .

# This creates a ~20MB Docker image (vs nginx + 2GB of files)
```

### 5. Stop the Old Service

```bash
# Stop the old nginx service
docker-compose down

# Backup your old docker-compose.yml if needed
cp docker-compose.yml docker-compose.nginx-backup.yml
```

### 6. Deploy the Go API

```bash
# Start the Go API with SSL
docker-compose -f docker-compose.go.yml up -d

# Check logs
docker-compose -f docker-compose.go.yml logs -f api
```

You should see:
```
Connected to database: /data/hamqrzdb.sqlite
Starting server on port 8080
```

### 7. Test Case-Insensitive Lookups

```bash
# Test uppercase (worked before)
curl https://lookup.kj5djc.com/v1/KJ5DJC/json/test | jq -r .hamdb.messages.status
# Should return: OK

# Test lowercase (this is the FIX!)
curl https://lookup.kj5djc.com/v1/kj5djc/json/test | jq -r .hamdb.messages.status
# Should return: OK (not NOT_FOUND!)

# Test mixed case
curl https://lookup.kj5djc.com/v1/Kj5dJc/json/test | jq -r .hamdb.messages.status
# Should return: OK

# Test health endpoint
curl https://lookup.kj5djc.com/health
# Should return: {"status":"healthy"}
```

### 8. Verify SSL and Homepage

```bash
# Check SSL is working
curl -I https://lookup.kj5djc.com/

# Check homepage
curl https://lookup.kj5djc.com/
# Should show your beautiful index.html
```

## Updating Data

With the Go API, updates are MUCH simpler:

### Daily Updates (Incremental)

```bash
# Update database with daily changes
python3 process_uls_db.py --daily

# That's it! No restart needed.
# Changes are available immediately on next query.
```

### Weekly Full Rebuild

```bash
# Full rebuild
python3 process_uls_db.py --full

# Optional: Add location data
python3 process_uls_locations.py --full

# Still no restart needed!
```

## Troubleshooting

### Case-Insensitive Lookup Not Working

```bash
# Check database has data
sqlite3 hamqrzdb.sqlite "SELECT COUNT(*) FROM callsigns;"
# Should return: 1575334 (or similar)

# Test database query directly
sqlite3 hamqrzdb.sqlite "SELECT * FROM callsigns WHERE UPPER(callsign) = 'KJ5DJC';"

# Check API logs
docker-compose -f docker-compose.go.yml logs api | tail -20
```

### API Returns NOT_FOUND for Valid Callsign

```bash
# Check database file exists
ls -lh hamqrzdb.sqlite

# Check Docker volume mount
docker-compose -f docker-compose.go.yml exec api ls -lh /data/

# Check API can read database
docker-compose -f docker-compose.go.yml exec api sqlite3 /data/hamqrzdb.sqlite "SELECT COUNT(*) FROM callsigns;"
```

### SSL Certificate Issues

Your existing Let's Encrypt certificates should work. If you need to renew:

```bash
# Check certificate expiry
docker-compose -f docker-compose.go.yml exec nginx openssl x509 -in /etc/letsencrypt/live/lookup.kj5djc.com/fullchain.pem -noout -dates

# Renew certificate
docker-compose -f docker-compose.go.yml run --rm certbot renew

# Restart nginx
docker-compose -f docker-compose.go.yml restart nginx
```

## Performance Comparison

### Before (nginx + static files):
- Disk usage: ~2GB (1.5M JSON files)
- Lookup time: ~5-10ms (file read)
- Updates: Regenerate ALL 1.5M files (20+ minutes)
- Case-sensitive: Only exact match works

### After (Go API + SQLite):
- Disk usage: ~500MB (single database)
- Lookup time: ~2-5ms (indexed query)
- Updates: Update database only (5-10 minutes)
- Case-insensitive: All variations work!

## Monitoring

```bash
# Check API health
curl https://lookup.kj5djc.com/health

# View API logs
docker-compose -f docker-compose.go.yml logs -f api

# View nginx logs
docker-compose -f docker-compose.go.yml logs -f nginx

# Check API stats
docker stats hamqrzdb-api
```

## Rollback Plan

If you need to rollback to the old nginx setup:

```bash
# Stop Go API
docker-compose -f docker-compose.go.yml down

# Start old nginx service
docker-compose -f docker-compose.nginx-backup.yml up -d
```

## Optional: Clean Up Old JSON Files

Once you verify the Go API works, you can delete the old JSON files to save space:

```bash
# Backup first!
tar -czf output-backup.tar.gz output/

# Delete JSON files (keep index.html)
find output -name "*.json" ! -name "404.json" -delete
find output -type d -empty -delete

# This frees up ~1.5GB of disk space!
```

## Support

- GitHub: https://github.com/chriskacerguis/hamqrzdb
- QRZ: https://www.qrz.com/db/KJ5DJC

73! 📻
