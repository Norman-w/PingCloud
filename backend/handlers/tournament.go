package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
	ID           int                    `json:"id"`
	TournamentID int                    `json:"tournament_id"`
	GroupName    string                 `json:"group_name"`
	TeamIndex    int                    `json:"team_index"`
	TeamName     string                 `json:"team_name"`
	KnockoutSeed *int                   `json:"knockout_seed"`
	GroupRank    *int                   `json:"group_rank"`
	GroupWins    int                    `json:"group_wins"`
	GroupLosses  int                    `json:"group_losses"`
	GroupPoints  int                    `json:"group_points"`
	GamesWon     int                    `json:"games_won"`
	GamesLost    int                    `json:"games_lost"`
	PointsScored int                    `json:"points_scored"`
	RankManual   bool                   `json:"rank_manual"`
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
		req.Name = "混合团体赛"
	}
	if req.GroupCount < 1 {
		req.GroupCount = 1
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

	// Backfill 决赛/三四名 if both semis already finished (upgrade path)
	if detail.Phase == "semifinal" || detail.Phase == "final" {
		if tx, err := db.DB.Begin(); err == nil {
			setupFinalMatch(tx, tournamentID)
			_ = tx.Commit()
			// refresh phase after possible upgrade
			db.DB.QueryRow(`SELECT phase FROM tournaments WHERE id=$1`, tournamentID).Scan(&detail.Phase)
		}
	}

	// 小组赛阶段每次打开刷新积分榜（含「已满3胜待提交」预览）
	if detail.Phase == "group" || detail.Status == "group_stage" {
		if tx, err := db.DB.Begin(); err == nil {
			if err := recalcAllGroupStandings(tx, tournamentID); err != nil {
				tx.Rollback()
			} else {
				_ = tx.Commit()
			}
		}
	}

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
			knockout_seed, group_rank, group_wins, group_losses, group_points,
			COALESCE(games_won, 0), COALESCE(games_lost, 0), COALESCE(points_scored, 0),
			COALESCE(rank_manual, false)
		FROM tournament_teams WHERE tournament_id=$1 ORDER BY group_name, team_index`, tournamentID)
	if err != nil {
		return []TournamentTeam{}
	}
	defer rows.Close()

	var teams []TournamentTeam
	for rows.Next() {
		var t TournamentTeam
		rows.Scan(&t.ID, &t.TournamentID, &t.GroupName, &t.TeamIndex, &t.TeamName,
			&t.KnockoutSeed, &t.GroupRank, &t.GroupWins, &t.GroupLosses, &t.GroupPoints,
			&t.GamesWon, &t.GamesLost, &t.PointsScored, &t.RankManual)
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
		ORDER BY CASE phase WHEN 'group' THEN 1 WHEN 'semifinal' THEN 2 WHEN 'final' THEN 3 WHEN 'third_place' THEN 4 ELSE 5 END, round, id`, tournamentID)
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

	var groupCount, playersPerTeam, seedCount int
	var seedEnabled bool
	err = tx.QueryRow(
		`SELECT group_count, players_per_team, seed_enabled, seed_count
		FROM tournaments WHERE id=$1 AND deleted=false`, tournamentID,
	).Scan(&groupCount, &playersPerTeam, &seedEnabled, &seedCount)
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

	totalTeams := len(players) / playersPerTeam
	if len(players)%playersPerTeam != 0 {
		http.Error(w, fmt.Sprintf("报名人数需为每队人数的整数倍（当前 %d 人，每队 %d 人）", len(players), playersPerTeam), http.StatusBadRequest)
		return
	}
	// 每组至少 2 队；单组模式也只需 ≥2 队即可循环
	minTeams := 2
	if groupCount > 1 {
		minTeams = groupCount * 2
	}
	if totalTeams < minTeams {
		need := minTeams * playersPerTeam
		http.Error(w, fmt.Sprintf("队伍数不足：至少需要 %d 人组成 %d 组（当前可组成 %d 队）", need, groupCount, totalTeams), http.StatusBadRequest)
		return
	}

	// 组间队数尽量均分；有余数时随机决定哪几组多 1 队（例如 7 队 → 3+4）
	groupSizes := make([]int, groupCount)
	base := totalTeams / groupCount
	extra := totalTeams % groupCount
	for i := 0; i < groupCount; i++ {
		groupSizes[i] = base
	}
	if extra > 0 {
		order := rand.Perm(groupCount)
		for i := 0; i < extra; i++ {
			groupSizes[order[i]]++
		}
	}

	// Create teams (临时名，抽签后改为 组字母+种子姓名)
	teamIDs := make([]int, 0, totalTeams)
	teamGroup := make(map[int]string) // teamID -> group letter
	groupLetters := "ABCDEFGH"
	for g := 0; g < groupCount; g++ {
		groupName := string(groupLetters[g])
		for t := 0; t < groupSizes[g]; t++ {
			teamName := groupName + strconv.Itoa(t+1)
			var tid int
			tx.QueryRow(
				`INSERT INTO tournament_teams (tournament_id, group_name, team_index, team_name) VALUES ($1, $2, $3, $4) RETURNING id`,
				tournamentID, groupName, t, teamName,
			).Scan(&tid)
			teamIDs = append(teamIDs, tid)
			teamGroup[tid] = groupName
		}
	}

	roles := make([]string, playersPerTeam)
	for i := 0; i < playersPerTeam; i++ {
		roles[i] = string(rune('A' + i))
	}

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

	// 队名：组字母+种子姓名（无种子则取队内积分最高者），如 A+任鑫
	for _, tid := range teamIDs {
		var anchor string
		tx.QueryRow(
			`SELECT p.name FROM tournament_team_players ttp
			JOIN players p ON p.id = ttp.player_id
			WHERE ttp.team_id=$1
			ORDER BY ttp.is_seed DESC, COALESCE(p.reference_rating, p.current_rating) DESC
			LIMIT 1`, tid,
		).Scan(&anchor)
		if anchor == "" {
			anchor = "队"
		}
		teamName := teamGroup[tid] + "+" + anchor
		tx.Exec(`UPDATE tournament_teams SET team_name=$1 WHERE id=$2`, teamName, tid)
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
	var tmPlayed bool
	var tmPhase string
	tx.QueryRow(
		`SELECT tournament_id, team_match_id, team_a_id, team_b_id, played
		FROM tournament_matches WHERE id=$1 FOR UPDATE`, matchID,
	).Scan(&tournamentID, &teamMatchID, &teamAID, &teamBID, &wasPlayed)

	tx.QueryRow(
		`SELECT played, phase FROM tournament_team_matches WHERE id=$1 FOR UPDATE`, teamMatchID,
	).Scan(&tmPlayed, &tmPhase)
	if tmPlayed {
		http.Error(w, "对阵已结束，请先「重新打开」后再改分", http.StatusBadRequest)
		return
	}

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
			winner_team_id=$7, played=true, forfeit=false
		WHERE id=$8`,
		req.Game1ScoreA, req.Game1ScoreB,
		req.Game2ScoreA, req.Game2ScoreB,
		g3a, g3b,
		winnerTeamID, matchID,
	)

	// Update team match score (do NOT auto-finalize — wait for manual complete)
	updateTeamMatchScore(tx, teamMatchID, teamAID, teamBID)

	var tmAWins, tmBWins int
	tx.QueryRow(
		`SELECT team_a_wins, team_b_wins FROM tournament_team_matches WHERE id=$1`,
		teamMatchID,
	).Scan(&tmAWins, &tmBWins)

	// Live standings preview only from completed team matches; unfinished scores don't affect table
	if tmPhase == "group" {
		var groupName string
		tx.QueryRow(`SELECT group_name FROM tournament_team_matches WHERE id=$1`, teamMatchID).Scan(&groupName)
		if err := recalcGroupStandings(tx, tournamentID, groupName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
		"ready_to_complete": tmAWins >= 3 || tmBWins >= 3,
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

	if sfCount < 2 || sfComplete < sfCount {
		return
	}

	type sfRow struct {
		TeamAID, TeamBID, WinnerID int
	}
	rows, err := tx.Query(
		`SELECT team_a_id, team_b_id, COALESCE(winner_team_id, 0)
		FROM tournament_team_matches WHERE tournament_id=$1 AND phase='semifinal' ORDER BY id`,
		tournamentID,
	)
	if err != nil {
		return
	}
	var sfs []sfRow
	for rows.Next() {
		var s sfRow
		rows.Scan(&s.TeamAID, &s.TeamBID, &s.WinnerID)
		sfs = append(sfs, s)
	}
	rows.Close()
	if len(sfs) < 2 || sfs[0].WinnerID == 0 || sfs[1].WinnerID == 0 {
		return
	}

	w1, w2 := sfs[0].WinnerID, sfs[1].WinnerID
	l1 := sfs[0].TeamBID
	if sfs[0].WinnerID == sfs[0].TeamBID {
		l1 = sfs[0].TeamAID
	}
	l2 := sfs[1].TeamBID
	if sfs[1].WinnerID == sfs[1].TeamBID {
		l2 = sfs[1].TeamAID
	}

	// 决赛：胜者决 1-2 名
	var finalExists int
	tx.QueryRow(
		`SELECT COUNT(*) FROM tournament_team_matches WHERE tournament_id=$1 AND phase='final'`, tournamentID,
	).Scan(&finalExists)
	if finalExists == 0 {
		var tmID int
		tx.QueryRow(
			`INSERT INTO tournament_team_matches (tournament_id, phase, round, team_a_id, team_b_id)
			VALUES ($1, 'final', 1, $2, $3) RETURNING id`,
			tournamentID, w1, w2,
		).Scan(&tmID)
		createSubMatches(tx, tmID, tournamentID, "final", 1, "", w1, w2)
	}

	// 三四名决赛：负者决 3-4 名
	var thirdExists int
	tx.QueryRow(
		`SELECT COUNT(*) FROM tournament_team_matches WHERE tournament_id=$1 AND phase='third_place'`, tournamentID,
	).Scan(&thirdExists)
	if thirdExists == 0 {
		var tmID int
		tx.QueryRow(
			`INSERT INTO tournament_team_matches (tournament_id, phase, round, team_a_id, team_b_id)
			VALUES ($1, 'third_place', 1, $2, $3) RETURNING id`,
			tournamentID, l1, l2,
		).Scan(&tmID)
		createSubMatches(tx, tmID, tournamentID, "third_place", 1, "", l1, l2)
	}

	tx.Exec(`UPDATE tournaments SET phase='final' WHERE id=$1`, tournamentID)
}

func ForfeitTournamentMatch(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "混合团体赛单局不支持弃权，请直接录入比分", http.StatusBadRequest)
}

// ClearTournamentMatch resets a sub-match to unplayed (score cleared).
func ClearTournamentMatch(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tournaments/")
	parts := strings.Split(path, "/")
	// expected: {id}/matches/{matchId}/clear
	if len(parts) < 4 || parts[1] != "matches" || parts[3] != "clear" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	matchID, _ := strconv.Atoi(parts[2])

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var tournamentID, teamMatchID, teamAID, teamBID int
	var tmPlayed bool
	var tmPhase string
	err = tx.QueryRow(
		`SELECT tournament_id, team_match_id, team_a_id, team_b_id
		FROM tournament_matches WHERE id=$1 FOR UPDATE`, matchID,
	).Scan(&tournamentID, &teamMatchID, &teamAID, &teamBID)
	if err != nil {
		http.Error(w, "比赛不存在", http.StatusNotFound)
		return
	}

	tx.QueryRow(
		`SELECT played, phase FROM tournament_team_matches WHERE id=$1 FOR UPDATE`, teamMatchID,
	).Scan(&tmPlayed, &tmPhase)
	if tmPlayed {
		http.Error(w, "对阵已结束，请先「重新打开」后再清除比分", http.StatusBadRequest)
		return
	}

	_, err = tx.Exec(
		`UPDATE tournament_matches SET
			game1_score_a=NULL, game1_score_b=NULL,
			game2_score_a=NULL, game2_score_b=NULL,
			game3_score_a=NULL, game3_score_b=NULL,
			winner_id=NULL, winner_team_id=NULL, played=false, forfeit=false
		WHERE id=$1`, matchID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updateTeamMatchScore(tx, teamMatchID, teamAID, teamBID)

	var tmAWins, tmBWins int
	tx.QueryRow(
		`SELECT team_a_wins, team_b_wins FROM tournament_team_matches WHERE id=$1`,
		teamMatchID,
	).Scan(&tmAWins, &tmBWins)

	if tmPhase == "group" {
		var groupName string
		tx.QueryRow(`SELECT group_name FROM tournament_team_matches WHERE id=$1`, teamMatchID).Scan(&groupName)
		if err := recalcGroupStandings(tx, tournamentID, groupName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status":         "ok",
		"team_a_wins":    tmAWins,
		"team_b_wins":    tmBWins,
	})
}

// --- Cards ---

func DrawTeamCard(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tournaments/")
	parts := strings.Split(path, "/")
	// expected: {id}/team-matches/{tmId}/draw-card
	if len(parts) < 4 || parts[1] != "team-matches" || parts[3] != "draw-card" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	tournamentID, _ := strconv.Atoi(parts[0])
	teamMatchID, _ := strconv.Atoi(parts[2])

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

	// Validate team match belongs to tournament
	var tmCount int
	db.DB.QueryRow(
		`SELECT COUNT(*) FROM tournament_team_matches WHERE id=$1 AND tournament_id=$2`,
		teamMatchID, tournamentID,
	).Scan(&tmCount)
	if tmCount == 0 {
		http.Error(w, "对阵不存在", http.StatusNotFound)
		return
	}

	// Random draw
	card := tournamentCardTypes[rand.Intn(len(tournamentCardTypes))]

	if _, err := db.DB.Exec(
		`INSERT INTO tournament_cards (tournament_id, team_match_id, team_id, card_type) VALUES ($1, $2, $3, $4)`,
		tournamentID, teamMatchID, req.TeamID, card.Type,
	); err != nil {
		http.Error(w, "抽卡失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{
		"card_type":   card.Type,
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
	if groupCount < 2 {
		http.Error(w, "单组循环赛无需晋级半决赛，请直接结束赛事公布排名", http.StatusBadRequest)
		return
	}

	groupLetters := "ABCDEFGH"

	// Rank teams within each group and get top 2
	type rankedTeam struct {
		TeamID int
		Rank   int
	}
	var groupTop2 []rankedTeam // [A1, A2, B1, B2, ...]

	for g := 0; g < groupCount; g++ {
		groupName := string(groupLetters[g])

		if err := recalcGroupStandings(tx, tournamentID, groupName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rows, err := tx.Query(
			`SELECT id, group_rank FROM tournament_teams
			WHERE tournament_id=$1 AND group_name=$2
			ORDER BY COALESCE(group_rank, 99), group_points DESC, games_won DESC, points_scored DESC`, tournamentID, groupName)
		if err != nil {
			continue
		}

		var teams []struct {
			ID   int
			Rank sql.NullInt64
		}
		for rows.Next() {
			var t struct {
				ID   int
				Rank sql.NullInt64
			}
			rows.Scan(&t.ID, &t.Rank)
			teams = append(teams, t)
		}
		rows.Close()

		for _, t := range teams {
			if !t.Rank.Valid {
				http.Error(w, groupName+"组存在无法自动判定的并列，请先手动设定名次", http.StatusBadRequest)
				return
			}
		}

		// Take top 2 by rank
		if len(teams) >= 2 {
			groupTop2 = append(groupTop2, rankedTeam{teams[0].ID, int(teams[0].Rank.Int64)}, rankedTeam{teams[1].ID, int(teams[1].Rank.Int64)})
		}
	}

	if len(groupTop2) < 4 {
		http.Error(w, "小组赛未完成，无法生成半决赛", http.StatusBadRequest)
		return
	}

	// Delete existing knockout matches if re-advancing (FK cascade handles sub-matches and cards)
	tx.Exec(`DELETE FROM tournament_team_matches WHERE tournament_id=$1 AND phase IN ('semifinal','final','third_place')`, tournamentID)

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

	// 单组循环赛：结束前再算一次积分榜，确保最终名次写入 group_rank
	var groupCount int
	db.DB.QueryRow(`SELECT group_count FROM tournaments WHERE id=$1`, tournamentID).Scan(&groupCount)
	if groupCount == 1 {
		tx, err := db.DB.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()
		if err := recalcGroupStandings(tx, tournamentID, "A"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var unresolved int
		tx.QueryRow(
			`SELECT COUNT(*) FROM tournament_teams
			WHERE tournament_id=$1 AND group_rank IS NULL`, tournamentID,
		).Scan(&unresolved)
		if unresolved > 0 {
			http.Error(w, "存在无法自动判定的并列，请先手动设定名次", http.StatusBadRequest)
			return
		}
		tx.Exec(
			`UPDATE tournaments SET status='completed', phase='completed', completed_at=NOW() WHERE id=$1`,
			tournamentID,
		)
		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]string{"status": "ok"})
		return
	}

	db.DB.Exec(
		`UPDATE tournaments SET status='completed', phase='completed', completed_at=NOW() WHERE id=$1`,
		tournamentID,
	)

	writeJSON(w, map[string]string{"status": "ok"})
}
