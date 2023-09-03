package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/quasilyte/roboden-game/sqliteutil"
)

func cmdArchiveExtract(args []string) error {
	fs := flag.NewFlagSet("runticks exec", flag.ExitOnError)
	dbPath := fs.String("queue", "", "path to the queue db file")
	replayID := fs.Uint("id", 0, "archived replay id")
	fs.Parse(args)

	if *dbPath == "" {
		return errors.New("queue filename can't be empty")
	}
	if *replayID == 0 {
		return errors.New("replay ID can't be 0")
	}

	db, err := sqliteutil.Connect(*dbPath)
	if err != nil {
		return fmt.Errorf("connect to %q: %w", *dbPath, err)
	}

	var data []byte
	querySQL := fmt.Sprintf(`
		SELECT replay_json
		FROM failed_replay_archive
		WHERE replay_ID = %d
	`, *replayID)
	if err := db.QueryRow(querySQL).Scan(&data); err != nil {
		return fmt.Errorf("fetch replay: %w", err)
	}
	fmt.Println("len", len(data))

	return nil
}
