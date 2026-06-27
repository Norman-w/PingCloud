package rating

import "math"

// PlayerK returns the K-factor for a player based on their match count and rating.
// Follows USATT/kaiqiuwang standard:
//   K=40 for new players (<30 matches) — fast convergence
//   K=20 for regular players (30+ matches) — normal
//   K=10 for elite players (2400+) — stability
func PlayerK(matches int, rating int) int {
	if rating >= 2400 {
		return 10
	}
	if matches < 30 {
		return 40
	}
	return 20
}

// CalculateRatingChanges computes USATT-style rating changes for a match.
// Each player uses their own K-factor.
// Returns (change_for_a, change_for_b, winner_id).
func CalculateRatingChanges(ratingA, ratingB, kA, kB int, scoreA, scoreB int) (int, int, int) {
	// Determine winner
	winnerID := 0 // 0 means A wins, 1 means B wins
	actualA := 1.0
	actualB := 0.0
	if scoreB > scoreA {
		winnerID = 1
		actualA = 0.0
		actualB = 1.0
	}

	// Expected win rate for player A
	expectedA := 1.0 / (1.0 + math.Pow(10, float64(ratingB-ratingA)/400.0))
	expectedB := 1.0 - expectedA

	changeA := int(math.Round(float64(kA) * (actualA - expectedA)))
	changeB := int(math.Round(float64(kB) * (actualB - expectedB)))

	return changeA, changeB, winnerID
}
