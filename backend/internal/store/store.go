package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"example.com/dlm/backend/internal/wiremodel"

	_ "modernc.org/sqlite"
)

// ErrNotFound is returned when a model id does not exist.
var ErrNotFound = errors.New("model not found")

// ErrDuplicateName is returned when a model name already exists.
var ErrDuplicateName = errors.New("model name already exists")

// Summary is list metadata for one model.
type Summary struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	LightCount int       `json:"light_count"`
}

// LightDTO is one light for API JSON.
type LightDTO struct {
	ID int     `json:"id"`
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	Z  float64 `json:"z"`
}

// Detail is full model for GET.
type Detail struct {
	Summary
	Lights []LightDTO `json:"lights"`
}

// Store persists models in SQLite.
type Store struct {
	db *sql.DB
}

// Open opens or creates a SQLite database at path (file URL fragment enables foreign keys).
func Open(path string) (*Store, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS models (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS lights (
			model_id TEXT NOT NULL,
			idx INTEGER NOT NULL,
			x REAL NOT NULL,
			y REAL NOT NULL,
			z REAL NOT NULL,
			PRIMARY KEY (model_id, idx),
			FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE
		)`,
	}
	for _, q := range stmts {
		if _, err := s.db.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

// Close releases the database handle.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// List returns all models with light counts, newest first.
func (s *Store) List(ctx context.Context) ([]Summary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT m.id, m.name, m.created_at, COUNT(l.idx) AS n
		FROM models m
		LEFT JOIN lights l ON l.model_id = m.id
		GROUP BY m.id
		ORDER BY m.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Summary
	for rows.Next() {
		var sum Summary
		var created string
		if err := rows.Scan(&sum.ID, &sum.Name, &created, &sum.LightCount); err != nil {
			return nil, err
		}
		t, err := time.Parse(time.RFC3339Nano, created)
		if err != nil {
			t, err = time.Parse(time.RFC3339, created)
			if err != nil {
				return nil, fmt.Errorf("parse created_at: %w", err)
			}
		}
		sum.CreatedAt = t.UTC()
		out = append(out, sum)
	}
	return out, rows.Err()
}

// Get returns model metadata and lights ordered by idx.
func (s *Store) Get(ctx context.Context, id string) (*Detail, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, name, created_at FROM models WHERE id = ?`, id)
	var d Detail
	var created string
	if err := row.Scan(&d.ID, &d.Name, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	t, err := time.Parse(time.RFC3339Nano, created)
	if err != nil {
		t, err = time.Parse(time.RFC3339, created)
		if err != nil {
			return nil, fmt.Errorf("parse created_at: %w", err)
		}
	}
	d.CreatedAt = t.UTC()

	rows, err := s.db.QueryContext(ctx, `
		SELECT idx, x, y, z FROM lights WHERE model_id = ? ORDER BY idx ASC
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var L LightDTO
		if err := rows.Scan(&L.ID, &L.X, &L.Y, &L.Z); err != nil {
			return nil, err
		}
		d.Lights = append(d.Lights, L)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	d.LightCount = len(d.Lights)
	return &d, nil
}

// Create inserts a model and its lights in one transaction.
func (s *Store) Create(ctx context.Context, name string, lights []wiremodel.Light) (*Summary, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	id := uuid.NewString()
	created := time.Now().UTC().Format(time.RFC3339Nano)

	if _, err := tx.ExecContext(ctx, `INSERT INTO models (id, name, created_at) VALUES (?, ?, ?)`,
		id, name, created); err != nil {
		if isUniqueConstraint(err) {
			return nil, ErrDuplicateName
		}
		return nil, err
	}

	for _, L := range lights {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO lights (model_id, idx, x, y, z) VALUES (?, ?, ?, ?, ?)`,
			id, L.ID, L.X, L.Y, L.Z,
		); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	t, _ := time.Parse(time.RFC3339Nano, created)
	return &Summary{
		ID:         id,
		Name:       name,
		CreatedAt:  t,
		LightCount: len(lights),
	}, nil
}

// Delete removes a model and its lights.
func (s *Store) Delete(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM models WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func isUniqueConstraint(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "unique constraint") || strings.Contains(s, "constraint failed")
}
