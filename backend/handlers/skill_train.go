package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"pingpong/db"
	"pingpong/models"
)

// GetSkillTrainHistory returns skill info + training history for this skill
func GetSkillTrainHistory(w http.ResponseWriter, r *http.Request) {
	pid := getPlayerIDFromCookie(r)

	// Extract skill ID from path: /api/skill-train/123
	idStr := strings.TrimPrefix(r.URL.Path, "/api/skill-train/")
	skillID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid skill id", http.StatusBadRequest)
		return
	}

	// Fetch skill info
	var skillName, skillCat string
	if err := db.DB.QueryRow(`SELECT name, category FROM skills WHERE id = $1`, skillID).Scan(&skillName, &skillCat); err != nil {
		http.Error(w, "skill not found", http.StatusNotFound)
		return
	}

	// Build response
	result := map[string]interface{}{
		"skill_id":   skillID,
		"skill_name": skillName,
		"category":   skillCat,
		"history":    []map[string]interface{}{},
	}

	if pid == 0 {
		writeJSON(w, result)
		return
	}

	// Fetch training history for this skill
	rows, err := db.DB.Query(
		`SELECT tl.id, tl.date, tl.duration_minutes, tl.location, tl.partner,
		        tl.energy_rating, tl.feel_rating, tl.notes, tl.created_at,
		        COALESCE(tls.practice_amount, ''), COALESCE(tls.notes, ''),
		        COALESCE(tls.indicators::text, '{}')
		 FROM training_log_skills tls
		 JOIN training_logs tl ON tl.id = tls.training_log_id
		 WHERE tls.skill_id = $1 AND tl.player_id = $2
		 ORDER BY tl.date DESC, tl.created_at DESC`, skillID, pid)
	if err == nil {
		defer rows.Close()
		history := make([]map[string]interface{}, 0)
		for rows.Next() {
			var id, dur, energy, feel int
			var date, loc, partner, notes, createdAt, amount, skillNotes, indicatorsStr string
			if err := rows.Scan(&id, &date, &dur, &loc, &partner, &energy, &feel, &notes, &createdAt, &amount, &skillNotes, &indicatorsStr); err == nil {
				entry := map[string]interface{}{
					"id":               id,
					"date":             date,
					"duration_minutes": dur,
					"location":         loc,
					"partner":          partner,
					"energy_rating":    energy,
					"feel_rating":      feel,
					"notes":            notes,
					"created_at":       createdAt,
					"practice_amount":  amount,
					"skill_notes":      skillNotes,
				}
				// Parse indicators JSON
				var indicators map[string]interface{}
				if json.Unmarshal([]byte(indicatorsStr), &indicators) == nil {
					entry["indicators"] = indicators
				} else {
					entry["indicators"] = map[string]interface{}{}
				}
				history = append(history, entry)
			}
		}
		result["history"] = history
	}

	writeJSON(w, result)
}

// CreateSkillTraining creates a training log with indicators for a specific skill
func CreateSkillTraining(w http.ResponseWriter, r *http.Request) {
	pid := getPlayerIDFromCookie(r)
	if pid == 0 {
		http.Error(w, "请先登录", http.StatusUnauthorized)
		return
	}

	var req struct {
		SkillID         int                    `json:"skill_id"`
		Date            string                 `json:"date"`
		DurationMinutes int                    `json:"duration_minutes"`
		Location        string                 `json:"location"`
		Partner         string                 `json:"partner"`
		EnergyRating    int                    `json:"energy_rating"`
		FeelRating      int                    `json:"feel_rating"`
		Notes           string                 `json:"notes"`
		PracticeAmount  string                 `json:"practice_amount"`
		SkillNotes      string                 `json:"skill_notes"`
		Indicators      map[string]interface{} `json:"indicators"`
		GoalValues      map[string]interface{} `json:"goal_values"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if req.SkillID == 0 || req.Date == "" || req.DurationMinutes <= 0 {
		http.Error(w, "skill_id, date, duration_minutes required", http.StatusBadRequest)
		return
	}

	// Clamp ratings
	if req.EnergyRating < 0 { req.EnergyRating = 0 }
	if req.EnergyRating > 5 { req.EnergyRating = 5 }
	if req.FeelRating < 0 { req.FeelRating = 0 }
	if req.FeelRating > 5 { req.FeelRating = 5 }

	// Marshal indicators to JSON string
	indicatorsJSON := "{}"
	if req.Indicators != nil && len(req.Indicators) > 0 {
		if b, err := json.Marshal(req.Indicators); err == nil {
			indicatorsJSON = string(b)
		}
	}

	// Create training log
	var logID int
	err := db.DB.QueryRow(
		`INSERT INTO training_logs (player_id, date, duration_minutes, location, partner, energy_rating, feel_rating, notes)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		pid, req.Date, req.DurationMinutes, req.Location, req.Partner,
		req.EnergyRating, req.FeelRating, req.Notes,
	).Scan(&logID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Marshal goal values
	goalValuesJSON := "{}"
	if req.GoalValues != nil && len(req.GoalValues) > 0 {
		if b, err := json.Marshal(req.GoalValues); err == nil {
			goalValuesJSON = string(b)
		}
	}

	// Link skill with indicators and goal values
	_, err = db.DB.Exec(
		`INSERT INTO training_log_skills (training_log_id, skill_id, practice_amount, notes, indicators, goal_values)
		 VALUES ($1,$2,$3,$4,$5::jsonb,$6::jsonb)`,
		logID, req.SkillID, req.PracticeAmount, req.SkillNotes, indicatorsJSON, goalValuesJSON)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build response
	var log models.TrainingLog
	db.DB.QueryRow(
		`SELECT id, player_id, date, duration_minutes, COALESCE(location,''), COALESCE(partner,''),
		        energy_rating, feel_rating, COALESCE(notes,''), created_at
		 FROM training_logs WHERE id = $1`, logID,
	).Scan(&log.ID, &log.PlayerID, &log.Date, &log.DurationMinutes, &log.Location,
		&log.Partner, &log.EnergyRating, &log.FeelRating, &log.Notes, &log.CreatedAt)

	writeJSON(w, log)
}
