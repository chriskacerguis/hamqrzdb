#!/usr/bin/env python3
"""
FCC ULS Amateur Radio Callsign Data Processor - Memory Optimized Version
Processes HD.dat, EN.dat, and AM.dat files using streaming to minimize memory usage

This version processes files in multiple passes to avoid loading everything into memory:
1. First pass: Build index of callsigns and their file positions
2. Second pass: Process and write JSON files one at a time
"""

import argparse
import csv
import json
import os
import sys
import zipfile
from datetime import datetime
from pathlib import Path
from typing import Dict, Optional, Tuple
import urllib.request
import shutil
import tempfile


class ULSProcessorOptimized:
    """Process ULS amateur radio data files with minimal memory usage"""
    
    def __init__(self, output_dir: str = "output"):
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(exist_ok=True)
        
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
    
    def build_callsign_index(self, hd_file: str) -> set:
        """Build a set of all valid callsigns from HD.dat"""
        print("Building callsign index from HD.dat...")
        callsigns = set()
        count = 0
        
        with open(hd_file, 'r', encoding='latin-1') as f:
            reader = csv.reader(f, delimiter='|')
            for row in reader:
                if len(row) < 5:
                    continue
                if row[0] != 'HD':
                    continue
                
                callsign = row[4].strip()
                if callsign:
                    callsigns.add(callsign)
                    count += 1
                    
                    if count % 50000 == 0:
                        print(f"  Indexed {count} callsigns...")
        
        print(f"Indexed {count} callsigns")
        return callsigns
    
    def get_record_for_callsign(self, file_path: str, callsign: str, record_type: str) -> Optional[Dict]:
        """Find and return a single record for a callsign"""
        with open(file_path, 'r', encoding='latin-1') as f:
            reader = csv.reader(f, delimiter='|')
            for row in reader:
                if len(row) < 5:
                    continue
                if row[0] != record_type:
                    continue
                if row[4].strip() == callsign:
                    return row
        return None
    
    def parse_hd_record(self, row: list) -> Dict:
        """Parse a single HD record"""
        return {
            'call_sign': row[4].strip() if len(row) > 4 else '',
            'license_status': row[5].strip() if len(row) > 5 else '',
            'radio_service_code': row[6].strip() if len(row) > 6 else '',
            'grant_date': row[7].strip() if len(row) > 7 else '',
            'expired_date': row[8].strip() if len(row) > 8 else '',
            'cancellation_date': row[9].strip() if len(row) > 9 else '',
        }
    
    def parse_en_record(self, row: list) -> Dict:
        """Parse a single EN record"""
        return {
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
    
    def parse_am_record(self, row: list) -> Dict:
        """Parse a single AM record"""
        return {
            'operator_class': row[5].strip() if len(row) > 5 else '',
            'group_code': row[6].strip() if len(row) > 6 else '',
            'region_code': row[7].strip() if len(row) > 7 else '',
        }
    
    def calculate_grid_square(self, lat: float, lon: float) -> str:
        """Calculate Maidenhead grid square from latitude and longitude"""
        lon = lon + 180
        lat = lat + 90
        
        field_lon = chr(int(lon / 20) + ord('A'))
        field_lat = chr(int(lat / 10) + ord('A'))
        square_lon = str(int((lon % 20) / 2))
        square_lat = str(int((lat % 10)))
        subsquare_lon = chr(int((lon % 2) * 12) + ord('a'))
        subsquare_lat = chr(int(((lat % 1) * 24)) + ord('a'))
        
        return f"{field_lon}{field_lat}{square_lon}{square_lat}{subsquare_lon}{subsquare_lat}"
    
    def format_expiration_date(self, date_str: str) -> str:
        """Format expiration date to MM/DD/YYYY"""
        if not date_str or len(date_str) != 10:
            return "NOT_FOUND"
        
        try:
            date_obj = datetime.strptime(date_str, '%m/%d/%Y')
            return date_obj.strftime('%m/%d/%Y')
        except ValueError:
            return "NOT_FOUND"
    
    def create_json_record(self, callsign: str, hd_data: Dict, en_data: Dict, am_data: Dict) -> Dict:
        """Create a JSON record from combined data"""
        status_map = {
            'A': 'A',  # Active
            'C': 'C',  # Canceled
            'E': 'E',  # Expired
            'T': 'T',  # Terminated
        }
        
        status = status_map.get(hd_data.get('license_status', ''), 'NOT_FOUND')
        operator_class = am_data.get('operator_class', 'NOT_FOUND')
        expires = self.format_expiration_date(hd_data.get('expired_date', ''))
        
        # Get name components
        first_name = en_data.get('first_name', '')
        mi = en_data.get('mi', '')
        last_name = en_data.get('last_name', '')
        suffix = en_data.get('suffix', '')
        
        # Use entity name if individual names are not available
        if not last_name and en_data.get('entity_name'):
            last_name = en_data.get('entity_name', '')
        
        return {
            "hamdb": {
                "version": "1",
                "callsign": {
                    "call": callsign,
                    "class": operator_class,
                    "expires": expires,
                    "status": status,
                    "grid": "NOT_FOUND",  # Would need LA.dat for actual coordinates
                    "lat": "NOT_FOUND",
                    "lon": "NOT_FOUND",
                    "fname": first_name,
                    "mi": mi,
                    "name": last_name,
                    "suffix": suffix,
                    "addr1": en_data.get('street_address', ''),
                    "addr2": en_data.get('city', ''),
                    "state": en_data.get('state', ''),
                    "zip": en_data.get('zip_code', ''),
                    "country": "United States"
                },
                "messages": {
                    "status": "OK"
                }
            }
        }
    
    def write_json_file(self, callsign: str, data: Dict):
        """Write JSON file to nested directory structure"""
        if len(callsign) < 3:
            print(f"Warning: Skipping callsign '{callsign}' (too short)")
            return
        
        first = callsign[0].upper()
        second = callsign[1].upper()
        third = callsign[2].upper()
        
        dir_path = self.output_dir / first / second / third
        dir_path.mkdir(parents=True, exist_ok=True)
        
        file_path = dir_path / f"{callsign}.json"
        with open(file_path, 'w') as f:
            json.dump(data, f, indent=2)
    
    def process_batch(self, callsigns: list, hd_file: str, en_file: str, am_file: str):
        """Process a batch of callsigns"""
        for i, callsign in enumerate(callsigns):
            if i % 1000 == 0 and i > 0:
                print(f"  Processed {i}/{len(callsigns)} callsigns...")
            
            # Get records for this callsign
            hd_row = self.get_record_for_callsign(hd_file, callsign, 'HD')
            en_row = self.get_record_for_callsign(en_file, callsign, 'EN')
            am_row = self.get_record_for_callsign(am_file, callsign, 'AM')
            
            if not hd_row:
                continue
            
            hd_data = self.parse_hd_record(hd_row)
            en_data = self.parse_en_record(en_row) if en_row else {}
            am_data = self.parse_am_record(am_row) if am_row else {}
            
            json_data = self.create_json_record(callsign, hd_data, en_data, am_data)
            self.write_json_file(callsign, json_data)
    
    def process_files_streaming(self, hd_file: str, en_file: str, am_file: str, 
                                filter_callsign: Optional[str] = None, batch_size: int = 10000):
        """Process files in streaming mode to minimize memory usage"""
        
        if filter_callsign:
            # Process single callsign
            print(f"Processing single callsign: {filter_callsign}")
            callsign = filter_callsign.upper()
            
            hd_row = self.get_record_for_callsign(hd_file, callsign, 'HD')
            if not hd_row:
                print(f"Callsign {callsign} not found in HD.dat")
                return
            
            en_row = self.get_record_for_callsign(en_file, callsign, 'EN')
            am_row = self.get_record_for_callsign(am_file, callsign, 'AM')
            
            hd_data = self.parse_hd_record(hd_row)
            en_data = self.parse_en_record(en_row) if en_row else {}
            am_data = self.parse_am_record(am_row) if am_row else {}
            
            json_data = self.create_json_record(callsign, hd_data, en_data, am_data)
            self.write_json_file(callsign, json_data)
            print(f"Created JSON for {callsign}")
        else:
            # Build index of all callsigns
            callsigns = self.build_callsign_index(hd_file)
            total = len(callsigns)
            print(f"Processing {total} callsigns in batches of {batch_size}...")
            
            # Process in batches
            callsign_list = list(callsigns)
            for i in range(0, len(callsign_list), batch_size):
                batch = callsign_list[i:i + batch_size]
                batch_num = (i // batch_size) + 1
                total_batches = (len(callsign_list) + batch_size - 1) // batch_size
                print(f"Processing batch {batch_num}/{total_batches}...")
                self.process_batch(batch, hd_file, en_file, am_file)
            
            print(f"Completed processing {total} callsigns")


def main():
    parser = argparse.ArgumentParser(description='Process FCC ULS Amateur Radio Data (Memory Optimized)')
    parser.add_argument('--full', action='store_true', help='Download and process full database')
    parser.add_argument('--daily', action='store_true', help='Download and process daily updates')
    parser.add_argument('--file', type=str, help='Process a specific ZIP file')
    parser.add_argument('--output', type=str, default='output', help='Output directory')
    parser.add_argument('--callsign', type=str, help='Process only a specific callsign')
    parser.add_argument('--batch-size', type=int, default=10000, 
                       help='Number of callsigns to process per batch (default: 10000)')
    
    args = parser.parse_args()
    
    if not (args.full or args.daily or args.file):
        parser.print_help()
        sys.exit(1)
    
    processor = ULSProcessorOptimized(args.output)
    
    # Create temporary directory for downloads
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_path = Path(temp_dir)
        
        if args.full:
            # Download full database
            url = "https://data.fcc.gov/download/pub/uls/complete/l_amat.zip"
            zip_file = temp_path / "l_amat.zip"
            
            if not processor.download_file(url, str(zip_file)):
                sys.exit(1)
            
            if not processor.extract_zip(str(zip_file), str(temp_path)):
                sys.exit(1)
        
        elif args.daily:
            # Download daily updates
            today = datetime.now().strftime('%m%d%Y')
            url = f"https://data.fcc.gov/download/pub/uls/daily/l_am_{today}.zip"
            zip_file = temp_path / f"l_am_{today}.zip"
            
            if not processor.download_file(url, str(zip_file)):
                print("Daily file not available. Try --full instead.")
                sys.exit(1)
            
            if not processor.extract_zip(str(zip_file), str(temp_path)):
                sys.exit(1)
        
        elif args.file:
            # Use provided file
            zip_file = Path(args.file)
            if not zip_file.exists():
                print(f"File not found: {args.file}")
                sys.exit(1)
            
            if not processor.extract_zip(str(zip_file), str(temp_path)):
                sys.exit(1)
        
        # Process the files
        hd_file = temp_path / "HD.dat"
        en_file = temp_path / "EN.dat"
        am_file = temp_path / "AM.dat"
        
        if not all(f.exists() for f in [hd_file, en_file, am_file]):
            print("Error: Required .dat files not found in archive")
            sys.exit(1)
        
        processor.process_files_streaming(
            str(hd_file), 
            str(en_file), 
            str(am_file),
            filter_callsign=args.callsign,
            batch_size=args.batch_size
        )
        
        print("Processing complete!")


if __name__ == "__main__":
    main()
