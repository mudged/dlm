package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DeviceTypeWLED is the only supported device type in MVP (REQ-035).
const DeviceTypeWLED = "wled"

// ErrDeviceNotFound is returned when a device id does not exist.
var ErrDeviceNotFound = errors.New("device not found")

// ErrDeviceAssignmentConflict is returned when assign would break REQ-036 1:1 rules.
var ErrDeviceAssignmentConflict = errors.New("device assignment conflict")

// ErrInvalidBaseURL is returned when a supplied device base_url fails the
// REQ-035 BR 6 allowlist (must be http:// or https:// with a non-empty host).
// HTTP handlers map this to 400 invalid_base_url per architecture §3.20.
var ErrInvalidBaseURL = errors.New("invalid base_url: must be http:// or https:// with a host")

// normalizeBaseURL enforces the REQ-035 BR 6 allowlist on a device base_url.
// Accepts: scheme http or https, non-empty host, parseable by net/url.Parse.
// Rejects: empty/whitespace input, other schemes (file, ftp, gopher, javascript, …),
// schemes with no host (e.g. "http:" alone). Returns the canonical string with
// any trailing slash trimmed so equality on stored values is stable.
func normalizeBaseURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrInvalidBaseURL
	}
	u, err := url.Parse(trimmed)
	if err != nil {
		return "", ErrInvalidBaseURL
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", ErrInvalidBaseURL
	}
	if u.Host == "" {
		return "", ErrInvalidBaseURL
	}
	return strings.TrimRight(u.String(), "/"), nil
}

// Device is persisted device registry metadata (REQ-035 / architecture §3.3).
type Device struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"`
	Name          string    `json:"name"`
	BaseURL       string    `json:"base_url"`
	WLEDPassword  string    `json:"-"` // never serialized to JSON from store; HTTP layer omits
	ModelID       *string   `json:"model_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// DeviceCreate holds fields for POST /api/v1/devices.
type DeviceCreate struct {
	Type         string
	Name         string
	BaseURL      string
	WLEDPassword string
}

// DevicePatch updates optional fields (non-empty strings overwrite).
type DevicePatch struct {
	Name         *string
	BaseURL      *string
	WLEDPassword *string // set to empty string to clear
}

func (s *Store) ensureDeviceTable(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS devices (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			name TEXT NOT NULL,
			base_url TEXT NOT NULL,
			wled_password TEXT,
			model_id TEXT UNIQUE,
			created_at TEXT NOT NULL,
			FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE SET NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_devices_model_id ON devices(model_id) WHERE model_id IS NOT NULL`,
	}
	for _, q := range stmts {
		if _, err := s.db.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("devices migrate: %w", err)
		}
	}
	return nil
}

func scanDevice(row interface{ Scan(dest ...any) error }) (Device, error) {
	var d Device
	var modelID sql.NullString
	var created string
	var pw sql.NullString
	if err := row.Scan(&d.ID, &d.Type, &d.Name, &d.BaseURL, &pw, &modelID, &created); err != nil {
		return Device{}, err
	}
	if modelID.Valid {
		d.ModelID = &modelID.String
	}
	if pw.Valid {
		d.WLEDPassword = pw.String
	}
	t, err := time.Parse(time.RFC3339Nano, created)
	if err != nil {
		t, err = time.Parse(time.RFC3339, created)
		if err != nil {
			return Device{}, fmt.Errorf("parse device created_at: %w", err)
		}
	}
	d.CreatedAt = t.UTC()
	return d, nil
}

// ListDevices returns all devices, newest first.
func (s *Store) ListDevices(ctx context.Context) ([]Device, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, type, name, base_url, wled_password, model_id, created_at
		FROM devices ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Device
	for rows.Next() {
		d, err := scanDevice(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// GetDevice returns one device by id.
func (s *Store) GetDevice(ctx context.Context, id string) (Device, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, type, name, base_url, wled_password, model_id, created_at
		FROM devices WHERE id = ?
	`, id)
	d, err := scanDevice(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Device{}, ErrDeviceNotFound
		}
		return Device{}, err
	}
	return d, nil
}

// GetDeviceForModel returns the device assigned to the model, or ErrDeviceNotFound if none.
func (s *Store) GetDeviceForModel(ctx context.Context, modelID string) (Device, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, type, name, base_url, wled_password, model_id, created_at
		FROM devices WHERE model_id = ?
	`, modelID)
	d, err := scanDevice(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Device{}, ErrDeviceNotFound
		}
		return Device{}, err
	}
	return d, nil
}

// CreateDevice registers a new device (REQ-035 manual MVP path).
func (s *Store) CreateDevice(ctx context.Context, in DeviceCreate) (Device, error) {
	in.Type = strings.TrimSpace(in.Type)
	in.Name = strings.TrimSpace(in.Name)
	if in.Type == "" {
		in.Type = DeviceTypeWLED
	}
	if in.Type != DeviceTypeWLED {
		return Device{}, fmt.Errorf("unsupported device type %q", in.Type)
	}
	if in.Name == "" {
		return Device{}, fmt.Errorf("name and base_url are required")
	}
	canonical, err := normalizeBaseURL(in.BaseURL)
	if err != nil {
		return Device{}, err
	}
	in.BaseURL = canonical
	id := uuid.NewString()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	var pw any
	if in.WLEDPassword != "" {
		pw = in.WLEDPassword
	} else {
		pw = nil
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO devices (id, type, name, base_url, wled_password, model_id, created_at)
		VALUES (?, ?, ?, ?, ?, NULL, ?)
	`, id, in.Type, in.Name, in.BaseURL, pw, now); err != nil {
		return Device{}, err
	}
	return s.GetDevice(ctx, id)
}

// PatchDevice updates metadata / connection fields.
func (s *Store) PatchDevice(ctx context.Context, id string, p DevicePatch) (Device, error) {
	if _, err := s.GetDevice(ctx, id); err != nil {
		return Device{}, err
	}
	sets := []string{}
	args := []any{}
	if p.Name != nil {
		n := strings.TrimSpace(*p.Name)
		if n == "" {
			return Device{}, fmt.Errorf("name cannot be empty")
		}
		sets = append(sets, "name = ?")
		args = append(args, n)
	}
	if p.BaseURL != nil {
		canonical, err := normalizeBaseURL(*p.BaseURL)
		if err != nil {
			return Device{}, err
		}
		sets = append(sets, "base_url = ?")
		args = append(args, canonical)
	}
	if p.WLEDPassword != nil {
		sets = append(sets, "wled_password = ?")
		if *p.WLEDPassword == "" {
			args = append(args, nil)
		} else {
			args = append(args, *p.WLEDPassword)
		}
	}
	if len(sets) == 0 {
		return s.GetDevice(ctx, id)
	}
	args = append(args, id)
	q := "UPDATE devices SET " + strings.Join(sets, ", ") + " WHERE id = ?"
	if _, err := s.db.ExecContext(ctx, q, args...); err != nil {
		return Device{}, err
	}
	return s.GetDevice(ctx, id)
}

// DeleteDevice removes a device and clears assignment (REQ-037 BR6).
func (s *Store) DeleteDevice(ctx context.Context, id string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM devices WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrDeviceNotFound
	}
	return nil
}

// AssignDevice links device to model with REQ-036 validation.
func (s *Store) AssignDevice(ctx context.Context, deviceID, modelID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var existingModel sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT model_id FROM devices WHERE id = ?`, deviceID).Scan(&existingModel)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrDeviceNotFound
		}
		return err
	}
	if existingModel.Valid && existingModel.String != "" {
		return fmt.Errorf("%w: device already assigned", ErrDeviceAssignmentConflict)
	}

	var mid string
	err = tx.QueryRowContext(ctx, `SELECT id FROM models WHERE id = ?`, modelID).Scan(&mid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	var other sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT id FROM devices WHERE model_id = ? AND id != ?`, modelID, deviceID).Scan(&other)
	if err == nil {
		return fmt.Errorf("%w: model already has a device", ErrDeviceAssignmentConflict)
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if _, err := tx.ExecContext(ctx, `UPDATE devices SET model_id = ? WHERE id = ?`, modelID, deviceID); err != nil {
		return err
	}
	return tx.Commit()
}

// UnassignDevice clears model_id for the device (REQ-036 / architecture §3.20).
func (s *Store) UnassignDevice(ctx context.Context, deviceID string) error {
	res, err := s.db.ExecContext(ctx, `UPDATE devices SET model_id = NULL WHERE id = ?`, deviceID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrDeviceNotFound
	}
	return nil
}

// ListModelIDsWithDevices returns every model_id currently assigned to any device.
func (s *Store) ListModelIDsWithDevices(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT model_id FROM devices WHERE model_id IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var mid string
		if err := rows.Scan(&mid); err != nil {
			return nil, err
		}
		out = append(out, mid)
	}
	return out, rows.Err()
}
