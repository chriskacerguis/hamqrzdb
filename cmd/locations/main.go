package main

import (
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// LocationData represents location information
type LocationData struct {
	Callsign   string
	Latitude   float64
	Longitude  float64
	GridSquare string
}

// LocationProcessor handles location data processing
type LocationProcessor struct {
	db *sql.DB
}

// NewLocationProcessor creates a new location processor
func NewLocationProcessor(dbPath string) (*LocationProcessor, error) {
	log.Printf("Connecting to database: %s", dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set pragmas for better performance
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=10000",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return nil, fmt.Errorf("failed to set pragma: %w", err)
		}
	}

	return &LocationProcessor{db: db}, nil
}

// Close closes the database connection
func (lp *LocationProcessor) Close() error {
	if lp.db != nil {
		log.Println("Database connection closed")
		return lp.db.Close()
	}
	return nil
}

// CalculateGridSquare calculates Maidenhead grid square from coordinates
// Returns 6-character grid square (e.g., EM10ci)
func CalculateGridSquare(lat, lon float64) string {
	if lat == 0 && lon == 0 {
		return ""
	}

	// Adjust longitude to 0-360 range
	adjLon := lon + 180
	adjLat := lat + 90

	// Field (20° longitude × 10° latitude)
	fieldLon := byte(int(adjLon/20) + 'A')
	fieldLat := byte(int(adjLat/10) + 'A')

	// Square (2° longitude × 1° latitude)
	squareLon := byte(int(math.Mod(adjLon, 20)/2) + '0')
	squareLat := byte(int(math.Mod(adjLat, 10)) + '0')

	// Subsquare (5' longitude × 2.5' latitude)
	subsquareLon := byte(int(math.Mod(adjLon, 2)*12) + 'a')
	subsquareLat := byte(int(math.Mod(adjLat, 1)*24) + 'a')

	return string([]byte{fieldLon, fieldLat, squareLon, squareLat, subsquareLon, subsquareLat})
}

// UpdateLocation updates a callsign with location data
func (lp *LocationProcessor) UpdateLocation(callsign string, latitude, longitude float64) error {
	gridSquare := CalculateGridSquare(latitude, longitude)

	query := `
		UPDATE callsigns
		SET latitude = ?,
		    longitude = ?,
		    grid_square = ?,
		    last_updated = CURRENT_TIMESTAMP
		WHERE callsign = ?
	`

	_, err := lp.db.Exec(query, latitude, longitude, gridSquare, callsign)
	return err
}

// ProcessLAFile processes LA.dat file and updates database with locations
func (lp *LocationProcessor) ProcessLAFile(filePath, filterCallsign string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	log.Printf("Processing location data from %s...", filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '|'
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	// Begin transaction for better performance
	tx, err := lp.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare update statement
	stmt, err := tx.Prepare(`
		UPDATE callsigns
		SET latitude = ?,
		    longitude = ?,
		    grid_square = ?,
		    last_updated = CURRENT_TIMESTAMP
		WHERE callsign = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	updated := 0
	batchSize := 1000

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		// Check if this is an LA record
		if len(row) < 5 || row[0] != "LA" {
			continue
		}

		callsign := strings.TrimSpace(row[4])
		if callsign == "" {
			continue
		}

		// Filter by callsign if specified
		if filterCallsign != "" && !strings.EqualFold(callsign, filterCallsign) {
			continue
		}

		// LA format fields:
		// [13] = lat_degrees, [14] = lat_minutes, [15] = lat_seconds, [16] = lat_direction
		// [17] = lon_degrees, [18] = lon_minutes, [19] = lon_seconds, [20] = lon_direction

		// Parse latitude
		latDeg := parseFloat(row, 13)
		latMin := parseFloat(row, 14)
		latSec := parseFloat(row, 15)
		latDir := getString(row, 16, "N")

		// Parse longitude
		lonDeg := parseFloat(row, 17)
		lonMin := parseFloat(row, 18)
		lonSec := parseFloat(row, 19)
		lonDir := getString(row, 20, "W")

		// Convert to decimal degrees
		latitude := latDeg + (latMin / 60.0) + (latSec / 3600.0)
		if latDir == "S" {
			latitude = -latitude
		}

		longitude := lonDeg + (lonMin / 60.0) + (lonSec / 3600.0)
		if lonDir == "W" {
			longitude = -longitude
		}

		// Skip if coordinates are 0,0 (invalid)
		if latitude == 0 && longitude == 0 {
			continue
		}

		// Calculate grid square
		gridSquare := CalculateGridSquare(latitude, longitude)

		// Update database
		if _, err := stmt.Exec(latitude, longitude, gridSquare, callsign); err != nil {
			log.Printf("Error updating location for %s: %v", callsign, err)
			continue
		}

		updated++
		count++

		// Commit batch
		if count%batchSize == 0 {
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit batch: %w", err)
			}

			// Start new transaction
			tx, err = lp.db.Begin()
			if err != nil {
				return fmt.Errorf("failed to begin new transaction: %w", err)
			}

			stmt, err = tx.Prepare(`
				UPDATE callsigns
				SET latitude = ?,
				    longitude = ?,
				    grid_square = ?,
				    last_updated = CURRENT_TIMESTAMP
				WHERE callsign = ?
			`)
			if err != nil {
				return fmt.Errorf("failed to prepare new statement: %w", err)
			}

			if count%10000 == 0 {
				log.Printf("  Processed %d LA records, updated %d callsigns...", count, updated)
			}
		}
	}

	// Final commit
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit final batch: %w", err)
	}

	log.Printf("Processed %d LA records", count)
	log.Printf("Updated %d callsigns with location data", updated)

	return nil
}

// Helper functions

func parseFloat(row []string, index int) float64 {
	if len(row) <= index {
		return 0
	}

	val := strings.TrimSpace(row[index])
	if val == "" {
		return 0
	}

	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0
	}

	return f
}

func getString(row []string, index int, defaultVal string) string {
	if len(row) <= index {
		return defaultVal
	}

	val := strings.TrimSpace(row[index])
	if val == "" {
		return defaultVal
	}

	return val
}

func main() {
	laFileFlag := flag.String("la-file", "", "Path to LA.dat file (required)")
	callsignFlag := flag.String("callsign", "", "Process only this specific callsign")
	regenerateFlag := flag.Bool("regenerate", false, "Regenerate JSON files after updating locations")
	dbFlag := flag.String("db", "hamqrzdb.sqlite", "Path to SQLite database")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Process FCC ULS location data and update database\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Process location data\n")
		fmt.Fprintf(os.Stderr, "  %s --la-file temp_uls/LA.dat\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Process and regenerate JSON files\n")
		fmt.Fprintf(os.Stderr, "  %s --la-file temp_uls/LA.dat --regenerate\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Process specific callsign\n")
		fmt.Fprintf(os.Stderr, "  %s --la-file temp_uls/LA.dat --callsign KJ5DJC\n", os.Args[0])
	}

	flag.Parse()

	if *laFileFlag == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Process locations
	processor, err := NewLocationProcessor(*dbFlag)
	if err != nil {
		log.Fatalf("Failed to create processor: %v", err)
	}
	defer processor.Close()

	if err := processor.ProcessLAFile(*laFileFlag, *callsignFlag); err != nil {
		log.Fatalf("Location processing failed: %v", err)
	}

	// Regenerate JSON files if requested
	if *regenerateFlag {
		log.Println("\nRegenerating JSON files...")
		log.Println("Note: Use hamqrzdb-process --generate to regenerate JSON files")
		log.Println("Example: ./bin/hamqrzdb-process --generate --db " + *dbFlag)
		if *callsignFlag != "" {
			log.Println("         --callsign " + *callsignFlag)
		}
	}

	log.Println("\nLocation processing complete!")
}
