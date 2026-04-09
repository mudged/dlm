package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Routine type identifiers (architecture §3.16, §3.17).
const (
	RoutineTypeRandomColourCycleAll = "random_colour_cycle_all"
	RoutineTypePythonSceneScript    = "python_scene_script"
	RoutineTypeShapeAnimation       = "shape_animation"
)

// RoutineStatusRunning / RoutineStatusStopped are persisted in routine_runs.status.
const (
	RoutineStatusRunning = "running"
	RoutineStatusStopped = "stopped"
)

// ErrRoutineNotFound is returned when a routine definition id does not exist.
var ErrRoutineNotFound = errors.New("routine not found")

// ErrRoutineRunActive is returned when deleting a routine that has a running instance.
var ErrRoutineRunActive = errors.New("routine has an active run")

// ErrRoutineUnknownType is returned for unsupported routine type strings.
var ErrRoutineUnknownType = errors.New("unknown routine type")

// ErrRoutineNotEditable is returned when PATCH uses fields incompatible with the routine type.
var ErrRoutineNotEditable = errors.New("routine type does not support this update")

// ErrRoutineRunNotFound is returned when stop targets a missing or non-running run for the scene.
var ErrRoutineRunNotFound = errors.New("routine run not found for this scene")

// SceneRoutineConflictError is returned when starting a routine while another runs on the same scene.
type SceneRoutineConflictError struct {
	RunID     string `json:"run_id"`
	RoutineID string `json:"routine_id"`
}

func (e *SceneRoutineConflictError) Error() string {
	return "a routine is already running on this scene"
}

// RoutineDTO is a persisted routine definition (REQ-021, REQ-022).
type RoutineDTO struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	Type            string          `json:"type"`
	PythonSource    string          `json:"python_source,omitempty"`
	DefinitionJSON  json.RawMessage `json:"definition_json,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

// RoutineRunDTO is exposed for list-runs API.
type RoutineRunDTO struct {
	ID          string `json:"id"`
	RoutineID   string `json:"routine_id"`
	RoutineName string `json:"routine_name"`
	RoutineType string `json:"routine_type"`
	Status      string `json:"status"`
}

func scanRoutineDTO(scanner interface {
	Scan(dest ...any) error
}) (RoutineDTO, error) {
	var r RoutineDTO
	var created string
	var def sql.NullString
	if err := scanner.Scan(&r.ID, &r.Name, &r.Description, &r.Type, &r.PythonSource, &def, &created); err != nil {
		return r, err
	}
	if def.Valid && strings.TrimSpace(def.String) != "" {
		r.DefinitionJSON = json.RawMessage(def.String)
	}
	t, err := time.Parse(time.RFC3339Nano, created)
	if err != nil {
		t, err = time.Parse(time.RFC3339, created)
		if err != nil {
			return r, fmt.Errorf("parse routine created_at: %w", err)
		}
	}
	r.CreatedAt = t.UTC()
	return r, nil
}

// ListRoutines returns all routine definitions, newest first.
func (s *Store) ListRoutines(ctx context.Context) ([]RoutineDTO, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, description, type, python_source, definition_json, created_at FROM routines ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RoutineDTO
	for rows.Next() {
		r, err := scanRoutineDTO(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetRoutine loads one routine by id.
func (s *Store) GetRoutine(ctx context.Context, id string) (*RoutineDTO, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, description, type, python_source, definition_json, created_at FROM routines WHERE id = ?
	`, id)
	r, err := scanRoutineDTO(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRoutineNotFound
		}
		return nil, err
	}
	return &r, nil
}

// CreateRoutine inserts a new definition. Empty type defaults to python_scene_script (REQ-023).
// For shape_animation, definitionJSON must validate; pythonSource is stored as "".
// For python_scene_script, definitionJSON should be "" (stored as NULL).
func (s *Store) CreateRoutine(ctx context.Context, name, description, typ, pythonSource, definitionJSON string) (*RoutineDTO, error) {
	name = strings.TrimSpace(name)
	typ = strings.TrimSpace(typ)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if typ == "" {
		typ = RoutineTypePythonSceneScript
	}
	if typ == RoutineTypeRandomColourCycleAll {
		return nil, fmt.Errorf("%w: random_colour_cycle_all is not a creatable type", ErrRoutineUnknownType)
	}
	if typ != RoutineTypePythonSceneScript && typ != RoutineTypeShapeAnimation {
		return nil, fmt.Errorf("%w: %q", ErrRoutineUnknownType, typ)
	}

	var defArg any
	switch typ {
	case RoutineTypePythonSceneScript:
		defArg = nil
	case RoutineTypeShapeAnimation:
		if err := ValidateShapeAnimationDefinitionJSON(definitionJSON); err != nil {
			return nil, err
		}
		defArg = strings.TrimSpace(definitionJSON)
	}

	src := pythonSource
	if typ == RoutineTypeShapeAnimation {
		src = ""
	}

	id := uuid.NewString()
	created := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO routines (id, name, description, type, python_source, definition_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, name, description, typ, src, defArg, created); err != nil {
		return nil, err
	}
	return s.GetRoutine(ctx, id)
}

// PatchRoutine updates fields by routine kind. python_source only for python_scene_script; definition_json only for shape_animation.
func (s *Store) PatchRoutine(ctx context.Context, id string, name, description, pythonSource, definitionJSON *string) (*RoutineDTO, error) {
	if name == nil && description == nil && pythonSource == nil && definitionJSON == nil {
		return nil, fmt.Errorf("at least one field is required")
	}
	r, err := s.GetRoutine(ctx, id)
	if err != nil {
		return nil, err
	}
	if r.Type != RoutineTypePythonSceneScript && r.Type != RoutineTypeShapeAnimation {
		return nil, ErrRoutineNotEditable
	}
	if r.Type == RoutineTypePythonSceneScript && definitionJSON != nil {
		return nil, ErrRoutineNotEditable
	}
	if r.Type == RoutineTypeShapeAnimation && pythonSource != nil {
		return nil, ErrRoutineNotEditable
	}
	var n int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM routine_runs WHERE routine_id = ? AND status = ?
	`, id, RoutineStatusRunning).Scan(&n); err != nil {
		return nil, err
	}
	if n > 0 {
		return nil, ErrRoutineRunActive
	}
	newName := r.Name
	newDesc := r.Description
	newSrc := r.PythonSource
	var newDef sql.NullString
	if len(r.DefinitionJSON) > 0 {
		newDef = sql.NullString{String: string(r.DefinitionJSON), Valid: true}
	}
	if name != nil {
		t := strings.TrimSpace(*name)
		if t == "" {
			return nil, fmt.Errorf("name is required")
		}
		newName = t
	}
	if description != nil {
		newDesc = *description
	}
	if pythonSource != nil {
		newSrc = *pythonSource
	}
	if definitionJSON != nil {
		trimmed := strings.TrimSpace(*definitionJSON)
		if err := ValidateShapeAnimationDefinitionJSON(trimmed); err != nil {
			return nil, err
		}
		newDef = sql.NullString{String: trimmed, Valid: true}
	}

	var execErr error
	if r.Type == RoutineTypePythonSceneScript {
		_, execErr = s.db.ExecContext(ctx, `
			UPDATE routines SET name = ?, description = ?, python_source = ? WHERE id = ?
		`, newName, newDesc, newSrc, id)
	} else {
		var defArg any
		if newDef.Valid {
			defArg = newDef.String
		} else {
			defArg = nil
		}
		_, execErr = s.db.ExecContext(ctx, `
			UPDATE routines SET name = ?, description = ?, definition_json = ? WHERE id = ?
		`, newName, newDesc, defArg, id)
	}
	if execErr != nil {
		return nil, execErr
	}
	return s.GetRoutine(ctx, id)
}

// DeleteRoutine removes a definition if no run is active for it.
func (s *Store) DeleteRoutine(ctx context.Context, id string) error {
	var n int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM routine_runs WHERE routine_id = ? AND status = ?
	`, id, RoutineStatusRunning).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return ErrRoutineRunActive
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM routine_runs WHERE routine_id = ?`, id); err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM routines WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrRoutineNotFound
	}
	return nil
}

// StartRoutineRun creates a running row. Light mutations during a run are performed by
// internal/routineengine (supervised python3 and/or Go shape ticker) calling store §3.15 paths (REQ-038).
// REQ-021 BR5: if any run is already active on this scene (including the same routine), returns
// *SceneRoutineConflictError and does not insert a second row.
func (s *Store) StartRoutineRun(ctx context.Context, sceneID, routineID string) (runID string, err error) {
	if _, err := s.GetRoutine(ctx, routineID); err != nil {
		return "", err
	}
	ok, err := s.sceneExistsTx(ctx, nil, sceneID)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrSceneNotFound
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()

	var existingRunID, existingRid string
	row := tx.QueryRowContext(ctx, `
		SELECT id, routine_id FROM routine_runs WHERE scene_id = ? AND status = ? LIMIT 1
	`, sceneID, RoutineStatusRunning)
	if err := row.Scan(&existingRunID, &existingRid); err == nil {
		return "", &SceneRoutineConflictError{RunID: existingRunID, RoutineID: existingRid}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	runID = uuid.NewString()
	started := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO routine_runs (id, routine_id, scene_id, status, started_at, stopped_at)
		VALUES (?, ?, ?, ?, ?, NULL)
	`, runID, routineID, sceneID, RoutineStatusRunning, started); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return runID, nil
}

// StopRoutineRun marks a run stopped if it belongs to the scene.
func (s *Store) StopRoutineRun(ctx context.Context, sceneID, runID string) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE routine_runs SET status = ?, stopped_at = ? WHERE id = ? AND scene_id = ? AND status = ?
	`, RoutineStatusStopped, time.Now().UTC().Format(time.RFC3339Nano), runID, sceneID, RoutineStatusRunning)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrRoutineRunNotFound
	}
	return nil
}

// ListRunningRoutineRunsForScene returns running runs for the scene (0 or 1 per architecture).
func (s *Store) ListRunningRoutineRunsForScene(ctx context.Context, sceneID string) ([]RoutineRunDTO, error) {
	ok, err := s.sceneExistsTx(ctx, nil, sceneID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrSceneNotFound
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT rr.id, rr.routine_id, r.name, r.type, rr.status
		FROM routine_runs rr
		INNER JOIN routines r ON r.id = rr.routine_id
		WHERE rr.scene_id = ? AND rr.status = ?
	`, sceneID, RoutineStatusRunning)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RoutineRunDTO
	for rows.Next() {
		var dto RoutineRunDTO
		if err := rows.Scan(&dto.ID, &dto.RoutineID, &dto.RoutineName, &dto.RoutineType, &dto.Status); err != nil {
			return nil, err
		}
		out = append(out, dto)
	}
	return out, rows.Err()
}

// StopAllRunningRoutineRuns marks every running row stopped (crash recovery before supervisors restart).
func (s *Store) StopAllRunningRoutineRuns(ctx context.Context) error {
	stoppedAt := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, `
		UPDATE routine_runs SET status = ?, stopped_at = ? WHERE status = ?
	`, RoutineStatusStopped, stoppedAt, RoutineStatusRunning)
	return err
}
