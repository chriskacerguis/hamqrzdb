#!/usr/bin/env python3
"""
DEPRECATED: This file is no longer maintained.

Please use process_uls_db.py instead, which provides:
- Memory-efficient SQLite database storage
- Incremental updates with upsert support
- Better handling of large datasets
- Separate location processing with process_uls_locations.py

This file is kept for reference only.
"""

import argparse
import csv
import json
import os
import sys
import zipfile
from datetime import datetime
from pathlib import Path
from typing import Dict, Optional
import urllib.request
import shutil
from collections import defaultdict


class EnhancedULSProcessor:
    """Process ULS amateur radio data files with location support"""
    
    def __init__(self, output_dir: str = "output"):
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(exist_ok=True)
        
        # Data storage
        self.hd_data = {}  # callsign -> HD record
        self.en_data = {}  # callsign -> EN record
        self.am_data = {}  # callsign -> AM record
        self.la_data = {}  # callsign -> LA record (location)
        
    def download_file(self, url: str, destination: str) -> bool:
        """Download a file from URL to destination"""
        try:
            print(f"Downloading {url}...")
            with urllib.request.urlopen(url) as response:
                with open(destination, 'wb') as out_file:
                    shutil.copyfileobj(response, out_file)
            print(f"Downloaded to {destination}")
            return True
        except Exception as e:
            print(f"Error downloading file: {e}")
            return False
    
    def extract_zip(self, zip_path: str, extract_to: str) -> bool:
        """Extract ZIP file to directory"""
        try:
            print(f"Extracting {zip_path}...")
            with zipfile.ZipFile(zip_path, 'r') as zip_ref:
                zip_ref.extractall(extract_to)
            print(f"Extracted to {extract_to}")
            return True
        except Exception as e:
            print(f"Error extracting ZIP: {e}")
            return False
    
    def parse_hd_file(self, file_path: str, filter_callsign: Optional[str] = None):
        """Parse HD.dat (Header) file"""
        print(f"Parsing HD.dat...")
        count = 0
        with open(file_path, 'r', encoding='latin-1') as f:
            reader = csv.reader(f, delimiter='|')
            for row in reader:
                if len(row) < 5:
                    continue
                
                record_type = row[0]
                if record_type != 'HD':
                    continue
                
                callsign = row[4].strip()
                
                # Filter if specific callsign requested
                if filter_callsign and callsign != filter_callsign.upper():
                    continue
                
                self.hd_data[callsign] = {
                    'unique_system_identifier': row[1].strip() if len(row) > 1 else '',
                    'uls_file_number': row[2].strip() if len(row) > 2 else '',
                    'call_sign': callsign,
                    'license_status': row[5].strip() if len(row) > 5 else '',
                    'radio_service_code': row[6].strip() if len(row) > 6 else '',
                    'grant_date': row[7].strip() if len(row) > 7 else '',
                    'expired_date': row[8].strip() if len(row) > 8 else '',
                    'cancellation_date': row[9].strip() if len(row) > 9 else '',
                }
                count += 1
                
                if count % 10000 == 0:
                    print(f"  Processed {count} HD records...")
        
        print(f"Parsed {count} HD records")
    
    def parse_en_file(self, file_path: str, filter_callsign: Optional[str] = None):
        """Parse EN.dat (Entity) file"""
        print(f"Parsing EN.dat...")
        count = 0
        with open(file_path, 'r', encoding='latin-1') as f:
            reader = csv.reader(f, delimiter='|')
            for row in reader:
                if len(row) < 5:
                    continue
                
                record_type = row[0]
                if record_type != 'EN':
                    continue
                
                callsign = row[4].strip()
                
                # Filter if specific callsign requested
                if filter_callsign and callsign != filter_callsign.upper():
                    continue
                
                self.en_data[callsign] = {
                    'entity_name': row[7].strip() if len(row) > 7 else '',
                    'first_name': row[8].strip() if len(row) > 8 else '',
                    'mi': row[9].strip() if len(row) > 9 else '',
                    'last_name': row[10].strip() if len(row) > 10 else '',
                    'suffix': row[11].strip() if len(row) > 11 else '',
                    'street_address': row[15].strip() if len(row) > 15 else '',
                    'city': row[16].strip() if len(row) > 16 else '',
                    'state': row[17].strip() if len(row) > 17 else '',
                    'zip_code': row[18].strip() if len(row) > 18 else '',
                }
                count += 1
                
                if count % 10000 == 0:
                    print(f"  Processed {count} EN records...")
        
        print(f"Parsed {count} EN records")
    
    def parse_am_file(self, file_path: str, filter_callsign: Optional[str] = None):
        """Parse AM.dat (Amateur) file"""
        print(f"Parsing AM.dat...")
        count = 0
        with open(file_path, 'r', encoding='latin-1') as f:
            reader = csv.reader(f, delimiter='|')
            for row in reader:
                if len(row) < 5:
                    continue
                
                record_type = row[0]
                if record_type != 'AM':
                    continue
                
                callsign = row[4].strip()
                
                # Filter if specific callsign requested
                if filter_callsign and callsign != filter_callsign.upper():
                    continue
                
                self.am_data[callsign] = {
                    'operator_class': row[5].strip() if len(row) > 5 else '',
                    'group_code': row[6].strip() if len(row) > 6 else '',
                    'region_code': row[7].strip() if len(row) > 7 else '',
                }
                count += 1
                
                if count % 10000 == 0:
                    print(f"  Processed {count} AM records...")
        
        print(f"Parsed {count} AM records")
    
    def parse_la_file(self, file_path: str, filter_callsign: Optional[str] = None):
        """Parse LA.dat (Location/Antenna) file for coordinates"""
        print(f"Parsing LA.dat...")
        count = 0
        
        if not Path(file_path).exists():
            print(f"Warning: LA.dat not found, coordinates will not be available")
            return
        
        with open(file_path, 'r', encoding='latin-1') as f:
            reader = csv.reader(f, delimiter='|')
            for row in reader:
                if len(row) < 5:
                    continue
                
                record_type = row[0]
                if record_type != 'LA':
                    continue
                
                callsign = row[4].strip()
                
                # Filter if specific callsign requested
                if filter_callsign and callsign != filter_callsign.upper():
                    continue
                
                # LA fields include location information
                # Get latitude and longitude (fields vary, typically around index 13-15)
                try:
                    lat_deg = row[13].strip() if len(row) > 13 else ''
                    lat_min = row[14].strip() if len(row) > 14 else ''
                    lat_sec = row[15].strip() if len(row) > 15 else ''
                    lat_dir = row[16].strip() if len(row) > 16 else ''
                    
                    lon_deg = row[17].strip() if len(row) > 17 else ''
                    lon_min = row[18].strip() if len(row) > 18 else ''
                    lon_sec = row[19].strip() if len(row) > 19 else ''
                    lon_dir = row[20].strip() if len(row) > 20 else ''
                    
                    # Convert to decimal degrees
                    lat = None
                    lon = None
                    
                    if lat_deg and lat_min and lat_sec:
                        lat = float(lat_deg) + float(lat_min)/60 + float(lat_sec)/3600
                        if lat_dir == 'S':
                            lat = -lat
                    
                    if lon_deg and lon_min and lon_sec:
                        lon = float(lon_deg) + float(lon_min)/60 + float(lon_sec)/3600
                        if lon_dir == 'W':
                            lon = -lon
                    
                    if lat is not None and lon is not None:
                        self.la_data[callsign] = {
                            'latitude': lat,
                            'longitude': lon
                        }
                        count += 1
                except (ValueError, IndexError):
                    # Skip records with invalid coordinates
                    pass
                
                if count % 10000 == 0:
                    print(f"  Processed {count} LA records with coordinates...")
        
        print(f"Parsed {count} LA records with valid coordinates")
    
    def calculate_grid_square(self, lat: float, lon: float) -> str:
        """Calculate Maidenhead grid square from latitude and longitude"""
        # Adjust longitude to 0-360 range
        adj_lon = lon + 180
        adj_lat = lat + 90
        
        # Calculate field (first two characters)
        field_lon = chr(int(adj_lon / 20) + ord('A'))
        field_lat = chr(int(adj_lat / 10) + ord('A'))
        
        # Calculate square (next two digits)
        square_lon = str(int((adj_lon % 20) / 2))
        square_lat = str(int((adj_lat % 10)))
        
        # Calculate subsquare (last two characters, lowercase)
        subsquare_lon = chr(int((adj_lon % 2) * 12) + ord('a'))
        subsquare_lat = chr(int((adj_lat % 1) * 24) + ord('a'))
        
        return f"{field_lon}{field_lat}{square_lon}{square_lat}{subsquare_lon}{subsquare_lat}"
    
    def format_date(self, date_str: str) -> str:
        """Format date from MMDDYYYY to MM/DD/YYYY"""
        if not date_str or len(date_str) != 8:
            return ""
        try:
            month = date_str[0:2]
            day = date_str[2:4]
            year = date_str[4:8]
            return f"{month}/{day}/{year}"
        except:
            return date_str
    
    def get_license_status(self, status_code: str) -> str:
        """Convert license status code to letter"""
        status_map = {
            'A': 'A',  # Active
            'C': 'C',  # Canceled
            'E': 'E',  # Expired
            'T': 'T',  # Terminated
        }
        return status_map.get(status_code, status_code)
    
    def generate_json_record(self, callsign: str) -> Optional[Dict]:
        """Generate JSON record for a callsign"""
        # Get data from all files
        hd = self.hd_data.get(callsign)
        en = self.en_data.get(callsign)
        am = self.am_data.get(callsign)
        la = self.la_data.get(callsign)
        
        # Need at least HD record
        if not hd:
            return None
        
        # Get coordinates from LA data if available
        lat = ""
        lon = ""
        grid = ""
        
        if la:
            lat = str(la['latitude'])
            lon = str(la['longitude'])
            grid = self.calculate_grid_square(la['latitude'], la['longitude'])
        
        # Build the JSON structure
        record = {
            "hamdb": {
                "version": "1",
                "callsign": {
                    "call": callsign,
                    "class": am['operator_class'] if am else "",
                    "expires": self.format_date(hd['expired_date']),
                    "status": self.get_license_status(hd['license_status']),
                    "grid": grid,
                    "lat": lat,
                    "lon": lon,
                    "fname": en['first_name'] if en else "",
                    "mi": en['mi'] if en else "",
                    "name": en['last_name'] if en else "",
                    "suffix": en['suffix'] if en else "",
                    "addr1": en['street_address'] if en else "",
                    "addr2": en['city'] if en else "",
                    "state": en['state'] if en else "",
                    "zip": en['zip_code'] if en else "",
                    "country": "United States"
                },
                "messages": {
                    "status": "OK"
                }
            }
        }
        
        return record
    
    def write_json_file(self, callsign: str, record: Dict):
        """Write JSON file for a callsign"""
        # Create directory structure
        if len(callsign) >= 3:
            dir_path = self.output_dir / callsign[0] / callsign[1] / callsign[2]
        elif len(callsign) >= 2:
            dir_path = self.output_dir / callsign[0] / callsign[1]
        else:
            dir_path = self.output_dir / callsign[0]
        
        dir_path.mkdir(parents=True, exist_ok=True)
        
        file_path = dir_path / f"{callsign}.json"
        with open(file_path, 'w') as f:
            json.dump(record, f, indent=2)
    
    def process_all_callsigns(self):
        """Generate JSON files for all callsigns"""
        print(f"\nGenerating JSON files...")
        
        all_callsigns = set(self.hd_data.keys())
        total = len(all_callsigns)
        count = 0
        
        for callsign in all_callsigns:
            record = self.generate_json_record(callsign)
            if record:
                self.write_json_file(callsign, record)
                count += 1
                
                if count % 1000 == 0:
                    print(f"  Generated {count}/{total} JSON files...")
        
        print(f"\nCompleted! Generated {count} JSON files in {self.output_dir}")
    
    def process_single_callsign(self, callsign: str):
        """Generate JSON file for a single callsign"""
        print(f"\nGenerating JSON for {callsign}...")
        
        callsign = callsign.upper()
        record = self.generate_json_record(callsign)
        
        if record:
            self.write_json_file(callsign, record)
            print(f"Generated {callsign}.json")
        else:
            print(f"No data found for {callsign}")


def main():
    parser = argparse.ArgumentParser(
        description='Enhanced FCC ULS processor with location support'
    )
    
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument('--daily', action='store_true',
                      help='Download and process daily update file')
    group.add_argument('--full', action='store_true',
                      help='Download and process full database')
    group.add_argument('--file', type=str,
                      help='Process a local ZIP file instead of downloading')
    
    parser.add_argument('--callsign', type=str,
                       help='Process only a specific callsign')
    parser.add_argument('--output', type=str, default='output',
                       help='Output directory for JSON files (default: output)')
    
    args = parser.parse_args()
    
    processor = EnhancedULSProcessor(output_dir=args.output)
    
    temp_dir = Path('temp_uls')
    temp_dir.mkdir(exist_ok=True)
    
    if args.file:
        zip_file = args.file
        print(f"Using provided file: {zip_file}")
    elif args.daily:
        today = datetime.now().strftime('%m%d%Y')
        url = f"https://data.fcc.gov/download/pub/uls/daily/l_am_{today}.zip"
        zip_file = temp_dir / f"l_am_{today}.zip"
        
        if not processor.download_file(url, str(zip_file)):
            print("Failed to download daily file")
            return 1
    elif args.full:
        url = "https://data.fcc.gov/download/pub/uls/complete/l_amat.zip"
        zip_file = temp_dir / "l_amat.zip"
        
        if not processor.download_file(url, str(zip_file)):
            print("Failed to download full file")
            return 1
    
    extract_dir = temp_dir / "extracted"
    extract_dir.mkdir(exist_ok=True)
    
    if not processor.extract_zip(str(zip_file), str(extract_dir)):
        print("Failed to extract ZIP file")
        return 1
    
    # Find and parse data files
    hd_file = extract_dir / "HD.dat"
    en_file = extract_dir / "EN.dat"
    am_file = extract_dir / "AM.dat"
    la_file = extract_dir / "LA.dat"
    
    if not hd_file.exists():
        print(f"HD.dat not found")
        return 1
    if not en_file.exists():
        print(f"EN.dat not found")
        return 1
    if not am_file.exists():
        print(f"AM.dat not found")
        return 1
    
    filter_call = args.callsign.upper() if args.callsign else None
    
    processor.parse_hd_file(str(hd_file), filter_call)
    processor.parse_en_file(str(en_file), filter_call)
    processor.parse_am_file(str(am_file), filter_call)
    processor.parse_la_file(str(la_file), filter_call)  # LA.dat may not exist
    
    if args.callsign:
        processor.process_single_callsign(args.callsign)
    else:
        processor.process_all_callsigns()
    
    print("\nCleaning up temporary files...")
    shutil.rmtree(temp_dir)
    
    print("\nDone!")
    return 0


if __name__ == "__main__":
    sys.exit(main())
