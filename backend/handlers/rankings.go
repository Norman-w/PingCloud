package handlers

import (
	"net/http"

	"pingpong/db"
	"pingpong/models"
)

func GetRankings(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(
		`SELECT p.id, p.name, p.initial_rating, p.current_rating, p.created_at,
		        COUNT(m.id) AS matches_played,
		        COALESCE(SUM(CASE WHEN m.winner_id = p.id THEN 1 ELSE 0 END), 0) AS wins,
		        COALESCE(SUM(CASE WHEN m.winner_id != p.id THEN 1 ELSE 0 END), 0) AS losses
		 FROM players p
		 LEFT JOIN matches m ON m.player_a_id = p.id OR m.player_b_id = p.id
		 GROUP BY p.id
		 ORDER BY p.current_rating DESC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type RankingEntry struct {
		models.Player
		MatchesPlayed int `json:"matches_played"`
		Wins          int `json:"wins"`
		Losses        int `json:"losses"`
		WinRate       float64 `json:"win_rate"`
	}

	rankings := make([]RankingEntry, 0)
	for rows.Next() {
		var re RankingEntry
		if err := rows.Scan(&re.ID, &re.Name, &re.InitialRating, &re.CurrentRating, &re.CreatedAt,
			&re.MatchesPlayed, &re.Wins, &re.Losses); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if re.MatchesPlayed > 0 {
			re.WinRate = float64(re.Wins) / float64(re.MatchesPlayed) * 100
		}
		rankings = append(rankings, re)
	}

	writeJSON(w, rankings)
}
