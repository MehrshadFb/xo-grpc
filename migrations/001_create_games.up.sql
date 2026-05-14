CREATE TABLE games (
    id TEXT PRIMARY KEY,
    join_code TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL,
    board JSONB NOT NULL,
    next_turn TEXT NOT NULL,
    winner TEXT NOT NULL,
    is_draw BOOLEAN NOT NULL DEFAULT FALSE,
    move_number BIGINT NOT NULL DEFAULT 0,
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE players (
    id TEXT PRIMARY KEY,
    game_id TEXT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    display_name TEXT NOT NULL,
    mark TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(game_id, mark)
);

CREATE TABLE sessions (
    token TEXT PRIMARY KEY,
    game_id TEXT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    player_id TEXT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    mark TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);