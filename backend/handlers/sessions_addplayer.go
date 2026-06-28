package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"pingpong/db"
)

func AddPlayerToSession(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	parts := strings.Split(path, "/")
	sessionID, _ := strconv.Atoi(parts[0])

	var req struct {
		PlayerID int `json:"player_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	var status string
	db.DB.QueryRow("SELECT status FROM sessions WHERE id=$1", sessionID).Scan(&status)
	if status != "active" {
		http.Error(w, "session is not active", http.StatusBadRequest)
		return
	}

	var exists bool
	db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM session_players WHERE session_id=$1 AND player_id=$2)", sessionID, req.PlayerID).Scan(&exists)
	if exists {
		http.Error(w, "球员已在活动中", http.StatusConflict)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO session_players (session_id, player_id) VALUES ($1, $2)", sessionID, req.PlayerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete unplayed matches
	_, err = tx.Exec("DELETE FROM matches WHERE session_id=$1 AND score_a IS NULL", sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get all session players
	rows, err := tx.Query("SELECT player_id FROM session_players WHERE session_id=$1", sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var playerIDs []int
	for rows.Next() {
		var pid int
		rows.Scan(&pid)
		playerIDs = append(playerIDs, pid)
	}

	// Regenerate unplayed matches from max round
	var maxRound int
	tx.QueryRow("SELECT COALESCE(MAX(round), 0) FROM matches WHERE session_id=$1", sessionID).Scan(&maxRound)
	err = regenerateUnplayed(tx, sessionID, playerIDs, maxRound)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"status": "ok"})
}
