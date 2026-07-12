package models

import "time"

// Skill represents a table tennis skill in the skill library
type Skill struct {
	ID        int       `json:"id"`
	Category  string    `json:"category"` // 基本功 / 技战术
	Name      string    `json:"name"`
	Style     string    `json:"style"`    // 双反 etc.
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

// TrainingLog represents a single training session record
type TrainingLog struct {
	ID              int              `json:"id"`
	PlayerID        int              `json:"player_id"`
	Date            string           `json:"date"`
	DurationMinutes int              `json:"duration_minutes"`
	Location        string           `json:"location"`
	Partner         string           `json:"partner"`
	EnergyRating    int              `json:"energy_rating"` // 1-5
	FeelRating      int              `json:"feel_rating"`   // 1-5
	Notes           string           `json:"notes"`
	CreatedAt       time.Time        `json:"created_at"`
	Skills          []TrainingLogSkill `json:"skills"`
}

// TrainingLogSkill links a training log to a skill with practice details
type TrainingLogSkill struct {
	SkillID        int    `json:"skill_id"`
	SkillName      string `json:"skill_name"`
	Category       string `json:"category"`
	PracticeAmount string `json:"practice_amount"`
	Notes          string `json:"notes"`
}

// CreateTrainingLogRequest is the request body for creating a training log
type CreateTrainingLogRequest struct {
	Date            string                       `json:"date"`
	DurationMinutes int                          `json:"duration_minutes"`
	Location        string                       `json:"location"`
	Partner         string                       `json:"partner"`
	EnergyRating    int                          `json:"energy_rating"`
	FeelRating      int                          `json:"feel_rating"`
	Notes           string                       `json:"notes"`
	Skills          []CreateTrainingLogSkillInput `json:"skills"`
}

// CreateTrainingLogSkillInput is a skill entry in the create request
type CreateTrainingLogSkillInput struct {
	SkillID        int    `json:"skill_id"`
	PracticeAmount string `json:"practice_amount"`
	Notes          string `json:"notes"`
}

// TrainingStats holds aggregated training statistics
type TrainingStats struct {
	TotalSessions      int                  `json:"total_sessions"`
	TotalMinutes       int                  `json:"total_minutes"`
	ThisMonthSessions  int                  `json:"this_month_sessions"`
	SkillFrequencies   []SkillFrequency     `json:"skill_frequencies"`
}

// SkillFrequency tracks how often a skill was practiced
type SkillFrequency struct {
	SkillID   int    `json:"skill_id"`
	SkillName string `json:"skill_name"`
	Category  string `json:"category"`
	Count     int    `json:"count"`
}

// SkillMasteryItem represents a skill with mastery status and tags
type SkillMasteryItem struct {
	ID                    int      `json:"id"`
	Name                  string   `json:"name"`
	Category              string   `json:"category"`
	Tags                  []string `json:"tags"`
	PracticeCount         int      `json:"practice_count"`
	TotalDurationMinutes  int      `json:"total_duration_minutes"`
	LastPracticed         string   `json:"last_practiced"`
	Status                string   `json:"status"` // none | practicing | mastered
}

// SkillMasteryGroup groups skills by an attribute tag
type SkillMasteryGroup struct {
	Label  string             `json:"label"`
	Skills []SkillMasteryItem `json:"skills"`
}

// UpdateSkillMasteryRequest is the body for PUT /api/skill-mastery/:id
type UpdateSkillMasteryRequest struct {
	Status string `json:"status"`
}
