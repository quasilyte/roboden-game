CREATE TABLE classic_scores (
    player_name TEXT NOT NULL PRIMARY KEY,
    score INTEGER NOT NULL,
    difficulty INTEGER NOT NULL,
    time_seconds INTEGER NOT NULL,
    drones TEXT,
    score_rank INTEGER
);

CREATE TABLE arena_scores (
    player_name TEXT NOT NULL PRIMARY KEY,
    score INTEGER NOT NULL,
    difficulty INTEGER NOT NULL,
    drones TEXT,
    score_rank INTEGER
);

CREATE TABLE inf_arena_scores (
    player_name TEXT NOT NULL PRIMARY KEY,
    score INTEGER NOT NULL,
    difficulty INTEGER NOT NULL,
    time_seconds INTEGER NOT NULL,
    drones TEXT,
    score_rank INTEGER
);
