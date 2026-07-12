-- Skill mastery tracking table
CREATE TABLE IF NOT EXISTS skill_mastery (
    id SERIAL PRIMARY KEY,
    player_id INT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    skill_id INT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    status VARCHAR(10) NOT NULL DEFAULT 'none',  -- 'none' | 'practicing' | 'mastered'
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(player_id, skill_id)
);
