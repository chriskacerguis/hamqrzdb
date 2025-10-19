# Deployment Guide

## Quick Fix (Build Locally on Server)

SSH into your server and run these commands:

```bash
cd /opt/hamqrzdb

# Pull latest code
git pull origin main

# Build Docker image locally
docker build -t hamqrzdb-api:latest .

# Stop the current container
docker-compose down

# Edit docker-compose.yml to use local image
sed -i 's|ghcr.io/chriskacerguis/hamqrzdb:latest|hamqrzdb-api:latest|' docker-compose.yml

# Start the new container
docker-compose up -d

# Check logs
docker-compose logs -f api

# Test the API
curl http://localhost:8080/v1/kj5djc/json/test
```

## Production Deployment (GitHub Actions)

Once you push your code to GitHub, the workflow will automatically build and push to GitHub Container Registry.

### Initial Setup (One Time)

1. Make sure your GitHub repository has package permissions enabled
2. Go to https://github.com/chriskacerguis/hamqrzdb/settings/actions
3. Under "Workflow permissions", ensure "Read and write permissions" is selected

### Deploy Process

```bash
# Commit and push your changes
git add .
git commit -m "Fix API database column mapping"
git push origin main

# Wait for GitHub Actions to build (2-3 minutes)
# Watch progress at: https://github.com/chriskacerguis/hamqrzdb/actions

# On your server, pull the new image
ssh root@157.230.133.121
cd /opt/hamqrzdb

# Make sure docker-compose.yml uses GHCR image
# image: ghcr.io/chriskacerguis/hamqrzdb:latest

# Pull and restart
docker-compose pull
docker-compose down
docker-compose up -d

# Check logs
docker-compose logs -f api

# Test
curl http://localhost:8080/v1/kj5djc/json/test
```

## Verify Deployment

After restarting the container, verify the API works:

```bash
# Test valid callsign (should return data)
curl http://localhost:8080/v1/kj5djc/json/test | jq .

# Test invalid callsign (should return NOT_FOUND)
curl http://localhost:8080/v1/INVALIDCALL/json/test | jq .

# Check health endpoint
curl http://localhost:8080/health

# View container logs
docker-compose logs --tail=50 api
```

## Troubleshooting

### Container won't start

```bash
# Check container logs
docker-compose logs api

# Check if database file is readable
ls -la hamqrzdb.sqlite

# Verify database has data
sqlite3 hamqrzdb.sqlite "SELECT COUNT(*) FROM callsigns"
```

### API returns NOT_FOUND for valid callsigns

```bash
# Check database columns match API expectations
sqlite3 hamqrzdb.sqlite ".schema callsigns"

# Verify data exists
sqlite3 hamqrzdb.sqlite "SELECT * FROM callsigns WHERE callsign = 'KJ5DJC'"

# Check container can access database
docker-compose exec api ls -la /data/
```

### Image pull fails from GHCR

```bash
# Login to GitHub Container Registry
docker login ghcr.io -u chriskacerguis

# Pull manually
docker pull ghcr.io/chriskacerguis/hamqrzdb:latest

# Or build locally instead (see Quick Fix above)
```

## Rollback

If the new version has issues:

```bash
# Use a specific tag instead of latest
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/hamqrzdb.sqlite:/data/hamqrzdb.sqlite:ro \
  --name hamqrzdb-api \
  ghcr.io/chriskacerguis/hamqrzdb:20231019-abc1234
```
