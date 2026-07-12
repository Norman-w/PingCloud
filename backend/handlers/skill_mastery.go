package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"pingpong/db"
	"pingpong/models"
)

// ── Skill attribute definitions ──

const (
	BodyNone = iota // 0 — knowledge/theory, no body side
	FH              // 1
	BH              // 2
	Both            // 3
)
const (
	Attack  = iota // 0
	Defense        // 1
	Neutral        // 2
)

type skillAttrs struct {
	Body     int    // BodyNone / FH / BH / Both
	Style    int    // Attack / Defense / Neutral
	Spin     string // "topspin" / "backspin" / "none"
	Distance string // "short" / "long" / "mixed"
	SideSpin string // "left" / "right" / "none"
}

func (a skillAttrs) Tags() []string {
	tags := make([]string, 0)
	switch a.Body {
	case FH:   tags = append(tags, "正手")
	case BH:   tags = append(tags, "反手")
	case Both: tags = append(tags, "正手", "反手")
	// BodyNone: no body tag
	}
	switch a.Style {
	case Attack:  tags = append(tags, "进攻")
	case Defense: tags = append(tags, "防守")
	// Neutral: no style tag
	}
	if a.Spin == "topspin"  { tags = append(tags, "上旋") }
	if a.Spin == "backspin" { tags = append(tags, "下旋") }
	if a.Distance == "short" { tags = append(tags, "短球") }
	if a.Distance == "long"  { tags = append(tags, "长球") }
	if a.SideSpin == "left"  { tags = append(tags, "左侧旋") }
	if a.SideSpin == "right" { tags = append(tags, "右侧旋") }
	return tags
}

// skillAttrMap maps skill ID to its attributes
var skillAttrMap = map[int]skillAttrs{
	// ── 基本功 (1-19) ──
	1:  {FH, Attack, "topspin",  "short", "none"},  // 正手攻球
	2:  {BH, Defense, "none",    "short", "none"},  // 反手拨球
	3:  {FH, Attack, "topspin",  "long",  "none"},  // 正手前冲弧圈
	4:  {FH, Attack, "topspin",  "mixed", "none"},  // 正手加转弧圈
	5:  {BH, Attack, "topspin",  "long",  "none"},  // 反手拉球
	6:  {FH, Defense, "backspin","short", "none"},  // 正手搓球
	7:  {BH, Defense, "backspin","short", "none"},  // 反手搓球
	8:  {Both, Defense, "backspin","short","none"}, // 摆短
	9:  {Both, Defense, "backspin","long", "none"}, // 劈长
	10: {FH, Attack, "topspin",  "short", "none"},  // 台内挑打
	11: {BH, Attack, "topspin",  "short", "none"},  // 反手拧拉
	12: {FH, Defense, "topspin", "short", "none"},  // 正手快带
	13: {BH, Attack, "none",     "short", "none"},  // 反手弹击
	14: {FH, Attack, "none",     "mixed", "none"},  // 正手扣杀
	15: {FH, Neutral, "mixed",   "short", "none"},  // 正手发球
	16: {BH, Neutral, "mixed",   "short", "none"},  // 反手发球
	17: {FH, Neutral, "mixed",   "short", "right"}, // 逆旋转发球
	18: {Both, Defense, "mixed", "short", "none"},  // 接发球
	19: {BodyNone, Neutral, "none","mixed","none"}, // 步法
	// ── 技战术 (20-29) ──
	20: {FH, Attack, "topspin",  "short", "none"},  // 发球抢攻
	21: {BH, Attack, "topspin",  "short", "none"},  // 接发球抢攻
	22: {BH, Attack, "topspin",  "long",  "none"},  // 反手相持对拉
	23: {FH, Attack, "topspin",  "long",  "none"},  // 反手相持转正手
	24: {FH, Attack, "topspin",  "mixed", "none"},  // 侧身抢攻
	25: {Both, Attack, "backspin","short","none"},  // 摆短控制+抢先上手
	26: {BH, Defense, "backspin", "long", "none"},  // 劈长压反手底线
	27: {Both, Attack, "none",   "mixed", "none"},  // 落点变化
	28: {Both, Attack, "topspin", "mixed","none"},  // 节奏旋转变化
	29: {Both, Defense, "topspin","mixed","none"},  // 相持转换与反拉防守
	// ── 基础入门 (30-34) ──
	30: {BodyNone, Neutral, "none","mixed","none"}, // 了解乒乓球
	31: {BodyNone, Neutral, "none","mixed","none"}, // 了解握拍方式
	32: {BodyNone, Neutral, "none","mixed","none"}, // 了解胶皮
	33: {BodyNone, Neutral, "none","mixed","none"}, // 了解球台
	34: {BodyNone, Neutral, "none","mixed","none"}, // 了解规则
	// ── 物理学原理 (35-39) ──
	35: {BodyNone, Neutral, "mixed","mixed","none"}, // 旋转原理
	36: {BodyNone, Neutral, "mixed","mixed","none"}, // 弧线原理
	37: {BodyNone, Neutral, "none", "mixed","none"}, // 速度力量
	38: {BodyNone, Neutral, "none", "mixed","none"}, // 落点角度
	39: {BodyNone, Neutral, "none", "mixed","none"}, // 借力发力
	// ── 初学技巧 (40-55) ──
	40: {BodyNone, Neutral, "mixed","mixed","none"}, // 颠球
	41: {BodyNone, Neutral, "mixed","mixed","none"}, // 抛球
	42: {BodyNone, Neutral, "mixed","mixed","none"}, // 停球
	43: {BodyNone, Neutral, "mixed","mixed","none"}, // 吊球
	44: {BodyNone, Neutral, "none", "mixed","none"}, // 躺姿开肩放松
	45: {BodyNone, Neutral, "none", "mixed","none"}, // 躺姿负重手小臂
	46: {FH, Neutral, "mixed", "mixed","none"},      // 正手拉球引手辅助
	47: {FH, Neutral, "mixed", "mixed","none"},      // 正手拉球等位辅助
	48: {BH, Neutral, "mixed", "mixed","none"},      // 反手拨球定肘技巧
	49: {Both, Neutral, "mixed","mixed","none"},     // 跟球技巧
	50: {Both, Neutral, "mixed","mixed","none"},     // 击球位置技巧
	51: {Both, Neutral, "mixed","mixed","none"},     // 发球跟球技巧
	52: {Both, Neutral, "mixed","mixed","none"},     // 发球弧线技巧
	53: {Both, Neutral, "mixed","mixed","none"},     // 跪姿跟弧线技巧
	54: {FH, Neutral, "mixed", "mixed","none"},      // 正手拉下旋防漏技巧
	55: {FH, Neutral, "mixed", "mixed","none"},      // 自动学翻挑技巧
	// ── 身体协调和灵敏训练 (56-59) ──
	56: {FH, Neutral, "mixed", "mixed","none"},      // 单脚正手拉球
	57: {BodyNone, Neutral, "none","mixed","none"},  // 并步摸台
	58: {BodyNone, Neutral, "none","mixed","none"},  // 接下落球
	59: {FH, Neutral, "mixed", "mixed","none"},      // 三脚架引手和回收
}

// GetSkillMastery returns all skills with mastery status, tags, and training stats
func GetSkillMastery(w http.ResponseWriter, r *http.Request) {
	pid := getPlayerIDFromCookie(r)
	if pid == 0 {
		items := buildMasteryItems(0)
		writeJSON(w, map[string]interface{}{
			"skills":     items,
			"stages":     buildStageGroups(items),
			"tagFilters": buildTagGroups(items),
		})
		return
	}

	items := buildMasteryItems(pid)
	writeJSON(w, map[string]interface{}{
		"skills":     items,
		"stages":     buildStageGroups(items),
		"tagFilters": buildTagGroups(items),
	})
}

func buildMasteryItems(pid int) []models.SkillMasteryItem {
	// Fetch all skills
	rows, err := db.DB.Query(`SELECT id, category, name, sort_order FROM skills ORDER BY category, sort_order`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	// Build training stats map for this player
	type trainData struct {
		count    int
		duration int
		lastDate string
	}
	trainingMap := make(map[int]trainData)
	if pid > 0 {
		tRows, err := db.DB.Query(
			`SELECT tls.skill_id, COUNT(*), COALESCE(SUM(tl.duration_minutes),0), MAX(tl.date)::text
			 FROM training_log_skills tls
			 JOIN training_logs tl ON tl.id = tls.training_log_id
			 WHERE tl.player_id = $1
			 GROUP BY tls.skill_id`, pid)
		if err == nil {
			defer tRows.Close()
			for tRows.Next() {
				var sid int
				var td trainData
				if err := tRows.Scan(&sid, &td.count, &td.duration, &td.lastDate); err == nil {
					trainingMap[sid] = td
				}
			}
		}
	}

	// Build mastery overrides map
	masteryMap := make(map[int]string)
	if pid > 0 {
		mRows, err := db.DB.Query(`SELECT skill_id, status FROM skill_mastery WHERE player_id = $1`, pid)
		if err == nil {
			defer mRows.Close()
			for mRows.Next() {
				var sid int; var st string
				if err := mRows.Scan(&sid, &st); err == nil {
					masteryMap[sid] = st
				}
			}
		}
	}

	items := make([]models.SkillMasteryItem, 0)
	for rows.Next() {
		var id int; var cat, name string; var sort int
		if err := rows.Scan(&id, &cat, &name, &sort); err != nil {
			continue
		}

		attrs, ok := skillAttrMap[id]
		if !ok {
			attrs = skillAttrs{Both, Neutral, "none", "mixed", "none"}
		}

		item := models.SkillMasteryItem{
			ID:       id,
			Name:     name,
			Category: cat,
			Tags:     attrs.Tags(),
		}

		if td, ok := trainingMap[id]; ok {
			item.PracticeCount = td.count
			item.TotalDurationMinutes = td.duration
			item.LastPracticed = td.lastDate
		}

		// Determine status
		if ms, ok := masteryMap[id]; ok && ms != "none" {
			item.Status = ms
		} else if item.PracticeCount > 0 {
			item.Status = "practicing"
		} else {
			item.Status = "none"
		}

		items = append(items, item)
	}

	return items
}

// buildTagGroups creates filter chips grouped by attribute tags
func buildTagGroups(items []models.SkillMasteryItem) []models.SkillMasteryGroup {
	groupDefs := []struct{ label, tag string }{
		{"正手技术", "正手"},
		{"反手技术", "反手"},
		{"进攻", "进攻"},
		{"防守", "防守"},
		{"上旋球", "上旋"},
		{"下旋球", "下旋"},
		{"短球", "短球"},
		{"长球", "长球"},
	}
	groups := make([]models.SkillMasteryGroup, 0)
	for _, gd := range groupDefs {
		filtered := make([]models.SkillMasteryItem, 0)
		for _, item := range items {
			for _, t := range item.Tags {
				if t == gd.tag {
					filtered = append(filtered, item)
					break
				}
			}
		}
		if len(filtered) > 0 {
			groups = append(groups, models.SkillMasteryGroup{Label: gd.label, Skills: filtered})
		}
	}
	return groups
}

func buildStageGroups(items []models.SkillMasteryItem) []models.SkillMasteryGroup {
	// Group by category (training stage)
	stageOrder := []string{"基础入门", "初学技巧", "身体协调和灵敏训练", "基本功", "技战术", "物理学原理"}
	stageMap := make(map[string][]models.SkillMasteryItem)
	for _, item := range items {
		stageMap[item.Category] = append(stageMap[item.Category], item)
	}

	groups := make([]models.SkillMasteryGroup, 0)
	for _, stage := range stageOrder {
		if skills, ok := stageMap[stage]; ok && len(skills) > 0 {
			groups = append(groups, models.SkillMasteryGroup{Label: stage, Skills: skills})
		}
	}
	return groups
}

// UpdateSkillMastery handles PUT /api/skill-mastery/:id
func UpdateSkillMastery(w http.ResponseWriter, r *http.Request) {
	pid := getPlayerIDFromCookie(r)
	if pid == 0 {
		http.Error(w, "请先登录", http.StatusUnauthorized)
		return
	}

	// Extract skill ID from path: /api/skill-mastery/123
	idStr := strings.TrimPrefix(r.URL.Path, "/api/skill-mastery/")
	skillID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid skill id", http.StatusBadRequest)
		return
	}

	var req models.UpdateSkillMasteryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if req.Status != "none" && req.Status != "practicing" && req.Status != "mastered" {
		http.Error(w, "status must be none, practicing, or mastered", http.StatusBadRequest)
		return
	}

	_, err = db.DB.Exec(
		`INSERT INTO skill_mastery (player_id, skill_id, status, updated_at)
		 VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (player_id, skill_id) DO UPDATE SET status=$3, updated_at=NOW()`,
		pid, skillID, req.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated item
	items := buildMasteryItems(pid)
	for _, item := range items {
		if item.ID == skillID {
			writeJSON(w, item)
			return
		}
	}
	http.Error(w, "skill not found", http.StatusNotFound)
}
