# UK Amateur Radio Data Importer

This tool imports amateur radio license data from the UK Office of Communications (Ofcom) into the HamQRZDB database.

## Data Source

- **Provider**: Ofcom (UK Office of Communications)
- **URL**: https://www.ofcom.org.uk/manage-your-licence/radiocommunication-licences/amateur-radio/amateur-radio-licence-data
- **Data File**: https://www.ofcom.org.uk/siteassets/resources/documents/manage-your-licence/amateur/callsign-030625.csv?v=398262
- **Format**: CSV (Comma-separated values)
- **License**: Open Government Licence

## CSV Format

The Ofcom CSV file contains the following columns:

1. **Licence Number** - Unique license identifier
2. **Call sign** - Amateur radio callsign (e.g., G0ABC, M0XYZ)
3. **First name** - Licensee's first name
4. **Surname** - Licensee's surname
5. **Full address** - Complete postal address
6. **Postcode** - UK postcode
7. **Licence status** - Current status (e.g., "Current", "Revoked", "Expired")
8. **Licence valid from** - Start date
9. **Licence valid to** - Expiry date

## Usage

### Build

```bash
# Build the importer
task build:import-uk

# Or build all tools
task build
```

### Run

**Note**: Ofcom's website uses Cloudflare protection which may block automated downloads with a 403 error. If automatic download fails, use the manual method below.

#### Automatic Download (may be blocked by Cloudflare)

```bash
# Download and import latest UK data
./bin/hamqrzdb-import-uk --db hamqrzdb.sqlite

# Using task
task db:import-uk
```

#### Manual Download (recommended if automatic fails)

```bash
# 1. Download the CSV file in your browser from:
#    https://www.ofcom.org.uk/siteassets/resources/documents/manage-your-licence/amateur/callsign-030625.csv?v=398262

# 2. Import the downloaded file
./bin/hamqrzdb-import-uk --db hamqrzdb.sqlite --file /path/to/callsign-030625.csv --download=false

# With Docker Compose
docker compose exec api /app/hamqrzdb-import-uk --db /data/hamqrzdb.sqlite --file /data/callsign-030625.csv --download=false
```

### Command-line Flags

- `--db <path>` - Path to SQLite database (default: `hamqrzdb.sqlite`)
- `--download` - Download fresh data from Ofcom (default: `true`)
- `--file <path>` - Use local CSV file instead of downloading

## Database Integration

UK callsigns are stored in the same `callsigns` table as US data, with the following fields populated:

- `callsign` - UK callsign
- `first_name` - Licensee's first name
- `last_name` - Licensee's surname (mapped from "Surname" column)
- `street_address` - Full postal address
- `zip_code` - UK postcode (mapped to zip_code field)
- `license_status` - Mapped to FCC-like codes:
  - `A` - Active/Current
  - `R` - Revoked
  - `E` - Expired
- `grant_date` - License valid from date
- `expired_date` - License valid to date
- `radio_service_code` - Set to "UK" to distinguish from US licenses

## API Access

UK callsigns can be queried through the same API endpoint:

```bash
# Query UK callsign
curl http://localhost:8080/v1/G0ABC/json/test

# Example response
{
  "hamdb": {
    "version": "1",
    "callsign": {
      "call": "G0ABC",
      "class": "",
      "expires": "2026-01-15",
      "status": "A",
      "grid": "",
      "lat": "",
      "lon": "",
      "fname": "John",
      "mi": "",
      "name": "Smith",
      "suffix": "",
      "addr1": "123 Main Street, London",
      "addr2": "",
      "state": "",
      "zip": "SW1A 1AA",
      "country": "United Kingdom"
    },
    "messages": {
      "status": "OK"
    }
  }
}
```

## UK Callsign Prefixes

Common UK amateur radio callsign prefixes include:

- **G** - UK Foundation, Intermediate, or Full licensees
- **M** - UK Foundation, Intermediate, or Full licensees (alternative prefix)
- **2E** - England
- **GW**, **MW**, **2W** - Wales
- **GM**, **MM**, **2M** - Scotland
- **GI**, **MI**, **2I** - Northern Ireland
- **GD**, **MD**, **2D** - Isle of Man
- **GJ**, **MJ**, **2J** - Jersey
- **GU**, **MU**, **2U** - Guernsey

## Notes

- UK data does not include grid square or latitude/longitude coordinates by default
- The importer uses UPSERT logic, so running it multiple times will update existing records
- UK licenses are marked with `radio_service_code = "UK"` to distinguish them from US licenses
- The API response sets `country` to "United Kingdom" for UK callsigns

## Updates

Ofcom updates the amateur radio license data regularly. To refresh your database:

```bash
# Re-run the importer to download latest data
task db:import-uk
```

## 73! 📻🇬🇧
