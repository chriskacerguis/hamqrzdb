#!/usr/bin/env python3
"""
FCC ULS Location Data Processor
Processes LA.dat (Location) records and updates the database with coordinates and grid squares.

This is a separate optional processor that can be run after the main data load.
Location data adds latitude, longitude, and grid square information to callsigns.

Usage:
    # Process locations from extracted ULS data
    python3 process_uls_locations.py --la-file temp_uls/LA.dat
    
    # Process locations and regenerate JSON files
    python3 process_uls_locations.py --la-file temp_uls/LA.dat --regenerate
    
    # Process specific callsign only
    python3 process_uls_locations.py --la-file temp_uls/LA.dat --callsign KJ5DJC
"""

import argparse
import csv
import sqlite3
import sys
from pathlib import Path
from typing import Optional
import math


class LocationProcessor:
    """Process FCC ULS location data and update database"""
    
    def __init__(self, db_path: str = "hamqrzdb.sqlite"):
        self.db_path = db_path
        self.conn = None
        self.cursor = None
    
    def connect(self):
        """Connect to database"""
        self.conn = sqlite3.connect(self.db_path)
        self.conn.row_factory = sqlite3.Row
        self.cursor = self.conn.cursor()
        print(f"Connected to database: {self.db_path}")
    
    def close(self):
        """Close database connection"""
        if self.conn:
            self.conn.close()
            print("Database connection closed")
    
    def commit(self):
        """Commit transaction"""
        if self.conn:
            self.conn.commit()
    
    def calculate_grid_square(self, lat: float, lon: float) -> str:
        """
        Calculate Maidenhead grid square from coordinates
        Returns 6-character grid square (e.g., EM10ci)
        """
        if lat is None or lon is None:
            return ""
        
        # Adjust longitude to 0-360 range
        adj_lon = lon + 180
        adj_lat = lat + 90
        
        # Field (20° longitude × 10° latitude)
        field_lon = chr(int(adj_lon / 20) + ord('A'))
        field_lat = chr(int(adj_lat / 10) + ord('A'))
        
        # Square (2° longitude × 1° latitude)
        square_lon = str(int((adj_lon % 20) / 2))
        square_lat = str(int(adj_lat % 10))
        
        # Subsquare (5' longitude × 2.5' latitude)
        subsquare_lon = chr(int((adj_lon % 2) * 12) + ord('a'))
        subsquare_lat = chr(int((adj_lat % 1) * 24) + ord('a'))
        
        return f"{field_lon}{field_lat}{square_lon}{square_lat}{subsquare_lon}{subsquare_lat}"
    
    def update_location(self, callsign: str, latitude: float, longitude: float):
        """Update callsign with location data"""
        grid_square = self.calculate_grid_square(latitude, longitude)
        
        self.cursor.execute('''
            UPDATE callsigns 
            SET latitude = ?, 
                longitude = ?, 
                grid_square = ?,
                last_updated = CURRENT_TIMESTAMP
            WHERE callsign = ?
        ''', (latitude, longitude, grid_square, callsign))
    
    def process_la_file(self, file_path: str, filter_callsign: Optional[str] = None):
        """Process LA.dat file and update database with locations"""
        if not Path(file_path).exists():
            print(f"Error: File not found: {file_path}")
            return False
        
        print(f"Processing location data from {file_path}...")
        count = 0
        updated = 0
        batch_size = 1000
        
        try:
            with open(file_path, 'r', encoding='latin-1') as f:
                reader = csv.reader(f, delimiter='|')
                for row in reader:
                    if len(row) < 5:
                        continue
                    if row[0] != 'LA':
                        continue
                    
                    callsign = row[4].strip()
                    if not callsign:
                        continue
                    
                    # Filter if specific callsign requested
                    if filter_callsign and callsign != filter_callsign.upper():
                        continue
                    
                    # Parse location data
                    # LA format: record_type|system_id|call_sign|location_action|
                    #            location_type_code|location_class_code|
                    #            location_number|site_status|
                    #            corresponding_fixed_location|
                    #            location_address|location_city|location_county|
                    #            location_state|radius_of_operation|
                    #            area_of_operation_code|clearance_indicator|
                    #            ground_elevation|lat_degrees|lat_minutes|lat_seconds|
                    #            lat_direction|long_degrees|long_minutes|long_seconds|
                    #            long_direction|max_lat_degrees|...
                    
                    try:
                        # Extract latitude
                        lat_deg = float(row[13]) if len(row) > 13 and row[13].strip() else 0
                        lat_min = float(row[14]) if len(row) > 14 and row[14].strip() else 0
                        lat_sec = float(row[15]) if len(row) > 15 and row[15].strip() else 0
                        lat_dir = row[16].strip() if len(row) > 16 else 'N'
                        
                        # Extract longitude
                        lon_deg = float(row[17]) if len(row) > 17 and row[17].strip() else 0
                        lon_min = float(row[18]) if len(row) > 18 and row[18].strip() else 0
                        lon_sec = float(row[19]) if len(row) > 19 and row[19].strip() else 0
                        lon_dir = row[20].strip() if len(row) > 20 else 'W'
                        
                        # Convert to decimal degrees
                        latitude = lat_deg + (lat_min / 60.0) + (lat_sec / 3600.0)
                        if lat_dir == 'S':
                            latitude = -latitude
                        
                        longitude = lon_deg + (lon_min / 60.0) + (lon_sec / 3600.0)
                        if lon_dir == 'W':
                            longitude = -longitude
                        
                        # Skip if coordinates are 0,0 (invalid)
                        if latitude == 0 and longitude == 0:
                            continue
                        
                        # Update database
                        self.update_location(callsign, latitude, longitude)
                        updated += 1
                        
                    except (ValueError, IndexError) as e:
                        # Skip records with invalid location data
                        continue
                    
                    count += 1
                    
                    if count % batch_size == 0:
                        self.commit()
                        if count % 10000 == 0:
                            print(f"  Processed {count} LA records, updated {updated} callsigns...")
            
            # Final commit
            self.commit()
            print(f"Processed {count} LA records")
            print(f"Updated {updated} callsigns with location data")
            return True
            
        except Exception as e:
            print(f"Error processing LA file: {e}")
            return False


def main():
    parser = argparse.ArgumentParser(
        description='Process FCC ULS location data and update database',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  # Process location data
  python3 process_uls_locations.py --la-file temp_uls/LA.dat
  
  # Process and regenerate JSON files
  python3 process_uls_locations.py --la-file temp_uls/LA.dat --regenerate
  
  # Process specific callsign
  python3 process_uls_locations.py --la-file temp_uls/LA.dat --callsign KJ5DJC
        """
    )
    
    parser.add_argument('--la-file', required=True,
                        help='Path to LA.dat file')
    parser.add_argument('--callsign',
                        help='Process only this specific callsign')
    parser.add_argument('--regenerate', action='store_true',
                        help='Regenerate JSON files after updating locations')
    parser.add_argument('--db', default='hamqrzdb.sqlite',
                        help='Path to SQLite database (default: hamqrzdb.sqlite)')
    
    args = parser.parse_args()
    
    # Process locations
    processor = LocationProcessor(args.db)
    processor.connect()
    
    success = processor.process_la_file(args.la_file, args.callsign)
    
    processor.close()
    
    if not success:
        print("Location processing failed")
        sys.exit(1)
    
    # Regenerate JSON files if requested
    if args.regenerate:
        print("\nRegenerating JSON files...")
        import subprocess
        cmd = ['python3', 'process_uls_db.py', '--generate']
        if args.callsign:
            cmd.extend(['--callsign', args.callsign])
        
        result = subprocess.run(cmd)
        if result.returncode != 0:
            print("JSON regeneration failed")
            sys.exit(1)
    
    print("\nLocation processing complete!")


if __name__ == '__main__':
    main()
