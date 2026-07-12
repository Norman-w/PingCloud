package handlers

import (
	"net/http"

	"pingpong/db"
	"pingpong/models"
)

// GetSkills returns all skills in the skill library, grouped by category
func GetSkills(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(
		`SELECT id, category, name, COALESCE(style,''), sort_order, created_at
		 FROM skills ORDER BY category, sort_order`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	skills := make([]models.Skill, 0)
	for rows.Next() {
		var s models.Skill
		if err := rows.Scan(&s.ID, &s.Category, &s.Name, &s.Style, &s.SortOrder, &s.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		skills = append(skills, s)
	}

	writeJSON(w, skills)
}
