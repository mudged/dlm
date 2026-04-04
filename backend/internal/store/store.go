package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"example.com/dlm/backend/internal/samples"
	"example.com/dlm/backend/internal/wiremodel"

	_ "modernc.org/sqlite"
)

// ErrNotFound is returned when a model id does not exist.
var ErrNotFound = errors.New("model not found")

// ErrDuplicateName is returned when a model name already exists.
var ErrDuplicateName = errors.New("model name already exists")

// ErrInvalidLightIndex is returned when a light id is not part of the model.
var ErrInvalidLightIndex = errors.New("light index out of range")

// ErrBatchEmptyIDs is returned when a batch patch has no ids.
var ErrBatchEmptyIDs = errors.New("batch ids must be non-empty")

// ErrBatchDuplicateIDs is returned when the same light id appears more than once in a batch.
var ErrBatchDuplicateIDs = errors.New("duplicate light ids in batch")

// Default light state for new rows (REQ-011 / REQ-014 / architecture §3.9).
const (
	DefaultLightOn            = false
	DefaultLightColor         = "#ffffff"
	DefaultLightBrightnessPct = 100.0
)

var hexColorRe = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// Summary is list metadata for one model.
type Summary struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	LightCount int       `json:"light_count"`
}

// LightDTO is one light for API JSON (positions + per-light state).
type LightDTO struct {
	ID            int     `json:"id"`
	X             float64 `json:"x"`
	Y             float64 `json:"y"`
	Z             float64 `json:"z"`
	On            bool    `json:"on"`
	Color         string  `json:"color"`
	BrightnessPct float64 `json:"brightness_pct"`
}

// LightStateDTO is the state-only payload for GET …/lights/state endpoints.
type LightStateDTO struct {
	ID            int     `json:"id"`
	On            bool    `json:"on"`
	Color         string  `json:"color"`
	BrightnessPct float64 `json:"brightness_pct"`
}

// LightStatePatch is a partial update; nil fields are unchanged.
type LightStatePatch struct {
	On            *bool
	Color         *string
	BrightnessPct *float64
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
			"on" INTEGER NOT NULL DEFAULT 0,
			color TEXT NOT NULL DEFAULT '#ffffff',
			brightness_pct REAL NOT NULL DEFAULT 100,
			PRIMARY KEY (model_id, idx),
			FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE
		)`,
	}
	for _, q := range stmts {
		if _, err := s.db.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	if err := s.ensureLightStateColumns(ctx); err != nil {
		return err
	}
	return s.ensureSceneTables(ctx)
}

func (s *Store) ensureSceneTables(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS scenes (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS scene_models (
			scene_id TEXT NOT NULL,
			model_id TEXT NOT NULL,
			offset_x INTEGER NOT NULL,
			offset_y INTEGER NOT NULL,
			offset_z INTEGER NOT NULL,
			ordinal INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (scene_id, model_id),
			FOREIGN KEY (scene_id) REFERENCES scenes(id) ON DELETE CASCADE,
			FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE RESTRICT
		)`,
	}
	for _, q := range stmts {
		if _, err := s.db.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("migrate scenes: %w", err)
		}
	}
	if err := s.ensureSceneModelOrdinal(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Store) ensureSceneModelOrdinal(ctx context.Context) error {
	cols, err := s.tableColumns(ctx, "scene_models")
	if err != nil {
		return err
	}
	if cols["ordinal"] {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, `ALTER TABLE scene_models ADD COLUMN ordinal INTEGER NOT NULL DEFAULT 0`); err != nil {
		return fmt.Errorf("migrate scene_models.ordinal: %w", err)
	}
	return s.backfillSceneModelOrdinals(ctx)
}

func (s *Store) backfillSceneModelOrdinals(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT scene_id FROM scene_models`)
	if err != nil {
		return err
	}
	var sceneIDs []string
	for rows.Next() {
		var sid string
		if err := rows.Scan(&sid); err != nil {
			_ = rows.Close()
			return err
		}
		sceneIDs = append(sceneIDs, sid)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, sid := range sceneIDs {
		r2, err := s.db.QueryContext(ctx, `
			SELECT model_id FROM scene_models WHERE scene_id = ? ORDER BY model_id ASC
		`, sid)
		if err != nil {
			return err
		}
		i := 0
		for r2.Next() {
			var mid string
			if err := r2.Scan(&mid); err != nil {
				_ = r2.Close()
				return err
			}
			if _, err := s.db.ExecContext(ctx, `
				UPDATE scene_models SET ordinal = ? WHERE scene_id = ? AND model_id = ?
			`, i, sid, mid); err != nil {
				_ = r2.Close()
				return err
			}
			i++
		}
		if err := r2.Err(); err != nil {
			_ = r2.Close()
			return err
		}
		if err := r2.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ensureLightStateColumns(ctx context.Context) error {
	// Add per-light state columns for existing databases (idempotent).
	cols, err := s.tableColumns(ctx, "lights")
	if err != nil {
		return err
	}
	if !cols["on"] {
		if _, err := s.db.ExecContext(ctx, `ALTER TABLE lights ADD COLUMN "on" INTEGER NOT NULL DEFAULT 0`); err != nil {
			return fmt.Errorf("migrate lights.on: %w", err)
		}
	}
	if !cols["color"] {
		if _, err := s.db.ExecContext(ctx, `ALTER TABLE lights ADD COLUMN color TEXT NOT NULL DEFAULT '#ffffff'`); err != nil {
			return fmt.Errorf("migrate lights.color: %w", err)
		}
	}
	if !cols["brightness_pct"] {
		if _, err := s.db.ExecContext(ctx, `ALTER TABLE lights ADD COLUMN brightness_pct REAL NOT NULL DEFAULT 100`); err != nil {
			return fmt.Errorf("migrate lights.brightness_pct: %w", err)
		}
	}
	return nil
}

func (s *Store) tableColumns(ctx context.Context, table string) (map[string]bool, error) {
	rows, err := s.db.QueryContext(ctx, `PRAGMA table_info(`+table+`)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		out[name] = true
	}
	return out, rows.Err()
}

// seedThreeCanonicalSamplesTx inserts the three REQ-009 samples inside an existing transaction (REQ-017 factory reset).
func (s *Store) seedThreeCanonicalSamplesTx(ctx context.Context, tx *sql.Tx) error {
	seed := []struct {
		name   string
		lights []wiremodel.Light
	}{
		{samples.NameSphere, samples.SphereLights()},
		{samples.NameCube, samples.CubeLights()},
		{samples.NameCone, samples.ConeLights()},
	}
	for _, m := range seed {
		id := uuid.NewString()
		created := time.Now().UTC().Format(time.RFC3339Nano)
		if _, err := tx.ExecContext(ctx, `INSERT INTO models (id, name, created_at) VALUES (?, ?, ?)`,
			id, m.name, created); err != nil {
			return err
		}
		for _, L := range m.lights {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO lights (model_id, idx, x, y, z, "on", color, brightness_pct) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				id, L.ID, L.X, L.Y, L.Z, boolToInt(DefaultLightOn), DefaultLightColor, DefaultLightBrightnessPct,
			); err != nil {
				return err
			}
		}
	}
	return nil
}

// SeedDefaultSamples inserts the three REQ-009 geometric samples when the models table is empty.
func (s *Store) SeedDefaultSamples(ctx context.Context) error {
	var n int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM models`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := s.seedThreeCanonicalSamplesTx(ctx, tx); err != nil {
		return err
	}
	return tx.Commit()
}

// FactoryReset deletes all scenes, models, and lights, then re-seeds the three default samples (REQ-017 / architecture §3.14).
func (s *Store) FactoryReset(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// scene_models rows CASCADE when scenes are removed (FK on scene_id).
	if _, err := tx.ExecContext(ctx, `DELETE FROM scenes`); err != nil {
		return fmt.Errorf("factory reset delete scenes: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM lights`); err != nil {
		return fmt.Errorf("factory reset delete lights: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM models`); err != nil {
		return fmt.Errorf("factory reset delete models: %w", err)
	}
	if err := s.seedThreeCanonicalSamplesTx(ctx, tx); err != nil {
		return fmt.Errorf("factory reset seed samples: %w", err)
	}
	return tx.Commit()
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
		SELECT idx, x, y, z, "on", color, brightness_pct FROM lights WHERE model_id = ? ORDER BY idx ASC
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var L LightDTO
		var onInt int
		if err := rows.Scan(&L.ID, &L.X, &L.Y, &L.Z, &onInt, &L.Color, &L.BrightnessPct); err != nil {
			return nil, err
		}
		L.On = onInt != 0
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
			`INSERT INTO lights (model_id, idx, x, y, z, "on", color, brightness_pct) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			id, L.ID, L.X, L.Y, L.Z, boolToInt(DefaultLightOn), DefaultLightColor, DefaultLightBrightnessPct,
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
	refs, err := s.listScenesReferencingModel(ctx, id)
	if err != nil {
		return err
	}
	if len(refs) > 0 {
		return &ModelInUseError{Scenes: refs}
	}
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

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ValidateColor returns a normalized lowercase hex color or an error.
func ValidateColor(s string) (string, error) {
	s = strings.TrimSpace(s)
	if !hexColorRe.MatchString(s) {
		return "", fmt.Errorf("color must be #RRGGBB with six hex digits")
	}
	return strings.ToLower(s), nil
}

// ValidateBrightnessPct checks range [0, 100] and finite.
func ValidateBrightnessPct(v float64) error {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return fmt.Errorf("brightness_pct must be a finite number")
	}
	if v < 0 || v > 100 {
		return fmt.Errorf("brightness_pct must be between 0 and 100")
	}
	return nil
}

func (s *Store) modelExists(ctx context.Context, modelID string) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM models WHERE id = ?`, modelID).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// ListLightStates returns state for all lights in a model (REQ-011).
func (s *Store) ListLightStates(ctx context.Context, modelID string) ([]LightStateDTO, error) {
	ok, err := s.modelExists(ctx, modelID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFound
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT idx, "on", color, brightness_pct FROM lights WHERE model_id = ? ORDER BY idx ASC
	`, modelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LightStateDTO
	for rows.Next() {
		var st LightStateDTO
		var onInt int
		if err := rows.Scan(&st.ID, &onInt, &st.Color, &st.BrightnessPct); err != nil {
			return nil, err
		}
		st.On = onInt != 0
		out = append(out, st)
	}
	return out, rows.Err()
}

// GetLightState returns one light's state (REQ-011).
func (s *Store) GetLightState(ctx context.Context, modelID string, lightID int) (*LightStateDTO, error) {
	ok, err := s.modelExists(ctx, modelID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFound
	}
	row := s.db.QueryRowContext(ctx, `
		SELECT idx, "on", color, brightness_pct FROM lights WHERE model_id = ? AND idx = ?
	`, modelID, lightID)
	var st LightStateDTO
	var onInt int
	if err := row.Scan(&st.ID, &onInt, &st.Color, &st.BrightnessPct); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidLightIndex
		}
		return nil, err
	}
	st.On = onInt != 0
	return &st, nil
}

// PatchLightState applies a partial update and returns the merged state (REQ-011).
func (s *Store) PatchLightState(ctx context.Context, modelID string, lightID int, patch LightStatePatch) (*LightStateDTO, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var n int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM models WHERE id = ?`, modelID).Scan(&n); err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, ErrNotFound
	}

	row := tx.QueryRowContext(ctx, `
		SELECT "on", color, brightness_pct FROM lights WHERE model_id = ? AND idx = ?
	`, modelID, lightID)
	var onInt int
	var color string
	var brightness float64
	if err := row.Scan(&onInt, &color, &brightness); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidLightIndex
		}
		return nil, err
	}
	on := onInt != 0

	if patch.On != nil {
		on = *patch.On
	}
	if patch.Color != nil {
		c, err := ValidateColor(*patch.Color)
		if err != nil {
			return nil, err
		}
		color = c
	}
	if patch.BrightnessPct != nil {
		if err := ValidateBrightnessPct(*patch.BrightnessPct); err != nil {
			return nil, err
		}
		brightness = *patch.BrightnessPct
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE lights SET "on" = ?, color = ?, brightness_pct = ? WHERE model_id = ? AND idx = ?
	`, boolToInt(on), color, brightness, modelID, lightID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &LightStateDTO{ID: lightID, On: on, Color: color, BrightnessPct: brightness}, nil
}

// BatchPatchLightStates applies the same partial update to many lights in one transaction (REQ-013 / architecture §3.10).
// ids may be in any order; returned states are sorted by id ascending. Duplicate ids yield ErrBatchDuplicateIDs.
func (s *Store) BatchPatchLightStates(ctx context.Context, modelID string, ids []int, patch LightStatePatch) ([]LightStateDTO, error) {
	if len(ids) == 0 {
		return nil, ErrBatchEmptyIDs
	}
	if patch.On == nil && patch.Color == nil && patch.BrightnessPct == nil {
		return nil, fmt.Errorf("batch patch requires at least one of on, color, brightness_pct")
	}
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if _, dup := seen[id]; dup {
			return nil, ErrBatchDuplicateIDs
		}
		seen[id] = struct{}{}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var modelCount int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM models WHERE id = ?`, modelID).Scan(&modelCount); err != nil {
		return nil, err
	}
	if modelCount == 0 {
		return nil, ErrNotFound
	}

	var lightCount int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM lights WHERE model_id = ?`, modelID).Scan(&lightCount); err != nil {
		return nil, err
	}

	sorted := slices.Clone(ids)
	slices.Sort(sorted)

	out := make([]LightStateDTO, 0, len(sorted))
	for _, lightID := range sorted {
		if lightID < 0 || lightID >= lightCount {
			return nil, ErrInvalidLightIndex
		}

		row := tx.QueryRowContext(ctx, `
			SELECT "on", color, brightness_pct FROM lights WHERE model_id = ? AND idx = ?
		`, modelID, lightID)
		var onInt int
		var color string
		var brightness float64
		if err := row.Scan(&onInt, &color, &brightness); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrInvalidLightIndex
			}
			return nil, err
		}
		on := onInt != 0

		if patch.On != nil {
			on = *patch.On
		}
		if patch.Color != nil {
			c, err := ValidateColor(*patch.Color)
			if err != nil {
				return nil, err
			}
			color = c
		}
		if patch.BrightnessPct != nil {
			if err := ValidateBrightnessPct(*patch.BrightnessPct); err != nil {
				return nil, err
			}
			brightness = *patch.BrightnessPct
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE lights SET "on" = ?, color = ?, brightness_pct = ? WHERE model_id = ? AND idx = ?
		`, boolToInt(on), color, brightness, modelID, lightID); err != nil {
			return nil, err
		}
		out = append(out, LightStateDTO{
			ID:            lightID,
			On:            on,
			Color:         color,
			BrightnessPct: brightness,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

// ResetAllLightStates sets every light in the model to REQ-014 defaults (off, #ffffff, 100% brightness).
func (s *Store) ResetAllLightStates(ctx context.Context, modelID string) ([]LightStateDTO, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var n int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM models WHERE id = ?`, modelID).Scan(&n); err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, ErrNotFound
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE lights SET "on" = 0, color = ?, brightness_pct = ? WHERE model_id = ?
	`, DefaultLightColor, DefaultLightBrightnessPct, modelID); err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT idx, "on", color, brightness_pct FROM lights WHERE model_id = ? ORDER BY idx ASC
	`, modelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LightStateDTO
	for rows.Next() {
		var st LightStateDTO
		var onInt int
		if err := rows.Scan(&st.ID, &onInt, &st.Color, &st.BrightnessPct); err != nil {
			return nil, err
		}
		st.On = onInt != 0
		out = append(out, st)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}
