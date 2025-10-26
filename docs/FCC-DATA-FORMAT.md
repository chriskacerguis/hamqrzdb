# FCC ULS Amateur Radio Data Format

This document explains the FCC Universal Licensing System (ULS) data files used by HamQRZDB, including file descriptions and field definitions.

## Data Sources

- **Full Database**: `https://data.fcc.gov/download/pub/uls/complete/l_amat.zip`
- **Daily Updates**: `https://data.fcc.gov/download/pub/uls/daily/l_am_MMDDYYYY.zip`

The amateur radio database is updated daily by the FCC and contains all amateur radio licenses in the United States.

## File Format

All files are pipe-delimited (`|`) text files with the `.dat` extension. Fields are positional, and the first field indicates the record type.

## Primary Data Files

### HD.dat - Header Data (License Information)

Contains the main license record for each callsign. This is the primary file and must be processed first.

**Record Type**: `HD`

**Key Fields** (0-indexed):
- Field 0: Record Type (`HD`)
- Field 1: Unique System Identifier
- Field 2: ULS File Number
- Field 3: EBF Number (Electronic Batch Filing)
- **Field 4: CALLSIGN** - The amateur radio callsign
- **Field 5: License Status** - Current status of the license
  - `A` = Active
  - `E` = Expired
  - `C` = Cancelled
  - `T` = Terminated
- Field 6: Radio Service Code (`HA` = Amateur)
- **Field 7: Grant Date** - Date license was granted (MM/DD/YYYY)
- **Field 8: Expired Date** - Date license expires (MM/DD/YYYY)
- **Field 9: Cancellation Date** - Date license was cancelled (MM/DD/YYYY)
- Field 10-19: Various administrative fields
- Field 20-29: Licensee name fields (first, middle, last, suffix)

**Example**:
```
HD|4186771|0011303942||KN6DQD|A|HA|08/06/2019|08/06/2029||||||||||N||||||||||N||Zoe||Downing||||||||||10/30/2024|10/30/2024|||||||||||||||
```

### EN.dat - Entity Data (Name and Address)

Contains the licensee's personal information including name and mailing address.

**Record Type**: `EN`

**Key Fields** (0-indexed):
- Field 0: Record Type (`EN`)
- Field 1: Unique System Identifier
- Field 2-3: Administrative fields
- **Field 4: CALLSIGN** - Links to HD.dat
- Field 5: Entity Type (`L` = Licensee)
- Field 6: Licensee ID
- **Field 7: Entity Name** - Full name or organization name
- **Field 8: First Name**
- Field 9: Middle Initial
- **Field 10: Last Name**
- **Field 11: Suffix** (Jr., Sr., III, etc.)
- Field 12-15: Reserved/Additional name fields
- **Field 16: Street Address**
- **Field 17: City**
- **Field 18: State** (2-letter abbreviation)
- **Field 19: ZIP Code**
- Field 20: ZIP+4 Extension
- Field 21-23: Phone and additional contact info

**Example**:
```
EN|4186771|||KN6DQD|L|L02283715|Downing, Zoe|Zoe||Downing||||||Montara|CA|94037|370545||000|0028710390|I||||||
```

**Important Notes**:
- Field 16 (Street Address) is often empty for privacy reasons
- Field 7 typically contains "Last, First" format
- This file must be processed AFTER HD.dat as it updates existing records

### AM.dat - Amateur Radio Data (Operator Class)

Contains amateur radio-specific information, primarily the operator class.

**Record Type**: `AM`

**Key Fields** (0-indexed):
- Field 0: Record Type (`AM`)
- Field 1: Unique System Identifier
- Field 2-3: Administrative fields
- **Field 4: CALLSIGN** - Links to HD.dat
- **Field 5: Operator Class** - License class
  - `N` = Novice (discontinued)
  - `T` = Technician
  - `G` = General
  - `A` = Advanced (discontinued)
  - `E` = Amateur Extra
- **Field 6: Group Code** - License group
  - `A` = Amateur
  - `B` = Novice/Technician
  - `C` = Technician Plus (discontinued)
  - `D` = General/Advanced/Amateur Extra
- **Field 7: Region Code** - FCC region (0-9, A-Z)
- Field 8-15: Additional administrative fields
- Field 16: Previous Callsign
- Field 17: Previous Operator Class

**Example**:
```
AM|4186771|||KN6DQD|G|D|6|||||||||T|
```

**Operator Class Hierarchy** (lowest to highest):
1. Technician (T)
2. General (G)
3. Amateur Extra (E)

### LA.dat - Location Data (Coordinates)

Contains geographic coordinates (latitude/longitude) for each license.

**Record Type**: `LA`

**Key Fields** (0-indexed):
- Field 0: Record Type (`LA`)
- Field 1: Unique System Identifier
- Field 2-3: Administrative fields
- **Field 4: CALLSIGN** - Links to HD.dat
- Field 5-12: Location type and administrative fields
- **Field 13: Latitude Degrees**
- **Field 14: Latitude Minutes**
- **Field 15: Latitude Seconds**
- **Field 16: Latitude Direction** (`N` or `S`)
- **Field 17: Longitude Degrees**
- **Field 18: Longitude Minutes**
- **Field 19: Longitude Seconds**
- **Field 20: Longitude Direction** (`E` or `W`)

**Example**:
```
LA|4186771|||KN6DQD||||||||37|32|11.9|N|122|31|17.4|W||||
```

**Coordinate Conversion**:
Coordinates are in DMS (Degrees, Minutes, Seconds) format and must be converted to decimal:
```
Decimal = Degrees + (Minutes/60) + (Seconds/3600)
If Direction is S or W, multiply by -1
```

**Grid Square Calculation**:
From the decimal coordinates, the Maidenhead grid square is calculated (e.g., `CM87wj`).

## Processing Order

Files must be processed in this specific order:

1. **HD.dat** - Creates initial callsign records (INSERT)
2. **EN.dat** - Adds name and address data (UPDATE)
3. **AM.dat** - Adds operator class (UPDATE)
4. **LA.dat** - Adds coordinates and grid square (UPDATE)

## Additional Files (Not Used by HamQRZDB)

The FCC ULS database contains other files that are not currently used:

- **HS.dat** - History Data (license history/changes)
- **CO.dat** - Comments
- **SC.dat** - Special Conditions
- **SF.dat** - Ship/Aircraft data
- **AD.dat** - Application Data

## Database Schema Mapping

### HamQRZDB SQLite Schema

```sql
CREATE TABLE callsigns (
    callsign TEXT PRIMARY KEY,              -- HD.dat field 4
    license_status TEXT,                     -- HD.dat field 5
    radio_service_code TEXT,                 -- HD.dat field 6
    grant_date TEXT,                         -- HD.dat field 7
    expired_date TEXT,                       -- HD.dat field 8
    cancellation_date TEXT,                  -- HD.dat field 9
    operator_class TEXT,                     -- AM.dat field 5
    group_code TEXT,                         -- AM.dat field 6
    region_code TEXT,                        -- AM.dat field 7
    first_name TEXT,                         -- EN.dat field 8
    mi TEXT,                                 -- EN.dat field 9
    last_name TEXT,                          -- EN.dat field 10
    suffix TEXT,                             -- EN.dat field 11
    entity_name TEXT,                        -- EN.dat field 7
    street_address TEXT,                     -- EN.dat field 16
    city TEXT,                               -- EN.dat field 17
    state TEXT,                              -- EN.dat field 18
    zip_code TEXT,                           -- EN.dat field 19
    latitude REAL,                           -- LA.dat fields 13-16 (converted)
    longitude REAL,                          -- LA.dat fields 17-20 (converted)
    grid_square TEXT,                        -- Calculated from lat/lon
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## API Response Mapping

The HamDB-compatible JSON API maps database fields to JSON response:

```json
{
  "hamdb": {
    "version": "1",
    "callsign": {
      "call": "callsign",
      "class": "operator_class",
      "expires": "expired_date",
      "status": "license_status",
      "grid": "grid_square",
      "lat": "latitude",
      "lon": "longitude",
      "fname": "first_name",
      "mi": "mi",
      "name": "last_name",
      "suffix": "suffix",
      "addr1": "street_address",
      "addr2": "city",
      "state": "state",
      "zip": "zip_code",
      "country": "United States"
    },
    "messages": {
      "status": "OK"
    }
  }
}
```

## Data Privacy Notes

1. **Street addresses** (EN.dat field 16) are often blank in the FCC database for privacy reasons
2. The FCC provides **mailing addresses** which may be PO boxes
3. **Coordinates** (LA.dat) represent the location associated with the license, not necessarily the licensee's home
4. All data is **public information** available from the FCC

## Performance Considerations

### File Sizes (Approximate)
- HD.dat: ~150 MB, ~1.4M records
- EN.dat: ~200 MB, ~1.4M records
- AM.dat: ~75 MB, ~1.4M records
- LA.dat: ~150 MB, ~1.4M records
- **Total**: ~575 MB uncompressed, ~180 MB compressed

### Processing Time
- Full database: 5-10 minutes
- Daily updates: 30-60 seconds
- EN.dat is the slowest (largest file with most fields)

### Memory Usage
- Processing uses **streaming/batch processing** to minimize memory
- Batch size: 1,000 records per transaction
- Peak memory usage: ~100 MB

## References

- [FCC ULS Database Downloads](https://www.fcc.gov/uls/transactions/daily-weekly)
- [FCC ULS Database Documentation](https://www.fcc.gov/wireless/systems-utilities/universal-licensing-system)
- [Amateur Radio Service](https://www.fcc.gov/wireless/bureau-divisions/mobility-division/amateur-radio-service)

## Updates

The FCC updates the database **daily** with new licenses, modifications, and expirations. Daily update files follow the naming pattern `l_am_MMDDYYYY.zip`.

To stay current:
- Run `--daily` updates daily via cron
- Run `--full` rebuild weekly or monthly to ensure data integrity
