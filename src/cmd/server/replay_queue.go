package main

import (
	"database/sql"
	"encoding/json"

	"github.com/quasilyte/roboden-game/serverapi"
)

type replayQueue struct {
	conn *sql.DB

	checksumOwner    *sql.Stmt
	addChecksum      *sql.Stmt
	countStmt        *sql.Stmt
	countForPlayer   *sql.Stmt
	pushStmt         *sql.Stmt
	selectNextStmt   *sql.Stmt
	deleteByIDStmt   *sql.Stmt
	addToArchiveStmt *sql.Stmt
}

func newReplayQueue(conn *sql.DB) *replayQueue {
	return &replayQueue{conn: conn}
}

func (q *replayQueue) PrepareQueries() error {
	{
		stmt, err := q.conn.Prepare("SELECT player_name FROM replay_checksums WHERE replay_hash = ?")
		if err != nil {
			return err
		}
		q.checksumOwner = stmt
	}

	{
		stmt, err := q.conn.Prepare(`
			INSERT INTO replay_checksums
			       ('replay_hash', 'player_name')
			VALUES (?, ?)
		`)
		if err != nil {
			return err
		}
		q.addChecksum = stmt
	}

	{
		stmt, err := q.conn.Prepare(`
			INSERT INTO replay_queue
			       ('player_name', 'created_at', 'replay_json')
			VALUES (?, ?, ?)
		`)
		if err != nil {
			return err
		}
		q.pushStmt = stmt
	}

	{
		stmt, err := q.conn.Prepare(`
			SELECT COUNT(*) FROM replay_queue
		`)
		if err != nil {
			return err
		}
		q.countStmt = stmt
	}

	{
		stmt, err := q.conn.Prepare(`
			SELECT COUNT(*)
			FROM replay_queue
			WHERE player_name = ?
		`)
		if err != nil {
			return err
		}
		q.countForPlayer = stmt
	}

	{
		stmt, err := q.conn.Prepare(`
			SELECT id, player_name, replay_json
			FROM replay_queue
			ORDER BY id
			LIMIT 1
		`)
		if err != nil {
			return err
		}
		q.selectNextStmt = stmt
	}

	{
		stmt, err := q.conn.Prepare(`
			DELETE FROM replay_queue
			WHERE id = ?
		`)
		if err != nil {
			return err
		}
		q.deleteByIDStmt = stmt
	}

	{
		stmt, err := q.conn.Prepare(`
			INSERT INTO failed_replay_archive
			       ('replay_id', 'player_name', 'created_at', 'replay_json', 'fail_reason')
			VALUES (?, ?, ?, ?, ?)
		`)
		if err != nil {
			return err
		}
		q.addToArchiveStmt = stmt
	}

	return nil
}

func (q *replayQueue) ChecksumOwner(h string) (string, error) {
	var name string
	err := q.checksumOwner.QueryRow(h).Scan(&name)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return name, err
}

func (q *replayQueue) Delete(id int, playerName string) error {
	_, err := q.deleteByIDStmt.Exec(id)
	return err
}

func (q *replayQueue) Get() (int, string, []byte, error) {
	var id int
	var playerName string
	var data []byte
	err := q.selectNextStmt.QueryRow().Scan(&id, &playerName, &data)
	return id, playerName, data, err
}

func (q *replayQueue) Count() (int, error) {
	var result int
	err := q.countStmt.QueryRow().Scan(&result)
	return result, err
}

func (q *replayQueue) CountForPlayer(name string) (int, error) {
	var result int
	err := q.countForPlayer.QueryRow(name).Scan(&result)
	return result, err
}

func (q *replayQueue) Archive(id int, playerName string, createdAt int64, compressedData []byte, reason archiveReason) error {
	return withTransaction(q.conn, func(tx *sql.Tx) error {
		_, err := tx.Stmt(q.addToArchiveStmt).Exec(id, playerName, createdAt, compressedData, int(reason))
		if err != nil {
			return err
		}
		_, err = tx.Stmt(q.deleteByIDStmt).Exec(id)
		return err
	})

}

func (q *replayQueue) PushRaw(checksum, playerName string, createdAt int64, replayData []byte, compressed bool) error {
	if !compressed {
		compressedReplayData, err := gzipCompress(replayData)
		if err != nil {
			return err
		}
		replayData = compressedReplayData
	}
	return withTransaction(q.conn, func(tx *sql.Tx) error {
		_, err := tx.Stmt(q.addChecksum).Exec(checksum, playerName)
		if err != nil {
			return err
		}
		_, err = tx.Stmt(q.pushStmt).Exec(playerName, createdAt, replayData)
		return err
	})
}

func (q *replayQueue) Push(checksum, playerName string, createdAt int64, replay serverapi.GameReplay) error {
	replayData, err := json.Marshal(replay)
	if err != nil {
		return err
	}
	return q.PushRaw(checksum, playerName, createdAt, replayData, false)
}
