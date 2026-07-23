package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"pingpong/db"
)

// ===== 分区：模型/类型 =====

type teamStanding struct {
	TeamID       int
	Points       int // 胜场×2 + 负场×1
	Wins         int
	Losses       int
	GamesWon     int
	GamesLost    int
	PointsScored int
	RankManual   bool
	ManualRank   int
}

type teamMatchResult struct {
	TeamAID      int
	TeamBID      int
	WinnerTeamID int
	TeamAWins    int
	TeamBWins    int
	Played       bool
}

type gameScoreLine struct {
	TeamAID int
	TeamBID int
	ScoreA  int
	ScoreB  int
}

// ===== 分区：公开 API =====

// SetGroupRanks allows manual override when auto ranking cannot break a tie.
func SetGroupRanks(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/tournaments/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "set-ranks" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	tournamentID, _ := strconv.Atoi(parts[0])

	var req struct {
		GroupName string `json:"group_name"`
		Ranks     []struct {
			TeamID int `json:"team_id"`
			Rank   int `json:"rank"`
		} `json:"ranks"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.GroupName == "" || len(req.Ranks) == 0 {
		http.Error(w, "group_name and ranks required", http.StatusBadRequest)
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	for _, item := range req.Ranks {
		if item.Rank < 1 {
			http.Error(w, "rank must be >= 1", http.StatusBadRequest)
			return
		}
		res, err := tx.Exec(
			`UPDATE tournament_teams SET group_rank=$1, rank_manual=true
			WHERE id=$2 AND tournament_id=$3 AND group_name=$4`,
			item.Rank, item.TeamID, tournamentID, req.GroupName,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			http.Error(w, "team not found in group", http.StatusBadRequest)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// ===== 分区：业务逻辑 =====

func recalcGroupStandings(tx *sql.Tx, tournamentID int, groupName string) error {
	rows, err := tx.Query(
		`SELECT id, COALESCE(rank_manual, false), COALESCE(group_rank, 0)
		FROM tournament_teams WHERE tournament_id=$1 AND group_name=$2`,
		tournamentID, groupName,
	)
	if err != nil {
		return err
	}
	standings := map[int]*teamStanding{}
	var teamOrder []int
	for rows.Next() {
		var id int
		var manual bool
		var rank int
		if err := rows.Scan(&id, &manual, &rank); err != nil {
			rows.Close()
			return err
		}
		standings[id] = &teamStanding{TeamID: id, RankManual: manual, ManualRank: rank}
		teamOrder = append(teamOrder, id)
	}
	rows.Close()
	if len(standings) == 0 {
		return nil
	}

	tmRows, err := tx.Query(
		`SELECT team_a_id, team_b_id, COALESCE(winner_team_id, 0), team_a_wins, team_b_wins, played
		FROM tournament_team_matches
		WHERE tournament_id=$1 AND group_name=$2 AND phase='group'`,
		tournamentID, groupName,
	)
	if err != nil {
		return err
	}
	var teamMatches []teamMatchResult
	for tmRows.Next() {
		var m teamMatchResult
		if err := tmRows.Scan(&m.TeamAID, &m.TeamBID, &m.WinnerTeamID, &m.TeamAWins, &m.TeamBWins, &m.Played); err != nil {
			tmRows.Close()
			return err
		}
		teamMatches = append(teamMatches, m)
	}
	tmRows.Close()

	for _, m := range teamMatches {
		if !m.Played || m.WinnerTeamID == 0 {
			continue
		}
		winner := standings[m.WinnerTeamID]
		loserID := m.TeamAID
		if m.WinnerTeamID == m.TeamAID {
			loserID = m.TeamBID
		}
		loser := standings[loserID]
		if winner == nil || loser == nil {
			continue
		}
		winner.Wins++
		loser.Losses++
	}

	gameRows, err := tx.Query(
		`SELECT m.team_a_id, m.team_b_id,
			COALESCE(m.game1_score_a, 0), COALESCE(m.game1_score_b, 0),
			COALESCE(m.game2_score_a, 0), COALESCE(m.game2_score_b, 0),
			m.game3_score_a, m.game3_score_b,
			m.played, COALESCE(m.forfeit, false)
		FROM tournament_matches m
		JOIN tournament_team_matches tm ON tm.id = m.team_match_id
		WHERE tm.tournament_id=$1 AND tm.group_name=$2 AND tm.phase='group'
			AND tm.played=true AND m.played=true`,
		tournamentID, groupName,
	)
	if err != nil {
		return err
	}
	var gameLines []gameScoreLine
	for gameRows.Next() {
		var teamAID, teamBID int
		var g1a, g1b, g2a, g2b int
		var g3a, g3b sql.NullInt64
		var played, forfeit bool
		if err := gameRows.Scan(&teamAID, &teamBID, &g1a, &g1b, &g2a, &g2b, &g3a, &g3b, &played, &forfeit); err != nil {
			gameRows.Close()
			return err
		}
		if !played {
			continue
		}
		lines := []gameScoreLine{
			{teamAID, teamBID, g1a, g1b},
			{teamAID, teamBID, g2a, g2b},
		}
		if g3a.Valid && g3b.Valid {
			lines = append(lines, gameScoreLine{teamAID, teamBID, int(g3a.Int64), int(g3b.Int64)})
		}
		for _, line := range lines {
			if forfeit {
				continue
			}
			gameLines = append(gameLines, line)
			applyGameScores(standings, line.TeamAID, line.TeamBID, line.ScoreA, line.ScoreB)
		}
	}
	gameRows.Close()

	for _, s := range standings {
		s.Points = s.Wins*2 + s.Losses*1
	}

	ranked := rankGroupTeams(standings, teamMatches, gameLines, teamOrder)

	for _, s := range standings {
		rank := ranked[s.TeamID]
		_, err := tx.Exec(
			`UPDATE tournament_teams SET
				group_wins=$1, group_losses=$2, group_points=$3,
				games_won=$4, games_lost=$5, points_scored=$6,
				group_rank=$7
			WHERE id=$8`,
			s.Wins, s.Losses, s.Points,
			s.GamesWon, s.GamesLost, s.PointsScored,
			nullInt(rank), s.TeamID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func recalcAllGroupStandings(tx *sql.Tx, tournamentID int) error {
	rows, err := tx.Query(
		`SELECT DISTINCT group_name FROM tournament_teams WHERE tournament_id=$1`,
		tournamentID,
	)
	if err != nil {
		return err
	}
	var groups []string
	for rows.Next() {
		var g string
		if err := rows.Scan(&g); err != nil {
			rows.Close()
			return err
		}
		groups = append(groups, g)
	}
	rows.Close()
	for _, g := range groups {
		if err := recalcGroupStandings(tx, tournamentID, g); err != nil {
			return err
		}
	}
	return nil
}

// ===== 分区：方法/工具 =====

func applyGameScores(standings map[int]*teamStanding, teamAID, teamBID, scoreA, scoreB int) {
	a := standings[teamAID]
	b := standings[teamBID]
	if a == nil || b == nil {
		return
	}
	a.PointsScored += scoreA
	b.PointsScored += scoreB
	if scoreA > scoreB {
		a.GamesWon++
		b.GamesLost++
	} else if scoreB > scoreA {
		b.GamesWon++
		a.GamesLost++
	}
}

func nullInt(v int) interface{} {
	if v <= 0 {
		return nil
	}
	return v
}

func rankGroupTeams(standings map[int]*teamStanding, matches []teamMatchResult, games []gameScoreLine, teamOrder []int) map[int]int {
	result := map[int]int{}
	manualUsed := map[int]bool{}
	for _, s := range standings {
		if s.RankManual && s.ManualRank > 0 {
			result[s.TeamID] = s.ManualRank
			manualUsed[s.ManualRank] = true
		}
	}

	type node struct {
		id int
		s  *teamStanding
	}
	var auto []node
	for _, id := range teamOrder {
		s := standings[id]
		if s.RankManual && s.ManualRank > 0 {
			continue
		}
		auto = append(auto, node{id: id, s: s})
	}

	sort.SliceStable(auto, func(i, j int) bool {
		return compareTeams(auto[i].s, auto[j].s, standings, matches, games) < 0
	})

	nextRank := 1
	for nextRank <= len(standings) && manualUsed[nextRank] {
		nextRank++
	}

	i := 0
	for i < len(auto) {
		j := i + 1
		for j < len(auto) && compareTeams(auto[i].s, auto[j].s, standings, matches, games) == 0 {
			j++
		}
		if j-i > 1 {
			for k := i; k < j; k++ {
				result[auto[k].id] = 0
			}
		} else {
			result[auto[i].id] = nextRank
			nextRank++
			for nextRank <= len(standings) && manualUsed[nextRank] {
				nextRank++
			}
		}
		i = j
	}
	return result
}

func compareTeams(a, b *teamStanding, all map[int]*teamStanding, matches []teamMatchResult, games []gameScoreLine) int {
	if a.Points != b.Points {
		if a.Points > b.Points {
			return -1
		}
		return 1
	}

	var tiedIDs []int
	for id, s := range all {
		if s.Points == a.Points && !(s.RankManual && s.ManualRank > 0) {
			tiedIDs = append(tiedIDs, id)
		}
	}
	sort.Ints(tiedIDs)

	if len(tiedIDs) == 2 {
		h2h := headToHeadWinner(a.TeamID, b.TeamID, matches)
		if h2h == a.TeamID {
			return -1
		}
		if h2h == b.TeamID {
			return 1
		}
	}

	if len(tiedIDs) >= 2 {
		miniA := miniStats(a.TeamID, tiedIDs, matches, games)
		miniB := miniStats(b.TeamID, tiedIDs, matches, games)
		if miniA.points != miniB.points {
			if miniA.points > miniB.points {
				return -1
			}
			return 1
		}
		// 互战场比（小场净胜）
		fDiffA := miniA.fieldsWon - miniA.fieldsLost
		fDiffB := miniB.fieldsWon - miniB.fieldsLost
		if fDiffA != fDiffB {
			if fDiffA > fDiffB {
				return -1
			}
			return 1
		}
		if miniA.fieldsWon != miniB.fieldsWon {
			if miniA.fieldsWon > miniB.fieldsWon {
				return -1
			}
			return 1
		}
		diffA := miniA.gamesWon - miniA.gamesLost
		diffB := miniB.gamesWon - miniB.gamesLost
		if diffA != diffB {
			if diffA > diffB {
				return -1
			}
			return 1
		}
		if miniA.gamesWon != miniB.gamesWon {
			if miniA.gamesWon > miniB.gamesWon {
				return -1
			}
			return 1
		}
		if miniA.pointsScored != miniB.pointsScored {
			if miniA.pointsScored > miniB.pointsScored {
				return -1
			}
			return 1
		}
		return 0
	}

	diffA := a.GamesWon - a.GamesLost
	diffB := b.GamesWon - b.GamesLost
	if diffA != diffB {
		if diffA > diffB {
			return -1
		}
		return 1
	}
	if a.GamesWon != b.GamesWon {
		if a.GamesWon > b.GamesWon {
			return -1
		}
		return 1
	}
	if a.PointsScored != b.PointsScored {
		if a.PointsScored > b.PointsScored {
			return -1
		}
		return 1
	}
	return 0
}

func headToHeadWinner(aID, bID int, matches []teamMatchResult) int {
	for _, m := range matches {
		if !m.Played || m.WinnerTeamID == 0 {
			continue
		}
		if (m.TeamAID == aID && m.TeamBID == bID) || (m.TeamAID == bID && m.TeamBID == aID) {
			return m.WinnerTeamID
		}
	}
	return 0
}

type miniStanding struct {
	points       int
	fieldsWon    int // 场（小场）胜
	fieldsLost   int
	gamesWon     int
	gamesLost    int
	pointsScored int
}

func miniStats(teamID int, tiedIDs []int, matches []teamMatchResult, games []gameScoreLine) miniStanding {
	tiedSet := map[int]bool{}
	for _, id := range tiedIDs {
		tiedSet[id] = true
	}
	var ms miniStanding
	for _, m := range matches {
		if !m.Played || m.WinnerTeamID == 0 {
			continue
		}
		if !tiedSet[m.TeamAID] || !tiedSet[m.TeamBID] {
			continue
		}
		if m.TeamAID != teamID && m.TeamBID != teamID {
			continue
		}
		if m.WinnerTeamID == teamID {
			ms.points += 2
		} else {
			ms.points += 1
		}
		if m.TeamAID == teamID {
			ms.fieldsWon += m.TeamAWins
			ms.fieldsLost += m.TeamBWins
		} else {
			ms.fieldsWon += m.TeamBWins
			ms.fieldsLost += m.TeamAWins
		}
	}
	for _, g := range games {
		if !tiedSet[g.TeamAID] || !tiedSet[g.TeamBID] {
			continue
		}
		if g.TeamAID != teamID && g.TeamBID != teamID {
			continue
		}
		if g.TeamAID == teamID {
			ms.pointsScored += g.ScoreA
			if g.ScoreA > g.ScoreB {
				ms.gamesWon++
			} else if g.ScoreB > g.ScoreA {
				ms.gamesLost++
			}
		} else {
			ms.pointsScored += g.ScoreB
			if g.ScoreB > g.ScoreA {
				ms.gamesWon++
			} else if g.ScoreA > g.ScoreB {
				ms.gamesLost++
			}
		}
	}
	return ms
}
