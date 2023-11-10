CREATE TABLE replay_checksums (
    replay_hash TEXT NOT NULL PRIMARY KEY,
    player_name TEXT NOT NULL
);

CREATE TABLE replay_queue (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    player_name TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    replay_json BLOB NOT NULL
);

CREATE INDEX replay_queue_player_name_index 
ON replay_queue(player_name);

CREATE TABLE good_replay_archive (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    replay_id INTEGER NOT NULL,
    player_name TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    replay_json BLOB NOT NULL
);

CREATE TABLE failed_replay_archive (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    replay_id INTEGER NOT NULL,
    player_name TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    replay_json BLOB NOT NULL,
    fail_reason INTEGER NOT NULL
);
