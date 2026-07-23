-- Group standings extras: games ratio + total points scored; manual rank override
ALTER TABLE tournament_teams ADD COLUMN IF NOT EXISTS games_won INT NOT NULL DEFAULT 0;
ALTER TABLE tournament_teams ADD COLUMN IF NOT EXISTS games_lost INT NOT NULL DEFAULT 0;
ALTER TABLE tournament_teams ADD COLUMN IF NOT EXISTS points_scored INT NOT NULL DEFAULT 0;
ALTER TABLE tournament_teams ADD COLUMN IF NOT EXISTS rank_manual BOOLEAN NOT NULL DEFAULT false;
