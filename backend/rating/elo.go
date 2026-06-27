package rating

import "math"

// CalculateRatingChanges computes USATT-style rating changes for a match.
// Returns (change_for_a, change_for_b, winner_id).
func CalculateRatingChanges(ratingA, ratingB, scoreA, scoreB int) (int, int, int) {
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

	// Dynamic K-factor based on rating difference
	kFactor := kFactorDynamic(ratingA, ratingB, actualA > 0.5)

	changeA := int(math.Round(kFactor * (actualA - expectedA)))
	changeB := int(math.Round(kFactor * (actualB - expectedB)))

	return changeA, changeB, winnerID
}

func kFactorDynamic(ratingA, ratingB int, isUpset bool) float64 {
	diff := math.Abs(float64(ratingA - ratingB))
	baseK := 32.0

	// Reduce K for large rating differences
	if diff > 200 {
		baseK = 24.0
	}
	if diff > 400 {
		baseK = 16.0
	}

	// Boost K slightly for upsets (underdog wins)
	if isUpset && diff > 100 {
		baseK *= 1.2
	}

	return baseK
}
