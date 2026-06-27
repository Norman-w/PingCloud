package rating

import (
	"math"
	"os"
	"strconv"
)

// winnerBonus is a small point injection to prevent rating stagnation in closed groups.
// Set RATING_BONUS env var to 1 (default 0 = USATT standard).
var winnerBonus = func() int {
	if v := os.Getenv("RATING_BONUS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 0
}()

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

	// Inject winner bonus (prevent closed-system rating stagnation)
	if winnerBonus > 0 {
		if winnerID == 0 {
			changeA += winnerBonus
		} else {
			changeB += winnerBonus
		}
	}

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
