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
			SELECT player_name, score, difficulty, drones, time_seconds
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
			('player_name', 'score', 'difficulty', 'drones', 'time_seconds')
		VALUES
			(?, ?, ?, ?, ?)
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
			SELECT player_name, score, difficulty, drones
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
			('player_name', 'score', 'difficulty', 'drones')
		VALUES
			(?, ?, ?, ?)
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
			SELECT player_name, score, difficulty, drones, time_seconds
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
			('player_name', 'score', 'difficulty', 'drones', 'time_seconds')
		VALUES
			(?, ?, ?, ?, ?)
		`
		stmt, err := db.conn.Prepare(q)
		if err != nil {
			return err
		}
		db.infArenaUpsert = stmt
	}

	return nil
}

func (db *seasonDB) UpdatePlayerScore(mode, name, drones string, score, difficulty, timeSeconds int) error {
	var err error
	switch mode {
	case "classic":
		_, err = db.classicUpsert.Exec(name, score, difficulty, drones, timeSeconds)
	case "arena":
		_, err = db.arenaUpsert.Exec(name, score, difficulty, drones)
	case "inf_arena":
		_, err = db.infArenaUpsert.Exec(name, score, difficulty, drones, timeSeconds)
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
			err = rows.Scan(&e.PlayerName, &e.Score, &e.Difficulty, &e.Drones, &e.Time)
		case "arena":
			err = rows.Scan(&e.PlayerName, &e.Score, &e.Difficulty, &e.Drones)
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
