package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"pingpong/db"
)

type SkillGoalDef struct {
	ID        int    `json:"id"`
	SkillID   int    `json:"skill_id"`
	GoalType  string `json:"goal_type"`  // "total_minutes", "session_count", "max_indicator"
	Indicator string `json:"indicator"`  // which radar indicator, if goal_type=="max_indicator"
	Label     string `json:"label"`
	Star1     int    `json:"star_1"`
	Star2     int    `json:"star_2"`
	Star3     int    `json:"star_3"`
	Star4     int    `json:"star_4"`
	Star5     int    `json:"star_5"`
	MinStars  int    `json:"min_stars"` // minimum stars to clear this goal
}

// Hardcoded goal definitions per skill
var skillGoals = map[int][]SkillGoalDef{
	// Each skill gets: 1) total time goal, 2) session count goal
	// Star thresholds escalate, min 2 stars to pass
}

func init() {
	// Generate default goals for all skills (IDs 1-59 based on existing data)
	for id := 1; id <= 59; id++ {
		skillGoals[id] = []SkillGoalDef{
			{SkillID: id, GoalType: "total_minutes", Label: "累计练习时长",
				Star1: 60, Star2: 200, Star3: 500, Star4: 1000, Star5: 2000, MinStars: 2},
			{SkillID: id, GoalType: "session_count", Label: "累计练习次数",
				Star1: 3, Star2: 10, Star3: 30, Star4: 80, Star5: 200, MinStars: 1},
		}
	}
}

// GetSkillGoals returns goals and computed progress for a skill
func GetSkillGoals(w http.ResponseWriter, r *http.Request) {
	pid := getPlayerIDFromCookie(r)
	idStr := strings.TrimPrefix(r.URL.Path, "/api/skill-goals/")
	skillID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid skill id", http.StatusBadRequest)
		return
	}

	goals, ok := skillGoals[skillID]
	if !ok {
		goals = []SkillGoalDef{
			{Label: "累计练习时长", GoalType: "total_minutes", Star1: 60, Star2: 200, Star3: 500, Star4: 1000, Star5: 2000, MinStars: 2},
			{Label: "累计练习次数", GoalType: "session_count", Star1: 3, Star2: 10, Star3: 30, Star4: 80, Star5: 200, MinStars: 1},
		}
	}

	type GoalWithProgress struct {
		SkillGoalDef
		CurrentValue int `json:"current_value"`
		Stars        int `json:"stars"`
		Passed       bool `json:"passed"`
	}

	result := make([]GoalWithProgress, 0)
	for _, g := range goals {
		gwp := GoalWithProgress{SkillGoalDef: g}
		if pid > 0 {
			switch g.GoalType {
			case "total_minutes":
				db.DB.QueryRow(`SELECT COALESCE(SUM(tl.duration_minutes),0) FROM training_log_skills tls
					JOIN training_logs tl ON tl.id=tls.training_log_id
					WHERE tls.skill_id=$1 AND tl.player_id=$2`, skillID, pid).Scan(&gwp.CurrentValue)
			case "session_count":
				db.DB.QueryRow(`SELECT COUNT(*) FROM training_log_skills tls
					JOIN training_logs tl ON tl.id=tls.training_log_id
					WHERE tls.skill_id=$1 AND tl.player_id=$2`, skillID, pid).Scan(&gwp.CurrentValue)
			}
		}
		// Calculate stars based on thresholds
		v := gwp.CurrentValue
		if v >= g.Star5 { gwp.Stars = 5 } else if v >= g.Star4 { gwp.Stars = 4 } else if v >= g.Star3 { gwp.Stars = 3 } else if v >= g.Star2 { gwp.Stars = 2 } else if v >= g.Star1 { gwp.Stars = 1 }
		gwp.Passed = gwp.Stars >= g.MinStars
		result = append(result, gwp)
	}

	writeJSON(w, result)
}
