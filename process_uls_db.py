#!/usr/bin/env python3
"""
FCC ULS Amateur Radio Callsign Data Processor - SQLite Database Version
Processes HD.dat, EN.dat, and AM.dat files into SQLite database, then generates JSON files

Architecture:
1. Load FCC data into SQLite database (normalized tables)
2. Generate JSON files from database queries
3. Support incremental updates efficiently

Memory usage: Minimal - streams data directly to/from SQLite
"""

import argparse
import csv
import json
import os
import sys
import sqlite3
import zipfile
from datetime import datetime
from pathlib import Path
from typing import Optional
import urllib.request
import shutil
import tempfile


class ULSDatabase:
    """Manage ULS data in SQLite database"""
    
    def __init__(self, db_path: str = "hamqrzdb.sqlite"):
        self.db_path = db_path
        self.conn = None
        self.cursor = None
        
    def connect(self):
        """Connect to database and create tables if needed"""
        print(f"Connecting to database: {self.db_path}")
        self.conn = sqlite3.connect(self.db_path)
        self.conn.row_factory = sqlite3.Row
        self.cursor = self.conn.cursor()
        self.create_tables()
        
    def create_tables(self):
        """Create database tables"""
        print("Creating/verifying database schema...")
        
        # Main callsigns table
        self.cursor.execute('''
            CREATE TABLE IF NOT EXISTS callsigns (
                callsign TEXT PRIMARY KEY,
                license_status TEXT,
                radio_service_code TEXT,
                grant_date TEXT,
                expired_date TEXT,
                cancellation_date TEXT,
                operator_class TEXT,
                group_code TEXT,
                region_code TEXT,
                first_name TEXT,
                mi TEXT,
                last_name TEXT,
                suffix TEXT,
                entity_name TEXT,
                street_address TEXT,
                city TEXT,
                state TEXT,
                zip_code TEXT,
                latitude REAL,
                longitude REAL,
                grid_square TEXT,
                last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        ''')
        
        # Create indexes for performance
        self.cursor.execute('''
            CREATE INDEX IF NOT EXISTS idx_callsign ON callsigns(callsign)
        ''')
        
        self.cursor.execute('''
            CREATE INDEX IF NOT EXISTS idx_status ON callsigns(license_status)
        ''')
        
        self.conn.commit()
        print("Database schema ready")
    
    def upsert_callsign(self, data: dict):
        """Insert or update a callsign record"""
        self.cursor.execute('''
            INSERT INTO callsigns (
                callsign, license_status, radio_service_code, grant_date, 
                expired_date, cancellation_date, operator_class, group_code, 
                region_code, first_name, mi, last_name, suffix, entity_name,
                street_address, city, state, zip_code, latitude, longitude, 
                grid_square, last_updated
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
            ON CONFLICT(callsign) DO UPDATE SET
                license_status = CASE WHEN excluded.license_status != '' THEN excluded.license_status ELSE callsigns.license_status END,
                radio_service_code = CASE WHEN excluded.radio_service_code != '' THEN excluded.radio_service_code ELSE callsigns.radio_service_code END,
                grant_date = CASE WHEN excluded.grant_date != '' THEN excluded.grant_date ELSE callsigns.grant_date END,
                expired_date = CASE WHEN excluded.expired_date != '' THEN excluded.expired_date ELSE callsigns.expired_date END,
                cancellation_date = CASE WHEN excluded.cancellation_date != '' THEN excluded.cancellation_date ELSE callsigns.cancellation_date END,
                operator_class = CASE WHEN excluded.operator_class != '' THEN excluded.operator_class ELSE callsigns.operator_class END,
                group_code = CASE WHEN excluded.group_code != '' THEN excluded.group_code ELSE callsigns.group_code END,
                region_code = CASE WHEN excluded.region_code != '' THEN excluded.region_code ELSE callsigns.region_code END,
                first_name = CASE WHEN excluded.first_name != '' THEN excluded.first_name ELSE callsigns.first_name END,
                mi = CASE WHEN excluded.mi != '' THEN excluded.mi ELSE callsigns.mi END,
                last_name = CASE WHEN excluded.last_name != '' THEN excluded.last_name ELSE callsigns.last_name END,
                suffix = CASE WHEN excluded.suffix != '' THEN excluded.suffix ELSE callsigns.suffix END,
                entity_name = CASE WHEN excluded.entity_name != '' THEN excluded.entity_name ELSE callsigns.entity_name END,
                street_address = CASE WHEN excluded.street_address != '' THEN excluded.street_address ELSE callsigns.street_address END,
                city = CASE WHEN excluded.city != '' THEN excluded.city ELSE callsigns.city END,
                state = CASE WHEN excluded.state != '' THEN excluded.state ELSE callsigns.state END,
                zip_code = CASE WHEN excluded.zip_code != '' THEN excluded.zip_code ELSE callsigns.zip_code END,
                latitude = CASE WHEN excluded.latitude IS NOT NULL THEN excluded.latitude ELSE callsigns.latitude END,
                longitude = CASE WHEN excluded.longitude IS NOT NULL THEN excluded.longitude ELSE callsigns.longitude END,
                grid_square = CASE WHEN excluded.grid_square != '' THEN excluded.grid_square ELSE callsigns.grid_square END,
                last_updated = CURRENT_TIMESTAMP
        ''', (
            data.get('callsign'),
            data.get('license_status'),
            data.get('radio_service_code'),
            data.get('grant_date'),
            data.get('expired_date'),
            data.get('cancellation_date'),
            data.get('operator_class'),
            data.get('group_code'),
            data.get('region_code'),
            data.get('first_name'),
            data.get('mi'),
            data.get('last_name'),
            data.get('suffix'),
            data.get('entity_name'),
            data.get('street_address'),
            data.get('city'),
            data.get('state'),
            data.get('zip_code'),
            data.get('latitude'),
            data.get('longitude'),
            data.get('grid_square')
        ))
    
    def get_callsign(self, callsign: str) -> Optional[dict]:
        """Retrieve a callsign record"""
        self.cursor.execute('SELECT * FROM callsigns WHERE callsign = ?', (callsign,))
        row = self.cursor.fetchone()
        if row:
            return dict(row)
        return None
    
    def get_all_callsigns(self):
        """Generator that yields all callsigns one at a time"""
        # Use a separate cursor to avoid conflicts with nested queries
        cursor = self.conn.cursor()
        cursor.execute('SELECT callsign FROM callsigns ORDER BY callsign')
        while True:
            row = cursor.fetchone()
            if row is None:
                break
            yield row[0]
        cursor.close()
    
    def get_callsign_count(self) -> int:
        """Get total number of callsigns in database"""
        self.cursor.execute('SELECT COUNT(*) FROM callsigns')
        return self.cursor.fetchone()[0]
    
    def commit(self):
        """Commit changes to database"""
        self.conn.commit()
    
    def close(self):
        """Close database connection"""
        if self.conn:
            self.conn.close()


class ULSProcessorDB:
    """Process ULS amateur radio data files using SQLite database"""
    
    def __init__(self, db_path: str = "hamqrzdb.sqlite", output_dir: str = "output"):
        self.db = ULSDatabase(db_path)
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
    
    def load_hd_file(self, file_path: str, filter_callsign: Optional[str] = None):
        """Load HD.dat into database"""
        print("Loading HD.dat into database...")
        count = 0
        batch = []
        batch_size = 1000
        
        with open(file_path, 'r', encoding='latin-1') as f:
            reader = csv.reader(f, delimiter='|')
            for row in reader:
                if len(row) < 5:
                    continue
                if row[0] != 'HD':
                    continue
                
                callsign = row[4].strip()
                if not callsign:
                    continue
                
                if filter_callsign and callsign != filter_callsign.upper():
                    continue
                
                data = {
                    'callsign': callsign,
                    'license_status': row[5].strip() if len(row) > 5 else '',
                    'radio_service_code': row[6].strip() if len(row) > 6 else '',
                    'grant_date': row[7].strip() if len(row) > 7 else '',
                    'expired_date': row[8].strip() if len(row) > 8 else '',
                    'cancellation_date': row[9].strip() if len(row) > 9 else '',
                }
                
                batch.append(data)
                count += 1
                
                if len(batch) >= batch_size:
                    for item in batch:
                        self.db.upsert_callsign(item)
                    self.db.commit()
                    batch = []
                    if count % 10000 == 0:
                        print(f"  Loaded {count} HD records...")
            
            # Process remaining batch
            if batch:
                for item in batch:
                    self.db.upsert_callsign(item)
                self.db.commit()
        
        print(f"Loaded {count} HD records")
    
    def update_en_data(self, file_path: str, filter_callsign: Optional[str] = None):
        """Update database with EN.dat data"""
        print("Updating database with EN.dat...")
        count = 0
        batch = []
        batch_size = 1000
        
        with open(file_path, 'r', encoding='latin-1') as f:
            reader = csv.reader(f, delimiter='|')
            for row in reader:
                if len(row) < 5:
                    continue
                if row[0] != 'EN':
                    continue
                
                callsign = row[4].strip()
                if not callsign:
                    continue
                
                if filter_callsign and callsign != filter_callsign.upper():
                    continue
                
                data = {
                    'callsign': callsign,
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
                
                batch.append(data)
                count += 1
                
                if len(batch) >= batch_size:
                    for item in batch:
                        self.db.upsert_callsign(item)
                    self.db.commit()
                    batch = []
                    if count % 10000 == 0:
                        print(f"  Updated {count} EN records...")
            
            # Process remaining batch
            if batch:
                for item in batch:
                    self.db.upsert_callsign(item)
                self.db.commit()
        
        print(f"Updated {count} EN records")
    
    def update_am_data(self, file_path: str, filter_callsign: Optional[str] = None):
        """Update database with AM.dat data"""
        print("Updating database with AM.dat...")
        count = 0
        batch = []
        batch_size = 1000
        
        with open(file_path, 'r', encoding='latin-1') as f:
            reader = csv.reader(f, delimiter='|')
            for row in reader:
                if len(row) < 5:
                    continue
                if row[0] != 'AM':
                    continue
                
                callsign = row[4].strip()
                if not callsign:
                    continue
                
                if filter_callsign and callsign != filter_callsign.upper():
                    continue
                
                data = {
                    'callsign': callsign,
                    'operator_class': row[5].strip() if len(row) > 5 else '',
                    'group_code': row[6].strip() if len(row) > 6 else '',
                    'region_code': row[7].strip() if len(row) > 7 else '',
                }
                
                batch.append(data)
                count += 1
                
                if len(batch) >= batch_size:
                    for item in batch:
                        self.db.upsert_callsign(item)
                    self.db.commit()
                    batch = []
                    if count % 10000 == 0:
                        print(f"  Updated {count} AM records...")
            
            # Process remaining batch
            if batch:
                for item in batch:
                    self.db.upsert_callsign(item)
                self.db.commit()
        
        print(f"Updated {count} AM records")
    
    def format_expiration_date(self, date_str: str) -> str:
        """Format expiration date to MM/DD/YYYY"""
        if not date_str or len(date_str) != 10:
            return "NOT_FOUND"
        
        try:
            date_obj = datetime.strptime(date_str, '%m/%d/%Y')
            return date_obj.strftime('%m/%d/%Y')
        except ValueError:
            return "NOT_FOUND"
    
    def create_json_from_db_record(self, record: dict) -> dict:
        """Create JSON structure from database record"""
        status_map = {
            'A': 'A',  # Active
            'C': 'C',  # Canceled
            'E': 'E',  # Expired
            'T': 'T',  # Terminated
        }
        
        status = status_map.get(record.get('license_status', ''), 'NOT_FOUND')
        operator_class = record.get('operator_class', 'NOT_FOUND') or 'NOT_FOUND'
        expires = self.format_expiration_date(record.get('expired_date', ''))
        
        # Get name components
        first_name = record.get('first_name', '') or ''
        mi = record.get('mi', '') or ''
        last_name = record.get('last_name', '') or ''
        suffix = record.get('suffix', '') or ''
        
        # Use entity name if individual names are not available
        if not last_name and record.get('entity_name'):
            last_name = record.get('entity_name', '')
        
        # Get coordinates
        grid = record.get('grid_square', 'NOT_FOUND') or 'NOT_FOUND'
        lat = str(record.get('latitude', 'NOT_FOUND')) if record.get('latitude') else 'NOT_FOUND'
        lon = str(record.get('longitude', 'NOT_FOUND')) if record.get('longitude') else 'NOT_FOUND'
        
        return {
            "hamdb": {
                "version": "1",
                "callsign": {
                    "call": record.get('callsign', ''),
                    "class": operator_class,
                    "expires": expires,
                    "status": status,
                    "grid": grid,
                    "lat": lat,
                    "lon": lon,
                    "fname": first_name,
                    "mi": mi,
                    "name": last_name,
                    "suffix": suffix,
                    "addr1": record.get('street_address', '') or '',
                    "addr2": record.get('city', '') or '',
                    "state": record.get('state', '') or '',
                    "zip": record.get('zip_code', '') or '',
                    "country": "United States"
                },
                "messages": {
                    "status": "OK"
                }
            }
        }
    
    def write_json_file(self, callsign: str, data: dict):
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
    
    def generate_json_files(self, filter_callsign: Optional[str] = None):
        """Generate JSON files from database"""
        if filter_callsign:
            print(f"Generating JSON for callsign: {filter_callsign}")
            record = self.db.get_callsign(filter_callsign.upper())
            if not record:
                print(f"Callsign {filter_callsign} not found in database")
                return
            
            json_data = self.create_json_from_db_record(record)
            self.write_json_file(filter_callsign.upper(), json_data)
            print(f"Created JSON for {filter_callsign}")
        else:
            total = self.db.get_callsign_count()
            print(f"Generating JSON files for {total} callsigns...")
            
            count = 0
            for callsign in self.db.get_all_callsigns():
                record = self.db.get_callsign(callsign)
                if record:
                    json_data = self.create_json_from_db_record(record)
                    self.write_json_file(callsign, json_data)
                    count += 1
                    
                    if count % 10000 == 0:
                        print(f"  Generated {count}/{total} JSON files...")
            
            print(f"Generated {count} JSON files")
    
    def load_data_files(self, hd_file: str, en_file: str, am_file: str, filter_callsign: Optional[str] = None):
        """Load all data files into database"""
        self.db.connect()
        
        try:
            self.load_hd_file(hd_file, filter_callsign)
            self.update_en_data(en_file, filter_callsign)
            self.update_am_data(am_file, filter_callsign)
            
            print(f"\nDatabase loaded successfully!")
            print(f"Total callsigns: {self.db.get_callsign_count()}")
        finally:
            self.db.close()


def main():
    parser = argparse.ArgumentParser(description='Process FCC ULS Amateur Radio Data with SQLite Database')
    parser.add_argument('--full', action='store_true', help='Download and process full database')
    parser.add_argument('--daily', action='store_true', help='Download and process daily updates')
    parser.add_argument('--file', type=str, help='Process a specific ZIP file')
    parser.add_argument('--generate', action='store_true', help='Generate JSON files from existing database')
    parser.add_argument('--output', type=str, default='output', help='Output directory for JSON files')
    parser.add_argument('--db', type=str, default='hamqrzdb.sqlite', help='SQLite database path')
    parser.add_argument('--callsign', type=str, help='Process only a specific callsign')
    
    args = parser.parse_args()
    
    if not (args.full or args.daily or args.file or args.generate):
        parser.print_help()
        sys.exit(1)
    
    processor = ULSProcessorDB(db_path=args.db, output_dir=args.output)
    
    # If just generating from existing database
    if args.generate:
        print("Generating JSON files from database...")
        processor.db.connect()
        try:
            processor.generate_json_files(filter_callsign=args.callsign)
        finally:
            processor.db.close()
        print("Generation complete!")
        sys.exit(0)
    
    # Otherwise, load data and generate
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
        
        # Load into database
        processor.load_data_files(str(hd_file), str(en_file), str(am_file), filter_callsign=args.callsign)
        
        # Generate JSON files
        print("\nGenerating JSON files from database...")
        processor.db.connect()
        try:
            processor.generate_json_files(filter_callsign=args.callsign)
        finally:
            processor.db.close()
        
        print("\nProcessing complete!")
        print(f"Database: {args.db}")
        print(f"JSON files: {args.output}/")


if __name__ == "__main__":
    main()
