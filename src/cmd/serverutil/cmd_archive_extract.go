package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/quasilyte/roboden-game/sqliteutil"
)

func cmdArchiveExtract(args []string) error {
	fs := flag.NewFlagSet("runticks exec", flag.ExitOnError)
	dbPath := fs.String("queue", "", "path to the queue db file")
	outputName := fs.String("o", "replay.json", "output file name")
	replayID := fs.Uint("id", 0, "archived replay id")
	fs.Parse(args)

	if *outputName == "" {
		return errors.New("output file name can't be empty")
	}
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

	var compressedData []byte
	querySQL := fmt.Sprintf(`
		SELECT replay_json
		FROM failed_replay_archive
		WHERE replay_ID = %d
	`, *replayID)
	if err := db.QueryRow(querySQL).Scan(&compressedData); err != nil {
		return fmt.Errorf("fetch replay: %w", err)
	}

	data, err := gzipUncompress(compressedData)
	if err != nil {
		return fmt.Errorf("uncompress replay: %w", err)
	}
	if err := os.WriteFile(*outputName, data, os.ModePerm); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}
