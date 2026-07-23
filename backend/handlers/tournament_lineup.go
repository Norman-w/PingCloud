package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"pingpong/db"
)

// ===== 分区：公开 API =====

// SetTeamMatchLineup configures A/B/C players for both sides of a team match.
func SetTeamMatchLineup(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tournaments/")
	parts := strings.Split(path, "/")
	// {id}/team-matches/{tmId}/lineup
	if len(parts) < 4 || parts[1] != "team-matches" || parts[3] != "lineup" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	tournamentID, _ := strconv.Atoi(parts[0])
	teamMatchID, _ := strconv.Atoi(parts[2])

	var req struct {
		TeamA map[string]int `json:"team_a"` // A/B/C -> player_id
		TeamB map[string]int `json:"team_b"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	for _, role := range []string{"A", "B", "C"} {
		if req.TeamA[role] == 0 || req.TeamB[role] == 0 {
			http.Error(w, "请为双方配置完整的 A/B/C 出场", http.StatusBadRequest)
			return
		}
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var teamAID, teamBID int
	var played bool
	err = tx.QueryRow(
		`SELECT team_a_id, team_b_id, played FROM tournament_team_matches
		WHERE id=$1 AND tournament_id=$2 FOR UPDATE`, teamMatchID, tournamentID,
	).Scan(&teamAID, &teamBID, &played)
	if err != nil {
		http.Error(w, "对阵不存在", http.StatusNotFound)
		return
	}
	if played {
		http.Error(w, "对阵已结束，无法修改出场", http.StatusBadRequest)
		return
	}

	if err := assertPlayersOnTeam(tx, teamAID, req.TeamA); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := assertPlayersOnTeam(tx, teamBID, req.TeamB); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	aA, aB, aC := req.TeamA["A"], req.TeamA["B"], req.TeamA["C"]
	bA, bB, bC := req.TeamB["A"], req.TeamB["B"], req.TeamB["C"]

	// Standard 5-match ABC assignment
	updates := []struct {
		order int
		a1, a2, b1, b2 interface{}
	}{
		{1, aA, nil, bA, nil},
		{2, aB, aC, bB, bC},
		{3, aC, nil, bC, nil},
		{4, aA, aB, bA, bB},
		{5, aB, nil, bB, nil},
	}
	for _, u := range updates {
		_, err := tx.Exec(
			`UPDATE tournament_matches SET
				player_a_id=$1, player_a2_id=$2, player_b_id=$3, player_b2_id=$4
			WHERE team_match_id=$5 AND match_order=$6`,
			u.a1, u.a2, u.b1, u.b2, teamMatchID, u.order,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// CompleteTeamMatch manually finalizes a team match after scores are confirmed.
func CompleteTeamMatch(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tournaments/")
	parts := strings.Split(path, "/")
	// {id}/team-matches/{tmId}/complete
	if len(parts) < 4 || parts[1] != "team-matches" || parts[3] != "complete" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	tournamentID, _ := strconv.Atoi(parts[0])
	teamMatchID, _ := strconv.Atoi(parts[2])

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var teamAID, teamBID, aWins, bWins int
	var phase, groupName string
	var played bool
	err = tx.QueryRow(
		`SELECT team_a_id, team_b_id, team_a_wins, team_b_wins, phase, COALESCE(group_name,''), played
		FROM tournament_team_matches WHERE id=$1 AND tournament_id=$2 FOR UPDATE`,
		teamMatchID, tournamentID,
	).Scan(&teamAID, &teamBID, &aWins, &bWins, &phase, &groupName, &played)
	if err != nil {
		http.Error(w, "对阵不存在", http.StatusNotFound)
		return
	}
	if played {
		http.Error(w, "对阵已结束", http.StatusBadRequest)
		return
	}

	// Refresh wins from sub-matches
	updateTeamMatchScore(tx, teamMatchID, teamAID, teamBID)
	tx.QueryRow(
		`SELECT team_a_wins, team_b_wins FROM tournament_team_matches WHERE id=$1`, teamMatchID,
	).Scan(&aWins, &bWins)

	if aWins < 3 && bWins < 3 {
		http.Error(w, "需一方先胜满 3 场才能提交结束", http.StatusBadRequest)
		return
	}

	winningTeam := teamAID
	if bWins >= 3 {
		winningTeam = teamBID
	}
	tx.Exec(
		`UPDATE tournament_team_matches SET winner_team_id=$1, played=true WHERE id=$2`,
		winningTeam, teamMatchID,
	)

	if phase == "group" && groupName != "" {
		if err := recalcGroupStandings(tx, tournamentID, groupName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if phase == "semifinal" {
		setupFinalMatch(tx, tournamentID)
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]interface{}{
		"status":         "ok",
		"winner_team_id": winningTeam,
		"team_a_wins":    aWins,
		"team_b_wins":    bWins,
	})
}

// ReopenTeamMatch unlocks a completed team match so scores can be corrected.
func ReopenTeamMatch(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tournaments/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[1] != "team-matches" || parts[3] != "reopen" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	tournamentID, _ := strconv.Atoi(parts[0])
	teamMatchID, _ := strconv.Atoi(parts[2])

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var phase, groupName string
	var played bool
	err = tx.QueryRow(
		`SELECT phase, COALESCE(group_name,''), played FROM tournament_team_matches
		WHERE id=$1 AND tournament_id=$2 FOR UPDATE`, teamMatchID, tournamentID,
	).Scan(&phase, &groupName, &played)
	if err != nil {
		http.Error(w, "对阵不存在", http.StatusNotFound)
		return
	}
	if !played {
		http.Error(w, "对阵尚未结束", http.StatusBadRequest)
		return
	}

	tx.Exec(`UPDATE tournament_team_matches SET played=false, winner_team_id=NULL WHERE id=$1`, teamMatchID)

	if phase == "group" && groupName != "" {
		if err := recalcGroupStandings(tx, tournamentID, groupName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// 重开半决赛时，作废已生成的决赛/三四名，避免对阵错乱
	if phase == "semifinal" {
		tx.Exec(`DELETE FROM tournament_team_matches WHERE tournament_id=$1 AND phase IN ('final','third_place')`, tournamentID)
		tx.Exec(`UPDATE tournaments SET phase='semifinal' WHERE id=$1`, tournamentID)
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// ===== 分区：方法/工具 =====

func assertPlayersOnTeam(tx *sql.Tx, teamID int, roles map[string]int) error {
	seen := map[int]bool{}
	for role, pid := range roles {
		if seen[pid] {
			return errBadRequest("同一队出场选手不能重复")
		}
		seen[pid] = true
		var n int
		tx.QueryRow(
			`SELECT COUNT(*) FROM tournament_team_players WHERE team_id=$1 AND player_id=$2`,
			teamID, pid,
		).Scan(&n)
		if n == 0 {
			return errBadRequest("选手不属于该队: " + role)
		}
	}
	return nil
}

type badRequestError string

func (e badRequestError) Error() string { return string(e) }

func errBadRequest(msg string) error { return badRequestError(msg) }
