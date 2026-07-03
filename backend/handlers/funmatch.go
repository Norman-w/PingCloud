package handlers

import (
	"database/sql"
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"pingpong/db"
)

// --- types ---

type FunSessionSummary struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	MaleCount      int    `json:"male_count"`
	FemaleCount    int    `json:"female_count"`
	Status         string `json:"status"`
	MaleWins       int    `json:"male_wins"`
	FemaleWins     int    `json:"female_wins"`
	MaleGameWins   int    `json:"male_game_wins"`
	FemaleGameWins int    `json:"female_game_wins"`
	MalePoints     int    `json:"male_points"`
	FemalePoints   int    `json:"female_points"`
	MatchCount     int    `json:"match_count"`
	UnplayedCount  int    `json:"unplayed_count"`
	CreatedAt      string `json:"created_at"`
}

type FunSessionPlayer struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	CurrentRating   int    `json:"current_rating"`
	ReferenceRating int    `json:"reference_rating"`
	Team            string `json:"team"`
}

type FunDrawRecord struct {
	ID         int    `json:"id"`
	CardType   string `json:"card_type"`
	CardValue  *int   `json:"card_value"`
	CardDetail string `json:"card_detail"`
	PlayerID   int    `json:"player_id"`
	Cancelled  bool   `json:"cancelled"`
	DrawnAt    string `json:"drawn_at"`
}

type FunMatchInfo struct {
	ID               int             `json:"id"`
	MalePlayerID     int             `json:"male_player_id"`
	FemalePlayerID   int             `json:"female_player_id"`
	MalePlayerName   string          `json:"male_player_name"`
	FemalePlayerName string          `json:"female_player_name"`
	Game1ScoreMale   *int            `json:"game1_score_male"`
	Game1ScoreFemale *int            `json:"game1_score_female"`
	Game2ScoreMale   *int            `json:"game2_score_male"`
	Game2ScoreFemale *int            `json:"game2_score_female"`
	Game3ScoreMale   *int            `json:"game3_score_male"`
	Game3ScoreFemale *int            `json:"game3_score_female"`
	WinnerID         *int            `json:"winner_id"`
	WinnerTeam       string          `json:"winner_team"`
	HandicapPoints   int             `json:"handicap_points"`
	Played           bool            `json:"played"`
	Draws            []FunDrawRecord `json:"draws"`
}

type FunSessionDetail struct {
	ID             int                `json:"id"`
	Name           string             `json:"name"`
	MaleCount      int                `json:"male_count"`
	FemaleCount    int                `json:"female_count"`
	Status         string             `json:"status"`
	WinningTeam    string             `json:"winning_team"`
	MaleWins       int                `json:"male_wins"`
	FemaleWins     int                `json:"female_wins"`
	MaleGameWins   int                `json:"male_game_wins"`
	FemaleGameWins int                `json:"female_game_wins"`
	MalePoints     int                `json:"male_points"`
	FemalePoints   int                `json:"female_points"`
	CreatedAt      string             `json:"created_at"`
	Players        []FunSessionPlayer `json:"players"`
	Matches        []FunMatchInfo     `json:"matches"`
}

// 9-card types (unlimited draws, no consumption)
var cardTypes = []struct {
	Type   string
	Value  *int
	Detail string
}{
	{"handicap", intPtr(2), ""},
	{"handicap", intPtr(3), ""},
	{"handicap", intPtr(4), ""},
	{"handicap", intPtr(5), ""},
	{"spin", nil, "topspin"},
	{"spin", nil, "backspin"},
	{"table", nil, "left"},
	{"table", nil, "right"},
	{"defense", nil, ""},
}

// --- handlers ---

func GetFunSessions(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT fs.id, fs.name, fs.male_count, fs.female_count, fs.status,
			fs.male_wins, fs.female_wins,
			COALESCE(fs.male_game_wins,0), COALESCE(fs.female_game_wins,0),
			COALESCE(fs.male_points,0), COALESCE(fs.female_points,0),
			fs.created_at,
			COUNT(fm.id) AS match_count,
			COUNT(CASE WHEN fm.played = false AND fm.deleted = false THEN 1 END) AS unplayed_count
		FROM fun_sessions fs
		LEFT JOIN fun_matches fm ON fm.session_id = fs.id AND fm.deleted = false
		WHERE fs.deleted = false
		GROUP BY fs.id
		ORDER BY fs.created_at DESC LIMIT 20
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var result []FunSessionSummary
	for rows.Next() {
		var s FunSessionSummary
		if err := rows.Scan(&s.ID, &s.Name, &s.MaleCount, &s.FemaleCount, &s.Status,
			&s.MaleWins, &s.FemaleWins, &s.MaleGameWins, &s.FemaleGameWins, &s.MalePoints, &s.FemalePoints,
			&s.CreatedAt, &s.MatchCount, &s.UnplayedCount); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		result = append(result, s)
	}
	if result == nil {
		result = []FunSessionSummary{}
	}
	writeJSON(w, result)
}

func CreateFunSession(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name            string `json:"name"`
		MalePlayerIDs   []int  `json:"male_player_ids"`
		FemalePlayerIDs []int  `json:"female_player_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		req.Name = "趣味赛"
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var sessionID int
	err = tx.QueryRow(
		`INSERT INTO fun_sessions (name, male_count, female_count) VALUES ($1, $2, $3) RETURNING id`,
		req.Name, len(req.MalePlayerIDs), len(req.FemalePlayerIDs),
	).Scan(&sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, pid := range req.MalePlayerIDs {
		_, err = tx.Exec(`INSERT INTO fun_session_players (session_id, player_id, team) VALUES ($1, $2, 'male')`, sessionID, pid)
		if err != nil {
			http.Error(w, "duplicate player: "+err.Error(), http.StatusBadRequest)
			return
		}
	}
	for _, pid := range req.FemalePlayerIDs {
		_, err = tx.Exec(`INSERT INTO fun_session_players (session_id, player_id, team) VALUES ($1, $2, 'female')`, sessionID, pid)
		if err != nil {
			http.Error(w, "duplicate player: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Round-robin schedule between teams: each round every player plays at most once.
	// Team A = males (n players), Team B = females (m players).
	// Round k: female F[j] plays male M[(k + j) % n] for j = 0..m-1, k = 0..n-1.
	// This covers all n*m cross-pairs with n rounds of m matches each.
	males := req.MalePlayerIDs
	females := req.FemalePlayerIDs
	n := len(males)
	m := len(females)

	type pair struct{ m, f int }
	var rounds [][]pair

	// If one team is larger, the smaller team rotates through the larger team
	if n >= m {
		for k := 0; k < n; k++ {
			var round []pair
			for j := 0; j < m; j++ {
				round = append(round, pair{males[(k+j)%n], females[j]})
			}
			// Shuffle within round so matches appear interleaved
			rand.Shuffle(len(round), func(i, j int) { round[i], round[j] = round[j], round[i] })
			rounds = append(rounds, round)
		}
	} else {
		for k := 0; k < m; k++ {
			var round []pair
			for j := 0; j < n; j++ {
				round = append(round, pair{males[j], females[(k+j)%m]})
			}
			rand.Shuffle(len(round), func(i, j int) { round[i], round[j] = round[j], round[i] })
			rounds = append(rounds, round)
		}
	}

	// Flatten rounds into insert order
	for _, round := range rounds {
		for _, p := range round {
			_, err = tx.Exec(`INSERT INTO fun_matches (session_id, male_player_id, female_player_id) VALUES ($1, $2, $3)`, sessionID, p.m, p.f)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]int{"id": sessionID})
}

func GetFunSession(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/fun-sessions/")
	sessionID, _ := strconv.Atoi(path)

	detail := FunSessionDetail{}

	err := db.DB.QueryRow(
		`SELECT id, name, male_count, female_count, status, COALESCE(winning_team,''), male_wins, female_wins,
			COALESCE(male_game_wins,0), COALESCE(female_game_wins,0), COALESCE(male_points,0), COALESCE(female_points,0), created_at
		FROM fun_sessions WHERE id=$1 AND deleted=false`, sessionID,
	).Scan(&detail.ID, &detail.Name, &detail.MaleCount, &detail.FemaleCount,
		&detail.Status, &detail.WinningTeam, &detail.MaleWins, &detail.FemaleWins,
		&detail.MaleGameWins, &detail.FemaleGameWins, &detail.MalePoints, &detail.FemalePoints, &detail.CreatedAt)
	if err != nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	prows, err := db.DB.Query(
		`SELECT p.id, p.name, p.current_rating, COALESCE(p.reference_rating,0), fsp.team
		FROM fun_session_players fsp JOIN players p ON p.id = fsp.player_id
		WHERE fsp.session_id = $1 ORDER BY fsp.team, p.name`, sessionID)
	if err == nil {
		defer prows.Close()
		for prows.Next() {
			var p FunSessionPlayer
			prows.Scan(&p.ID, &p.Name, &p.CurrentRating, &p.ReferenceRating, &p.Team)
			detail.Players = append(detail.Players, p)
		}
	}
	if detail.Players == nil {
		detail.Players = []FunSessionPlayer{}
	}

	mrows, err := db.DB.Query(
		`SELECT fm.id, fm.male_player_id, fm.female_player_id,
			COALESCE(ma.name,''), COALESCE(fa.name,''),
			fm.game1_score_male, fm.game1_score_female,
			fm.game2_score_male, fm.game2_score_female,
			fm.game3_score_male, fm.game3_score_female,
			fm.winner_id, COALESCE(fm.winner_team,''),
			fm.handicap_points, fm.played
		FROM fun_matches fm
		JOIN players ma ON ma.id = fm.male_player_id
		JOIN players fa ON fa.id = fm.female_player_id
		WHERE fm.session_id=$1 AND fm.deleted=false
		ORDER BY fm.id`, sessionID)
	if err == nil {
		defer mrows.Close()
		for mrows.Next() {
			var m FunMatchInfo
			mrows.Scan(&m.ID, &m.MalePlayerID, &m.FemalePlayerID,
				&m.MalePlayerName, &m.FemalePlayerName,
				&m.Game1ScoreMale, &m.Game1ScoreFemale,
				&m.Game2ScoreMale, &m.Game2ScoreFemale,
				&m.Game3ScoreMale, &m.Game3ScoreFemale,
				&m.WinnerID, &m.WinnerTeam, &m.HandicapPoints, &m.Played)
			detail.Matches = append(detail.Matches, m)
		}
	}
	if detail.Matches == nil {
		detail.Matches = []FunMatchInfo{}
	}

	// Load draws for all matches in this session
	drows, err := db.DB.Query(
		`SELECT id, match_id, player_id, card_type, card_value, COALESCE(card_detail,''), COALESCE(cancelled,false), created_at
		FROM fun_match_draws WHERE session_id=$1 ORDER BY id`, sessionID)
	if err == nil {
		defer drows.Close()
		drawsByMatch := map[int][]FunDrawRecord{}
		for drows.Next() {
			var d FunDrawRecord
			var mid int
			drows.Scan(&d.ID, &mid, &d.PlayerID, &d.CardType, &d.CardValue, &d.CardDetail, &d.Cancelled, &d.DrawnAt)
			drawsByMatch[mid] = append(drawsByMatch[mid], d)
		}
		for i := range detail.Matches {
			if draws, ok := drawsByMatch[detail.Matches[i].ID]; ok {
				detail.Matches[i].Draws = draws
			} else {
				detail.Matches[i].Draws = []FunDrawRecord{}
			}
		}
	}

	writeJSON(w, detail)
}

func UpdateFunSession(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/fun-sessions/")
	sessionID, _ := strconv.Atoi(path)

	var req struct{ Name string `json:"name"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	_, err := db.DB.Exec(`UPDATE fun_sessions SET name=$1 WHERE id=$2`, req.Name, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func DeleteFunSession(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/fun-sessions/")
	sessionID, _ := strconv.Atoi(path)
	_, err := db.DB.Exec(`UPDATE fun_sessions SET deleted=true WHERE id=$1`, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func DrawFunCard(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/fun-sessions/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	sessionID, _ := strconv.Atoi(parts[0])
	matchID, _ := strconv.Atoi(parts[2])

	var req struct{ PlayerID int `json:"player_id"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	var malePID, femalePID int
	var maleRating, femaleRating int
	err := db.DB.QueryRow(
		`SELECT fm.male_player_id, fm.female_player_id, COALESCE(pm.reference_rating, pm.current_rating), COALESCE(pf.reference_rating, pf.current_rating)
		FROM fun_matches fm
		JOIN players pm ON pm.id = fm.male_player_id
		JOIN players pf ON pf.id = fm.female_player_id
		WHERE fm.id=$1 AND fm.session_id=$2 AND fm.deleted=false`, matchID, sessionID,
	).Scan(&malePID, &femalePID, &maleRating, &femaleRating)
	if err != nil {
		http.Error(w, "match not found", http.StatusNotFound)
		return
	}

	diff := maleRating - femaleRating
	if diff < 0 {
		diff = -diff
	}
	if diff < 50 {
		http.Error(w, "分差不足50，无需抽卡", http.StatusBadRequest)
		return
	}

	higherPID := malePID
	if femaleRating > maleRating {
		higherPID = femalePID
	}
	if req.PlayerID != higherPID {
		http.Error(w, "只有高分选手才能抽卡", http.StatusBadRequest)
		return
	}

	// Random draw from 9 card types (no consumption, unlimited draws)
	c := cardTypes[rand.Intn(len(cardTypes))]

	// Cancel all previous draws for this match
	db.DB.Exec(`UPDATE fun_match_draws SET cancelled=true WHERE match_id=$1`, matchID)

	// Record the new draw
	var val interface{}
	if c.Value != nil { val = *c.Value } else { val = nil }
	var detail interface{}
	if c.Detail != "" { detail = c.Detail } else { detail = nil }
	db.DB.Exec(
		`INSERT INTO fun_match_draws (session_id, match_id, player_id, card_type, card_value, card_detail)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		sessionID, matchID, req.PlayerID, c.Type, val, detail,
	)

	// If handicap card, accumulate (add to existing)
	if c.Type == "handicap" && c.Value != nil {
		db.DB.Exec(`UPDATE fun_matches SET handicap_points = handicap_points + $1 WHERE id=$2`, *c.Value, matchID)
	}

	writeJSON(w, map[string]interface{}{
		"card_type":   c.Type,
		"card_value":  c.Value,
		"card_detail": c.Detail,
	})
}

func ScoreFunMatch(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/fun-sessions/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	sessionID, _ := strconv.Atoi(parts[0])
	matchID, _ := strconv.Atoi(parts[2])

	var req struct {
		Game1ScoreMale   int  `json:"game1_score_male"`
		Game1ScoreFemale int  `json:"game1_score_female"`
		Game2ScoreMale   int  `json:"game2_score_male"`
		Game2ScoreFemale int  `json:"game2_score_female"`
		Game3ScoreMale   *int `json:"game3_score_male"`
		Game3ScoreFemale *int `json:"game3_score_female"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	maleWins := 0
	femaleWins := 0
	if req.Game1ScoreMale > req.Game1ScoreFemale {
		maleWins++
	} else {
		femaleWins++
	}
	if req.Game2ScoreMale > req.Game2ScoreFemale {
		maleWins++
	} else {
		femaleWins++
	}

	needsGame3 := maleWins == 1 && femaleWins == 1
	if needsGame3 && req.Game3ScoreMale == nil {
		http.Error(w, "需要第三局比分", http.StatusBadRequest)
		return
	}
	if !needsGame3 && req.Game3ScoreMale != nil {
		http.Error(w, "不需要第三局", http.StatusBadRequest)
		return
	}

	var winnerID int
	var winnerTeam string
	if maleWins == 2 || (needsGame3 && *req.Game3ScoreMale > *req.Game3ScoreFemale) {
		db.DB.QueryRow(`SELECT male_player_id FROM fun_matches WHERE id=$1`, matchID).Scan(&winnerID)
		winnerTeam = "male"
	} else {
		db.DB.QueryRow(`SELECT female_player_id FROM fun_matches WHERE id=$1`, matchID).Scan(&winnerID)
		winnerTeam = "female"
	}

	var g3m, g3f interface{}
	if req.Game3ScoreMale != nil {
		g3m = *req.Game3ScoreMale
	} else {
		g3m = nil
	}
	if req.Game3ScoreFemale != nil {
		g3f = *req.Game3ScoreFemale
	} else {
		g3f = nil
	}

	// Calculate game wins and points for this match
	newMaleGames := maleWins
	newFemaleGames := femaleWins
	if needsGame3 {
		if *req.Game3ScoreMale > *req.Game3ScoreFemale { newMaleGames++ } else { newFemaleGames++ }
	}
	newMalePoints := req.Game1ScoreMale + req.Game2ScoreMale
	newFemalePoints := req.Game1ScoreFemale + req.Game2ScoreFemale
	if req.Game3ScoreMale != nil { newMalePoints += *req.Game3ScoreMale }
	if req.Game3ScoreFemale != nil { newFemalePoints += *req.Game3ScoreFemale }

	// Undo old stats if previously scored
	var wasPlayed bool
	var oldWinnerTeam sql.NullString
	var oldG1m, oldG1f, oldG2m, oldG2f sql.NullInt64
	var oldG3m, oldG3f sql.NullInt64
	db.DB.QueryRow(`SELECT played, winner_team, game1_score_male, game1_score_female, game2_score_male, game2_score_female, game3_score_male, game3_score_female FROM fun_matches WHERE id=$1`, matchID).
		Scan(&wasPlayed, &oldWinnerTeam, &oldG1m, &oldG1f, &oldG2m, &oldG2f, &oldG3m, &oldG3f)

	if wasPlayed && oldWinnerTeam.String != "" {
		if oldWinnerTeam.String == "male" {
			db.DB.Exec(`UPDATE fun_sessions SET male_wins = GREATEST(male_wins - 1, 0) WHERE id=$1`, sessionID)
		} else {
			db.DB.Exec(`UPDATE fun_sessions SET female_wins = GREATEST(female_wins - 1, 0) WHERE id=$1`, sessionID)
		}
		// Undo old game wins and points
		oldMG := 0; oldFG := 0
		if oldG1m.Valid && oldG1f.Valid {
			if oldG1m.Int64 > oldG1f.Int64 { oldMG++ } else { oldFG++ }
		}
		if oldG2m.Valid && oldG2f.Valid {
			if oldG2m.Int64 > oldG2f.Int64 { oldMG++ } else { oldFG++ }
		}
		if oldG3m.Valid && oldG3f.Valid {
			if oldG3m.Int64 > oldG3f.Int64 { oldMG++ } else { oldFG++ }
		}
		oldMP := int(oldG1m.Int64 + oldG2m.Int64 + oldG3m.Int64)
		oldFP := int(oldG1f.Int64 + oldG2f.Int64 + oldG3f.Int64)
		db.DB.Exec(`UPDATE fun_sessions SET male_game_wins = GREATEST(male_game_wins - $1, 0), female_game_wins = GREATEST(female_game_wins - $2, 0), male_points = GREATEST(male_points - $3, 0), female_points = GREATEST(female_points - $4, 0) WHERE id=$5`,
			oldMG, oldFG, oldMP, oldFP, sessionID)
	}

	_, err := db.DB.Exec(
		`UPDATE fun_matches SET
			game1_score_male=$1, game1_score_female=$2,
			game2_score_male=$3, game2_score_female=$4,
			game3_score_male=$5, game3_score_female=$6,
			winner_id=$7, winner_team=$8, played=true
		WHERE id=$9`,
		req.Game1ScoreMale, req.Game1ScoreFemale,
		req.Game2ScoreMale, req.Game2ScoreFemale,
		g3m, g3f,
		winnerID, winnerTeam, matchID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if winnerTeam == "male" {
		db.DB.Exec(`UPDATE fun_sessions SET male_wins = male_wins + 1 WHERE id=$1`, sessionID)
	} else {
		db.DB.Exec(`UPDATE fun_sessions SET female_wins = female_wins + 1 WHERE id=$1`, sessionID)
	}
	db.DB.Exec(`UPDATE fun_sessions SET male_game_wins = male_game_wins + $1, female_game_wins = female_game_wins + $2, male_points = male_points + $3, female_points = female_points + $4 WHERE id=$5`,
		newMaleGames, newFemaleGames, newMalePoints, newFemalePoints, sessionID)

	writeJSON(w, map[string]interface{}{
		"winner_id":   winnerID,
		"winner_team": winnerTeam,
	})
}

func CompleteFunSession(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/fun-sessions/")
	parts := strings.Split(path, "/")
	sessionID, _ := strconv.Atoi(parts[0])

	var unplayed int
	db.DB.QueryRow(
		`SELECT COUNT(*) FROM fun_matches WHERE session_id=$1 AND played=false AND deleted=false`, sessionID,
	).Scan(&unplayed)
	if unplayed > 0 {
		http.Error(w, "还有未录入的场次", http.StatusBadRequest)
		return
	}

	var maleWins, femaleWins int
	db.DB.QueryRow(`SELECT male_wins, female_wins FROM fun_sessions WHERE id=$1`, sessionID).Scan(&maleWins, &femaleWins)
	winningTeam := ""
	if maleWins > femaleWins {
		winningTeam = "male"
	} else if femaleWins > maleWins {
		winningTeam = "female"
	}

	_, err := db.DB.Exec(
		`UPDATE fun_sessions SET status='completed', winning_team=$1, completed_at=NOW() WHERE id=$2`,
		winningTeam, sessionID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"status": "ok", "winning_team": winningTeam})
}

func intPtr(v int) *int { return &v }
