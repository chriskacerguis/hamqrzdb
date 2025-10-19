package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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

var db *sql.DB

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

	// Open database connection
	var err error
	db, err = sql.Open("sqlite3", dbPath+"?cache=shared&mode=ro")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Printf("Connected to database: %s", dbPath)

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
	query := `
		SELECT 
			callsign, operator_class, expired_date, license_status,
			grid_square, latitude, longitude,
			first_name, mi, last_name, suffix,
			street_address, city, state, zip_code, 'United States' as country
		FROM callsigns
		WHERE UPPER(callsign) = UPPER(?)
		LIMIT 1
	`

	var data CallsignData
	var lat, lon sql.NullFloat64
	var gridSquare, expiredDate sql.NullString
	
	err := db.QueryRow(query, callsign).Scan(
		&data.Call, &data.Class, &expiredDate, &data.Status,
		&gridSquare, &lat, &lon,
		&data.FName, &data.MI, &data.Name, &data.Suffix,
		&data.Addr1, &data.Addr2, &data.State, &data.Zip, &data.Country,
	)

	if err == sql.ErrNoRows {
		return CallsignData{}, false
	}

	if err != nil {
		log.Printf("Database error looking up %s: %v", callsign, err)
		return CallsignData{}, false
	}

	// Convert nullable fields to strings
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

	return data, true
}

// writeNotFound writes a NOT_FOUND response
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
	if err := db.Ping(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
			"error":  err.Error(),
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
