CREATE TABLE game_library (
    id             SERIAL PRIMARY KEY,
    rawg_id        INTEGER NOT NULL UNIQUE,
    title          VARCHAR(255) NOT NULL,
    genre          VARCHAR(100),
    platform       VARCHAR(100),
    cover_url      TEXT,
    personal_note  TEXT,
    personal_score INTEGER,
    status         VARCHAR(20),
    added_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

SELECT * FROM game_vault_db.game_library;