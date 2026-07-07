package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"pingpong/db"
)

type TeamBattleSummary struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Mode      string `json:"mode"`
	GroupA    string `json:"group_a_name"`
	GroupB    string `json:"group_b_name"`
	Status    string `json:"status"`
	AWins     int    `json:"a_wins"`
	BWins     int    `json:"b_wins"`
	CreatedAt string `json:"created_at"`
}

type TBPlayer struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	CurrentRating int    `json:"current_rating"`
	Team          string `json:"team"`
}

type TBMatch struct {
	ID        int    `json:"id"`
	MatchType string `json:"match_type"`
	A1ID      int    `json:"a1_id"`
	A2ID      *int   `json:"a2_id"`
	B1ID      int    `json:"b1_id"`
	B2ID      *int   `json:"b2_id"`
	A1Name    string `json:"a1_name"`
	A2Name    string `json:"a2_name"`
	B1Name    string `json:"b1_name"`
	B2Name    string `json:"b2_name"`
	ScoreA    *int   `json:"score_a"`
	ScoreB    *int   `json:"score_b"`
	Winner    string `json:"winner_team"`
	Played    bool   `json:"played"`
}

type TeamBattleDetail struct {
	ID      int        `json:"id"`
	Name    string     `json:"name"`
	Mode    string     `json:"mode"`
	GroupA  string     `json:"group_a_name"`
	GroupB  string     `json:"group_b_name"`
	Status  string     `json:"status"`
	AWins   int        `json:"a_wins"`
	BWins   int        `json:"b_wins"`
	Players []TBPlayer `json:"players"`
	Matches []TBMatch  `json:"matches"`
}

func CreateTeamBattle(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name            string `json:"name"`
		GroupAName      string `json:"group_a_name"`
		GroupBName      string `json:"group_b_name"`
		GroupAPlayerIDs []int  `json:"group_a_player_ids"`
		GroupBPlayerIDs []int  `json:"group_b_player_ids"`
		SinglesOn       bool   `json:"singles_on"`
		DoublesOn       bool   `json:"doubles_on"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		req.Name = "团体对抗赛"
	}
	if req.GroupAName == "" {
		req.GroupAName = "A组"
	}
	if req.GroupBName == "" {
		req.GroupBName = "B组"
	}
	if !req.SinglesOn && !req.DoublesOn {
		req.SinglesOn = true
	}

	if len(req.GroupAPlayerIDs) == 0 || len(req.GroupBPlayerIDs) == 0 {
		http.Error(w, "每组至少需要1名选手", http.StatusBadRequest)
		return
	}

	// Determine mode string
	mode := "singles"
	if req.DoublesOn && req.SinglesOn {
		mode = "mixed"
	} else if req.DoublesOn {
		mode = "doubles"
	}

	tx, err := db.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var id int
	err = tx.QueryRow(`INSERT INTO team_battles (name, mode, group_a_name, group_b_name, deleted) VALUES ($1,$2,$3,$4,false) RETURNING id`,
		req.Name, mode, req.GroupAName, req.GroupBName).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, pid := range req.GroupAPlayerIDs {
		_, err = tx.Exec(`INSERT INTO team_battle_players (battle_id, player_id, team) VALUES ($1,$2,'A')`, id, pid)
		if err != nil {
			http.Error(w, fmt.Sprintf("添加A组选手失败: %v", err), http.StatusBadRequest)
			return
		}
	}
	for _, pid := range req.GroupBPlayerIDs {
		_, err = tx.Exec(`INSERT INTO team_battle_players (battle_id, player_id, team) VALUES ($1,$2,'B')`, id, pid)
		if err != nil {
			http.Error(w, fmt.Sprintf("添加B组选手失败: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Generate match schedule: round-robin ensuring each player plays
	a := req.GroupAPlayerIDs
	b := req.GroupBPlayerIDs
	an, bn := len(a), len(b)

	// Build match list (indices into a/b arrays)
	type matchPlan struct {
		mt       string
		ai1, ai2 int // indices into a; -1 = unused
		bi1, bi2 int // indices into b; -1 = unused
	}
	var plans []matchPlan

	total := an + bn
	for i := 0; i < total; i++ {
		if req.SinglesOn && !req.DoublesOn {
			// Singles only
			plans = append(plans, matchPlan{"singles", i % an, -1, i % bn, -1})
		} else if req.DoublesOn && an >= 2 && bn >= 2 && i%2 == 0 {
			// Doubles
			plans = append(plans, matchPlan{"doubles", i % an, (i + 1) % an, i % bn, (i + 1) % bn})
		} else {
			// Singles
			plans = append(plans, matchPlan{"singles", i % an, -1, i % bn, -1})
		}
	}

	// Shuffle for variety
	rand.Shuffle(len(plans), func(i, j int) { plans[i], plans[j] = plans[j], plans[i] })

	for _, p := range plans {
		var execErr error
		if p.mt == "doubles" {
			_, execErr = tx.Exec(`INSERT INTO team_battle_matches (battle_id, match_type, a1_id, a2_id, b1_id, b2_id) VALUES ($1,'doubles',$2,$3,$4,$5)`,
				id, a[p.ai1], a[p.ai2], b[p.bi1], b[p.bi2])
		} else {
			_, execErr = tx.Exec(`INSERT INTO team_battle_matches (battle_id, match_type, a1_id, b1_id) VALUES ($1,'singles',$2,$3)`,
				id, a[p.ai1], b[p.bi1])
		}
		if execErr != nil {
			http.Error(w, execErr.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]int{"id": id})
}

func GetTeamBattles(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`SELECT tb.id, tb.name, COALESCE(tb.mode,''), tb.group_a_name, tb.group_b_name, tb.status, tb.a_wins, tb.b_wins, tb.created_at FROM team_battles tb WHERE tb.deleted=false ORDER BY tb.created_at DESC LIMIT 20`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var list []TeamBattleSummary
	for rows.Next() {
		var s TeamBattleSummary
		if err := rows.Scan(&s.ID, &s.Name, &s.Mode, &s.GroupA, &s.GroupB, &s.Status, &s.AWins, &s.BWins, &s.CreatedAt); err != nil {
			continue
		}
		list = append(list, s)
	}
	if list == nil {
		list = []TeamBattleSummary{}
	}
	writeJSON(w, list)
}

func GetTeamBattle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/team-battles/"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var d TeamBattleDetail
	err = db.DB.QueryRow(`SELECT id, name, COALESCE(mode,''), group_a_name, group_b_name, status, a_wins, b_wins FROM team_battles WHERE id=$1 AND deleted=false`, id).
		Scan(&d.ID, &d.Name, &d.Mode, &d.GroupA, &d.GroupB, &d.Status, &d.AWins, &d.BWins)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Players
	pr, err := db.DB.Query(`SELECT p.id, p.name, p.current_rating, tbp.team FROM team_battle_players tbp JOIN players p ON p.id=tbp.player_id WHERE tbp.battle_id=$1 ORDER BY tbp.team, p.name`, id)
	if err == nil {
		defer pr.Close()
		for pr.Next() {
			var p TBPlayer
			if err := pr.Scan(&p.ID, &p.Name, &p.CurrentRating, &p.Team); err == nil {
				d.Players = append(d.Players, p)
			}
		}
	}
	if d.Players == nil {
		d.Players = []TBPlayer{}
	}

	// Matches
	mr, err := db.DB.Query(`SELECT m.id, m.match_type, m.a1_id, m.a2_id, m.b1_id, m.b2_id, COALESCE(pa1.name,''), COALESCE(pa2.name,''), COALESCE(pb1.name,''), COALESCE(pb2.name,''), m.score_a, m.score_b, COALESCE(m.winner_team,''), m.played FROM team_battle_matches m JOIN players pa1 ON pa1.id=m.a1_id JOIN players pb1 ON pb1.id=m.b1_id LEFT JOIN players pa2 ON pa2.id=m.a2_id LEFT JOIN players pb2 ON pb2.id=m.b2_id WHERE m.battle_id=$1 ORDER BY m.id`, id)
	if err == nil {
		defer mr.Close()
		for mr.Next() {
			var m TBMatch
			if err := mr.Scan(&m.ID, &m.MatchType, &m.A1ID, &m.A2ID, &m.B1ID, &m.B2ID, &m.A1Name, &m.A2Name, &m.B1Name, &m.B2Name, &m.ScoreA, &m.ScoreB, &m.Winner, &m.Played); err == nil {
				d.Matches = append(d.Matches, m)
			}
		}
	}
	if d.Matches == nil {
		d.Matches = []TBMatch{}
	}

	writeJSON(w, d)
}

func ScoreTeamBattleMatch(w http.ResponseWriter, r *http.Request) {
	// Path: /api/team-battles/{battleID}/matches/{matchID}
	path := strings.TrimPrefix(r.URL.Path, "/api/team-battles/")
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	battleID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "invalid battle id", http.StatusBadRequest)
		return
	}
	matchID, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, "invalid match id", http.StatusBadRequest)
		return
	}

	var req struct {
		ScoreA int `json:"score_a"`
		ScoreB int `json:"score_b"`
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

	// Undo old winner stats if previously scored
	var oldWinner string
	var wasPlayed bool
	tx.QueryRow(`SELECT played, COALESCE(winner_team,'') FROM team_battle_matches WHERE id=$1 FOR UPDATE`, matchID).Scan(&wasPlayed, &oldWinner)
	if wasPlayed && oldWinner != "" {
		if oldWinner == "A" {
			tx.Exec(`UPDATE team_battles SET a_wins=GREATEST(a_wins-1,0) WHERE id=$1`, battleID)
		} else if oldWinner == "B" {
			tx.Exec(`UPDATE team_battles SET b_wins=GREATEST(b_wins-1,0) WHERE id=$1`, battleID)
		}
	}

	winner := "B"
	if req.ScoreA > req.ScoreB {
		winner = "A"
	}

	_, err = tx.Exec(`UPDATE team_battle_matches SET score_a=$1, score_b=$2, winner_team=$3, played=true WHERE id=$4`,
		req.ScoreA, req.ScoreB, winner, matchID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if winner == "A" {
		tx.Exec(`UPDATE team_battles SET a_wins=a_wins+1 WHERE id=$1`, battleID)
	} else {
		tx.Exec(`UPDATE team_battles SET b_wins=b_wins+1 WHERE id=$1`, battleID)
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"winner_team": winner})
}

func CompleteTeamBattle(w http.ResponseWriter, r *http.Request) {
	// Path: /api/team-battles/{id}/complete
	path := strings.TrimPrefix(r.URL.Path, "/api/team-battles/")
	path = strings.TrimSuffix(path, "/complete")
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	// Check all matches are played
	var unplayed int
	db.DB.QueryRow(`SELECT COUNT(*) FROM team_battle_matches WHERE battle_id=$1 AND played=false`, id).Scan(&unplayed)
	if unplayed > 0 {
		http.Error(w, fmt.Sprintf("还有%d场未录入比分", unplayed), http.StatusBadRequest)
		return
	}

	_, err = db.DB.Exec(`UPDATE team_battles SET status='completed' WHERE id=$1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func DeleteTeamBattle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/team-battles/"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	_, err = db.DB.Exec(`UPDATE team_battles SET deleted=true WHERE id=$1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

var _ = rand.Int
