package models

import "time"

type Player struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	InitialRating int       `json:"initial_rating"`
	CurrentRating int       `json:"current_rating"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreatePlayerRequest struct {
	Name          string `json:"name"`
	InitialRating int    `json:"initial_rating"`
}
