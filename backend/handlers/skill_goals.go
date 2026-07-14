package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"pingpong/db"
)

// GetSkillGoals returns goal definitions + accumulated progress for a skill
func GetSkillGoals(w http.ResponseWriter, r *http.Request) {
	pid := getPlayerIDFromCookie(r)
	idStr := strings.TrimPrefix(r.URL.Path, "/api/skill-goals/")
	skillID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid skill id", http.StatusBadRequest)
		return
	}

	// Fetch goal definitions
	rows, err := db.DB.Query(`SELECT id, skill_id, label, unit, tier_1, tier_2, tier_3, tier_4, tier_5, min_stars, sort_order
		FROM skill_goals WHERE skill_id=$1 ORDER BY sort_order`, skillID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type GoalDef struct {
		ID       int    `json:"id"`
		SkillID  int    `json:"skill_id"`
		Label    string `json:"label"`
		Unit     string `json:"unit"`
		Tier1    int    `json:"tier_1"`
		Tier2    int    `json:"tier_2"`
		Tier3    int    `json:"tier_3"`
		Tier4    int    `json:"tier_4"`
		Tier5    int    `json:"tier_5"`
		MinStars int    `json:"min_stars"`
	}

	type GoalProgress struct {
		GoalDef
		CurrentValue int  `json:"current_value"`
		Stars        int  `json:"stars"`
		Passed       bool `json:"passed"`
	}

	result := make([]GoalProgress, 0)

	for rows.Next() {
		var g GoalDef
		var sort int
		rows.Scan(&g.ID, &g.SkillID, &g.Label, &g.Unit, &g.Tier1, &g.Tier2, &g.Tier3, &g.Tier4, &g.Tier5, &g.MinStars, &sort)
		gp := GoalProgress{GoalDef: g}

		if pid > 0 {
			if g.Label == "累计时长" {
				db.DB.QueryRow(`SELECT COALESCE(SUM(tl.duration_minutes),0) FROM training_log_skills tls
					JOIN training_logs tl ON tl.id=tls.training_log_id
					WHERE tls.skill_id=$1 AND tl.player_id=$2`, skillID, pid).Scan(&gp.CurrentValue)
			} else {
				// Sum goal_values JSONB field across all sessions for this skill+player
				var total int
				jsonRows, err := db.DB.Query(`SELECT tls.goal_values FROM training_log_skills tls
					JOIN training_logs tl ON tl.id=tls.training_log_id
					WHERE tls.skill_id=$1 AND tl.player_id=$2`, skillID, pid)
				if err == nil {
					defer jsonRows.Close()
					for jsonRows.Next() {
						var gvStr string
						jsonRows.Scan(&gvStr)
						var gv map[string]interface{}
						if json.Unmarshal([]byte(gvStr), &gv) == nil {
							if v, ok := gv[g.Label]; ok {
								if n, ok := v.(float64); ok {
									total += int(n)
								}
							}
						}
					}
				}
				gp.CurrentValue = total
			}
		}

		// Calculate stars
		v := gp.CurrentValue
		if v >= g.Tier5 { gp.Stars = 5 } else if v >= g.Tier4 { gp.Stars = 4 } else if v >= g.Tier3 { gp.Stars = 3 } else if v >= g.Tier2 { gp.Stars = 2 } else if v >= g.Tier1 { gp.Stars = 1 }
		gp.Passed = gp.Stars >= g.MinStars
		result = append(result, gp)
	}

	writeJSON(w, result)
}

// GetGoalDefinitions returns raw goal definitions (for admin/frontend seeding)
func GetGoalDefinitions(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/skill-goals/def/")
	skillID, _ := strconv.Atoi(idStr)
	rows, err := db.DB.Query(`SELECT id, skill_id, label, unit, tier_1, tier_2, tier_3, tier_4, tier_5, min_stars, sort_order
		FROM skill_goals WHERE skill_id=$1 ORDER BY sort_order`, skillID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type GoalDef struct {
		ID       int    `json:"id"`
		SkillID  int    `json:"skill_id"`
		Label    string `json:"label"`
		Unit     string `json:"unit"`
		Tier1    int    `json:"tier_1"`
		Tier2    int    `json:"tier_2"`
		Tier3    int    `json:"tier_3"`
		Tier4    int    `json:"tier_4"`
		Tier5    int    `json:"tier_5"`
		MinStars int    `json:"min_stars"`
	}
	defs := make([]GoalDef, 0)
	for rows.Next() {
		var g GoalDef
		var sort int
		rows.Scan(&g.ID, &g.SkillID, &g.Label, &g.Unit, &g.Tier1, &g.Tier2, &g.Tier3, &g.Tier4, &g.Tier5, &g.MinStars, &sort)
		defs = append(defs, g)
	}
	writeJSON(w, defs)
}
