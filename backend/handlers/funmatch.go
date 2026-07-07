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
	Mode           string `json:"mode"`
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
	Wins            int    `json:"wins"`
	Losses          int    `json:"losses"`
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
	Mode           string             `json:"mode"`
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
	TopPlayerID    int                `json:"top_player_id"`
	TopPlayerName  string             `json:"top_player_name"`
	CreatedAt      string             `json:"created_at"`
	Players        []FunSessionPlayer `json:"players"`
	Matches        []FunMatchInfo     `json:"matches"`
}

type pair struct{ m, f int }

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

// 7-card types for wheel_rr: generic spin/table cards (low-rated player specifies details)
var wheelCardTypes = []struct {
	Type   string
	Value  *int
	Detail string
}{
	{"handicap", intPtr(2), ""},
	{"handicap", intPtr(3), ""},
	{"handicap", intPtr(4), ""},
	{"handicap", intPtr(5), ""},
	{"spin", nil, ""},
	{"table", nil, ""},
	{"defense", nil, ""},
}

// --- handlers ---

func GetFunSessions(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT fs.id, fs.name, COALESCE(fs.mode,'gender'), fs.male_count, fs.female_count, fs.status,
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
		if err := rows.Scan(&s.ID, &s.Name, &s.Mode, &s.MaleCount, &s.FemaleCount, &s.Status,
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
		Mode            string `json:"mode"`
		MalePlayerIDs   []int  `json:"male_player_ids"`
		FemalePlayerIDs []int  `json:"female_player_ids"`
	}
	if req.Mode == "" { req.Mode = "gender" }
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
		`INSERT INTO fun_sessions (name, mode, male_count, female_count) VALUES ($1, $2, $3, $4) RETURNING id`,
		req.Name, req.Mode, len(req.MalePlayerIDs), len(req.FemalePlayerIDs),
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

	males := req.MalePlayerIDs
	females := req.FemalePlayerIDs
	n := len(males)
	m := len(females)

	var allPairs []pair

	if req.Mode == "pimple_rr" || req.Mode == "wheel_rr" || (n > 0 && m == 0) {
		// Single group round-robin using circle method for fair scheduling
		// Each round: every player plays at most once, even rest intervals
		allPairs = generateFunRoundRobin(males)
	} else {
		// Default: cross-team round-robin
		type rpair struct{ m, f int }
		var rounds [][]rpair
		if n >= m {
			for k := 0; k < n; k++ {
				var round []rpair
				for j := 0; j < m; j++ {
					round = append(round, rpair{males[(k+j)%n], females[j]})
				}
				rand.Shuffle(len(round), func(i, j int) { round[i], round[j] = round[j], round[i] })
				rounds = append(rounds, round)
			}
		} else {
			for k := 0; k < m; k++ {
				var round []rpair
				for j := 0; j < n; j++ {
					round = append(round, rpair{males[j], females[(k+j)%m]})
				}
				rand.Shuffle(len(round), func(i, j int) { round[i], round[j] = round[j], round[i] })
				rounds = append(rounds, round)
			}
		}
		for _, round := range rounds {
			for _, p := range round {
				allPairs = append(allPairs, pair{p.m, p.f})
			}
		}
	}

	for _, p := range allPairs {
		_, err = tx.Exec(`INSERT INTO fun_matches (session_id, male_player_id, female_player_id) VALUES ($1, $2, $3)`, sessionID, p.m, p.f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
		`SELECT id, name, COALESCE(mode,'gender'), male_count, female_count, status, COALESCE(winning_team,''), male_wins, female_wins,
			COALESCE(male_game_wins,0), COALESCE(female_game_wins,0), COALESCE(male_points,0), COALESCE(female_points,0), created_at
		FROM fun_sessions WHERE id=$1 AND deleted=false`, sessionID,
	).Scan(&detail.ID, &detail.Name, &detail.Mode, &detail.MaleCount, &detail.FemaleCount,
		&detail.Status, &detail.WinningTeam, &detail.MaleWins, &detail.FemaleWins,
		&detail.MaleGameWins, &detail.FemaleGameWins, &detail.MalePoints, &detail.FemalePoints, &detail.CreatedAt)
	if err != nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}

	prows, err := db.DB.Query(
		`SELECT p.id, p.name, p.current_rating, COALESCE(p.reference_rating,0), fsp.team,
			COALESCE(SUM(CASE WHEN fm.winner_id = p.id AND fm.played = true AND fm.deleted = false THEN 1 ELSE 0 END), 0) AS wins,
			COALESCE(SUM(CASE WHEN fm.winner_id IS NOT NULL AND fm.winner_id != p.id AND fm.played = true AND fm.deleted = false AND (fm.male_player_id = p.id OR fm.female_player_id = p.id) THEN 1 ELSE 0 END), 0) AS losses
		FROM fun_session_players fsp JOIN players p ON p.id = fsp.player_id
		LEFT JOIN fun_matches fm ON (fm.male_player_id = p.id OR fm.female_player_id = p.id) AND fm.session_id = $1
		WHERE fsp.session_id = $1
		GROUP BY p.id, p.name, p.current_rating, p.reference_rating, fsp.team
		ORDER BY wins DESC, p.reference_rating DESC, p.name`, sessionID)
	if err == nil {
		defer prows.Close()
		for prows.Next() {
			var p FunSessionPlayer
			prows.Scan(&p.ID, &p.Name, &p.CurrentRating, &p.ReferenceRating, &p.Team, &p.Wins, &p.Losses)
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

	// Pick card set based on mode: wheel_rr uses generic spin/table cards
	cards := cardTypes
	var sessionMode string
	db.DB.QueryRow(`SELECT COALESCE(mode,'gender') FROM fun_sessions WHERE id=$1`, sessionID).Scan(&sessionMode)
	if sessionMode == "wheel_rr" {
		cards = wheelCardTypes
	}

	// Random draw (no consumption, unlimited draws)
	c := cards[rand.Intn(len(cards))]

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

	// Check if this is a wheel_rr (individual) session — skip team stat updates if so
	var sessionMode string
	db.DB.QueryRow(`SELECT COALESCE(mode,'gender') FROM fun_sessions WHERE id=$1`, sessionID).Scan(&sessionMode)
	isWheel := sessionMode == "wheel_rr"

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

	// Transaction with row lock to prevent race conditions (double-click, concurrent scoring)
	tx, err := db.DB.Begin()
	if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
	defer tx.Rollback()

	// Lock the match row for update
	var wasPlayed bool
	var oldWinnerTeam sql.NullString
	var oldG1m, oldG1f, oldG2m, oldG2f sql.NullInt64
	var oldG3m, oldG3f sql.NullInt64
	tx.QueryRow(`SELECT played, winner_team, game1_score_male, game1_score_female, game2_score_male, game2_score_female, game3_score_male, game3_score_female FROM fun_matches WHERE id=$1 FOR UPDATE`, matchID).
		Scan(&wasPlayed, &oldWinnerTeam, &oldG1m, &oldG1f, &oldG2m, &oldG2f, &oldG3m, &oldG3f)

	// Undo old stats if previously scored (skip for wheel_rr — no team stats)
	if !isWheel && wasPlayed && oldWinnerTeam.String != "" {
		if oldWinnerTeam.String == "male" {
			tx.Exec(`UPDATE fun_sessions SET male_wins = GREATEST(male_wins - 1, 0) WHERE id=$1`, sessionID)
		} else {
			tx.Exec(`UPDATE fun_sessions SET female_wins = GREATEST(female_wins - 1, 0) WHERE id=$1`, sessionID)
		}
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
		tx.Exec(`UPDATE fun_sessions SET male_game_wins = GREATEST(male_game_wins - $1, 0), female_game_wins = GREATEST(female_game_wins - $2, 0), male_points = GREATEST(male_points - $3, 0), female_points = GREATEST(female_points - $4, 0) WHERE id=$5`,
			oldMG, oldFG, oldMP, oldFP, sessionID)
	}

	tx.Exec(
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

	// Team stats update (skip for wheel_rr — individual standings computed from matches)
	if !isWheel {
		if winnerTeam == "male" {
			tx.Exec(`UPDATE fun_sessions SET male_wins = male_wins + 1 WHERE id=$1`, sessionID)
		} else {
			tx.Exec(`UPDATE fun_sessions SET female_wins = female_wins + 1 WHERE id=$1`, sessionID)
		}
		tx.Exec(`UPDATE fun_sessions SET male_game_wins = male_game_wins + $1, female_game_wins = female_game_wins + $2, male_points = male_points + $3, female_points = female_points + $4 WHERE id=$5`,
			newMaleGames, newFemaleGames, newMalePoints, newFemalePoints, sessionID)
	}

	if err := tx.Commit(); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }

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

	var sessionMode string
	db.DB.QueryRow(`SELECT COALESCE(mode,'gender') FROM fun_sessions WHERE id=$1`, sessionID).Scan(&sessionMode)

	winningTeam := ""
	if sessionMode == "wheel_rr" {
		winningTeam = "individual"
	} else {
		var maleWins, femaleWins int
		db.DB.QueryRow(`SELECT male_wins, female_wins FROM fun_sessions WHERE id=$1`, sessionID).Scan(&maleWins, &femaleWins)
		if maleWins > femaleWins {
			winningTeam = "male"
		} else if femaleWins > maleWins {
			winningTeam = "female"
		}
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

// generateFunRoundRobin creates a round-robin schedule using the circle method.
// Ensures fair scheduling: each player plays at most once per round, even rest intervals.
// Returns pairs as (male_player_id, female_player_id) for fun_matches insertion.
func generateFunRoundRobin(playerIDs []int) []pair {
	n := len(playerIDs)
	if n < 2 {
		return nil
	}

	// Use circle method: fix first player, rotate the rest
	ids := make([]int, n)
	copy(ids, playerIDs)

	// If odd, add a dummy bye spot (0)
	hasBye := n%2 != 0
	if hasBye {
		ids = append(ids, 0)
		n++
	}

	var result []pair
	half := n / 2

	for r := 0; r < n-1; r++ {
		for i := 0; i < half; i++ {
			a := ids[i]
			b := ids[n-1-i]

			if a == 0 || b == 0 {
				continue // bye
			}

			// Alternate home/away for variety
			if i%2 == 1 {
				a, b = b, a
			}
			result = append(result, pair{a, b})
		}

		// Rotate: fix first element, shift the rest right
		last := ids[n-1]
		for i := n - 1; i > 1; i-- {
			ids[i] = ids[i-1]
		}
		ids[1] = last
	}

	return result
}
