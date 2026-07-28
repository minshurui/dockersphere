package audit

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Record represents an audit log entry.
type Record struct {
	ID        int64     `json:"id"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	User      string    `json:"user"`
	Detail    string    `json:"detail,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Store manages audit records in SQLite.
type Store struct {
	db *sql.DB
}

// NewStore creates a new audit store.
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open audit db: %w", err)
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate audit db: %w", err)
	}
	return s, nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS audit_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			action TEXT NOT NULL,
			target TEXT NOT NULL,
			user TEXT NOT NULL DEFAULT 'system',
			detail TEXT,
			created_at DATETIME NOT NULL
		)
	`)
	return err
}

// Record logs an audit entry.
func (s *Store) Record(ctx context.Context, action, target, user, detail string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO audit_logs (action, target, user, detail, created_at) VALUES (?, ?, ?, ?, ?)`,
		action, target, user, detail, time.Now(),
	)
	return err
}

// List returns recent audit records.
func (s *Store) List(ctx context.Context, limit int) ([]*Record, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, action, target, user, detail, created_at FROM audit_logs ORDER BY created_at DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*Record
	for rows.Next() {
		r := &Record{}
		if err := rows.Scan(&r.ID, &r.Action, &r.Target, &r.User, &r.Detail, &r.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}
