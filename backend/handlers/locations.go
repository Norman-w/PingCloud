package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"pingpong/db"
)

type location struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GetLocations returns all locations, optionally filtered by query
func GetLocations(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	var query string
	var args []interface{}
	if q != "" {
		query = `SELECT id, name FROM locations WHERE name ILIKE $1 ORDER BY name LIMIT 20`
		args = []interface{}{"%" + q + "%"}
	} else {
		query = `SELECT id, name FROM locations ORDER BY name`
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	locs := make([]location, 0)
	for rows.Next() {
		var l location
		if err := rows.Scan(&l.ID, &l.Name); err == nil {
			locs = append(locs, l)
		}
	}
	writeJSON(w, locs)
}

// CreateLocation creates a new location
func CreateLocation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	var id int
	err := db.DB.QueryRow(
		`INSERT INTO locations (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name RETURNING id`,
		strings.TrimSpace(req.Name),
	).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, location{ID: id, Name: strings.TrimSpace(req.Name)})
}
