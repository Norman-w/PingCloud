package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"pingpong/db"
	"pingpong/models"
)

// getPlayerIDFromCookie extracts the player ID from the ping_id cookie
func getPlayerIDFromCookie(r *http.Request) int {
	c, err := r.Cookie("ping_id")
	if err != nil || c.Value == "" {
		return 0
	}
	var pid int
	fmt.Sscanf(c.Value, "%d:", &pid)
	return pid
}

// GetTrainingLogs returns training logs for the current user
func GetTrainingLogs(w http.ResponseWriter, r *http.Request) {
	pid := getPlayerIDFromCookie(r)
	if pid == 0 {
		writeJSON(w, []models.TrainingLog{})
		return
	}

	rows, err := db.DB.Query(
		`SELECT id, player_id, date, duration_minutes, COALESCE(location,''), COALESCE(partner,''),
		        energy_rating, feel_rating, COALESCE(notes,''), created_at
		 FROM training_logs WHERE player_id = $1 ORDER BY date DESC, created_at DESC`, pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	logs := make([]models.TrainingLog, 0)
	for rows.Next() {
		var l models.TrainingLog
		if err := rows.Scan(&l.ID, &l.PlayerID, &l.Date, &l.DurationMinutes, &l.Location,
			&l.Partner, &l.EnergyRating, &l.FeelRating, &l.Notes, &l.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		l.Skills = loadLogSkills(l.ID)
		logs = append(logs, l)
	}

	writeJSON(w, logs)
}

// GetTrainingLog returns a single training log by ID
func GetTrainingLog(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/training-logs/")
	idStr = strings.Split(idStr, "/")[0] // remove any sub-paths
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var l models.TrainingLog
	err = db.DB.QueryRow(
		`SELECT id, player_id, date, duration_minutes, COALESCE(location,''), COALESCE(partner,''),
		        energy_rating, feel_rating, COALESCE(notes,''), created_at
		 FROM training_logs WHERE id = $1`, id,
	).Scan(&l.ID, &l.PlayerID, &l.Date, &l.DurationMinutes, &l.Location,
		&l.Partner, &l.EnergyRating, &l.FeelRating, &l.Notes, &l.CreatedAt)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	l.Skills = loadLogSkills(l.ID)

	writeJSON(w, l)
}

// CreateTrainingLog creates a new training log with associated skills
func CreateTrainingLog(w http.ResponseWriter, r *http.Request) {
	pid := getPlayerIDFromCookie(r)
	if pid == 0 {
		http.Error(w, "请先登录", http.StatusUnauthorized)
		return
	}

	var req models.CreateTrainingLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Date == "" {
		http.Error(w, "date is required", http.StatusBadRequest)
		return
	}
	if req.DurationMinutes <= 0 {
		http.Error(w, "duration_minutes is required", http.StatusBadRequest)
		return
	}

	// Clamp ratings
	if req.EnergyRating < 0 { req.EnergyRating = 0 }
	if req.EnergyRating > 5 { req.EnergyRating = 5 }
	if req.FeelRating < 0 { req.FeelRating = 0 }
	if req.FeelRating > 5 { req.FeelRating = 5 }

	var l models.TrainingLog
	err := db.DB.QueryRow(
		`INSERT INTO training_logs (player_id, date, duration_minutes, location, partner, energy_rating, feel_rating, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, player_id, date, duration_minutes, COALESCE(location,''), COALESCE(partner,''),
		           energy_rating, feel_rating, COALESCE(notes,''), created_at`,
		pid, req.Date, req.DurationMinutes, req.Location, req.Partner,
		req.EnergyRating, req.FeelRating, req.Notes,
	).Scan(&l.ID, &l.PlayerID, &l.Date, &l.DurationMinutes, &l.Location,
		&l.Partner, &l.EnergyRating, &l.FeelRating, &l.Notes, &l.CreatedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert skill associations
	for _, sk := range req.Skills {
		_, err := db.DB.Exec(
			`INSERT INTO training_log_skills (training_log_id, skill_id, practice_amount, notes)
			 VALUES ($1, $2, $3, $4)`,
			l.ID, sk.SkillID, sk.PracticeAmount, sk.Notes)
		if err != nil {
			// Log but don't fail the whole request
			fmt.Printf("Warning: failed to insert skill %d for log %d: %v\n", sk.SkillID, l.ID, err)
		}
	}

	l.Skills = loadLogSkills(l.ID)
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, l)
}

// UpdateTrainingLog updates an existing training log
func UpdateTrainingLog(w http.ResponseWriter, r *http.Request) {
	pid := getPlayerIDFromCookie(r)
	if pid == 0 {
		http.Error(w, "请先登录", http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/training-logs/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	// Verify ownership
	var ownerID int
	if err := db.DB.QueryRow("SELECT player_id FROM training_logs WHERE id = $1", id).Scan(&ownerID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if ownerID != pid {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var req models.CreateTrainingLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.EnergyRating < 0 { req.EnergyRating = 0 }
	if req.EnergyRating > 5 { req.EnergyRating = 5 }
	if req.FeelRating < 0 { req.FeelRating = 0 }
	if req.FeelRating > 5 { req.FeelRating = 5 }

	_, err = db.DB.Exec(
		`UPDATE training_logs SET date=$1, duration_minutes=$2, location=$3, partner=$4,
		 energy_rating=$5, feel_rating=$6, notes=$7 WHERE id=$8`,
		req.Date, req.DurationMinutes, req.Location, req.Partner,
		req.EnergyRating, req.FeelRating, req.Notes, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Replace skills: delete old, insert new
	db.DB.Exec("DELETE FROM training_log_skills WHERE training_log_id = $1", id)
	for _, sk := range req.Skills {
		db.DB.Exec(
			`INSERT INTO training_log_skills (training_log_id, skill_id, practice_amount, notes)
			 VALUES ($1, $2, $3, $4)`,
			id, sk.SkillID, sk.PracticeAmount, sk.Notes)
	}

	// Return updated log
	var l models.TrainingLog
	db.DB.QueryRow(
		`SELECT id, player_id, date, duration_minutes, COALESCE(location,''), COALESCE(partner,''),
		        energy_rating, feel_rating, COALESCE(notes,''), created_at
		 FROM training_logs WHERE id = $1`, id,
	).Scan(&l.ID, &l.PlayerID, &l.Date, &l.DurationMinutes, &l.Location,
		&l.Partner, &l.EnergyRating, &l.FeelRating, &l.Notes, &l.CreatedAt)
	l.Skills = loadLogSkills(l.ID)

	writeJSON(w, l)
}

// DeleteTrainingLog deletes a training log
func DeleteTrainingLog(w http.ResponseWriter, r *http.Request) {
	pid := getPlayerIDFromCookie(r)
	if pid == 0 {
		http.Error(w, "请先登录", http.StatusUnauthorized)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/training-logs/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	// Verify ownership
	var ownerID int
	if err := db.DB.QueryRow("SELECT player_id FROM training_logs WHERE id = $1", id).Scan(&ownerID); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if ownerID != pid {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	db.DB.Exec("DELETE FROM training_log_skills WHERE training_log_id = $1", id)
	db.DB.Exec("DELETE FROM training_logs WHERE id = $1", id)

	w.WriteHeader(http.StatusNoContent)
}

// GetTrainingStats returns aggregated statistics for the current user
func GetTrainingStats(w http.ResponseWriter, r *http.Request) {
	pid := getPlayerIDFromCookie(r)
	if pid == 0 {
		writeJSON(w, models.TrainingStats{SkillFrequencies: make([]models.SkillFrequency, 0)})
		return
	}

	var stats models.TrainingStats
	stats.SkillFrequencies = make([]models.SkillFrequency, 0)

	db.DB.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(duration_minutes), 0) FROM training_logs WHERE player_id = $1`, pid,
	).Scan(&stats.TotalSessions, &stats.TotalMinutes)

	db.DB.QueryRow(
		`SELECT COUNT(*) FROM training_logs
		 WHERE player_id = $1 AND date >= date_trunc('month', CURRENT_DATE)::text`, pid,
	).Scan(&stats.ThisMonthSessions)

	// Skill frequencies
	rows, err := db.DB.Query(
		`SELECT s.id, s.name, s.category, COUNT(*) as cnt
		 FROM training_log_skills tls
		 JOIN skills s ON s.id = tls.skill_id
		 JOIN training_logs tl ON tl.id = tls.training_log_id
		 WHERE tl.player_id = $1
		 GROUP BY s.id, s.name, s.category
		 ORDER BY cnt DESC LIMIT 20`, pid)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var sf models.SkillFrequency
			if err := rows.Scan(&sf.SkillID, &sf.SkillName, &sf.Category, &sf.Count); err == nil {
				stats.SkillFrequencies = append(stats.SkillFrequencies, sf)
			}
		}
	}

	writeJSON(w, stats)
}

// loadLogSkills loads the skills associated with a training log
func loadLogSkills(logID int) []models.TrainingLogSkill {
	rows, err := db.DB.Query(
		`SELECT s.id, s.name, s.category, COALESCE(tls.practice_amount,''), COALESCE(tls.notes,'')
		 FROM training_log_skills tls
		 JOIN skills s ON s.id = tls.skill_id
		 WHERE tls.training_log_id = $1
		 ORDER BY s.category, s.sort_order`, logID)
	if err != nil {
		return []models.TrainingLogSkill{}
	}
	defer rows.Close()

	skills := make([]models.TrainingLogSkill, 0)
	for rows.Next() {
		var s models.TrainingLogSkill
		if err := rows.Scan(&s.SkillID, &s.SkillName, &s.Category, &s.PracticeAmount, &s.Notes); err == nil {
			skills = append(skills, s)
		}
	}
	if skills == nil {
		skills = []models.TrainingLogSkill{}
	}
	return skills
}
