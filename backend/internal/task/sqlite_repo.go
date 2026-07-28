package task

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteRepo is a SQLite-backed implementation of Repository.
type SQLiteRepo struct {
	db *sql.DB
}

// NewSQLiteRepo creates a new SQLite task repository.
func NewSQLiteRepo(dbPath string) (*SQLiteRepo, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	r := &SQLiteRepo{db: db}
	if err := r.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return r, nil
}

func (r *SQLiteRepo) migrate() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			action TEXT NOT NULL,
			target TEXT NOT NULL,
			status TEXT NOT NULL,
			result TEXT,
			error TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`)
	return err
}

func (r *SQLiteRepo) Save(_ context.Context, t *Task) error {
	_, err := r.db.Exec(
		`INSERT INTO tasks (id, action, target, status, result, error, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.Action, t.Target, t.Status, t.Result, t.Err, t.CreatedAt, t.UpdatedAt,
	)
	return err
}

func (r *SQLiteRepo) FindByID(_ context.Context, id string) (*Task, error) {
	row := r.db.QueryRow(
		`SELECT id, action, target, status, result, error, created_at, updated_at FROM tasks WHERE id = ?`, id,
	)
	return scanTask(row)
}

func (r *SQLiteRepo) List(_ context.Context) ([]*Task, error) {
	rows, err := r.db.Query(
		`SELECT id, action, target, status, result, error, created_at, updated_at FROM tasks ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		t, err := scanTaskRow(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (r *SQLiteRepo) Update(_ context.Context, t *Task) error {
	res, err := r.db.Exec(
		`UPDATE tasks SET action=?, target=?, status=?, result=?, error=?, updated_at=? WHERE id=?`,
		t.Action, t.Target, t.Status, t.Result, t.Err, t.UpdatedAt, t.ID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("task %s not found", t.ID)
	}
	return nil
}

func scanTask(row *sql.Row) (*Task, error) {
	t := &Task{}
	err := row.Scan(&t.ID, &t.Action, &t.Target, &t.Status, &t.Result, &t.Err, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func scanTaskRow(rows *sql.Rows) (*Task, error) {
	t := &Task{}
	err := rows.Scan(&t.ID, &t.Action, &t.Target, &t.Status, &t.Result, &t.Err, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}
