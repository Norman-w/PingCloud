-- Training Log System: skills library, training logs, training log skills (many-to-many)

-- Skills library table
CREATE TABLE IF NOT EXISTS skills (
    id SERIAL PRIMARY KEY,
    category VARCHAR(20) NOT NULL,  -- '基本功' or '技战术'
    name VARCHAR(50) NOT NULL,
    style VARCHAR(20) NOT NULL DEFAULT '双反',
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(category, name, style)
);

-- Training logs table
CREATE TABLE IF NOT EXISTS training_logs (
    id SERIAL PRIMARY KEY,
    player_id INT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    duration_minutes INT NOT NULL DEFAULT 0,
    location VARCHAR(100) NOT NULL DEFAULT '',
    partner VARCHAR(50) NOT NULL DEFAULT '',
    energy_rating INT NOT NULL DEFAULT 0 CHECK (energy_rating >= 0 AND energy_rating <= 5),
    feel_rating INT NOT NULL DEFAULT 0 CHECK (feel_rating >= 0 AND feel_rating <= 5),
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_training_logs_player_date ON training_logs(player_id, date DESC);

-- Training log skills (many-to-many)
CREATE TABLE IF NOT EXISTS training_log_skills (
    id SERIAL PRIMARY KEY,
    training_log_id INT NOT NULL REFERENCES training_logs(id) ON DELETE CASCADE,
    skill_id INT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    practice_amount VARCHAR(100) NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    UNIQUE(training_log_id, skill_id)
);

-- ============================================
-- Seed data: 双反打法技能库 (29 skills)
-- ============================================

-- 一、基本功 (19 items)
INSERT INTO skills (category, name, style, sort_order) VALUES
('基本功', '正手攻球', '双反', 1),
('基本功', '反手拨球', '双反', 2),
('基本功', '正手前冲弧圈', '双反', 3),
('基本功', '正手加转弧圈', '双反', 4),
('基本功', '反手拉球', '双反', 5),
('基本功', '正手搓球', '双反', 6),
('基本功', '反手搓球', '双反', 7),
('基本功', '摆短', '双反', 8),
('基本功', '劈长', '双反', 9),
('基本功', '台内挑打', '双反', 10),
('基本功', '反手拧拉', '双反', 11),
('基本功', '正手快带', '双反', 12),
('基本功', '反手弹击', '双反', 13),
('基本功', '正手扣杀', '双反', 14),
('基本功', '正手发球', '双反', 15),
('基本功', '反手发球', '双反', 16),
('基本功', '逆旋转发球', '双反', 17),
('基本功', '接发球', '双反', 18),
('基本功', '步法', '双反', 19)
ON CONFLICT (category, name, style) DO UPDATE SET sort_order = EXCLUDED.sort_order;

-- 二、技战术 (10 items)
INSERT INTO skills (category, name, style, sort_order) VALUES
('技战术', '发球抢攻', '双反', 1),
('技战术', '接发球抢攻', '双反', 2),
('技战术', '反手相持对拉', '双反', 3),
('技战术', '反手相持转正手', '双反', 4),
('技战术', '侧身抢攻', '双反', 5),
('技战术', '摆短控制+抢先上手', '双反', 6),
('技战术', '劈长压反手底线', '双反', 7),
('技战术', '落点变化', '双反', 8),
('技战术', '节奏旋转变化', '双反', 9),
('技战术', '相持转换与反拉防守', '双反', 10)
ON CONFLICT (category, name, style) DO UPDATE SET sort_order = EXCLUDED.sort_order;
