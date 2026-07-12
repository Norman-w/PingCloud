package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"pingpong/db"
)

type location struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Address      string `json:"address"`
	Phone        string `json:"phone"`
	Notes        string `json:"notes"`
	LocationType string `json:"location_type"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

// GetLocations returns all locations, optionally filtered by query
func GetLocations(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	var query string
	var args []interface{}
	if q != "" {
		query = `SELECT id, name, COALESCE(address,''), COALESCE(phone,''), COALESCE(notes,''), COALESCE(location_type,'球馆')
		         FROM locations WHERE name ILIKE $1 ORDER BY name LIMIT 20`
		args = []interface{}{"%" + q + "%"}
	} else {
		query = `SELECT id, name, COALESCE(address,''), COALESCE(phone,''), COALESCE(notes,''), COALESCE(location_type,'球馆')
		         FROM locations ORDER BY name`
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
		if err := rows.Scan(&l.ID, &l.Name, &l.Address, &l.Phone, &l.Notes, &l.LocationType); err == nil {
			locs = append(locs, l)
		}
	}
	writeJSON(w, locs)
}

// GetLocation returns a single location by ID
func GetLocation(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/locations/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var l location
	err = db.DB.QueryRow(
		`SELECT id, name, COALESCE(address,''), COALESCE(phone,''), COALESCE(notes,''), COALESCE(location_type,'球馆'),
		        COALESCE(created_at::text,''), COALESCE(updated_at::text,'')
		 FROM locations WHERE id=$1`, id,
	).Scan(&l.ID, &l.Name, &l.Address, &l.Phone, &l.Notes, &l.LocationType, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, l)
}

// CreateLocation creates a new location with full details
func CreateLocation(w http.ResponseWriter, r *http.Request) {
	var req location
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Name) == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	var id int
	err := db.DB.QueryRow(
		`INSERT INTO locations (name, address, phone, notes, location_type)
		 VALUES ($1,$2,$3,$4,$5)
		 ON CONFLICT (name) DO UPDATE SET address=EXCLUDED.address, phone=EXCLUDED.phone, notes=EXCLUDED.notes, location_type=EXCLUDED.location_type, updated_at=NOW()
		 RETURNING id`,
		strings.TrimSpace(req.Name), req.Address, req.Phone, req.Notes, req.LocationType,
	).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.ID = id
	req.Name = strings.TrimSpace(req.Name)
	writeJSON(w, req)
}

// UpdateLocation updates an existing location
func UpdateLocation(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/locations/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req location
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}
	_, err = db.DB.Exec(
		`UPDATE locations SET name=$1, address=$2, phone=$3, notes=$4, location_type=$5, updated_at=NOW() WHERE id=$6`,
		strings.TrimSpace(req.Name), req.Address, req.Phone, req.Notes, req.LocationType, id,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.ID = id
	writeJSON(w, req)
}

// DeleteLocation deletes a location
func DeleteLocation(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/locations/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	db.DB.Exec(`DELETE FROM locations WHERE id=$1`, id)
	w.WriteHeader(http.StatusNoContent)
}
