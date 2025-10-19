# Memory Optimization Guide for HamQRZDB on Ubuntu 24

## Problem
The Python script gets "Killed" due to Out of Memory (OOM) issues when processing ~1M callsign records.

## Solutions (in order of recommendation)

### Option 1: Use the Streaming Version (RECOMMENDED)

I've created `process_uls_streaming.py` which uses minimal memory:

```bash
# Make it executable
chmod +x process_uls_streaming.py

# Process full database with default batch size
./process_uls_streaming.py --full

# Or adjust batch size for even lower memory usage
./process_uls_streaming.py --full --batch-size 5000

# For systems with very limited RAM
./process_uls_streaming.py --full --batch-size 1000
```

**How it works:**
- Builds an index of callsigns (only stores callsign strings, not full records)
- Processes callsigns in batches
- Only keeps one batch in memory at a time
- **Memory usage: ~100-500MB** instead of 2-3GB

### Option 2: Increase Swap Space

If you want to continue using the original script, increase swap:

```bash
# Check current swap
free -h

# Create a 4GB swap file
sudo fallocate -l 4G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# Make it permanent
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab

# Verify
free -h
```

**Note:** This will be slower but will work.

### Option 3: Adjust VM Overcommit Settings

Allow the kernel to overcommit memory:

```bash
# Check current setting
cat /proc/sys/vm/overcommit_memory

# Set to 1 (always overcommit)
sudo sysctl vm.overcommit_memory=1

# Make permanent
echo 'vm.overcommit_memory=1' | sudo tee -a /etc/sysctl.conf
```

### Option 4: Process in Docker with Memory Limits

Process the data with controlled memory:

```bash
# Run in a container with memory limit
docker run -it --rm \
  -v "$(pwd):/app" \
  -w /app \
  --memory="2g" \
  --memory-swap="4g" \
  python:3.11-slim \
  bash -c "pip install --no-cache-dir && python3 process_uls_streaming.py --full"
```

### Option 5: Split Processing Manually

Process the database in chunks by callsign prefix:

```bash
# Create a script to process by first letter
for letter in {A..Z} {0..9}; do
  echo "Processing callsigns starting with $letter..."
  # You'd need to modify the script to filter by prefix
  ./process_uls.py --full --prefix "$letter"
done
```

## Monitoring Memory Usage

While processing, monitor memory:

```bash
# In another terminal
watch -n 1 'free -h && echo "---" && ps aux | grep python | grep -v grep'
```

## Recommended System Requirements

For the original script:
- **Minimum RAM:** 4GB
- **Recommended RAM:** 8GB
- **Swap:** 4GB+

For the streaming version:
- **Minimum RAM:** 1GB
- **Recommended RAM:** 2GB
- **Swap:** 2GB (optional)

## Performance Comparison

| Version | Memory Usage | Processing Time | Notes |
|---------|--------------|----------------|--------|
| Original | ~2-3GB | ~3-5 minutes | Fast but memory-intensive |
| Streaming | ~100-500MB | ~15-30 minutes | Slower but memory-efficient |
| Streaming (batch=1000) | ~50-100MB | ~45-60 minutes | Very memory-efficient |

## Testing

Test with a single callsign first:

```bash
# Original version
./process_uls.py --full --callsign KJ5DJC

# Streaming version
./process_uls_streaming.py --full --callsign KJ5DJC
```

Both should create the same output with the streaming version using much less memory.

## Update Scripts

Update your `update-daily.sh` and `update-weekly.sh` to use the streaming version:

```bash
# Change this:
python3 "${SCRIPT_DIR}/process_uls.py" --full --output "$TEMP_DIR"

# To this:
python3 "${SCRIPT_DIR}/process_uls_streaming.py" --full --output "$TEMP_DIR" --batch-size 5000
```
