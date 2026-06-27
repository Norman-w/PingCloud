package rating

import "math"

// usattTable maps rating difference ranges to (expected_win, upset_win) changes.
// This is the original USATT lookup table used by 开球网.
// Key: [minDiff, maxDiff] → [high_win_change, low_win_change]
var usattTable = []struct {
	minDiff, maxDiff int
	highWinChange    int // higher-rated player wins
	lowWinChange     int // lower-rated player wins (upset)
}{
	{0, 12, 8, 8},
	{13, 37, 7, 12},
	{38, 62, 6, 16},
	{63, 87, 5, 20},
	{88, 112, 4, 25},
	{113, 137, 3, 30},
	{138, 162, 2, 35},
	{163, 187, 2, 40},
	{188, 212, 1, 45},
	{213, 237, 1, 50},
	{238, 9999, 0, 50},
}

// PlayerK returns the K-factor multiplier for a player.
// Follows USATT/开球网 standard: 40(new)/20(regular)/10(elite).
func PlayerK(matches int, rating int) float64 {
	if rating >= 2400 {
		return 0.5 // K=10 → 0.5× table value
	}
	if matches < 30 {
		return 2.0 // K=40 → 2.0× table value
	}
	return 1.0 // K=20 → 1.0× table value (baseline)
}

// CalculateRatingChanges computes USATT lookup-table based rating changes.
// Base table is for K=20. Other K values scale proportionally.
// Returns (change_for_a, change_for_b, winner_id).
func CalculateRatingChanges(ratingA, ratingB int, kA, kB float64, scoreA, scoreB int) (int, int, int) {
	diff := int(math.Abs(float64(ratingA - ratingB)))

	// Look up base changes from USATT table
	var highWin, lowWin int
	for _, row := range usattTable {
		if diff >= row.minDiff && diff <= row.maxDiff {
			highWin = row.highWinChange
			lowWin = row.lowWinChange
			break
		}
	}
	if lowWin == 0 {
		lowWin = 50
		highWin = 0
	}

	// Determine who is higher rated
	higherIsA := ratingA >= ratingB

	// Determine winner
	winnerID := 0 // 0=A wins, 1=B wins
	if scoreB > scoreA {
		winnerID = 1
	}

	// Determine base change value
	base := highWin // default: higher-rated won (small change)
	if (winnerID == 0 && !higherIsA) || (winnerID == 1 && higherIsA) {
		base = lowWin // lower-rated won (big upset change)
	}

	// Winner gets +change, loser gets -change
	var changeA, changeB int
	if winnerID == 0 {
		changeA = int(math.Round(float64(base) * kA))
		changeB = -int(math.Round(float64(base) * kB))
	} else {
		changeB = int(math.Round(float64(base) * kB))
		changeA = -int(math.Round(float64(base) * kA))
	}

	return changeA, changeB, winnerID
}
