package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// HamDBResponse represents the HamDB API response format
type HamDBResponse struct {
	HamDB HamDBData `json:"hamdb"`
}

type HamDBData struct {
	Version  string            `json:"version"`
	Callsign CallsignData      `json:"callsign"`
	Messages map[string]string `json:"messages"`
}

type CallsignData struct {
	Call    string `json:"call"`
	Class   string `json:"class"`
	Expires string `json:"expires"`
	Status  string `json:"status"`
	Grid    string `json:"grid"`
	Lat     string `json:"lat"`
	Lon     string `json:"lon"`
	FName   string `json:"fname"`
	MI      string `json:"mi"`
	Name    string `json:"name"`
	Suffix  string `json:"suffix"`
	Addr1   string `json:"addr1"`
	Addr2   string `json:"addr2"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
	Country string `json:"country"`
}

var (
	db   *sql.DB
	dbMu sync.RWMutex
)

func setDB(d *sql.DB) {
	dbMu.Lock()
	db = d
	dbMu.Unlock()
}

func getDB() *sql.DB {
	dbMu.RLock()
	defer dbMu.RUnlock()
	return db
}

func main() {
	// Get configuration from environment
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "/data/hamqrzdb.sqlite"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Ensure database exists (create schema if missing) and open read-only connection
	var err error
	conn, err := ensureDatabase(dbPath)
	if err != nil {
		// Don't exit; start without DB and allow it to be created/populated later
		log.Printf("Database not ready: %v", err)
		setDB(nil)
	} else {
		setDB(conn)
	}
	defer func() {
		if d := getDB(); d != nil {
			_ = d.Close()
		}
	}()

	// If DB is connected, configure pool; otherwise begin background connector
	if d := getDB(); d != nil {
		d.SetMaxOpenConns(25)
		d.SetMaxIdleConns(5)
		d.SetConnMaxLifetime(5 * time.Minute)
		if err := d.Ping(); err != nil {
			log.Printf("Failed to connect to database: %v", err)
		} else {
			log.Printf("Connected to database: %s", dbPath)
		}
	}

	// Start background connector to attach when DB becomes available
	startDBConnector(dbPath)

	// Setup HTTP handlers
	http.HandleFunc("/v1/", corsMiddleware(handleCallsignLookup))
	http.HandleFunc("/health", corsMiddleware(handleHealth))
	http.HandleFunc("/", corsMiddleware(handleIndex))

	// Start server
	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// ensureDatabase verifies the database file exists at path. If it doesn't,
// it creates a new SQLite database with the required schema, then returns a
// read-only connection suitable for serving API traffic.
func ensureDatabase(dbPath string) (*sql.DB, error) {
	// If file doesn't exist, attempt to create it with the schema
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		// If missing, don't force-create; allow container to start and DB to be built separately
		// Ensure parent directory exists if it's not a file bind mount
		if dir := filepath.Dir(dbPath); dir != "." && dir != "" {
			_ = os.MkdirAll(dir, 0o755)
		}
		return nil, fmt.Errorf("database file not found at %s", dbPath)
	}

	// Open read-only connection for serving
	ro, err := sql.Open("sqlite3", dbPath+"?cache=shared&mode=ro")
	if err != nil {
		// Provide a clearer hint if the failure is due to read-only mount on first start
		return nil, fmt.Errorf("failed to open database (read-only). If this is first start, ensure the DB file is writable or pre-created at %s: %w", dbPath, err)
	}
	return ro, nil
}

// Note: Schema creation is handled by the processor; the API attaches in
// read-only mode and will connect once the DB file exists.

// startDBConnector periodically attempts to connect to the database in read-only
// mode. This allows the API to start before the DB exists and attach later once
// the database file is created/populated by a separate process.
func startDBConnector(dbPath string) {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if getDB() != nil {
				// Optionally verify connection remains healthy
				if err := getDB().Ping(); err != nil {
					log.Printf("Database connection lost: %v", err)
					d := getDB()
					if d != nil {
						_ = d.Close()
					}
					setDB(nil)
				}
				continue
			}
			// Attempt to connect
			conn, err := sql.Open("sqlite3", dbPath+"?cache=shared&mode=ro")
			if err != nil {
				continue
			}
			if err := conn.Ping(); err != nil {
				_ = conn.Close()
				continue
			}
			setDB(conn)
			log.Printf("Database connected: %s", dbPath)
		}
	}()
}

// corsMiddleware adds CORS headers to all responses
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// handleCallsignLookup handles /v1/{callsign}/json/{app} requests
func handleCallsignLookup(w http.ResponseWriter, r *http.Request) {
	// Parse URL path: /v1/{callsign}/json/{app}
	path := strings.TrimPrefix(r.URL.Path, "/v1/")
	parts := strings.Split(path, "/")

	if len(parts) < 3 || parts[1] != "json" {
		writeNotFound(w, "INVALID_URL")
		return
	}

	callsign := strings.ToUpper(parts[0])

	// Look up callsign in database
	data, found := lookupCallsign(callsign)
	if !found {
		writeNotFound(w, callsign)
		return
	}

	// Return successful response
	response := HamDBResponse{
		HamDB: HamDBData{
			Version:  "1",
			Callsign: data,
			Messages: map[string]string{"status": "OK"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// lookupCallsign queries the database for a callsign (case-insensitive)
func lookupCallsign(callsign string) (CallsignData, bool) {
	if getDB() == nil {
		// DB not ready yet
		return CallsignData{}, false
	}
	query := `
		SELECT 
			callsign, license_status, expired_date, operator_class,
			grid_square, latitude, longitude,
			first_name, mi, last_name, suffix,
			street_address, city, state, zip_code, 'United States' as country
		FROM callsigns
		WHERE UPPER(callsign) = UPPER(?)
		LIMIT 1
	`

	var data CallsignData
	var lat, lon sql.NullFloat64
	var gridSquare, expiredDate, mi, suffix, streetAddress, city, state, zipCode sql.NullString
	var firstName, lastName sql.NullString

	err := getDB().QueryRow(query, callsign).Scan(
		&data.Call, &data.Status, &expiredDate, &data.Class,
		&gridSquare, &lat, &lon,
		&firstName, &mi, &lastName, &suffix,
		&streetAddress, &city, &state, &zipCode, &data.Country,
	)

	if err == sql.ErrNoRows {
		log.Printf("No rows found for callsign: %s", callsign)
		return CallsignData{}, false
	}

	if err != nil {
		log.Printf("Database error looking up %s: %v", callsign, err)
		return CallsignData{}, false
	}

	log.Printf("Successfully found callsign: %s (status: %s, class: %s)", data.Call, data.Status, data.Class)

	// Convert nullable fields to strings
	if firstName.Valid {
		data.FName = firstName.String
	}
	if lastName.Valid {
		data.Name = lastName.String
	}
	if expiredDate.Valid {
		data.Expires = expiredDate.String
	}
	if gridSquare.Valid {
		data.Grid = gridSquare.String
	}
	if lat.Valid {
		data.Lat = fmt.Sprintf("%.7f", lat.Float64)
	}
	if lon.Valid {
		data.Lon = fmt.Sprintf("%.7f", lon.Float64)
	}
	if mi.Valid {
		data.MI = mi.String
	}
	if suffix.Valid {
		data.Suffix = suffix.String
	}
	if streetAddress.Valid {
		data.Addr1 = streetAddress.String
	}
	if city.Valid {
		data.Addr2 = city.String
	}
	if state.Valid {
		data.State = state.String
	}
	if zipCode.Valid {
		data.Zip = zipCode.String
	}

	return data, true
} // writeNotFound writes a NOT_FOUND response
func writeNotFound(w http.ResponseWriter, callsign string) {
	response := HamDBResponse{
		HamDB: HamDBData{
			Version: "1",
			Callsign: CallsignData{
				Call:    "NOT_FOUND",
				Class:   "NOT_FOUND",
				Expires: "NOT_FOUND",
				Status:  "NOT_FOUND",
				Grid:    "NOT_FOUND",
				Lat:     "NOT_FOUND",
				Lon:     "NOT_FOUND",
				FName:   "NOT_FOUND",
				MI:      "NOT_FOUND",
				Name:    "NOT_FOUND",
				Suffix:  "NOT_FOUND",
				Addr1:   "NOT_FOUND",
				Addr2:   "NOT_FOUND",
				State:   "NOT_FOUND",
				Zip:     "NOT_FOUND",
				Country: "NOT_FOUND",
			},
			Messages: map[string]string{"status": "NOT_FOUND"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleHealth handles /health requests
func handleHealth(w http.ResponseWriter, r *http.Request) {
	// Test database connection
	d := getDB()
	if d == nil || d.Ping() != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
			"error":  "database not connected",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// handleIndex serves the index.html file
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Try to serve index.html from current directory or /app directory
	indexPaths := []string{
		"html/index.html",
		"/app/index.html",
		"index.html",
	}

	var content []byte
	var err error

	for _, path := range indexPaths {
		content, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		// Fallback to a simple HTML response
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
	<title>HamQRZDB API</title>
	<style>
		body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
		code { background: #f4f4f4; padding: 2px 6px; border-radius: 3px; }
	</style>
</head>
<body>
	<h1>HamQRZDB API</h1>
	<p>Welcome to the HamQRZDB callsign lookup API!</p>
	<h2>Usage</h2>
	<p>Look up a callsign:</p>
	<code>GET /v1/{callsign}/json/{appname}</code>
	<h2>Example</h2>
	<p><a href="/v1/KJ5DJC/json/demo">https://lookup.kj5djc.com/v1/KJ5DJC/json/demo</a></p>
	<h2>Health Check</h2>
	<p><a href="/health">https://lookup.kj5djc.com/health</a></p>
</body>
</html>`)
		return
	}

	// Serve the index.html file
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}
