package main

import (
	"database/sql"
	"fmt"

	"github.com/quasilyte/roboden-game/serverapi"
)

type seasonDB struct {
	id   int
	conn *sql.DB

	classicPlayerScore *sql.Stmt
	classicFetchAll    *sql.Stmt
	classicUpsert      *sql.Stmt

	arenaPlayerScore *sql.Stmt
	arenaFetchAll    *sql.Stmt
	arenaUpsert      *sql.Stmt

	infArenaPlayerScore *sql.Stmt
	infArenaFetchAll    *sql.Stmt
	infArenaUpsert      *sql.Stmt

	reversePlayerScore *sql.Stmt
	reverseFetchAll    *sql.Stmt
	reverseUpsert      *sql.Stmt
}

func withTransaction(conn *sql.DB, f func(tx *sql.Tx) error) (err error) {
	var tx *sql.Tx
	tx, err = conn.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			err = tx.Commit()
		} else {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				err = fmt.Errorf("rollback error (%v) after %w", rollbackErr, err)
			}
		}
	}()

	err = f(tx)
	return err
}

func (db *seasonDB) PrepareQueries() error {
	{
		q := "SELECT score FROM classic_scores WHERE player_name = ?"
		stmt, err := db.conn.Prepare(q)
		if err != nil {
			return err
		}
		db.classicPlayerScore = stmt
	}

	if db.id == currentSeason {
		q := `
			SELECT player_name, score, difficulty, drones, time_seconds, platform
			FROM classic_scores
			ORDER BY score DESC
		`
		stmt, err := db.conn.Prepare(q)
		if err != nil {
			return err
		}
		db.classicFetchAll = stmt
	}

	{
		q := `
		INSERT OR REPLACE INTO classic_scores
			('player_name', 'score', 'difficulty', 'drones', 'time_seconds', 'platform')
		VALUES
			(?, ?, ?, ?, ?, ?)
		`
		stmt, err := db.conn.Prepare(q)
		if err != nil {
			return err
		}
		db.classicUpsert = stmt
	}

	{
		q := "SELECT score FROM arena_scores WHERE player_name = ?"
		stmt, err := db.conn.Prepare(q)
		if err != nil {
			return err
		}
		db.arenaPlayerScore = stmt
	}

	if db.id == currentSeason {
		q := `
			SELECT player_name, score, difficulty, drones, platform
			FROM arena_scores
			ORDER BY score DESC
		`
		stmt, err := db.conn.Prepare(q)
		if err != nil {
			return err
		}
		db.arenaFetchAll = stmt
	}

	{
		q := `
		INSERT OR REPLACE INTO arena_scores
			('player_name', 'score', 'difficulty', 'drones', 'platform')
		VALUES
			(?, ?, ?, ?, ?)
		`
		stmt, err := db.conn.Prepare(q)
		if err != nil {
			return err
		}
		db.arenaUpsert = stmt
	}

	{
		q := "SELECT score FROM inf_arena_scores WHERE player_name = ?"
		stmt, err := db.conn.Prepare(q)
		if err != nil {
			return err
		}
		db.infArenaPlayerScore = stmt
	}

	if db.id == currentSeason {
		q := `
			SELECT player_name, score, difficulty, drones, time_seconds, platform
			FROM inf_arena_scores
			ORDER BY score DESC
		`
		stmt, err := db.conn.Prepare(q)
		if err != nil {
			return err
		}
		db.infArenaFetchAll = stmt
	}

	{
		q := `
		INSERT OR REPLACE INTO inf_arena_scores
			('player_name', 'score', 'difficulty', 'drones', 'time_seconds', 'platform')
		VALUES
			(?, ?, ?, ?, ?, ?)
		`
		stmt, err := db.conn.Prepare(q)
		if err != nil {
			return err
		}
		db.infArenaUpsert = stmt
	}

	{
		q := "SELECT score FROM reverse_scores WHERE player_name = ?"
		stmt, err := db.conn.Prepare(q)
		if err != nil {
			return err
		}
		db.reversePlayerScore = stmt
	}

	if db.id == currentSeason {
		q := `
			SELECT player_name, score, difficulty, time_seconds, platform
			FROM reverse_scores
			ORDER BY score DESC
		`
		stmt, err := db.conn.Prepare(q)
		if err != nil {
			return err
		}
		db.reverseFetchAll = stmt
	}

	{
		q := `
		INSERT OR REPLACE INTO reverse_scores
			('player_name', 'score', 'difficulty', 'time_seconds', 'platform')
		VALUES
			(?, ?, ?, ?, ?)
		`
		stmt, err := db.conn.Prepare(q)
		if err != nil {
			return err
		}
		db.reverseUpsert = stmt
	}

	return nil
}

func (db *seasonDB) UpdatePlayerScore(mode, name, drones string, score, difficulty, timeSeconds int, platform string) error {
	var err error
	switch mode {
	case "classic":
		_, err = db.classicUpsert.Exec(name, score, difficulty, drones, timeSeconds, platform)
	case "arena":
		_, err = db.arenaUpsert.Exec(name, score, difficulty, drones, platform)
	case "inf_arena":
		_, err = db.infArenaUpsert.Exec(name, score, difficulty, drones, timeSeconds, platform)
	case "reverse":
		_, err = db.reverseUpsert.Exec(name, score, difficulty, timeSeconds, platform)
	}
	return err
}

func (db *seasonDB) PlayerScore(mode, name string) int {
	var result int
	var err error
	switch mode {
	case "classic":
		err = db.classicPlayerScore.QueryRow(name).Scan(&result)
	case "arena":
		err = db.arenaPlayerScore.QueryRow(name).Scan(&result)
	case "inf_arena":
		err = db.infArenaPlayerScore.QueryRow(name).Scan(&result)
	case "reverse":
		err = db.reversePlayerScore.QueryRow(name).Scan(&result)
	}
	if err != nil {
		return -1
	}
	return result
}

func (db *seasonDB) AllScores(mode string) ([]serverapi.LeaderboardEntry, error) {
	var rows *sql.Rows
	var err error
	switch mode {
	case "classic":
		rows, err = db.classicFetchAll.Query()
	case "arena":
		rows, err = db.arenaFetchAll.Query()
	case "inf_arena":
		rows, err = db.infArenaFetchAll.Query()
	case "reverse":
		rows, err = db.reverseFetchAll.Query()
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]serverapi.LeaderboardEntry, 0, 512)
	for rows.Next() {
		var e serverapi.LeaderboardEntry
		var err error
		switch mode {
		case "classic", "inf_arena":
			err = rows.Scan(&e.PlayerName, &e.Score, &e.Difficulty, &e.Drones, &e.Time, &e.Platform)
		case "arena":
			err = rows.Scan(&e.PlayerName, &e.Score, &e.Difficulty, &e.Drones, &e.Platform)
		case "reverse":
			err = rows.Scan(&e.PlayerName, &e.Score, &e.Difficulty, &e.Time, &e.Platform)
		}
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return entries[:len(entries):len(entries)], nil
}
