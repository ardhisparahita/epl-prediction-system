CREATE TABLE matches (
    id BIGSERIAL PRIMARY KEY,
    date DATE NOT NULL,
    home_team VARCHAR(100) NOT NULL,
    away_team VARCHAR(100) NOT NULL,
    fthg INTEGER NOT NULL CHECK (fthg >= 0),
    ftag INTEGER NOT NULL CHECK (ftag >= 0),
    hst INTEGER CHECK (hst >= 0),
    ast INTEGER CHECK (ast >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT idx_matches_fixture UNIQUE (date, home_team, away_team)
);

CREATE INDEX idx_matches_date ON matches (date);

CREATE INDEX idx_matches_home_date ON matches (home_team, date DESC);

CREATE INDEX idx_matches_away_date ON matches (away_team, date DESC);
