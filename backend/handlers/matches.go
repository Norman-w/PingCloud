package handlers

import (
	"encoding/json"
	"net/http"

	"pingpong/db"
	"pingpong/models"
	"pingpong/rating"
)

func CreateMatch(w http.ResponseWriter, r *http.Request) {
	var req models.CreateMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.PlayerAID == req.PlayerBID {
		http.Error(w, "cannot play against yourself", http.StatusBadRequest)
		return
	}
	if req.ScoreA < 0 || req.ScoreB < 0 {
		http.Error(w, "scores must be non-negative", http.StatusBadRequest)
		return
	}
	if req.ScoreA == req.ScoreB {
		http.Error(w, "match cannot be a draw", http.StatusBadRequest)
		return
	}

	// Fetch current ratings
	var ratingA, ratingB int
	err := db.DB.QueryRow("SELECT current_rating FROM players WHERE id = $1", req.PlayerAID).Scan(&ratingA)
	if err != nil {
		http.Error(w, "player A not found", http.StatusNotFound)
		return
	}
	err = db.DB.QueryRow("SELECT current_rating FROM players WHERE id = $1", req.PlayerBID).Scan(&ratingB)
	if err != nil {
		http.Error(w, "player B not found", http.StatusNotFound)
		return
	}

	// Get match counts for K-factor calculation
	var countA, countB int
	db.DB.QueryRow("SELECT COUNT(*) FROM matches WHERE (player_a_id=$1 OR player_b_id=$1) AND score_a IS NOT NULL", req.PlayerAID).Scan(&countA)
	db.DB.QueryRow("SELECT COUNT(*) FROM matches WHERE (player_a_id=$1 OR player_b_id=$1) AND score_a IS NOT NULL", req.PlayerBID).Scan(&countB)
	kA := rating.PlayerK(countA, ratingA)
	kB := rating.PlayerK(countB, ratingB)

	// Calculate rating changes with per-player K
	changeA, changeB, winnerIdx := rating.CalculateRatingChanges(ratingA, ratingB, kA, kB, req.ScoreA, req.ScoreB)

	winnerID := req.PlayerAID
	if winnerIdx == 1 {
		winnerID = req.PlayerBID
	}

	// Insert match
	var m models.Match
	err = db.DB.QueryRow(
		`INSERT INTO matches (player_a_id, player_b_id, score_a, score_b, rating_change_a, rating_change_b, winner_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, player_a_id, player_b_id, score_a, score_b, rating_change_a, rating_change_b, winner_id, played_at`,
		req.PlayerAID, req.PlayerBID, req.ScoreA, req.ScoreB, changeA, changeB, winnerID,
	).Scan(&m.ID, &m.PlayerAID, &m.PlayerBID, &m.ScoreA, &m.ScoreB,
		&m.RatingChangeA, &m.RatingChangeB, &m.WinnerID, &m.PlayedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update player ratings
	_, err = db.DB.Exec("UPDATE players SET current_rating = current_rating + $1 WHERE id = $2", changeA, req.PlayerAID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = db.DB.Exec("UPDATE players SET current_rating = current_rating + $1 WHERE id = $2", changeB, req.PlayerBID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch names for response
	db.DB.QueryRow("SELECT name FROM players WHERE id = $1", m.PlayerAID).Scan(&m.PlayerAName)
	db.DB.QueryRow("SELECT name FROM players WHERE id = $1", m.PlayerBID).Scan(&m.PlayerBName)

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, m)
}

func GetMatches(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(
		`SELECT m.id, m.player_a_id, m.player_b_id, m.score_a, m.score_b,
		        m.rating_change_a, m.rating_change_b, m.winner_id, m.played_at,
		        pa.name, pb.name
		 FROM matches m
		 JOIN players pa ON pa.id = m.player_a_id
		 JOIN players pb ON pb.id = m.player_b_id
		 ORDER BY m.played_at DESC LIMIT 50`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	matches := make([]models.Match, 0)
	for rows.Next() {
		var m models.Match
		if err := rows.Scan(&m.ID, &m.PlayerAID, &m.PlayerBID, &m.ScoreA, &m.ScoreB,
			&m.RatingChangeA, &m.RatingChangeB, &m.WinnerID, &m.PlayedAt,
			&m.PlayerAName, &m.PlayerBName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		matches = append(matches, m)
	}

	writeJSON(w, matches)
}
