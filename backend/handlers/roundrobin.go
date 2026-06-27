package handlers

import (
	"database/sql"
	"sort"

	"pingpong/db"
)

// sortPlayersByRating returns player IDs sorted by current_rating descending.
func sortPlayersByRating(ids []int) []int {
	if len(ids) <= 1 {
		return ids
	}
	sorted := make([]int, len(ids))
	copy(sorted, ids)
	// Fetch ratings and sort
	type pr struct {
		id     int
		rating int
	}
	items := make([]pr, len(ids))
	for i, id := range ids {
		items[i].id = id
		db.DB.QueryRow("SELECT current_rating FROM players WHERE id=$1", id).Scan(&items[i].rating)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].rating > items[j].rating })
	for i, item := range items {
		sorted[i] = item.id
	}
	return sorted
}

// generateRoundRobin creates a round-robin schedule using the circle method.
// Players are seeded by rating (highest first) so top players meet in later rounds.
// If the number of players is odd, a dummy (0 = bye) is added.
func generateRoundRobin(tx *sql.Tx, sessionID int, playerIDs []int) error {
	// Sort by rating for seeding
	playerIDs = sortPlayersByRating(playerIDs)
	n := len(playerIDs)
	if n < 2 {
		return nil
	}

	ids := make([]int, n)
	copy(ids, playerIDs)

	// If odd, add a dummy (0 = bye)
	if n%2 != 0 {
		ids = append(ids, 0)
		n++
	}

	half := n / 2
	roundNum := 1

	for r := 0; r < n-1; r++ {
		for i := 0; i < half; i++ {
			a := ids[i]
			b := ids[n-1-i]

			if a == 0 || b == 0 {
				continue
			}

			// Alternate home/away
			if i%2 == 1 {
				a, b = b, a
			}

			_, err := tx.Exec(`
				INSERT INTO matches (player_a_id, player_b_id, session_id, round)
				VALUES ($1, $2, $3, $4)`,
				a, b, sessionID, roundNum,
			)
			if err != nil {
				return err
			}
		}

		// Rotate: fix first element, shift the rest
		last := ids[n-1]
		for i := n - 1; i > 1; i-- {
			ids[i] = ids[i-1]
		}
		ids[1] = last

		roundNum++
	}

	return nil
}

// regenerateUnplayed creates a fresh round-robin schedule for all players,
// starting from round offset+1 (to continue after already-played rounds).
func regenerateUnplayed(tx *sql.Tx, sessionID int, playerIDs []int, offset int) error {
	// Sort by rating for seeding
	playerIDs = sortPlayersByRating(playerIDs)
	n := len(playerIDs)
	if n < 2 {
		return nil
	}

	ids := make([]int, n)
	copy(ids, playerIDs)

	if n%2 != 0 {
		ids = append(ids, 0)
		n++
	}

	half := n / 2
	roundNum := offset + 1

	for r := 0; r < n-1; r++ {
		for i := 0; i < half; i++ {
			a := ids[i]
			b := ids[n-1-i]

			if a == 0 || b == 0 {
				continue
			}

			// Skip if this pair already played in this session
			var already int
			tx.QueryRow(`SELECT COUNT(*) FROM matches WHERE session_id=$1 AND
				((player_a_id=$2 AND player_b_id=$3) OR (player_a_id=$3 AND player_b_id=$2))
				AND score_a IS NOT NULL`, sessionID, a, b).Scan(&already)
			if already > 0 {
				continue
			}

			if i%2 == 1 {
				a, b = b, a
			}

			_, err := tx.Exec(`
				INSERT INTO matches (player_a_id, player_b_id, session_id, round)
				VALUES ($1, $2, $3, $4)`,
				a, b, sessionID, roundNum,
			)
			if err != nil {
				return err
			}
		}

		last := ids[n-1]
		for i := n - 1; i > 1; i-- {
			ids[i] = ids[i-1]
		}
		ids[1] = last

		roundNum++
	}

	return nil
}
