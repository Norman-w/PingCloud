package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"pingpong/db"
	"pingpong/models"
)

func GetPlayers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(
		"SELECT id, name, COALESCE(gender,''), initial_rating, current_rating, COALESCE(reference_rating,0), created_at FROM players ORDER BY current_rating DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	players := make([]models.Player, 0)
	for rows.Next() {
		var p models.Player
		if err := rows.Scan(&p.ID, &p.Name, &p.Gender, &p.InitialRating, &p.CurrentRating, &p.ReferenceRating, &p.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		players = append(players, p)
	}

	writeJSON(w, players)
}

func GetPlayer(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/players/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid player id", http.StatusBadRequest)
		return
	}

	var p models.Player
	err = db.DB.QueryRow(
		"SELECT id, name, COALESCE(gender,''), initial_rating, current_rating, COALESCE(reference_rating,0), created_at FROM players WHERE id = $1", id,
	).Scan(&p.ID, &p.Name, &p.Gender, &p.InitialRating, &p.CurrentRating, &p.ReferenceRating, &p.CreatedAt)
	if err != nil {
		http.Error(w, "player not found", http.StatusNotFound)
		return
	}

	// Fetch match history with forfeit info
	rows, err := db.DB.Query(
		`SELECT m.id, m.player_a_id, m.player_b_id, m.score_a, m.score_b,
		        m.rating_change_a, m.rating_change_b, m.winner_id, m.played_at,
		        COALESCE(m.forfeit, false),
		        pa.name, pb.name
		 FROM matches m
		 JOIN players pa ON pa.id = m.player_a_id
		 JOIN players pb ON pb.id = m.player_b_id
		 WHERE (m.player_a_id = $1 OR m.player_b_id = $1) AND m.score_a IS NOT NULL
		 ORDER BY m.played_at DESC LIMIT 20`, id)
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
			&m.Forfeit, &m.PlayerAName, &m.PlayerBName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		matches = append(matches, m)
	}

	// Fetch stats (excluding forfeits)
	var wins, losses, forfeitWins, forfeits int
	db.DB.QueryRow(`SELECT
		COALESCE(SUM(CASE WHEN winner_id=$1 AND forfeit=false THEN 1 ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN winner_id IS NOT NULL AND winner_id!=$1 AND forfeit=false THEN 1 ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN winner_id=$1 AND forfeit=true THEN 1 ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN winner_id IS NOT NULL AND winner_id!=$1 AND forfeit=true THEN 1 ELSE 0 END),0)
		FROM matches WHERE (player_a_id=$1 OR player_b_id=$1) AND score_a IS NOT NULL`,
		id).Scan(&wins, &losses, &forfeitWins, &forfeits)

	writeJSON(w, map[string]interface{}{
		"player":       p,
		"matches":      matches,
		"wins":         wins,
		"losses":       losses,
		"forfeit_wins": forfeitWins,
		"forfeits":     forfeits,
	})
}

func CreatePlayer(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	rating := req.InitialRating
	if rating <= 0 {
		rating = 1500
	}
	gender := req.Gender
	if gender != "male" && gender != "female" {
		gender = ""
	}
	refRating := req.ReferenceRating

	var p models.Player
	err := db.DB.QueryRow(
		"INSERT INTO players (name, gender, initial_rating, current_rating, reference_rating) VALUES ($1, $2, $3, $4, $5) RETURNING id, name, COALESCE(gender,''), initial_rating, current_rating, COALESCE(reference_rating,0), created_at",
		strings.TrimSpace(req.Name), gender, rating, rating, refRating,
	).Scan(&p.ID, &p.Name, &p.Gender, &p.InitialRating, &p.CurrentRating, &p.ReferenceRating, &p.CreatedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, p)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
