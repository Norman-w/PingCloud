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

type TournamentSummary struct {
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	GroupCount         int    `json:"group_count"`
	TeamsPerGroup      int    `json:"teams_per_group"`
	PlayersPerTeam     int    `json:"players_per_team"`
	MaxParticipants    int    `json:"max_participants"`
	SeedEnabled        bool   `json:"seed_enabled"`
	SeedCount          int    `json:"seed_count"`
	Status             string `json:"status"`
	Phase              string `json:"phase"`
	ConfirmedCount     int    `json:"confirmed_count"`
	WaitlistedCount    int    `json:"waitlisted_count"`
	RegistrationDeadline string `json:"registration_deadline"`
	CreatedAt          string `json:"created_at"`
}

type TournamentTeam struct {
	ID           int                   `json:"id"`
	TournamentID int                   `json:"tournament_id"`
	GroupName    string                `json:"group_name"`
	TeamIndex    int                   `json:"team_index"`
	TeamName     string                `json:"team_name"`
	KnockoutSeed *int                  `json:"knockout_seed"`
	GroupRank    *int                  `json:"group_rank"`
	GroupWins    int                   `json:"group_wins"`
	GroupLosses  int                   `json:"group_losses"`
	GroupPoints  int                   `json:"group_points"`
	Players      []TournamentTeamPlayer `json:"players"`
}

type TournamentTeamPlayer struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	CurrentRating  int    `json:"current_rating"`
	ReferenceRating int   `json:"reference_rating"`
	Role           string `json:"role"`
	IsSeed         bool   `json:"is_seed"`
}

type TournamentRegistration struct {
	ID           int    `json:"id"`
	PlayerID     int    `json:"player_id"`
	PlayerName   string `json:"player_name"`
	Status       string `json:"status"`
	WaitlistPos  *int   `json:"waitlist_pos"`
	RegisteredAt string `json:"registered_at"`
}

type TournamentTeamMatch struct {
	ID           int                `json:"id"`
	TournamentID int                `json:"tournament_id"`
	Phase        string             `json:"phase"`
	Round        int                `json:"round"`
	GroupName    string             `json:"group_name"`
	TeamAID      int                `json:"team_a_id"`
	TeamBID      int                `json:"team_b_id"`
	TeamAName    string             `json:"team_a_name"`
	TeamBName    string             `json:"team_b_name"`
	TeamAWins    int                `json:"team_a_wins"`
	TeamBWins    int                `json:"team_b_wins"`
	WinnerTeamID *int               `json:"winner_team_id"`
	Played       bool               `json:"played"`
	Matches      []TournamentMatch  `json:"matches"`
	Cards        []TournamentCard   `json:"cards"`
}

type TournamentMatch struct {
	ID             int    `json:"id"`
	TeamMatchID    int    `json:"team_match_id"`
	Phase          string `json:"phase"`
	Round          int    `json:"round"`
	GroupName      string `json:"group_name"`
	TeamAID        int    `json:"team_a_id"`
	TeamBID        int    `json:"team_b_id"`
	MatchOrder     int    `json:"match_order"`
	MatchType      string `json:"match_type"`
	PlayerAID      int    `json:"player_a_id"`
	PlayerBID      int    `json:"player_b_id"`
	PlayerAName    string `json:"player_a_name"`
	PlayerBName    string `json:"player_b_name"`
	PlayerA2ID     *int   `json:"player_a2_id"`
	PlayerB2ID     *int   `json:"player_b2_id"`
	PlayerA2Name   string `json:"player_a2_name"`
	PlayerB2Name   string `json:"player_b2_name"`
	Game1ScoreA    *int   `json:"game1_score_a"`
	Game1ScoreB    *int   `json:"game1_score_b"`
	Game2ScoreA    *int   `json:"game2_score_a"`
	Game2ScoreB    *int   `json:"game2_score_b"`
	Game3ScoreA    *int   `json:"game3_score_a"`
	Game3ScoreB    *int   `json:"game3_score_b"`
	WinnerID       *int   `json:"winner_id"`
	WinnerTeamID   *int   `json:"winner_team_id"`
	Played         bool   `json:"played"`
	Forfeit        bool   `json:"forfeit"`
}

type TournamentCard struct {
	ID           int    `json:"id"`
	TeamMatchID  int    `json:"team_match_id"`
	TeamID       int    `json:"team_id"`
	CardType     string `json:"card_type"`
	DrawnAt      string `json:"drawn_at"`
}

type TournamentDetail struct {
	ID                   int                    `json:"id"`
	Name                 string                 `json:"name"`
	GroupCount           int                    `json:"group_count"`
	TeamsPerGroup        int                    `json:"teams_per_group"`
	PlayersPerTeam       int                    `json:"players_per_team"`
	MaxParticipants      int                    `json:"max_participants"`
	SeedEnabled          bool                   `json:"seed_enabled"`
	SeedCount            int                    `json:"seed_count"`
	RegistrationDeadline string                 `json:"registration_deadline"`
	Status               string                 `json:"status"`
	Phase                string                 `json:"phase"`
	CreatedAt            string                 `json:"created_at"`
	CompletedAt          string                 `json:"completed_at"`
	Teams                []TournamentTeam       `json:"teams"`
	Registrations        []TournamentRegistration `json:"registrations"`
	TeamMatches          []TournamentTeamMatch  `json:"team_matches"`
}

type rolePlayer struct {
	PlayerID int
	Role     string
}

var tournamentCardTypes = []struct {
	Type   string
	Detail string
}{
	{"edge_double", "擦边翻倍卡"},
	{"net_deduction", "擦网扣分卡"},
}

// --- CRUD handlers ---

func CreateTournament(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name                 string `json:"name"`
		GroupCount           int    `json:"group_count"`
		TeamsPerGroup        int    `json:"teams_per_group"`
		PlayersPerTeam       int    `json:"players_per_team"`
		MaxParticipants      int    `json:"max_participants"`
		SeedEnabled          bool   `json:"seed_enabled"`
		SeedCount            int    `json:"seed_count"`
		RegistrationDeadline string `json:"registration_deadline"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		req.Name = "锦标赛"
	}
	if req.GroupCount < 1 {
		req.GroupCount = 2
	}
	if req.TeamsPerGroup < 2 {
		req.TeamsPerGroup = 3
	}
	if req.PlayersPerTeam < 2 {
		req.PlayersPerTeam = 3
	}
	if req.MaxParticipants < 1 {
		req.MaxParticipants = req.GroupCount * req.TeamsPerGroup * req.PlayersPerTeam
	}

	var regDeadline interface{}
	if req.RegistrationDeadline != "" {
		regDeadline = req.RegistrationDeadline
	} else {
		regDeadline = nil
	}

	var id int
	err := db.DB.QueryRow(
		`INSERT INTO tournaments (name, group_count, teams_per_group, players_per_team, max_participants, seed_enabled, seed_count, registration_deadline)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		req.Name, req.GroupCount, req.TeamsPerGroup, req.PlayersPerTeam, req.MaxParticipants, req.SeedEnabled, req.SeedCount, regDeadline,
	).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]int{"id": id})
}

func ListTournaments(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT t.id, t.name, t.group_count, t.teams_per_group, t.players_per_team,
			t.max_participants, t.seed_enabled, t.seed_count, t.status, t.phase,
			COALESCE(to_char(t.registration_deadline, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), ''),
			COALESCE(t.created_at::text, ''),
			COUNT(CASE WHEN tr.status = 'confirmed' THEN 1 END) AS confirmed_count,
			COUNT(CASE WHEN tr.status = 'waitlisted' THEN 1 END) AS waitlisted_count
		FROM tournaments t
		LEFT JOIN tournament_registrations tr ON tr.tournament_id = t.id
		WHERE t.deleted = false
		GROUP BY t.id
		ORDER BY t.created_at DESC LIMIT 20
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var result []TournamentSummary
	for rows.Next() {
		var s TournamentSummary
		if err := rows.Scan(&s.ID, &s.Name, &s.GroupCount, &s.TeamsPerGroup, &s.PlayersPerTeam,
			&s.MaxParticipants, &s.SeedEnabled, &s.SeedCount, &s.Status, &s.Phase,
			&s.RegistrationDeadline, &s.CreatedAt, &s.ConfirmedCount, &s.WaitlistedCount); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		result = append(result, s)
	}
	if result == nil {
		result = []TournamentSummary{}
	}
	writeJSON(w, result)
}

func GetTournament(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tournaments/")
	tournamentID, _ := strconv.Atoi(path)

	var detail TournamentDetail
	var completedAt sql.NullString
	err := db.DB.QueryRow(
		`SELECT id, name, group_count, teams_per_group, players_per_team, max_participants,
			seed_enabled, seed_count,
			COALESCE(to_char(registration_deadline, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), ''),
			status, phase, COALESCE(created_at::text, ''),
			CASE WHEN completed_at IS NOT NULL THEN completed_at::text ELSE '' END
		FROM tournaments WHERE id=$1 AND deleted=false`, tournamentID,
	).Scan(&detail.ID, &detail.Name, &detail.GroupCount, &detail.TeamsPerGroup, &detail.PlayersPerTeam,
		&detail.MaxParticipants, &detail.SeedEnabled, &detail.SeedCount,
		&detail.RegistrationDeadline, &detail.Status, &detail.Phase, &detail.CreatedAt, &completedAt)
	if err != nil {
		http.Error(w, "tournament not found", http.StatusNotFound)
		return
	}
	detail.CompletedAt = completedAt.String

	// Load teams with players
	detail.Teams = loadTournamentTeams(tournamentID)

	// Load registrations
	detail.Registrations = loadTournamentRegistrations(tournamentID)

	// Load team matches with sub-matches and cards
	detail.TeamMatches = loadTournamentTeamMatches(tournamentID)

	writeJSON(w, detail)
}

func loadTournamentTeams(tournamentID int) []TournamentTeam {
	rows, err := db.DB.Query(
		`SELECT id, tournament_id, group_name, team_index, team_name,
			knockout_seed, group_rank, group_wins, group_losses, group_points
		FROM tournament_teams WHERE tournament_id=$1 ORDER BY group_name, team_index`, tournamentID)
	if err != nil {
		return []TournamentTeam{}
	}
	defer rows.Close()

	var teams []TournamentTeam
	for rows.Next() {
		var t TournamentTeam
		rows.Scan(&t.ID, &t.TournamentID, &t.GroupName, &t.TeamIndex, &t.TeamName,
			&t.KnockoutSeed, &t.GroupRank, &t.GroupWins, &t.GroupLosses, &t.GroupPoints)
		teams = append(teams, t)
	}
	if teams == nil {
		return []TournamentTeam{}
	}

	// Load players for each team
	for i := range teams {
		prows, err := db.DB.Query(
			`SELECT p.id, p.name, p.current_rating, COALESCE(p.reference_rating, 0), ttp.role, ttp.is_seed
			FROM tournament_team_players ttp
			JOIN players p ON p.id = ttp.player_id
			WHERE ttp.team_id=$1 ORDER BY ttp.role`, teams[i].ID)
		if err != nil {
			continue
		}
		for prows.Next() {
			var p TournamentTeamPlayer
			prows.Scan(&p.ID, &p.Name, &p.CurrentRating, &p.ReferenceRating, &p.Role, &p.IsSeed)
			teams[i].Players = append(teams[i].Players, p)
		}
		prows.Close()
		if teams[i].Players == nil {
			teams[i].Players = []TournamentTeamPlayer{}
		}
	}
	return teams
}

func loadTournamentRegistrations(tournamentID int) []TournamentRegistration {
	rows, err := db.DB.Query(
		`SELECT tr.id, tr.player_id, p.name, tr.status, tr.waitlist_pos,
			COALESCE(tr.registered_at::text, '')
		FROM tournament_registrations tr
		JOIN players p ON p.id = tr.player_id
		WHERE tr.tournament_id=$1
		ORDER BY tr.status, tr.waitlist_pos NULLS FIRST, tr.registered_at`, tournamentID)
	if err != nil {
		return []TournamentRegistration{}
	}
	defer rows.Close()

	var result []TournamentRegistration
	for rows.Next() {
		var reg TournamentRegistration
		rows.Scan(&reg.ID, &reg.PlayerID, &reg.PlayerName, &reg.Status, &reg.WaitlistPos, &reg.RegisteredAt)
		result = append(result, reg)
	}
	if result == nil {
		result = []TournamentRegistration{}
	}
	return result
}

func loadTournamentTeamMatches(tournamentID int) []TournamentTeamMatch {
	tmRows, err := db.DB.Query(
		`SELECT id, tournament_id, phase, round, COALESCE(group_name,''),
			team_a_id, team_b_id, team_a_wins, team_b_wins, winner_team_id, played
		FROM tournament_team_matches WHERE tournament_id=$1
		ORDER BY CASE phase WHEN 'group' THEN 1 WHEN 'semifinal' THEN 2 WHEN 'final' THEN 3 END, round, id`, tournamentID)
	if err != nil {
		return []TournamentTeamMatch{}
	}
	defer tmRows.Close()

	var result []TournamentTeamMatch
	for tmRows.Next() {
		var tm TournamentTeamMatch
		tmRows.Scan(&tm.ID, &tm.TournamentID, &tm.Phase, &tm.Round, &tm.GroupName,
			&tm.TeamAID, &tm.TeamBID, &tm.TeamAWins, &tm.TeamBWins, &tm.WinnerTeamID, &tm.Played)

		// Get team names
		db.DB.QueryRow(`SELECT team_name FROM tournament_teams WHERE id=$1`, tm.TeamAID).Scan(&tm.TeamAName)
		db.DB.QueryRow(`SELECT team_name FROM tournament_teams WHERE id=$1`, tm.TeamBID).Scan(&tm.TeamBName)

		// Load sub-matches
		tm.Matches = loadTournamentMatches(tm.ID)
		// Load cards
		tm.Cards = loadTournamentCards(tm.ID)

		result = append(result, tm)
	}
	if result == nil {
		result = []TournamentTeamMatch{}
	}
	return result
}

func loadTournamentMatches(teamMatchID int) []TournamentMatch {
	rows, err := db.DB.Query(
		`SELECT tm.id, tm.team_match_id, tm.phase, tm.round, COALESCE(tm.group_name,''),
			tm.team_a_id, tm.team_b_id, tm.match_order, tm.match_type,
			tm.player_a_id, tm.player_b_id, tm.player_a2_id, tm.player_b2_id,
			COALESCE(pa.name,''), COALESCE(pb.name,''),
			COALESCE(pa2.name,''), COALESCE(pb2.name,''),
			tm.game1_score_a, tm.game1_score_b,
			tm.game2_score_a, tm.game2_score_b,
			tm.game3_score_a, tm.game3_score_b,
			tm.winner_id, tm.winner_team_id, tm.played, tm.forfeit
		FROM tournament_matches tm
		JOIN players pa ON pa.id = tm.player_a_id
		JOIN players pb ON pb.id = tm.player_b_id
		LEFT JOIN players pa2 ON pa2.id = tm.player_a2_id
		LEFT JOIN players pb2 ON pb2.id = tm.player_b2_id
		WHERE tm.team_match_id=$1
		ORDER BY tm.match_order`, teamMatchID)
	if err != nil {
		return []TournamentMatch{}
	}
	defer rows.Close()

	var result []TournamentMatch
	for rows.Next() {
		var m TournamentMatch
		rows.Scan(&m.ID, &m.TeamMatchID, &m.Phase, &m.Round, &m.GroupName,
			&m.TeamAID, &m.TeamBID, &m.MatchOrder, &m.MatchType,
			&m.PlayerAID, &m.PlayerBID, &m.PlayerA2ID, &m.PlayerB2ID,
			&m.PlayerAName, &m.PlayerBName, &m.PlayerA2Name, &m.PlayerB2Name,
			&m.Game1ScoreA, &m.Game1ScoreB, &m.Game2ScoreA, &m.Game2ScoreB,
			&m.Game3ScoreA, &m.Game3ScoreB, &m.WinnerID, &m.WinnerTeamID, &m.Played, &m.Forfeit)
		result = append(result, m)
	}
	if result == nil {
		result = []TournamentMatch{}
	}
	return result
}

func loadTournamentCards(teamMatchID int) []TournamentCard {
	rows, err := db.DB.Query(
		`SELECT id, team_match_id, team_id, card_type, COALESCE(drawn_at::text, '')
		FROM tournament_cards WHERE team_match_id=$1 ORDER BY id`, teamMatchID)
	if err != nil {
		return []TournamentCard{}
	}
	defer rows.Close()

	var result []TournamentCard
	for rows.Next() {
		var c TournamentCard
		rows.Scan(&c.ID, &c.TeamMatchID, &c.TeamID, &c.CardType, &c.DrawnAt)
		result = append(result, c)
	}
	if result == nil {
		result = []TournamentCard{}
	}
	return result
}

func UpdateTournament(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tournaments/")
	tournamentID, _ := strconv.Atoi(path)

	var req struct {
		Name                 string `json:"name"`
		MaxParticipants      *int   `json:"max_participants"`
		SeedEnabled          *bool  `json:"seed_enabled"`
		SeedCount            *int   `json:"seed_count"`
		RegistrationDeadline string `json:"registration_deadline"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if req.Name != "" {
		db.DB.Exec(`UPDATE tournaments SET name=$1 WHERE id=$2`, req.Name, tournamentID)
	}
	if req.MaxParticipants != nil {
		db.DB.Exec(`UPDATE tournaments SET max_participants=$1 WHERE id=$2`, *req.MaxParticipants, tournamentID)
	}
	if req.SeedEnabled != nil {
		db.DB.Exec(`UPDATE tournaments SET seed_enabled=$1 WHERE id=$2`, *req.SeedEnabled, tournamentID)
	}
	if req.SeedCount != nil {
		db.DB.Exec(`UPDATE tournaments SET seed_count=$1 WHERE id=$2`, *req.SeedCount, tournamentID)
	}
	if req.RegistrationDeadline != "" {
		db.DB.Exec(`UPDATE tournaments SET registration_deadline=$1 WHERE id=$2`, req.RegistrationDeadline, tournamentID)
	}

	writeJSON(w, map[string]string{"status": "ok"})
}

func DeleteTournament(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tournaments/")
	tournamentID, _ := strconv.Atoi(path)

	_, err := db.DB.Exec(`UPDATE tournaments SET deleted=true WHERE id=$1`, tournamentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// --- Registration ---

func RegisterForTournament(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tournaments/")
	tournamentID, _ := strconv.Atoi(strings.Split(path, "/")[0])

	var req struct {
		PlayerID int `json:"player_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Check tournament status
	var status string
	tx.QueryRow(`SELECT status FROM tournaments WHERE id=$1 AND deleted=false`, tournamentID).Scan(&status)
	if status != "registration" {
		http.Error(w, "报名已截止", http.StatusBadRequest)
		return
	}

	// Check already registered
	var existingID int
	var existingStatus string
	err = tx.QueryRow(
		`SELECT id, status FROM tournament_registrations WHERE tournament_id=$1 AND player_id=$2`,
		tournamentID, req.PlayerID,
	).Scan(&existingID, &existingStatus)
	if err == nil {
		if existingStatus == "withdrawn" || existingStatus == "cancelled" {
			// Re-register at end of waitlist
			tx.Exec(`UPDATE tournament_registrations SET status='waitlisted', waitlist_pos=(SELECT COALESCE(MAX(waitlist_pos),0)+1 FROM tournament_registrations WHERE tournament_id=$1 AND status='waitlisted'), cancelled_at=NULL WHERE id=$2`, tournamentID, existingID)
			tx.Commit()
			writeJSON(w, map[string]interface{}{"status": "waitlisted", "message": "已重新加入候补"})
			return
		}
		http.Error(w, "已报名", http.StatusBadRequest)
		return
	}

	var maxParticipants int
	tx.QueryRow(`SELECT max_participants FROM tournaments WHERE id=$1`, tournamentID).Scan(&maxParticipants)

	var confirmedCount int
	tx.QueryRow(`SELECT COUNT(*) FROM tournament_registrations WHERE tournament_id=$1 AND status='confirmed'`, tournamentID).Scan(&confirmedCount)

	if confirmedCount < maxParticipants {
		_, err = tx.Exec(
			`INSERT INTO tournament_registrations (tournament_id, player_id, status) VALUES ($1, $2, 'confirmed')`,
			tournamentID, req.PlayerID,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		tx.Commit()
		writeJSON(w, map[string]interface{}{"status": "confirmed"})
	} else {
		var maxPos sql.NullInt64
		tx.QueryRow(`SELECT MAX(waitlist_pos) FROM tournament_registrations WHERE tournament_id=$1 AND status='waitlisted'`, tournamentID).Scan(&maxPos)
		nextPos := 1
		if maxPos.Valid {
			nextPos = int(maxPos.Int64) + 1
		}
		_, err = tx.Exec(
			`INSERT INTO tournament_registrations (tournament_id, player_id, status, waitlist_pos) VALUES ($1, $2, 'waitlisted', $3)`,
			tournamentID, req.PlayerID, nextPos,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		tx.Commit()
		writeJSON(w, map[string]interface{}{"status": "waitlisted", "waitlist_position": nextPos})
	}
}

func CancelRegistration(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tournaments/")
	tournamentID, _ := strconv.Atoi(strings.Split(path, "/")[0])

	var req struct {
		PlayerID int `json:"player_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var regID int
	var regStatus string
	var waitlistPos sql.NullInt64
	err = tx.QueryRow(
		`SELECT id, status, waitlist_pos FROM tournament_registrations WHERE tournament_id=$1 AND player_id=$2`,
		tournamentID, req.PlayerID,
	).Scan(&regID, &regStatus, &waitlistPos)
	if err != nil {
		http.Error(w, "未报名", http.StatusNotFound)
		return
	}

	if regStatus == "confirmed" {
		// Cancel and promote first waitlisted
		tx.Exec(`UPDATE tournament_registrations SET status='cancelled', cancelled_at=NOW() WHERE id=$1`, regID)

		var nextID int
		err = tx.QueryRow(
			`SELECT id FROM tournament_registrations WHERE tournament_id=$1 AND status='waitlisted' ORDER BY waitlist_pos LIMIT 1`,
			tournamentID,
		).Scan(&nextID)
		if err == nil {
			tx.Exec(`UPDATE tournament_registrations SET status='confirmed', waitlist_pos=NULL WHERE id=$1`, nextID)
			// Shift remaining waitlist positions
			tx.Exec(`UPDATE tournament_registrations SET waitlist_pos = waitlist_pos - 1 WHERE tournament_id=$1 AND status='waitlisted' AND waitlist_pos > 1`, tournamentID)
		}
	} else if regStatus == "waitlisted" {
		cancelledPos := int(waitlistPos.Int64)
		tx.Exec(`UPDATE tournament_registrations SET status='withdrawn', cancelled_at=NOW() WHERE id=$1`, regID)
		// Close gap
		tx.Exec(`UPDATE tournament_registrations SET waitlist_pos = waitlist_pos - 1 WHERE tournament_id=$1 AND status='waitlisted' AND waitlist_pos > $2`, tournamentID, cancelledPos)
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func ListRegistrations(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tournaments/")
	parts := strings.Split(path, "/")
	tournamentID, _ := strconv.Atoi(parts[0])

	result := loadTournamentRegistrations(tournamentID)
	writeJSON(w, result)
}

// --- Team formation ---

func DrawTeams(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tournaments/")
	tournamentID, _ := strconv.Atoi(strings.Split(path, "/")[0])

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var groupCount, teamsPerGroup, playersPerTeam, seedCount int
	var seedEnabled bool
	err = tx.QueryRow(
		`SELECT group_count, teams_per_group, players_per_team, seed_enabled, seed_count
		FROM tournaments WHERE id=$1 AND deleted=false`, tournamentID,
	).Scan(&groupCount, &teamsPerGroup, &playersPerTeam, &seedEnabled, &seedCount)
	if err != nil {
		http.Error(w, "tournament not found", http.StatusNotFound)
		return
	}

	// Delete existing teams if re-drawing
	tx.Exec(`DELETE FROM tournament_team_players WHERE tournament_id=$1`, tournamentID)
	tx.Exec(`DELETE FROM tournament_team_matches WHERE tournament_id=$1`, tournamentID)
	tx.Exec(`DELETE FROM tournament_cards WHERE tournament_id=$1`, tournamentID)
	tx.Exec(`DELETE FROM tournament_teams WHERE tournament_id=$1`, tournamentID)

	// Get confirmed players ordered by rating
	rows, err := tx.Query(
		`SELECT p.id, p.name, COALESCE(p.reference_rating, p.current_rating) as rating
		FROM tournament_registrations tr
		JOIN players p ON p.id = tr.player_id
		WHERE tr.tournament_id=$1 AND tr.status='confirmed'
		ORDER BY rating DESC`, tournamentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type playerInfo struct {
		ID     int
		Name   string
		Rating int
	}
	var players []playerInfo
	for rows.Next() {
		var p playerInfo
		rows.Scan(&p.ID, &p.Name, &p.Rating)
		players = append(players, p)
	}

	totalTeams := groupCount * teamsPerGroup
	totalSlots := totalTeams * playersPerTeam

	if len(players) < totalSlots {
		http.Error(w, "报名人数不足", http.StatusBadRequest)
		return
	}

	// Create teams
	teamIDs := make([]int, 0, totalTeams)
	groupLetters := "ABCDEFGH"
	for g := 0; g < groupCount; g++ {
		groupName := string(groupLetters[g])
		for t := 0; t < teamsPerGroup; t++ {
			teamName := groupName + strconv.Itoa(t+1)
			var tid int
			tx.QueryRow(
				`INSERT INTO tournament_teams (tournament_id, group_name, team_index, team_name) VALUES ($1, $2, $3, $4) RETURNING id`,
				tournamentID, groupName, t, teamName,
			).Scan(&tid)
			teamIDs = append(teamIDs, tid)
		}
	}

	roles := []string{"A", "B", "C"}

	seeds := []playerInfo{}
	remaining := players
	if seedEnabled && seedCount > 0 {
		if seedCount > len(players) {
			seedCount = len(players)
		}
		seeds = players[:seedCount]
		remaining = players[seedCount:]
	}

	// Assign seeds: snake draft across all teams
	// Seed 1 -> team 0 (A1), Seed 2 -> team 1 (B1), Seed 3 -> team 2 (B2), Seed 4 -> team 3 (A2), etc.
	seedAssignments := make(map[int]int) // playerID -> teamID
	for i, seed := range seeds {
		teamIdx := i % totalTeams
		seedAssignments[seed.ID] = teamIDs[teamIdx]
	}

	// Build slot map: teamID -> remaining slots count
	teamSlots := make(map[int]int)
	for _, tid := range teamIDs {
		teamSlots[tid] = playersPerTeam
	}

	// Place seeds first
	for playerID, teamID := range seedAssignments {
		roleIdx := playersPerTeam - teamSlots[teamID]
		if roleIdx < len(roles) {
			role := roles[roleIdx]
			tx.Exec(
				`INSERT INTO tournament_team_players (tournament_id, team_id, player_id, role, is_seed) VALUES ($1, $2, $3, $4, true)`,
				tournamentID, teamID, playerID, role,
			)
			teamSlots[teamID]--
		}
	}

	// Shuffle remaining players
	rand.Shuffle(len(remaining), func(i, j int) { remaining[i], remaining[j] = remaining[j], remaining[i] })

	// Fill remaining slots round-robin
	playerIdx := 0
	for playerIdx < len(remaining) {
		placed := false
		for _, tid := range teamIDs {
			if playerIdx >= len(remaining) {
				break
			}
			if teamSlots[tid] > 0 {
				roleIdx := playersPerTeam - teamSlots[tid]
				role := roles[roleIdx]
				tx.Exec(
					`INSERT INTO tournament_team_players (tournament_id, team_id, player_id, role, is_seed) VALUES ($1, $2, $3, $4, false)`,
					tournamentID, tid, remaining[playerIdx].ID, role,
				)
				teamSlots[tid]--
				playerIdx++
				placed = true
			}
		}
		if !placed {
			break
		}
	}

	// Update tournament status
	tx.Exec(`UPDATE tournaments SET status='group_stage', phase='group' WHERE id=$1`, tournamentID)

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return teams
	teams := getTournamentTeams(tournamentID)
	writeJSON(w, map[string]interface{}{"teams": teams})
}

func getTournamentTeams(tournamentID int) []TournamentTeam {
	return loadTournamentTeams(tournamentID)
}

// --- Match generation ---

func GenerateGroupMatches(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tournaments/")
	tournamentID, _ := strconv.Atoi(strings.Split(path, "/")[0])

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var groupCount int
	tx.QueryRow(`SELECT group_count FROM tournaments WHERE id=$1`, tournamentID).Scan(&groupCount)

	// Delete existing matches if re-generating
	tx.Exec(`DELETE FROM tournament_cards WHERE tournament_id=$1`, tournamentID)
	tx.Exec(`DELETE FROM tournament_matches WHERE tournament_id=$1`, tournamentID)
	tx.Exec(`DELETE FROM tournament_team_matches WHERE tournament_id=$1`, tournamentID)

	groupLetters := "ABCDEFGH"
	round := 1

	for g := 0; g < groupCount; g++ {
		groupName := string(groupLetters[g])

		// Get teams in this group
		rows, err := tx.Query(
			`SELECT id FROM tournament_teams WHERE tournament_id=$1 AND group_name=$2 ORDER BY team_index`,
			tournamentID, groupName)
		if err != nil {
			continue
		}

		var teamIDs []int
		for rows.Next() {
			var tid int
			rows.Scan(&tid)
			teamIDs = append(teamIDs, tid)
		}
		rows.Close()

		// Generate round-robin pairs using circle method
		pairs := generateRoundRobinPairsInt(teamIDs)

		for _, pair := range pairs {
			// Create team match
			var tmID int
			tx.QueryRow(
				`INSERT INTO tournament_team_matches (tournament_id, phase, round, group_name, team_a_id, team_b_id)
				VALUES ($1, 'group', $2, $3, $4, $5) RETURNING id`,
				tournamentID, round, groupName, pair[0], pair[1],
			).Scan(&tmID)

			// Generate 5 sub-matches
			createSubMatches(tx, tmID, tournamentID, "group", round, groupName, pair[0], pair[1])
			round++
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"status": "ok"})
}

func generateRoundRobinPairsInt(ids []int) [][2]int {
	n := len(ids)
	if n < 2 {
		return nil
	}

	// Circle method
	items := make([]int, n)
	copy(items, ids)

	hasBye := n%2 != 0
	if hasBye {
		items = append(items, -1)
		n++
	}

	var result [][2]int
	half := n / 2

	for r := 0; r < n-1; r++ {
		for i := 0; i < half; i++ {
			a := items[i]
			b := items[n-1-i]
			if a == -1 || b == -1 {
				continue
			}
			if i%2 == 1 {
				a, b = b, a
			}
			result = append(result, [2]int{a, b})
		}
		// Rotate
		last := items[n-1]
		for i := n - 1; i > 1; i-- {
			items[i] = items[i-1]
		}
		items[1] = last
	}

	return result
}

func createSubMatches(tx *sql.Tx, teamMatchID, tournamentID int, phase string, round int, groupName string, teamAID, teamBID int) {
	// Get team players by role
	teamAPlayers := getTeamRolePlayers(tx, teamAID)
	teamBPlayers := getTeamRolePlayers(tx, teamBID)

	getPlayer := func(players []rolePlayer, role string) int {
		for _, p := range players {
			if p.Role == role {
				return p.PlayerID
			}
		}
		return 0
	}

	aA := getPlayer(teamAPlayers, "A")
	aB := getPlayer(teamAPlayers, "B")
	aC := getPlayer(teamAPlayers, "C")
	bA := getPlayer(teamBPlayers, "A")
	bB := getPlayer(teamBPlayers, "B")
	bC := getPlayer(teamBPlayers, "C")

	// 5-match order
	matchDefs := []struct {
		order     int
		matchType string
		a1, a2    int
		b1, b2    int
	}{
		{1, "singles", aA, 0, bA, 0},
		{2, "doubles", aB, aC, bB, bC},
		{3, "singles", aC, 0, bC, 0},
		{4, "doubles", aA, aB, bA, bB},
		{5, "singles", aB, 0, bB, 0},
	}

	for _, def := range matchDefs {
		var a2id, b2id interface{}
		if def.a2 != 0 {
			a2id = def.a2
		} else {
			a2id = nil
		}
		if def.b2 != 0 {
			b2id = def.b2
		} else {
			b2id = nil
		}

		tx.Exec(
			`INSERT INTO tournament_matches (tournament_id, team_match_id, phase, round, group_name,
				team_a_id, team_b_id, match_order, match_type,
				player_a_id, player_b_id, player_a2_id, player_b2_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
			tournamentID, teamMatchID, phase, round, groupName,
			teamAID, teamBID, def.order, def.matchType,
			def.a1, def.b1, a2id, b2id,
		)
	}
}

func getTeamRolePlayers(tx *sql.Tx, teamID int) []rolePlayer {
	rows, err := tx.Query(`SELECT player_id, role FROM tournament_team_players WHERE team_id=$1`, teamID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var result []rolePlayer
	for rows.Next() {
		var p rolePlayer
		rows.Scan(&p.PlayerID, &p.Role)
		result = append(result, p)
	}
	return result
}

// --- Scoring ---

func ScoreTournamentMatch(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tournaments/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "matches" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	matchID, _ := strconv.Atoi(parts[2])

	var req struct {
		Game1ScoreA int  `json:"game1_score_a"`
		Game1ScoreB int  `json:"game1_score_b"`
		Game2ScoreA int  `json:"game2_score_a"`
		Game2ScoreB int  `json:"game2_score_b"`
		Game3ScoreA *int `json:"game3_score_a"`
		Game3ScoreB *int `json:"game3_score_b"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	// Determine game winner (best-of-3)
	aWins := 0
	bWins := 0
	if req.Game1ScoreA > req.Game1ScoreB {
		aWins++
	} else {
		bWins++
	}
	if req.Game2ScoreA > req.Game2ScoreB {
		aWins++
	} else {
		bWins++
	}

	needsGame3 := aWins == 1 && bWins == 1
	if needsGame3 && req.Game3ScoreA == nil {
		http.Error(w, "需要第三局比分", http.StatusBadRequest)
		return
	}
	if !needsGame3 && req.Game3ScoreA != nil {
		http.Error(w, "不需要第三局", http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var tournamentID, teamMatchID, teamAID, teamBID int
	var wasPlayed bool
	tx.QueryRow(
		`SELECT tournament_id, team_match_id, team_a_id, team_b_id, played
		FROM tournament_matches WHERE id=$1 FOR UPDATE`, matchID,
	).Scan(&tournamentID, &teamMatchID, &teamAID, &teamBID, &wasPlayed)

	var winnerTeamID int
	if aWins == 2 || (needsGame3 && *req.Game3ScoreA > *req.Game3ScoreB) {
		winnerTeamID = teamAID
	} else {
		winnerTeamID = teamBID
	}

	var g3a, g3b interface{}
	if req.Game3ScoreA != nil {
		g3a = *req.Game3ScoreA
		g3b = *req.Game3ScoreB
	} else {
		g3a = nil
		g3b = nil
	}

	tx.Exec(
		`UPDATE tournament_matches SET
			game1_score_a=$1, game1_score_b=$2,
			game2_score_a=$3, game2_score_b=$4,
			game3_score_a=$5, game3_score_b=$6,
			winner_team_id=$7, played=true
		WHERE id=$8`,
		req.Game1ScoreA, req.Game1ScoreB,
		req.Game2ScoreA, req.Game2ScoreB,
		g3a, g3b,
		winnerTeamID, matchID,
	)

	// Update team match score
	updateTeamMatchScore(tx, teamMatchID, teamAID, teamBID)

	// Check if team match is complete (first to 3 wins)
	var tmAWins, tmBWins int
	var tmWinnerID sql.NullInt64
	var tmPhase string
	tx.QueryRow(
		`SELECT team_a_wins, team_b_wins, winner_team_id, phase FROM tournament_team_matches WHERE id=$1`,
		teamMatchID,
	).Scan(&tmAWins, &tmBWins, &tmWinnerID, &tmPhase)

	if tmAWins >= 3 || tmBWins >= 3 {
		winningTeam := teamAID
		if tmBWins >= 3 {
			winningTeam = teamBID
		}
		tx.Exec(`UPDATE tournament_team_matches SET winner_team_id=$1, played=true WHERE id=$2`, winningTeam, teamMatchID)

		// Update group standings if group phase
		if tmPhase == "group" {
			tx.Exec(`UPDATE tournament_teams SET group_wins = group_wins + 1 WHERE id=$1`, winningTeam)
			losingTeam := teamAID
			if winningTeam == teamAID {
				losingTeam = teamBID
			}
			tx.Exec(`UPDATE tournament_teams SET group_losses = group_losses + 1 WHERE id=$1`, losingTeam)
		}

		// If semifinal, set up final
		if tmPhase == "semifinal" {
			setupFinalMatch(tx, tournamentID)
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status":         "ok",
		"winner_team_id": winnerTeamID,
		"team_a_wins":    tmAWins,
		"team_b_wins":    tmBWins,
	})
}

func updateTeamMatchScore(tx *sql.Tx, teamMatchID, teamAID, teamBID int) {
	// Count played matches won by each team
	var aWins, bWins int
	tx.QueryRow(
		`SELECT COUNT(*) FROM tournament_matches WHERE team_match_id=$1 AND winner_team_id=$2 AND played=true`,
		teamMatchID, teamAID,
	).Scan(&aWins)
	tx.QueryRow(
		`SELECT COUNT(*) FROM tournament_matches WHERE team_match_id=$1 AND winner_team_id=$2 AND played=true`,
		teamMatchID, teamBID,
	).Scan(&bWins)

	tx.Exec(`UPDATE tournament_team_matches SET team_a_wins=$1, team_b_wins=$2 WHERE id=$3`, aWins, bWins, teamMatchID)
}

func setupFinalMatch(tx *sql.Tx, tournamentID int) {
	// Check if both semifinals are complete
	var sfCount, sfComplete int
	tx.QueryRow(
		`SELECT COUNT(*), COUNT(CASE WHEN played=true THEN 1 END)
		FROM tournament_team_matches WHERE tournament_id=$1 AND phase='semifinal'`, tournamentID,
	).Scan(&sfCount, &sfComplete)

	if sfComplete < sfCount {
		return
	}

	// Get semifinal winners
	var sf1Winner, sf2Winner int
	tx.QueryRow(
		`SELECT winner_team_id FROM tournament_team_matches WHERE tournament_id=$1 AND phase='semifinal' ORDER BY id LIMIT 1`,
		tournamentID,
	).Scan(&sf1Winner)
	tx.QueryRow(
		`SELECT winner_team_id FROM tournament_team_matches WHERE tournament_id=$1 AND phase='semifinal' ORDER BY id OFFSET 1 LIMIT 1`,
		tournamentID,
	).Scan(&sf2Winner)

	// Check if final already exists
	var finalExists int
	tx.QueryRow(
		`SELECT COUNT(*) FROM tournament_team_matches WHERE tournament_id=$1 AND phase='final'`, tournamentID,
	).Scan(&finalExists)

	if finalExists == 0 {
		var tmID int
		tx.QueryRow(
			`INSERT INTO tournament_team_matches (tournament_id, phase, round, team_a_id, team_b_id)
			VALUES ($1, 'final', 1, $2, $3) RETURNING id`,
			tournamentID, sf1Winner, sf2Winner,
		).Scan(&tmID)
		createSubMatches(tx, tmID, tournamentID, "final", 1, "", sf1Winner, sf2Winner)
	}

	tx.Exec(`UPDATE tournaments SET phase='final' WHERE id=$1`, tournamentID)
}

func ForfeitTournamentMatch(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tournaments/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[1] != "matches" || parts[3] != "forfeit" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	matchID, _ := strconv.Atoi(parts[2])

	var req struct {
		WinnerTeamID int `json:"winner_team_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var tournamentID, teamMatchID, teamAID, teamBID int
	var played bool
	tx.QueryRow(
		`SELECT tournament_id, team_match_id, team_a_id, team_b_id, played
		FROM tournament_matches WHERE id=$1 FOR UPDATE`, matchID,
	).Scan(&tournamentID, &teamMatchID, &teamAID, &teamBID, &played)

	if played {
		http.Error(w, "match already played", http.StatusBadRequest)
		return
	}

	tx.Exec(
		`UPDATE tournament_matches SET game1_score_a=0, game1_score_b=0, game2_score_a=0, game2_score_b=0, winner_team_id=$1, played=true, forfeit=true WHERE id=$2`,
		req.WinnerTeamID, matchID,
	)

	updateTeamMatchScore(tx, teamMatchID, teamAID, teamBID)

	// Check team match completion
	var tmAWins, tmBWins int
	var tmPhase string
	tx.QueryRow(
		`SELECT team_a_wins, team_b_wins, phase FROM tournament_team_matches WHERE id=$1`,
		teamMatchID,
	).Scan(&tmAWins, &tmBWins, &tmPhase)

	if tmAWins >= 3 || tmBWins >= 3 {
		winningTeam := teamAID
		if tmBWins >= 3 {
			winningTeam = teamBID
		}
		tx.Exec(`UPDATE tournament_team_matches SET winner_team_id=$1, played=true WHERE id=$2`, winningTeam, teamMatchID)

		if tmPhase == "group" {
			tx.Exec(`UPDATE tournament_teams SET group_wins = group_wins + 1 WHERE id=$1`, winningTeam)
			losingTeam := teamAID
			if winningTeam == teamAID {
				losingTeam = teamBID
			}
			tx.Exec(`UPDATE tournament_teams SET group_losses = group_losses + 1 WHERE id=$1`, losingTeam)
		}
		if tmPhase == "semifinal" {
			setupFinalMatch(tx, tournamentID)
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"status": "ok"})
}

// --- Cards ---

func DrawTeamCard(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tournaments/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[1] != "team-matches" || parts[2] != "draw-card" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	tournamentID, _ := strconv.Atoi(parts[0])
	teamMatchID, _ := strconv.Atoi(parts[1])

	var req struct {
		TeamID int `json:"team_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	// Check team hasn't already drawn for this team match
	var existingCount int
	db.DB.QueryRow(
		`SELECT COUNT(*) FROM tournament_cards WHERE team_match_id=$1 AND team_id=$2`,
		teamMatchID, req.TeamID,
	).Scan(&existingCount)
	if existingCount > 0 {
		http.Error(w, "已抽过卡", http.StatusBadRequest)
		return
	}

	// Random draw
	card := tournamentCardTypes[rand.Intn(len(tournamentCardTypes))]

	db.DB.Exec(
		`INSERT INTO tournament_cards (tournament_id, team_match_id, team_id, card_type) VALUES ($1, $2, $3, $4)`,
		tournamentID, teamMatchID, req.TeamID, card.Type,
	)

	writeJSON(w, map[string]interface{}{
		"card_type":  card.Type,
		"card_detail": card.Detail,
	})
}

// --- Knockout ---

func AdvanceToKnockout(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tournaments/")
	tournamentID, _ := strconv.Atoi(strings.Split(path, "/")[0])

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var groupCount int
	tx.QueryRow(`SELECT group_count FROM tournaments WHERE id=$1`, tournamentID).Scan(&groupCount)

	groupLetters := "ABCDEFGH"

	// Rank teams within each group and get top 2
	type rankedTeam struct {
		TeamID int
		Rank   int
	}
	var groupTop2 []rankedTeam // [A1, A2, B1, B2, ...]

	for g := 0; g < groupCount; g++ {
		groupName := string(groupLetters[g])

		rows, err := tx.Query(
			`SELECT id, group_wins, group_losses FROM tournament_teams
			WHERE tournament_id=$1 AND group_name=$2
			ORDER BY group_wins DESC, (group_wins - group_losses) DESC`, tournamentID, groupName)
		if err != nil {
			continue
		}

		var teams []struct {
			ID    int
			Wins  int
			Losses int
		}
		for rows.Next() {
			var t struct {
				ID    int
				Wins  int
				Losses int
			}
			rows.Scan(&t.ID, &t.Wins, &t.Losses)
			teams = append(teams, t)
		}
		rows.Close()

		// Update group ranks
		for i, t := range teams {
			rank := i + 1
			tx.Exec(`UPDATE tournament_teams SET group_rank=$1 WHERE id=$2`, rank, t.ID)
		}

		// Take top 2
		if len(teams) >= 2 {
			groupTop2 = append(groupTop2, rankedTeam{teams[0].ID, 1}, rankedTeam{teams[1].ID, 2})
		}
	}

	if len(groupTop2) < 4 {
		http.Error(w, "小组赛未完成，无法生成淘汰赛", http.StatusBadRequest)
		return
	}

	// Delete existing knockout matches if re-advancing (FK cascade handles sub-matches and cards)
	tx.Exec(`DELETE FROM tournament_team_matches WHERE tournament_id=$1 AND phase IN ('semifinal','final')`, tournamentID)

	// Cross semi-finals: A1 vs B2, B1 vs A2
	a1 := groupTop2[0].TeamID // A组第1
	a2 := groupTop2[1].TeamID // A组第2
	b1 := groupTop2[2].TeamID // B组第1
	b2 := groupTop2[3].TeamID // B组第2

	// Set knockout seeds
	tx.Exec(`UPDATE tournament_teams SET knockout_seed=1 WHERE id=$1`, a1)
	tx.Exec(`UPDATE tournament_teams SET knockout_seed=2 WHERE id=$1`, b1)
	tx.Exec(`UPDATE tournament_teams SET knockout_seed=3 WHERE id=$1`, a2)
	tx.Exec(`UPDATE tournament_teams SET knockout_seed=4 WHERE id=$1`, b2)

	// SF1: A1 vs B2
	var sf1ID int
	tx.QueryRow(
		`INSERT INTO tournament_team_matches (tournament_id, phase, round, team_a_id, team_b_id)
		VALUES ($1, 'semifinal', 1, $2, $3) RETURNING id`,
		tournamentID, a1, b2,
	).Scan(&sf1ID)
	createSubMatches(tx, sf1ID, tournamentID, "semifinal", 1, "", a1, b2)

	// SF2: B1 vs A2
	var sf2ID int
	tx.QueryRow(
		`INSERT INTO tournament_team_matches (tournament_id, phase, round, team_a_id, team_b_id)
		VALUES ($1, 'semifinal', 2, $2, $3) RETURNING id`,
		tournamentID, b1, a2,
	).Scan(&sf2ID)
	createSubMatches(tx, sf2ID, tournamentID, "semifinal", 2, "", b1, a2)

	tx.Exec(`UPDATE tournaments SET phase='semifinal' WHERE id=$1`, tournamentID)

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"status": "ok"})
}

// --- Complete ---

func CompleteTournament(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tournaments/")
	parts := strings.Split(path, "/")
	tournamentID, _ := strconv.Atoi(parts[0])

	// Check all team matches played
	var unplayed int
	db.DB.QueryRow(
		`SELECT COUNT(*) FROM tournament_team_matches WHERE tournament_id=$1 AND played=false`, tournamentID,
	).Scan(&unplayed)
	if unplayed > 0 {
		http.Error(w, "还有未结束的比赛", http.StatusBadRequest)
		return
	}

	db.DB.Exec(
		`UPDATE tournaments SET status='completed', phase='completed', completed_at=NOW() WHERE id=$1`,
		tournamentID,
	)

	writeJSON(w, map[string]string{"status": "ok"})
}
