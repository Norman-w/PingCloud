-- Add indicators JSONB to training_log_skills
ALTER TABLE training_log_skills ADD COLUMN IF NOT EXISTS indicators JSONB DEFAULT '{}';

-- Locations table
CREATE TABLE IF NOT EXISTS locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT NOW()
);
