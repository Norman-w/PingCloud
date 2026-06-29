package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"pingpong/db"
	"pingpong/rating"
)

// ---- Request types ----

type CreateSessionRequest struct {
	Name      string `json:"name"`
	PlayerIDs []int  `json:"player_ids"`
}

type ScoreMatchRequest struct {
	ScoreA int `json:"score_a"`
	ScoreB int `json:"score_b"`
}

// ---- Response types ----

type SessionInfo struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type SessionPlayer struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	StartingRating int    `json:"starting_rating"`
	Wins           int    `json:"wins"`
	Losses         int    `json:"losses"`
	ForfeitWins    int    `json:"forfeit_wins"`
	Forfeits       int    `json:"forfeits"`
}

type SessionMatch struct {
	ID             int    `json:"id"`
	Round          int    `json:"round"`
	PlayerAID      int    `json:"player_a_id"`
	PlayerBID      int    `json:"player_b_id"`
	PlayerAName    string `json:"player_a_name"`
	PlayerBName    string `json:"player_b_name"`
	ScoreA         int    `json:"score_a"`
	ScoreB         int    `json:"score_b"`
	RatingChangeA  int    `json:"rating_change_a"`
	RatingChangeB  int    `json:"rating_change_b"`
	WinnerID       int    `json:"winner_id"`
	Played         bool   `json:"played"`
	Forfeit        bool   `json:"forfeit"`
}

type SessionDetail struct {
	SessionInfo
	Players []SessionPlayer `json:"players"`
	Matches []SessionMatch  `json:"matches"`
}

// ---- Handlers ----

func CreateSession(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if len(req.PlayerIDs) < 2 {
		http.Error(w, "至少需要2名球员", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "乒乒活动"
	}

	// Dedup player IDs
	seen := map[int]bool{}
	unique := []int{}
	for _, pid := range req.PlayerIDs {
		if !seen[pid] {
			seen[pid] = true
			unique = append(unique, pid)
		}
	}

	// Start transaction
	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Create session
	var sessionID int
	err = tx.QueryRow(
		"INSERT INTO sessions (name) VALUES ($1) RETURNING id",
		name,
	).Scan(&sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Add session players with frozen starting ratings
	for _, pid := range unique {
		var sr int
		db.DB.QueryRow("SELECT current_rating FROM players WHERE id=$1", pid).Scan(&sr)
		_, err = tx.Exec(
			"INSERT INTO session_players (session_id, player_id, starting_rating) VALUES ($1, $2, $3)",
			sessionID, pid, sr,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Generate round-robin matches
	err = generateRoundRobin(tx, sessionID, unique)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]int{"id": sessionID})
}

func GetSessions(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT s.id, s.name, s.status, s.created_at,
			COUNT(DISTINCT sp.player_id) AS player_count,
			COUNT(DISTINCT m.id) AS match_count,
			COUNT(DISTINCT CASE WHEN m.score_a IS NULL THEN m.id END) AS unplayed_count
		FROM sessions s
		LEFT JOIN session_players sp ON sp.session_id = s.id
		LEFT JOIN matches m ON m.session_id = s.id
		GROUP BY s.id
		ORDER BY s.created_at DESC LIMIT 20
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type sessionRow struct {
		SessionInfo
		PlayerCount   int `json:"player_count"`
		MatchCount    int `json:"match_count"`
		UnplayedCount int `json:"unplayed_count"`
	}

	sessions := make([]sessionRow, 0)
	for rows.Next() {
		var s sessionRow
		var createdAt interface{}
		if err := rows.Scan(&s.ID, &s.Name, &s.Status, &createdAt,
			&s.PlayerCount, &s.MatchCount, &s.UnplayedCount); err != nil {
			continue
		}
		if t, ok := createdAt.(interface{ String() string }); ok {
			s.CreatedAt = t.String()
		}
		sessions = append(sessions, s)
	}
	writeJSON(w, sessions)
}

func GetSession(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from URL: /api/sessions/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	// Handle sub-paths like /api/sessions/1/matches/2
	parts := strings.Split(path, "/")
	sessionID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}

	// Get session info
	var detail SessionDetail
	var createdAt interface{}
	err = db.DB.QueryRow(
		"SELECT id, name, status, created_at FROM sessions WHERE id = $1",
		sessionID,
	).Scan(&detail.ID, &detail.Name, &detail.Status, &createdAt)
	if err != nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}
	if t, ok := createdAt.(interface{ String() string }); ok {
		detail.CreatedAt = t.String()
	}

	// Get session players with stats (excluding forfeits from losses)
	rows, err := db.DB.Query(`
		SELECT p.id, p.name, sp.starting_rating,
			COALESCE(SUM(CASE WHEN m.winner_id = p.id AND m.forfeit = false AND m.session_id = $1 THEN 1 ELSE 0 END), 0) AS wins,
			COALESCE(SUM(CASE WHEN m.winner_id IS NOT NULL AND m.winner_id != p.id AND m.forfeit = false AND m.session_id = $1 THEN 1 ELSE 0 END), 0) AS losses,
			COALESCE(SUM(CASE WHEN m.winner_id = p.id AND m.forfeit = true AND m.session_id = $1 THEN 1 ELSE 0 END), 0) AS forfeit_wins,
			COALESCE(SUM(CASE WHEN m.winner_id IS NOT NULL AND m.winner_id != p.id AND m.forfeit = true AND m.session_id = $1 THEN 1 ELSE 0 END), 0) AS forfeits
		FROM session_players sp
		JOIN players p ON p.id = sp.player_id
		LEFT JOIN matches m ON (m.player_a_id = p.id OR m.player_b_id = p.id) AND m.session_id = $1
		WHERE sp.session_id = $1
		GROUP BY p.id, sp.starting_rating
		ORDER BY wins DESC, sp.starting_rating DESC
	`, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	detail.Players = make([]SessionPlayer, 0)
	for rows.Next() {
		var sp SessionPlayer
		if err := rows.Scan(&sp.ID, &sp.Name, &sp.StartingRating, &sp.Wins, &sp.Losses, &sp.ForfeitWins, &sp.Forfeits); err != nil {
			continue
		}
		detail.Players = append(detail.Players, sp)
	}

	// Get session matches
	mRows, err := db.DB.Query(`
		SELECT m.id, m.round, m.player_a_id, m.player_b_id,
			COALESCE(m.score_a, -1), COALESCE(m.score_b, -1),
			COALESCE(m.rating_change_a, 0), COALESCE(m.rating_change_b, 0),
			COALESCE(m.winner_id, 0), COALESCE(m.forfeit, false),
			pa.name, pb.name
		FROM matches m
		JOIN players pa ON pa.id = m.player_a_id
		JOIN players pb ON pb.id = m.player_b_id
		WHERE m.session_id = $1
		ORDER BY m.round, m.id
	`, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer mRows.Close()

	detail.Matches = make([]SessionMatch, 0)
	for mRows.Next() {
		var sm SessionMatch
		var scoreA, scoreB, winnerID int
		if err := mRows.Scan(&sm.ID, &sm.Round, &sm.PlayerAID, &sm.PlayerBID,
			&scoreA, &scoreB, &sm.RatingChangeA, &sm.RatingChangeB, &winnerID, &sm.Forfeit,
			&sm.PlayerAName, &sm.PlayerBName); err != nil {
			continue
		}
		if scoreA >= 0 {
			sm.ScoreA = scoreA
			sm.ScoreB = scoreB
			sm.WinnerID = winnerID
			sm.Played = true
		}
		detail.Matches = append(detail.Matches, sm)
	}

	writeJSON(w, detail)
}

func ScoreSessionMatch(w http.ResponseWriter, r *http.Request) {
	// URL: /api/sessions/{sessionID}/matches/{matchID}
	path := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[1] != "matches" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	sessionID, _ := strconv.Atoi(parts[0])
	matchID, _ := strconv.Atoi(parts[2])

	var req ScoreMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.ScoreA == req.ScoreB {
		http.Error(w, "不能平局", http.StatusBadRequest)
		return
	}

	// Verify match belongs to session
	var playerAID, playerBID int
	var oldChangeA, oldChangeB int
	var existingScoreA sql.NullInt64
	err := db.DB.QueryRow(
		"SELECT player_a_id, player_b_id, rating_change_a, rating_change_b, score_a FROM matches WHERE id = $1 AND session_id = $2",
		matchID, sessionID,
	).Scan(&playerAID, &playerBID, &oldChangeA, &oldChangeB, &existingScoreA)
	if err != nil {
		http.Error(w, "match not found in this session", http.StatusNotFound)
		return
	}

	// Get frozen starting ratings for this session
	var srA, srB int
	db.DB.QueryRow("SELECT starting_rating FROM session_players WHERE session_id=$1 AND player_id=$2", sessionID, playerAID).Scan(&srA)
	db.DB.QueryRow("SELECT starting_rating FROM session_players WHERE session_id=$1 AND player_id=$2", sessionID, playerBID).Scan(&srB)

	// Get match counts for K-factor calculation
	var countA, countB int
	db.DB.QueryRow("SELECT COUNT(*) FROM matches WHERE (player_a_id=$1 OR player_b_id=$1) AND score_a IS NOT NULL", playerAID).Scan(&countA)
	db.DB.QueryRow("SELECT COUNT(*) FROM matches WHERE (player_a_id=$1 OR player_b_id=$1) AND score_a IS NOT NULL", playerBID).Scan(&countB)
	kA := rating.PlayerK(countA, srA)
	kB := rating.PlayerK(countB, srB)

	// Calculate rating changes based on FROZEN session-start ratings (not live current_rating)
	changeA, changeB, winnerIdx := rating.CalculateRatingChanges(srA, srB, kA, kB, req.ScoreA, req.ScoreB)
	winnerID := playerAID
	if winnerIdx == 1 {
		winnerID = playerBID
	}

	// If correcting, reverse old changes from current_rating first
	if existingScoreA.Valid {
		db.DB.Exec("UPDATE players SET current_rating = current_rating - $1 WHERE id = $2", oldChangeA, playerAID)
		db.DB.Exec("UPDATE players SET current_rating = current_rating - $1 WHERE id = $2", oldChangeB, playerBID)
	}

	// Update match record
	_, err = db.DB.Exec(`
		UPDATE matches SET score_a=$1, score_b=$2, rating_change_a=$3, rating_change_b=$4, winner_id=$5
		WHERE id=$6 AND session_id=$7`,
		req.ScoreA, req.ScoreB, changeA, changeB, winnerID, matchID, sessionID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Do NOT update current_rating now — only at session complete

	writeJSON(w, map[string]interface{}{
		"id":              matchID,
		"rating_change_a": changeA,
		"rating_change_b": changeB,
		"winner_id":       winnerID,
		"corrected":       existingScoreA.Valid,
	})
}

func ForfeitMatch(w http.ResponseWriter, r *http.Request) {
	// URL: /api/sessions/{sessionID}/matches/{matchID}/forfeit
	path := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[1] != "matches" || parts[3] != "forfeit" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	sessionID, _ := strconv.Atoi(parts[0])
	matchID, _ := strconv.Atoi(parts[2])

	var req struct {
		WinnerID int `json:"winner_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Verify match belongs to session
	var pa, pb int
	var existing sql.NullInt64
	err := db.DB.QueryRow(
		"SELECT player_a_id, player_b_id, score_a FROM matches WHERE id=$1 AND session_id=$2",
		matchID, sessionID,
	).Scan(&pa, &pb, &existing)
	if err != nil {
		http.Error(w, "match not found", http.StatusNotFound)
		return
	}

	// If correcting, clear old result first
	if existing.Valid {
		var oldA, oldB int
		db.DB.QueryRow("SELECT rating_change_a, rating_change_b FROM matches WHERE id=$1", matchID).Scan(&oldA, &oldB)
		db.DB.Exec("UPDATE players SET current_rating = current_rating - $1 WHERE id = $2", oldA, pa)
		db.DB.Exec("UPDATE players SET current_rating = current_rating - $1 WHERE id = $2", oldB, pb)
	}

	// Forfeit: no rating changes, set winner
	_, err = db.DB.Exec(`
		UPDATE matches SET score_a=0, score_b=0, rating_change_a=0, rating_change_b=0,
		winner_id=$1, forfeit=true WHERE id=$2`,
		req.WinnerID, matchID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{
		"id":       matchID,
		"forfeit":  true,
		"winner_id": req.WinnerID,
	})
}

func UpdateSession(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	parts := strings.Split(path, "/")
	sessionID, _ := strconv.Atoi(parts[0])

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	_, err := db.DB.Exec("UPDATE sessions SET name=$1 WHERE id=$2", strings.TrimSpace(req.Name), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func CompleteSession(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	parts := strings.Split(path, "/")
	sessionID, _ := strconv.Atoi(parts[0])

	// Guard: only apply once
	var currentStatus string
	db.DB.QueryRow("SELECT status FROM sessions WHERE id=$1", sessionID).Scan(&currentStatus)
	if currentStatus == "completed" {
		writeJSON(w, map[string]string{"status": "already_completed"})
		return
	}

	// Apply all match rating changes to current_rating
	_, err := db.DB.Exec(`
		UPDATE players p SET current_rating = p.current_rating + (
			SELECT COALESCE(SUM(
				CASE WHEN m.player_a_id = p.id THEN m.rating_change_a
				     WHEN m.player_b_id = p.id THEN m.rating_change_b
				     ELSE 0 END), 0)
			FROM matches m
			WHERE m.session_id = $1 AND m.score_a IS NOT NULL
			AND (m.player_a_id = p.id OR m.player_b_id = p.id)
		) WHERE p.id IN (SELECT player_id FROM session_players WHERE session_id = $1)
	`, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = db.DB.Exec("UPDATE sessions SET status='completed' WHERE id=$1", sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "completed"})
}
