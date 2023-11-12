CREATE TABLE blitz_scores (
    player_name TEXT NOT NULL PRIMARY KEY,
    replay_id INTEGER,
    score INTEGER NOT NULL,
    difficulty INTEGER NOT NULL,
    time_seconds INTEGER NOT NULL,
    drones TEXT,
    score_rank INTEGER,
    platform TEXT
);

CREATE TABLE classic_scores (
    player_name TEXT NOT NULL PRIMARY KEY,
    replay_id INTEGER,
    score INTEGER NOT NULL,
    difficulty INTEGER NOT NULL,
    time_seconds INTEGER NOT NULL,
    drones TEXT,
    score_rank INTEGER,
    platform TEXT
);

CREATE TABLE arena_scores (
    player_name TEXT NOT NULL PRIMARY KEY,
    replay_id INTEGER,
    score INTEGER NOT NULL,
    difficulty INTEGER NOT NULL,
    drones TEXT,
    score_rank INTEGER,
    platform TEXT
);

CREATE TABLE inf_arena_scores (
    player_name TEXT NOT NULL PRIMARY KEY,
    replay_id INTEGER,
    score INTEGER NOT NULL,
    difficulty INTEGER NOT NULL,
    time_seconds INTEGER NOT NULL,
    drones TEXT,
    score_rank INTEGER,
    platform TEXT
);

CREATE TABLE reverse_scores (
    player_name TEXT NOT NULL PRIMARY KEY,
    replay_id INTEGER,
    score INTEGER NOT NULL,
    difficulty INTEGER NOT NULL,
    time_seconds INTEGER NOT NULL,
    score_rank INTEGER,
    platform TEXT
);
