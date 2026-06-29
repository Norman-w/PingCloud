package handlers

import (
	"net/http"

	"pingpong/db"
)

func GetHeadToHead(w http.ResponseWriter, r *http.Request) {
	// Get all players ordered by rating
	rows, err := db.DB.Query(`SELECT id, name FROM players ORDER BY current_rating DESC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type H2HRecord struct {
		OpponentID   int    `json:"opponent_id"`
		OpponentName string `json:"opponent_name"`
		Wins         int    `json:"wins"`
		Losses       int    `json:"losses"`
	}

	type PlayerRow struct {
		ID      int         `json:"id"`
		Name    string      `json:"name"`
		Records []H2HRecord `json:"records"`
	}

	var players []PlayerRow
	for rows.Next() {
		var p PlayerRow
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			continue
		}
		players = append(players, p)
	}

	// For each player, get records against all other players
	for i := range players {
		players[i].Records = make([]H2HRecord, 0)

		for j := range players {
			if players[j].ID == players[i].ID {
				continue
			}
			oppID := players[j].ID

			var wins, losses int
			db.DB.QueryRow(`
				SELECT
					COALESCE(SUM(CASE WHEN m.winner_id = $1 THEN 1 ELSE 0 END), 0),
					COALESCE(SUM(CASE WHEN m.winner_id IS NOT NULL AND m.winner_id != $1 THEN 1 ELSE 0 END), 0)
				FROM matches m
				WHERE ((m.player_a_id = $1 AND m.player_b_id = $2) OR (m.player_a_id = $2 AND m.player_b_id = $1))
					AND m.score_a IS NOT NULL AND m.forfeit = false
			`, players[i].ID, oppID).Scan(&wins, &losses)

			players[i].Records = append(players[i].Records, H2HRecord{
				OpponentID:   oppID,
				OpponentName: players[j].Name,
				Wins:         wins,
				Losses:       losses,
			})
		}
	}

	writeJSON(w, players)
}
