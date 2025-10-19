# Ham Radio Callsign Lookup System

A high-performance system for creating and serving static JSON files for amateur radio callsign lookups using FCC ULS data.

## Architecture

1. **Data Processing Script** (`process_uls.py`) - Downloads and processes FCC ULS data into static JSON files
2. **AWS S3** - Stores the static JSON files
3. **CloudFront** - CDN for fast global access
4. **CloudFront Function** - Routes requests and handles the API format

## Setup

### Prerequisites

- Python 3.7+
- AWS CLI configured with appropriate credentials
- AWS S3 bucket created
- CloudFront distribution set up

### Installation

```bash
# Make the script executable
chmod +x process_uls.py
```

## Usage

### Download and Process Full Database

```bash
./process_uls.py --full
```

This will:
1. Download the complete ULS amateur radio database (~500MB)
2. Extract and process HD.dat, EN.dat, and AM.dat files
3. Generate JSON files in the `output/` directory with structure: `output/K/J/5/KJ5DJC.json`

### Download and Process Daily Updates

```bash
./process_uls.py --daily
```

This downloads only today's incremental changes and processes them.

### Process a Specific Callsign

```bash
./process_uls.py --full --callsign KJ5DJC
```

Downloads the full database but only generates a JSON file for the specified callsign.

### Process a Local File

```bash
./process_uls.py --file /path/to/l_amat.zip
```

Processes a ZIP file you've already downloaded.

### Custom Output Directory

```bash
./process_uls.py --full --output /path/to/output
```

## Upload to S3

After processing, sync the files to S3:

```bash
aws s3 sync output/ s3://your-bucket-name/ --delete
```

For better performance, you can:

```bash
# Use parallel uploads
aws s3 sync output/ s3://your-bucket-name/ --delete --only-show-errors

# Or with s3-dist-cp for massive datasets
```

## File Structure

The script creates a nested directory structure to avoid too many files in one directory:

```
output/
├── K/
│   └── J/
│       └── 5/
│           └── KJ5DJC.json
├── W/
│   └── 1/
│       └── A/
│           └── W1AW.json
...
```

## JSON Output Format

Each JSON file follows this structure:

```json
{
  "hamdb": {
    "version": "1",
    "callsign": {
      "call": "KJ5DJC",
      "class": "G",
      "expires": "11/18/2033",
      "status": "A",
      "grid": "EM10ci",
      "lat": "30.3416503",
      "lon": "-97.7548379",
      "fname": "CHRIS",
      "mi": "",
      "name": "KACERGUIS",
      "suffix": "",
      "addr1": "5900 Balcones Drive STE 26811",
      "addr2": "AUSTIN",
      "state": "TX",
      "zip": "78731",
      "country": "United States"
    },
    "messages": {
      "status": "OK"
    }
  }
}
```

## CloudFront Setup

### 1. Create S3 Bucket

```bash
aws s3 mb s3://hamqrz-callsigns
aws s3api put-bucket-policy --bucket hamqrz-callsigns --policy file://bucket-policy.json
```

### 2. Create CloudFront Distribution

- Origin: Your S3 bucket
- Viewer Protocol Policy: Redirect HTTP to HTTPS
- Allowed HTTP Methods: GET, HEAD, OPTIONS
- Cache Policy: CachingOptimized or custom

### 3. Add CloudFront Function

See `cloudfront-function.js` for the routing function that transforms:
- `/v1/KJ5DJC/json/myapp` → `/K/J/5/KJ5DJC.json`

## API URL Format

```
https://your-distribution.cloudfront.net/v1/:callsign/json/:appname
```

Example:
```
https://api.hamqrz.com/v1/KJ5DJC/json/myapp

https://lookup.kj5djc.com/v1/KJ5DJC/json/hamdb


```

## Automation

### Daily Updates Cron Job

Add to crontab:

```bash
# Run daily at 2 AM
0 2 * * * cd /path/to/hamqrz && ./process_uls.py --daily && aws s3 sync output/ s3://lookup.kj5djc.com/ --delete
```

### Weekly Full Rebuild

```bash
# Run weekly on Sunday at 3 AM
0 3 * * 0 cd /path/to/hamqrz && ./process_uls.py --full && aws s3 sync output/ s3://lookup.kj5djc.com/ --delete
```

## Performance Notes

- **Processing Speed**: The script processes ~1M records in a few minutes
- **Disk Space**: Full database generates ~1-2GB of JSON files
- **Memory**: Script uses minimal memory with streaming CSV parsing
- **S3 Sync**: Initial upload may take time; incremental syncs are fast

## Important Notes

### Missing Coordinates

The current script doesn't include latitude/longitude data because that information is in additional files (LA.dat - Location/Antenna). To add coordinate support:

1. Modify the script to also parse LA.dat
2. Match location records to callsigns
3. Calculate grid squares from coordinates

### License Status Codes

- `A` = Active
- `C` = Canceled  
- `E` = Expired
- `T` = Terminated

### Operator Classes

- `N` = Novice
- `T` = Technician
- `G` = General
- `A` = Amateur Extra
- `P` = Technician Plus

## Troubleshooting

### Download Fails

Check the FCC website for changes to URLs or file formats:
- https://www.fcc.gov/uls/transactions/daily-weekly

### Daily File Not Available

Daily files may not be available on weekends or holidays. Use `--full` or provide a specific `--file`.

### Out of Memory

If processing fails due to memory, consider processing in batches or increasing system memory.

## License

MIT License - Feel free to use and modify for your needs.

## Credits

Data source: FCC Universal Licensing System (ULS)
- https://www.fcc.gov/uls/
