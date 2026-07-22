-- Tournament system: configurable team tournament with group stage + knockout

-- Tournament configuration
CREATE TABLE IF NOT EXISTS tournaments (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(100) NOT NULL DEFAULT '锦标赛',
    group_count     INT NOT NULL DEFAULT 2,
    teams_per_group INT NOT NULL DEFAULT 3,
    players_per_team INT NOT NULL DEFAULT 3,
    max_participants INT NOT NULL DEFAULT 24,
    seed_enabled    BOOLEAN NOT NULL DEFAULT false,
    seed_count      INT NOT NULL DEFAULT 0,
    registration_deadline TIMESTAMP,
    status          VARCHAR(20) NOT NULL DEFAULT 'registration',
    phase           VARCHAR(20) NOT NULL DEFAULT 'registration',
    created_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMP,
    deleted         BOOLEAN NOT NULL DEFAULT false
);

-- Teams within a tournament
CREATE TABLE IF NOT EXISTS tournament_teams (
    id              SERIAL PRIMARY KEY,
    tournament_id   INT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    group_name      VARCHAR(10) NOT NULL,
    team_index      INT NOT NULL DEFAULT 0,
    team_name       VARCHAR(50) NOT NULL,
    knockout_seed   INT,
    group_rank      INT,
    group_wins      INT NOT NULL DEFAULT 0,
    group_losses    INT NOT NULL DEFAULT 0,
    group_points    INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tt_tournament ON tournament_teams(tournament_id);

-- Player registrations with waitlist
CREATE TABLE IF NOT EXISTS tournament_registrations (
    id              SERIAL PRIMARY KEY,
    tournament_id   INT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    player_id       INT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    status          VARCHAR(20) NOT NULL DEFAULT 'confirmed',
    waitlist_pos    INT,
    registered_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    cancelled_at    TIMESTAMP,
    UNIQUE(tournament_id, player_id)
);

CREATE INDEX IF NOT EXISTS idx_tr_tournament ON tournament_registrations(tournament_id);

-- Player-to-team assignment with fixed roles
CREATE TABLE IF NOT EXISTS tournament_team_players (
    id              SERIAL PRIMARY KEY,
    tournament_id   INT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    team_id         INT NOT NULL REFERENCES tournament_teams(id) ON DELETE CASCADE,
    player_id       INT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    role            VARCHAR(5) NOT NULL,
    is_seed         BOOLEAN NOT NULL DEFAULT false,
    UNIQUE(tournament_id, player_id)
);

CREATE INDEX IF NOT EXISTS idx_ttp_team ON tournament_team_players(team_id);

-- Parent container: one team match (best-of-5 sub-matches)
CREATE TABLE IF NOT EXISTS tournament_team_matches (
    id              SERIAL PRIMARY KEY,
    tournament_id   INT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    phase           VARCHAR(20) NOT NULL,
    round           INT NOT NULL DEFAULT 0,
    group_name      VARCHAR(10),
    team_a_id       INT NOT NULL REFERENCES tournament_teams(id) ON DELETE CASCADE,
    team_b_id       INT NOT NULL REFERENCES tournament_teams(id) ON DELETE CASCADE,
    team_a_wins     INT NOT NULL DEFAULT 0,
    team_b_wins     INT NOT NULL DEFAULT 0,
    winner_team_id  INT REFERENCES tournament_teams(id),
    played          BOOLEAN NOT NULL DEFAULT false,
    card_type       VARCHAR(20),
    card_detail     VARCHAR(50),
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ttm_tournament ON tournament_team_matches(tournament_id);

-- Individual matches (5 per team match, best-of-3 games)
CREATE TABLE IF NOT EXISTS tournament_matches (
    id              SERIAL PRIMARY KEY,
    tournament_id   INT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    team_match_id   INT NOT NULL REFERENCES tournament_team_matches(id) ON DELETE CASCADE,
    phase           VARCHAR(20) NOT NULL,
    round           INT NOT NULL DEFAULT 0,
    group_name      VARCHAR(10),
    team_a_id       INT NOT NULL REFERENCES tournament_teams(id) ON DELETE CASCADE,
    team_b_id       INT NOT NULL REFERENCES tournament_teams(id) ON DELETE CASCADE,
    match_order     INT NOT NULL,
    match_type      VARCHAR(10) NOT NULL,
    player_a_id     INT NOT NULL REFERENCES players(id),
    player_b_id     INT NOT NULL REFERENCES players(id),
    player_a2_id    INT REFERENCES players(id),
    player_b2_id    INT REFERENCES players(id),
    game1_score_a   INT,
    game1_score_b   INT,
    game2_score_a   INT,
    game2_score_b   INT,
    game3_score_a   INT,
    game3_score_b   INT,
    winner_id       INT REFERENCES players(id),
    winner_team_id  INT REFERENCES tournament_teams(id),
    played          BOOLEAN NOT NULL DEFAULT false,
    forfeit         BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tm_tournament ON tournament_matches(tournament_id);
CREATE INDEX IF NOT EXISTS idx_tm_team_match ON tournament_matches(team_match_id);

-- Cards drawn per team per team match
CREATE TABLE IF NOT EXISTS tournament_cards (
    id              SERIAL PRIMARY KEY,
    tournament_id   INT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    team_match_id   INT NOT NULL REFERENCES tournament_team_matches(id) ON DELETE CASCADE,
    team_id         INT NOT NULL REFERENCES tournament_teams(id) ON DELETE CASCADE,
    card_type       VARCHAR(20) NOT NULL,
    drawn_at        TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tc_team_match ON tournament_cards(team_match_id);
