package main

import (
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	FullDatabaseURL   = "https://data.fcc.gov/download/pub/uls/complete/l_amat.zip"
	DailyUpdateURLFmt = "https://data.fcc.gov/download/pub/uls/daily/l_am_%s.zip"
	BatchSize         = 1000
)

// CallsignRecord represents a complete callsign record
type CallsignRecord struct {
	Callsign         string
	LicenseStatus    string
	RadioServiceCode string
	GrantDate        string
	ExpiredDate      string
	CancellationDate string
	OperatorClass    string
	GroupCode        string
	RegionCode       string
	FirstName        string
	MI               string
	LastName         string
	Suffix           string
	EntityName       string
	StreetAddress    string
	City             string
	State            string
	ZipCode          string
	Latitude         float64
	Longitude        float64
	GridSquare       string
}

// Database handles SQLite operations
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

	d := &Database{db: db}
	if err := d.createTables(); err != nil {
		return nil, err
	}

	return d, nil
}

// createTables creates the database schema
func (d *Database) createTables() error {
	log.Println("Creating/verifying database schema...")

	schema := `
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
	);

	CREATE INDEX IF NOT EXISTS idx_callsign ON callsigns(callsign);
	CREATE INDEX IF NOT EXISTS idx_status ON callsigns(license_status);
	`

	if _, err := d.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	log.Println("Database schema ready")
	return nil
}

// UpsertCallsign inserts or updates a callsign record
func (d *Database) UpsertCallsign(record CallsignRecord) error {
	query := `
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
			latitude = CASE WHEN excluded.latitude != 0 THEN excluded.latitude ELSE callsigns.latitude END,
			longitude = CASE WHEN excluded.longitude != 0 THEN excluded.longitude ELSE callsigns.longitude END,
			grid_square = CASE WHEN excluded.grid_square != '' THEN excluded.grid_square ELSE callsigns.grid_square END,
			last_updated = CURRENT_TIMESTAMP
	`

	_, err := d.db.Exec(query,
		record.Callsign, record.LicenseStatus, record.RadioServiceCode, record.GrantDate,
		record.ExpiredDate, record.CancellationDate, record.OperatorClass, record.GroupCode,
		record.RegionCode, record.FirstName, record.MI, record.LastName, record.Suffix,
		record.EntityName, record.StreetAddress, record.City, record.State, record.ZipCode,
		record.Latitude, record.Longitude, record.GridSquare,
	)

	return err
}

// GetCallsign retrieves a callsign record
func (d *Database) GetCallsign(callsign string) (*CallsignRecord, error) {
	query := `
		SELECT callsign, license_status, radio_service_code, grant_date,
			expired_date, cancellation_date, operator_class, group_code,
			region_code, first_name, mi, last_name, suffix, entity_name,
			street_address, city, state, zip_code, latitude, longitude, grid_square
		FROM callsigns
		WHERE UPPER(callsign) = UPPER(?)
	`

	var record CallsignRecord
	var lat, lon sql.NullFloat64
	var mi, suffix, firstName, lastName, entityName, streetAddress, city, state, zipCode, gridSquare sql.NullString

	err := d.db.QueryRow(query, callsign).Scan(
		&record.Callsign, &record.LicenseStatus, &record.RadioServiceCode, &record.GrantDate,
		&record.ExpiredDate, &record.CancellationDate, &record.OperatorClass, &record.GroupCode,
		&record.RegionCode, &firstName, &mi, &lastName, &suffix,
		&entityName, &streetAddress, &city, &state, &zipCode,
		&lat, &lon, &gridSquare,
	)

	if err != nil {
		return nil, err
	}

	// Handle nullable string fields
	if firstName.Valid {
		record.FirstName = firstName.String
	}
	if mi.Valid {
		record.MI = mi.String
	}
	if lastName.Valid {
		record.LastName = lastName.String
	}
	if suffix.Valid {
		record.Suffix = suffix.String
	}
	if entityName.Valid {
		record.EntityName = entityName.String
	}
	if streetAddress.Valid {
		record.StreetAddress = streetAddress.String
	}
	if city.Valid {
		record.City = city.String
	}
	if state.Valid {
		record.State = state.String
	}
	if zipCode.Valid {
		record.ZipCode = zipCode.String
	}
	if gridSquare.Valid {
		record.GridSquare = gridSquare.String
	}

	if lat.Valid {
		record.Latitude = lat.Float64
	}
	if lon.Valid {
		record.Longitude = lon.Float64
	}

	return &record, nil
}

// GetCallsignCount returns the total number of callsigns
func (d *Database) GetCallsignCount() (int, error) {
	var count int
	err := d.db.QueryRow("SELECT COUNT(*) FROM callsigns").Scan(&count)
	return count, err
}

// GetAllCallsigns returns all callsigns (for JSON generation)
func (d *Database) GetAllCallsigns() ([]string, error) {
	rows, err := d.db.Query("SELECT callsign FROM callsigns ORDER BY callsign")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var callsigns []string
	for rows.Next() {
		var callsign string
		if err := rows.Scan(&callsign); err != nil {
			return nil, err
		}
		callsigns = append(callsigns, callsign)
	}

	return callsigns, rows.Err()
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// Processor handles FCC data processing
type Processor struct {
	db *Database
}

// NewProcessor creates a new processor
func NewProcessor(dbPath string) (*Processor, error) {
	db, err := NewDatabase(dbPath)
	if err != nil {
		return nil, err
	}

	return &Processor{
		db: db,
	}, nil
}

// DownloadFile downloads a file from URL
func (p *Processor) DownloadFile(url, destination string) error {
	log.Printf("Downloading %s...", url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	log.Printf("Downloaded to %s", destination)
	return nil
}

// ExtractZip extracts a ZIP file
func (p *Processor) ExtractZip(zipPath, destDir string) error {
	log.Printf("Extracting %s...", zipPath)

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	log.Printf("Extracted to %s", destDir)
	return nil
}

// LoadHDFile loads HD.dat into database
func (p *Processor) LoadHDFile(filePath, filterCallsign string) error {
	log.Println("Loading HD.dat into database...")

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '|'
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	tx, err := p.db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO callsigns (callsign, license_status, radio_service_code, grant_date, expired_date, cancellation_date)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(callsign) DO UPDATE SET
			license_status = CASE WHEN excluded.license_status != '' THEN excluded.license_status ELSE callsigns.license_status END,
			radio_service_code = CASE WHEN excluded.radio_service_code != '' THEN excluded.radio_service_code ELSE callsigns.radio_service_code END,
			grant_date = CASE WHEN excluded.grant_date != '' THEN excluded.grant_date ELSE callsigns.grant_date END,
			expired_date = CASE WHEN excluded.expired_date != '' THEN excluded.expired_date ELSE callsigns.expired_date END,
			cancellation_date = CASE WHEN excluded.cancellation_date != '' THEN excluded.cancellation_date ELSE callsigns.cancellation_date END,
			last_updated = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	count := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if len(row) < 5 || row[0] != "HD" {
			continue
		}

		callsign := strings.TrimSpace(row[4])
		if callsign == "" {
			continue
		}

		if filterCallsign != "" && !strings.EqualFold(callsign, filterCallsign) {
			continue
		}

		licenseStatus := ""
		radioServiceCode := ""
		grantDate := ""
		expiredDate := ""
		cancellationDate := ""
		if len(row) > 5 {
			licenseStatus = strings.TrimSpace(row[5])
		}
		if len(row) > 6 {
			radioServiceCode = strings.TrimSpace(row[6])
		}
		if len(row) > 7 {
			grantDate = strings.TrimSpace(row[7])
		}
		if len(row) > 8 {
			expiredDate = strings.TrimSpace(row[8])
		}
		if len(row) > 9 {
			cancellationDate = strings.TrimSpace(row[9])
		}
		if _, err := stmt.Exec(callsign, licenseStatus, radioServiceCode, grantDate, expiredDate, cancellationDate); err != nil {
			log.Printf("Error inserting HD record: %v", err)
			continue
		}

		count++
		if count%10000 == 0 {
			log.Printf("  Loaded %d HD records...", count)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Loaded %d HD records", count)
	return nil
}

// UpdateENData updates database with EN.dat
func (p *Processor) UpdateENData(filePath, filterCallsign string) error {
	log.Println("Updating database with EN.dat...")

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '|'
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	tx, err := p.db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		UPDATE callsigns SET
			entity_name = CASE WHEN ? != '' THEN ? ELSE entity_name END,
			first_name = CASE WHEN ? != '' THEN ? ELSE first_name END,
			mi = CASE WHEN ? != '' THEN ? ELSE mi END,
			last_name = CASE WHEN ? != '' THEN ? ELSE last_name END,
			suffix = CASE WHEN ? != '' THEN ? ELSE suffix END,
			street_address = CASE WHEN ? != '' THEN ? ELSE street_address END,
			city = CASE WHEN ? != '' THEN ? ELSE city END,
			state = CASE WHEN ? != '' THEN ? ELSE state END,
			zip_code = CASE WHEN ? != '' THEN ? ELSE zip_code END,
			last_updated = CURRENT_TIMESTAMP
		WHERE callsign = ?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	count := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if len(row) < 5 || row[0] != "EN" {
			continue
		}

		callsign := strings.TrimSpace(row[4])
		if callsign == "" {
			continue
		}

		if filterCallsign != "" && !strings.EqualFold(callsign, filterCallsign) {
			continue
		}

		entityName := ""
		firstName := ""
		mi := ""
		lastName := ""
		suffix := ""
		streetAddress := ""
		city := ""
		state := ""
		zipCode := ""

		if len(row) > 7 {
			entityName = strings.TrimSpace(row[7])
		}
		if len(row) > 8 {
			firstName = strings.TrimSpace(row[8])
		}
		if len(row) > 9 {
			mi = strings.TrimSpace(row[9])
		}
		if len(row) > 10 {
			lastName = strings.TrimSpace(row[10])
		}
		if len(row) > 11 {
			suffix = strings.TrimSpace(row[11])
		}
		if len(row) > 15 {
			streetAddress = strings.TrimSpace(row[15])
		}
		if len(row) > 16 {
			city = strings.TrimSpace(row[16])
		}
		if len(row) > 17 {
			state = strings.TrimSpace(row[17])
		}
		if len(row) > 18 {
			zipCode = strings.TrimSpace(row[18])
		}

		if _, err := stmt.Exec(
			entityName, entityName,
			firstName, firstName,
			mi, mi,
			lastName, lastName,
			suffix, suffix,
			streetAddress, streetAddress,
			city, city,
			state, state,
			zipCode, zipCode,
			callsign,
		); err != nil {
			log.Printf("Error updating EN record: %v", err)
			continue
		}

		count++
		if count%10000 == 0 {
			log.Printf("  Updated %d EN records...", count)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Updated %d EN records", count)
	return nil
}

// UpdateAMData updates database with AM.dat
func (p *Processor) UpdateAMData(filePath, filterCallsign string) error {
	log.Println("Updating database with AM.dat...")

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '|'
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true

	tx, err := p.db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		UPDATE callsigns SET
			operator_class = CASE WHEN ? != '' THEN ? ELSE operator_class END,
			group_code = CASE WHEN ? != '' THEN ? ELSE group_code END,
			region_code = CASE WHEN ? != '' THEN ? ELSE region_code END,
			last_updated = CURRENT_TIMESTAMP
		WHERE callsign = ?
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	count := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if len(row) < 5 || row[0] != "AM" {
			continue
		}

		callsign := strings.TrimSpace(row[4])
		if callsign == "" {
			continue
		}

		if filterCallsign != "" && !strings.EqualFold(callsign, filterCallsign) {
			continue
		}

		operatorClass := ""
		groupCode := ""
		regionCode := ""

		if len(row) > 5 {
			operatorClass = strings.TrimSpace(row[5])
		}
		if len(row) > 6 {
			groupCode = strings.TrimSpace(row[6])
		}
		if len(row) > 7 {
			regionCode = strings.TrimSpace(row[7])
		}

		if _, err := stmt.Exec(
			operatorClass, operatorClass,
			groupCode, groupCode,
			regionCode, regionCode,
			callsign,
		); err != nil {
			log.Printf("Error updating AM record: %v", err)
			continue
		}

		count++
		if count%10000 == 0 {
			log.Printf("  Updated %d AM records...", count)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Updated %d AM records", count)
	return nil
}

// FormatExpirationDate formats date to MM/DD/YYYY
func FormatExpirationDate(dateStr string) string {
	if dateStr == "" || len(dateStr) != 10 {
		return "NOT_FOUND"
	}

	t, err := time.Parse("01/02/2006", dateStr)
	if err != nil {
		return "NOT_FOUND"
	}

	return t.Format("01/02/2006")
}

// LoadDataFiles loads all data files into database
func (p *Processor) LoadDataFiles(hdFile, enFile, amFile, filterCallsign string) error {
	if err := p.LoadHDFile(hdFile, filterCallsign); err != nil {
		return fmt.Errorf("failed to load HD file: %w", err)
	}

	if err := p.UpdateENData(enFile, filterCallsign); err != nil {
		return fmt.Errorf("failed to load EN file: %w", err)
	}

	if err := p.UpdateAMData(amFile, filterCallsign); err != nil {
		return fmt.Errorf("failed to load AM file: %w", err)
	}

	total, err := p.db.GetCallsignCount()
	if err != nil {
		return err
	}

	log.Printf("\nDatabase loaded successfully!")
	log.Printf("Total callsigns: %d", total)
	return nil
}

// Close closes the processor
func (p *Processor) Close() error {
	return p.db.Close()
}

// CalculateGridSquare calculates the Maidenhead grid square from latitude and longitude.
// Returns a 6-character grid square (e.g., "EM10ci").
func CalculateGridSquare(lat, lon float64) string {
	// Adjust longitude and latitude to be in the range [0, 360) and [0, 180)
	adjustedLon := lon + 180.0
	adjustedLat := lat + 90.0

	// Calculate field (first pair - letters A-R)
	fieldLon := int(adjustedLon / 20.0)
	fieldLat := int(adjustedLat / 10.0)
	if fieldLon < 0 || fieldLon >= 18 || fieldLat < 0 || fieldLat >= 18 {
		return ""
	}

	// Calculate square (second pair - digits 0-9)
	squareLon := int((adjustedLon - float64(fieldLon)*20.0) / 2.0)
	squareLat := int((adjustedLat - float64(fieldLat)*10.0) / 1.0)
	if squareLon < 0 || squareLon >= 10 || squareLat < 0 || squareLat >= 10 {
		return ""
	}

	// Calculate subsquare (third pair - letters a-x)
	subsquareLon := int((adjustedLon - float64(fieldLon)*20.0 - float64(squareLon)*2.0) / (2.0 / 24.0))
	subsquareLat := int((adjustedLat - float64(fieldLat)*10.0 - float64(squareLat)*1.0) / (1.0 / 24.0))
	if subsquareLon < 0 || subsquareLon >= 24 || subsquareLat < 0 || subsquareLat >= 24 {
		return ""
	}

	// Build the grid square string
	return fmt.Sprintf("%c%c%d%d%c%c",
		'A'+byte(fieldLon),
		'A'+byte(fieldLat),
		squareLon,
		squareLat,
		'a'+byte(subsquareLon),
		'a'+byte(subsquareLat),
	)
}

// parseCoordinate parses FCC coordinate format (degrees, minutes, seconds, direction)
// into a decimal coordinate.
func parseCoordinate(degrees, minutes, seconds, direction string) (float64, error) {
	deg, err := strconv.ParseFloat(degrees, 64)
	if err != nil {
		return 0, err
	}

	min, err := strconv.ParseFloat(minutes, 64)
	if err != nil {
		return 0, err
	}

	sec, err := strconv.ParseFloat(seconds, 64)
	if err != nil {
		return 0, err
	}

	decimal := deg + (min / 60.0) + (sec / 3600.0)

	// South and West are negative
	if direction == "S" || direction == "W" {
		decimal = -decimal
	}

	return decimal, nil
}

// ProcessLAFile processes the FCC LA.dat file and updates location data in the database.
// LA.dat contains latitude/longitude coordinates for callsigns.
func (p *Processor) ProcessLAFile(laFile, filterCallsign string) error {
	file, err := os.Open(laFile)
	if err != nil {
		return fmt.Errorf("failed to open LA file: %w", err)
	}
	defer file.Close()

	log.Printf("Processing location data from: %s", laFile)

	reader := csv.NewReader(file)
	reader.Comma = '|'
	reader.FieldsPerRecord = -1 // Variable number of fields

	updateStmt, err := p.db.db.Prepare(`
		UPDATE callsigns
		SET latitude = ?,
		    longitude = ?,
		    grid_square = ?,
		    last_updated = CURRENT_TIMESTAMP
		WHERE call = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer updateStmt.Close()

	tx, err := p.db.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	count := 0
	updated := 0
	batchSize := 1000

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Warning: Error reading LA record: %v", err)
			continue
		}

		if len(record) < 21 {
			continue
		}

		callsign := strings.TrimSpace(record[4])

		// If filtering by callsign, skip non-matching records
		if filterCallsign != "" && !strings.EqualFold(callsign, filterCallsign) {
			continue
		}

		// Parse latitude: fields 13-16 (degrees, minutes, seconds, direction)
		lat, err := parseCoordinate(record[13], record[14], record[15], record[16])
		if err != nil {
			log.Printf("Warning: Failed to parse latitude for %s: %v", callsign, err)
			continue
		}

		// Parse longitude: fields 17-20 (degrees, minutes, seconds, direction)
		lon, err := parseCoordinate(record[17], record[18], record[19], record[20])
		if err != nil {
			log.Printf("Warning: Failed to parse longitude for %s: %v", callsign, err)
			continue
		}

		// Calculate grid square
		gridSquare := CalculateGridSquare(lat, lon)

		// Update database
		result, err := tx.Stmt(updateStmt).Exec(lat, lon, gridSquare, callsign)
		if err != nil {
			log.Printf("Warning: Failed to update %s: %v", callsign, err)
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			updated++
		}

		count++

		// Commit batch
		if count%batchSize == 0 {
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit batch: %w", err)
			}

			log.Printf("Processed %d records, updated %d callsigns...", count, updated)

			// Start new transaction
			tx, err = p.db.db.Begin()
			if err != nil {
				return fmt.Errorf("failed to begin transaction: %w", err)
			}
		}
	}

	// Commit final batch
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit final batch: %w", err)
	}

	log.Printf("Location processing complete: %d records processed, %d callsigns updated", count, updated)
	return nil
}

func main() {
	fullFlag := flag.Bool("full", false, "Download and process full database")
	dailyFlag := flag.Bool("daily", false, "Download and process daily updates")
	fileFlag := flag.String("file", "", "Process a specific ZIP file")
	dbFlag := flag.String("db", "hamqrzdb.sqlite", "SQLite database path")
	callsignFlag := flag.String("callsign", "", "Process only a specific callsign (requires -full, -daily, or -file)")

	flag.Parse()

	if !*fullFlag && !*dailyFlag && *fileFlag == "" {
		fmt.Fprintln(os.Stderr, "Error: You must specify one of: -full, -daily, or -file")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Examples:")
		fmt.Fprintln(os.Stderr, "  hamqrzdb-process -full                       # Download and process full database")
		fmt.Fprintln(os.Stderr, "  hamqrzdb-process -full -callsign KJ5DJC      # Process only KJ5DJC")
		fmt.Fprintln(os.Stderr, "  hamqrzdb-process -daily                      # Download and process daily updates")
		fmt.Fprintln(os.Stderr, "  hamqrzdb-process -file l_amat.zip            # Process specific ZIP file")
		fmt.Fprintln(os.Stderr, "")
		flag.Usage()
		os.Exit(1)
	}

	processor, err := NewProcessor(*dbFlag)
	if err != nil {
		log.Fatalf("Failed to create processor: %v", err)
	}
	defer processor.Close()

	// Create temporary directory for downloads
	tempDir, err := os.MkdirTemp("", "uls-*")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	var zipFile string

	if *fullFlag {
		// Download full database
		zipFile = filepath.Join(tempDir, "l_amat.zip")
		if err := processor.DownloadFile(FullDatabaseURL, zipFile); err != nil {
			log.Fatalf("Failed to download: %v", err)
		}
	} else if *dailyFlag {
		// Download daily updates
		today := time.Now().Format("01022006")
		url := fmt.Sprintf(DailyUpdateURLFmt, today)
		zipFile = filepath.Join(tempDir, fmt.Sprintf("l_am_%s.zip", today))

		if err := processor.DownloadFile(url, zipFile); err != nil {
			log.Fatalf("Daily file not available. Try --full instead: %v", err)
		}
	} else if *fileFlag != "" {
		zipFile = *fileFlag
		if _, err := os.Stat(zipFile); os.IsNotExist(err) {
			log.Fatalf("File not found: %s", zipFile)
		}
	}

	// Extract ZIP file
	extractDir := filepath.Join(tempDir, "extracted")
	if err := processor.ExtractZip(zipFile, extractDir); err != nil {
		log.Fatalf("Failed to extract: %v", err)
	}

	// Check for required files
	hdFile := filepath.Join(extractDir, "HD.dat")
	enFile := filepath.Join(extractDir, "EN.dat")
	amFile := filepath.Join(extractDir, "AM.dat")

	for _, f := range []string{hdFile, enFile, amFile} {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			log.Fatalf("Required file not found: %s", f)
		}
	}

	// Load into database
	if err := processor.LoadDataFiles(hdFile, enFile, amFile, *callsignFlag); err != nil {
		log.Fatalf("Failed to load data: %v", err)
	}

	log.Println("ULS data processing complete!")

	// Process location data if LA.dat exists
	laFile := filepath.Join(extractDir, "LA.dat")
	if _, err := os.Stat(laFile); err == nil {
		log.Println("LA.dat found, processing location data...")
		if err := processor.ProcessLAFile(laFile, *callsignFlag); err != nil {
			log.Printf("Warning: Failed to process location data: %v", err)
		} else {
			log.Println("Location data processing complete!")
		}
	} else {
		log.Println("LA.dat not found in archive, skipping location data")
	}

	// Final summary
	log.Println("\nProcessing complete!")
	log.Printf("Database: %s", *dbFlag)

	total, err := processor.db.GetCallsignCount()
	if err == nil {
		log.Printf("Total callsigns in database: %d", total)
	}
}
