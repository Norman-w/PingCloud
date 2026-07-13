-- Skill goal definitions (preset per skill, tiered star system)
CREATE TABLE IF NOT EXISTS skill_goals (
    id SERIAL PRIMARY KEY,
    skill_id INT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    label VARCHAR(50) NOT NULL,       -- "累计时长" "对攻次数" "连续对攻"
    unit VARCHAR(20) NOT NULL,        -- "分钟" "次" "组"
    tier_1 INT NOT NULL DEFAULT 0,    -- ★ threshold
    tier_2 INT NOT NULL DEFAULT 0,    -- ★★
    tier_3 INT NOT NULL DEFAULT 0,    -- ★★★
    tier_4 INT NOT NULL DEFAULT 0,    -- ★★★★
    tier_5 INT NOT NULL DEFAULT 0,    -- ★★★★★
    min_stars INT NOT NULL DEFAULT 0, -- minimum stars to pass this goal
    sort_order INT NOT NULL DEFAULT 0,
    UNIQUE(skill_id, label)
);

-- Per-session goal progress (user records after each training)
ALTER TABLE training_log_skills ADD COLUMN IF NOT EXISTS goal_values JSONB DEFAULT '{}';

-- Seed: 正手攻球 (skill_id=1)
INSERT INTO skill_goals (skill_id, label, unit, tier_1, tier_2, tier_3, tier_4, tier_5, min_stars, sort_order) VALUES
(1, '累计时长', '分钟', 60, 200, 500, 1000, 2000, 1, 1),
(1, '单次连续对攻', '次', 20, 50, 100, 200, 500, 1, 2),
(1, '累计对攻次数', '次', 100, 500, 2000, 5000, 10000, 1, 3)
ON CONFLICT (skill_id, label) DO NOTHING;
