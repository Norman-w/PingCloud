package models

import "time"

type Match struct {
	ID             int       `json:"id"`
	PlayerAID      int       `json:"player_a_id"`
	PlayerBID      int       `json:"player_b_id"`
	PlayerAName    string    `json:"player_a_name"`
	PlayerBName    string    `json:"player_b_name"`
	ScoreA         int       `json:"score_a"`
	ScoreB         int       `json:"score_b"`
	RatingChangeA  int       `json:"rating_change_a"`
	RatingChangeB  int       `json:"rating_change_b"`
	WinnerID       int       `json:"winner_id"`
	PlayedAt       time.Time `json:"played_at"`
}

type CreateMatchRequest struct {
	PlayerAID int `json:"player_a_id"`
	PlayerBID int `json:"player_b_id"`
	ScoreA    int `json:"score_a"`
	ScoreB    int `json:"score_b"`
}
