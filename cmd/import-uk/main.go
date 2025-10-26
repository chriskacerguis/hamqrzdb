package main

import (
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const (
	// Ofcom Amateur Radio License data
	// URL: https://www.ofcom.org.uk/manage-your-licence/radiocommunication-licences/amateur-radio/amateur-radio-licence-data
	OfcomDataURL = "https://www.ofcom.org.uk/siteassets/resources/documents/manage-your-licence/amateur/callsign-030625.csv?v=398262"
)

var (
	dbFlag       = flag.String("db", "hamqrzdb.sqlite", "Path to SQLite database")
	downloadFlag = flag.Bool("download", true, "Download fresh data from Ofcom")
	fileFlag     = flag.String("file", "", "Use local CSV file instead of downloading")
)

type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase(dbPath string) (*Database, error) {
	log.Printf("Connecting to database: %s", dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Optimize SQLite for bulk inserts
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=10000",
		"PRAGMA temp_store=MEMORY",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return nil, fmt.Errorf("failed to set pragma: %w", err)
		}
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

// DownloadFile downloads a file from URL to filepath
func DownloadFile(url, filepath string) error {
	log.Printf("Downloading %s...", url)

	// Create request with browser-like headers to bypass Cloudflare protection
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Add browser-like headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s (status code: %d)", resp.Status, resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	log.Printf("Downloaded to %s", filepath)
	return nil
}

// ProcessOfcomCSV processes the Ofcom amateur radio CSV file
// Format: Licence Number,Call sign,First name,Surname,Full address,Postcode,Licence status,Licence valid from,Licence valid to
func (d *Database) ProcessOfcomCSV(csvPath string) error {
	log.Println("Processing Ofcom amateur radio data...")

	file, err := os.Open(csvPath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Read header
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}
	log.Printf("CSV Header: %v", header)

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO callsigns (
			callsign, license_status, grant_date, expired_date,
			first_name, last_name, street_address, zip_code,
			radio_service_code, last_updated
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(callsign) DO UPDATE SET
			license_status = CASE WHEN excluded.license_status != '' THEN excluded.license_status ELSE callsigns.license_status END,
			grant_date = CASE WHEN excluded.grant_date != '' THEN excluded.grant_date ELSE callsigns.grant_date END,
			expired_date = CASE WHEN excluded.expired_date != '' THEN excluded.expired_date ELSE callsigns.expired_date END,
			first_name = CASE WHEN excluded.first_name != '' THEN excluded.first_name ELSE callsigns.first_name END,
			last_name = CASE WHEN excluded.last_name != '' THEN excluded.last_name ELSE callsigns.last_name END,
			street_address = CASE WHEN excluded.street_address != '' THEN excluded.street_address ELSE callsigns.street_address END,
			zip_code = CASE WHEN excluded.zip_code != '' THEN excluded.zip_code ELSE callsigns.zip_code END,
			radio_service_code = CASE WHEN excluded.radio_service_code != '' THEN excluded.radio_service_code ELSE callsigns.radio_service_code END,
			last_updated = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	count := 0
	skipped := 0

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Warning: CSV parse error (row skipped): %v", err)
			skipped++
			continue
		}

		// Expected columns: Licence Number,Call sign,First name,Surname,Full address,Postcode,Licence status,Licence valid from,Licence valid to
		if len(row) < 9 {
			continue
		}

		// licenceNumber := strings.TrimSpace(row[0]) // Not currently used
		callsign := strings.TrimSpace(row[1])
		firstName := strings.TrimSpace(row[2])
		surname := strings.TrimSpace(row[3])
		fullAddress := strings.TrimSpace(row[4])
		postcode := strings.TrimSpace(row[5])
		status := strings.TrimSpace(row[6])
		validFrom := strings.TrimSpace(row[7])
		validTo := strings.TrimSpace(row[8])

		if callsign == "" {
			continue
		}

		// Map UK status to FCC-like status (A=Active, E=Expired, etc.)
		licenseStatus := "A"
		if strings.Contains(strings.ToLower(status), "revoked") {
			licenseStatus = "R"
		} else if strings.Contains(strings.ToLower(status), "expired") {
			licenseStatus = "E"
		}

		_, err = stmt.Exec(
			callsign,
			licenseStatus,
			validFrom,
			validTo,
			firstName,
			surname,
			fullAddress,
			postcode,
			"UK", // Mark as UK license
		)
		if err != nil {
			log.Printf("Error inserting UK record for %s: %v", callsign, err)
			continue
		}

		count++
		if count%1000 == 0 {
			log.Printf("  Loaded %d UK records...", count)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Loaded %d UK amateur radio records", count)
	if skipped > 0 {
		log.Printf("Skipped %d records due to parse errors", skipped)
	}

	return nil
}

func main() {
	flag.Parse()

	log.SetFlags(log.LstdFlags)

	// Connect to database
	db, err := NewDatabase(*dbFlag)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	var csvFile string

	if *fileFlag != "" {
		// Use provided file
		csvFile = *fileFlag
		if _, err := os.Stat(csvFile); os.IsNotExist(err) {
			log.Fatalf("File not found: %s", csvFile)
		}
	} else if *downloadFlag {
		// Download from Ofcom
		tempDir, err := os.MkdirTemp("", "uk-amateur-*")
		if err != nil {
			log.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		csvFile = filepath.Join(tempDir, "amateur-current.csv")
		if err := DownloadFile(OfcomDataURL, csvFile); err != nil {
			log.Fatalf("Failed to download: %v", err)
		}
	} else {
		log.Fatal("Either --download or --file must be specified")
	}

	// Process the CSV
	if err := db.ProcessOfcomCSV(csvFile); err != nil {
		log.Fatalf("Failed to process UK data: %v", err)
	}

	log.Println("\nUK import complete!")
	log.Printf("Database: %s", *dbFlag)
}
