package sqliteutil

import (
	"database/sql"
	"fmt"
)

func Connect(dbPath string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", dbPath, err)
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping %q: %w", dbPath, err)
	}
	return conn, nil
}
