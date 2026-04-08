package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const sceneCoordEps = 1e-9

// DefaultSceneBoundaryMarginM is REQ-034 / REQ-015 default padding (30 cm).
const DefaultSceneBoundaryMarginM = 0.3

// MaxSceneBoundaryMarginM caps user-editable scene boundary padding (architecture §3.13).
const MaxSceneBoundaryMarginM = 10.0

// ErrSceneNotFound is returned when a scene id does not exist.
var ErrSceneNotFound = errors.New("scene not found")

// ErrDuplicateSceneName is returned when a scene name already exists.
var ErrDuplicateSceneName = errors.New("scene name already exists")

// ErrSceneLastModel is returned when removing the last model from a scene (client should confirm scene delete).
var ErrSceneLastModel = errors.New("last model in scene requires scene deletion")

// ErrModelAlreadyInScene is returned when adding a model that is already in the scene.
var ErrModelAlreadyInScene = errors.New("model already in scene")

// ErrSceneInvalidGeometry is returned when scene cuboid/sphere geometry is invalid.
var ErrSceneInvalidGeometry = errors.New("invalid scene geometry")

// SceneRef is a minimal scene reference for error payloads.
type SceneRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ModelInUseError is returned from Delete(model) when the model appears in one or more scenes.
type ModelInUseError struct {
	Scenes []SceneRef
}

func (e *ModelInUseError) Error() string {
	return "model is assigned to one or more scenes"
}

// SceneSummary is list/create metadata for one scene.
type SceneSummary struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	ModelCount int       `json:"model_count"`
}

// SceneModelPlacement is one model’s offsets in a scene (meters, integers).
type SceneModelPlacement struct {
	ModelID string `json:"model_id"`
	OffsetX int    `json:"offset_x"`
	OffsetY int    `json:"offset_y"`
	OffsetZ int    `json:"offset_z"`
}

// SceneLightDTO is one light in scene detail (model-local + scene-space).
type SceneLightDTO struct {
	ID            int     `json:"id"`
	X             float64 `json:"x"`
	Y             float64 `json:"y"`
	Z             float64 `json:"z"`
	Sx            float64 `json:"sx"`
	Sy            float64 `json:"sy"`
	Sz            float64 `json:"sz"`
	On            bool    `json:"on"`
	Color         string  `json:"color"`
	BrightnessPct float64 `json:"brightness_pct"`
}

// SceneItemDetail is one model block inside GET scene.
type SceneItemDetail struct {
	ModelID string          `json:"model_id"`
	Name    string          `json:"name"`
	OffsetX int             `json:"offset_x"`
	OffsetY int             `json:"offset_y"`
	OffsetZ int             `json:"offset_z"`
	Lights  []SceneLightDTO `json:"lights"`
}

// SceneDetail is the full GET /scenes/{id} payload.
type SceneDetail struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	CreatedAt          time.Time         `json:"created_at"`
	BoundaryMarginM    float64           `json:"boundary_margin_m"`
	Items              []SceneItemDetail `json:"items"`
}

// ScenePoint is a point in scene space.
type ScenePoint struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// SceneDimensionsSize is a width/height/depth triple in scene space.
type SceneDimensionsSize struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Depth  float64 `json:"depth"`
}

// SceneDimensions describes scene-space extents used for spatial querying.
type SceneDimensions struct {
	Origin  ScenePoint          `json:"origin"`
	Size    SceneDimensionsSize `json:"size"`
	Max     ScenePoint          `json:"max"`
	MarginM float64             `json:"margin_m"`
}

// SceneCuboid is a cuboid query/update geometry in scene space.
type SceneCuboid struct {
	Position   ScenePoint          `json:"position"`
	Dimensions SceneDimensionsSize `json:"dimensions"`
}

// SceneSphere is a sphere query/update geometry in scene space.
type SceneSphere struct {
	Center ScenePoint `json:"center"`
	Radius float64    `json:"radius"`
}

// SceneLightFlat is a flattened scene light record for scene-space endpoints.
type SceneLightFlat struct {
	SceneID       string  `json:"scene_id"`
	ModelID       string  `json:"model_id"`
	LightID       int     `json:"light_id"`
	X             float64 `json:"x"`
	Y             float64 `json:"y"`
	Z             float64 `json:"z"`
	Sx            float64 `json:"sx"`
	Sy            float64 `json:"sy"`
	Sz            float64 `json:"sz"`
	On            bool    `json:"on"`
	Color         string  `json:"color"`
	BrightnessPct float64 `json:"brightness_pct"`
}

// ScenePatchedState is one updated state item for scene-space bulk updates.
type ScenePatchedState struct {
	ModelID       string  `json:"model_id"`
	ID            int     `json:"id"`
	On            bool    `json:"on"`
	Color         string  `json:"color"`
	BrightnessPct float64 `json:"brightness_pct"`
	Sx            float64 `json:"sx"`
	Sy            float64 `json:"sy"`
	Sz            float64 `json:"sz"`
}

// SceneBulkPatchResult is the response payload for scene-space bulk updates.
type SceneBulkPatchResult struct {
	UpdatedCount int                 `json:"updated_count"`
	States       []ScenePatchedState `json:"states"`
	// UnchangedAll is true when every matched light was already equivalent to the merged state (REQ-031).
	UnchangedAll bool `json:"unchanged_all,omitempty"`
}

// SceneBatchLightUpdate is one per-light patch inside PATCH …/lights/state/batch (REQ-020 / REQ-021).
type SceneBatchLightUpdate struct {
	ModelID string
	LightID int
	Patch   LightStatePatch
}

func isFiniteFloat64(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

func validateSceneCuboid(c SceneCuboid) error {
	if !isFiniteFloat64(c.Position.X) || !isFiniteFloat64(c.Position.Y) || !isFiniteFloat64(c.Position.Z) {
		return fmt.Errorf("%w: cuboid position must contain finite numbers", ErrSceneInvalidGeometry)
	}
	if !isFiniteFloat64(c.Dimensions.Width) || !isFiniteFloat64(c.Dimensions.Height) || !isFiniteFloat64(c.Dimensions.Depth) {
		return fmt.Errorf("%w: cuboid dimensions must contain finite numbers", ErrSceneInvalidGeometry)
	}
	if c.Dimensions.Width <= 0 || c.Dimensions.Height <= 0 || c.Dimensions.Depth <= 0 {
		return fmt.Errorf("%w: cuboid dimensions must be positive", ErrSceneInvalidGeometry)
	}
	return nil
}

func validateSceneSphere(sph SceneSphere) error {
	if !isFiniteFloat64(sph.Center.X) || !isFiniteFloat64(sph.Center.Y) || !isFiniteFloat64(sph.Center.Z) {
		return fmt.Errorf("%w: sphere center must contain finite numbers", ErrSceneInvalidGeometry)
	}
	if !isFiniteFloat64(sph.Radius) {
		return fmt.Errorf("%w: sphere radius must be a finite number", ErrSceneInvalidGeometry)
	}
	if sph.Radius <= 0 {
		return fmt.Errorf("%w: sphere radius must be positive", ErrSceneInvalidGeometry)
	}
	return nil
}

func validateScenePatch(patch LightStatePatch) error {
	if patch.On == nil && patch.Color == nil && patch.BrightnessPct == nil {
		return fmt.Errorf("at least one of on, color, brightness_pct is required")
	}
	if patch.Color != nil {
		if _, err := ValidateColor(*patch.Color); err != nil {
			return err
		}
	}
	if patch.BrightnessPct != nil {
		if err := ValidateBrightnessPct(*patch.BrightnessPct); err != nil {
			return err
		}
	}
	return nil
}

func cuboidContains(c SceneCuboid, L SceneLightFlat) bool {
	maxX := c.Position.X + c.Dimensions.Width
	maxY := c.Position.Y + c.Dimensions.Height
	maxZ := c.Position.Z + c.Dimensions.Depth
	return L.Sx >= c.Position.X-sceneCoordEps && L.Sx <= maxX+sceneCoordEps &&
		L.Sy >= c.Position.Y-sceneCoordEps && L.Sy <= maxY+sceneCoordEps &&
		L.Sz >= c.Position.Z-sceneCoordEps && L.Sz <= maxZ+sceneCoordEps
}

func sphereContains(sph SceneSphere, L SceneLightFlat) bool {
	dx := L.Sx - sph.Center.X
	dy := L.Sy - sph.Center.Y
	dz := L.Sz - sph.Center.Z
	dist2 := dx*dx + dy*dy + dz*dz
	return dist2 <= sph.Radius*sph.Radius+sceneCoordEps
}

func (s *Store) sceneExistsTx(ctx context.Context, tx *sql.Tx, sceneID string) (bool, error) {
	var row *sql.Row
	if tx != nil {
		row = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM scenes WHERE id = ?`, sceneID)
	} else {
		row = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM scenes WHERE id = ?`, sceneID)
	}
	var n int
	if err := row.Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

func (s *Store) listSceneLightsTx(ctx context.Context, tx *sql.Tx, sceneID string) ([]SceneLightFlat, error) {
	ok, err := s.sceneExistsTx(ctx, tx, sceneID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrSceneNotFound
	}
	pls, err := s.loadScenePlacementsWithLights(ctx, tx, sceneID)
	if err != nil {
		return nil, err
	}
	out := make([]SceneLightFlat, 0)
	for _, p := range pls {
		for _, L := range p.lights {
			sx := sceneSpaceCoord(L.X, p.offsetX)
			sy := sceneSpaceCoord(L.Y, p.offsetY)
			sz := sceneSpaceCoord(L.Z, p.offsetZ)
			out = append(out, SceneLightFlat{
				SceneID:       sceneID,
				ModelID:       p.modelID,
				LightID:       L.ID,
				X:             L.X,
				Y:             L.Y,
				Z:             L.Z,
				Sx:            sx,
				Sy:            sy,
				Sz:            sz,
				On:            L.On,
				Color:         L.Color,
				BrightnessPct: L.BrightnessPct,
			})
		}
	}
	return out, nil
}

func (s *Store) listScenesReferencingModel(ctx context.Context, modelID string) ([]SceneRef, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.id, s.name FROM scenes s
		INNER JOIN scene_models sm ON sm.scene_id = s.id
		WHERE sm.model_id = ?
		ORDER BY s.name ASC
	`, modelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []SceneRef
	for rows.Next() {
		var r SceneRef
		if err := rows.Scan(&r.ID, &r.Name); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func sceneSpaceCoord(modelCoord float64, offset int) float64 {
	return modelCoord + float64(offset)
}

func validateLightsContainment(lights []LightDTO, ox, oy, oz int) error {
	for _, L := range lights {
		sx := sceneSpaceCoord(L.X, ox)
		sy := sceneSpaceCoord(L.Y, oy)
		sz := sceneSpaceCoord(L.Z, oz)
		if sx < -sceneCoordEps || sy < -sceneCoordEps || sz < -sceneCoordEps {
			return fmt.Errorf("light %d would lie outside the non-negative scene region (scene-space %.6f,%.6f,%.6f)", L.ID, sx, sy, sz)
		}
	}
	return nil
}

type scenePlacementLights struct {
	modelID     string
	offsetX     int
	offsetY     int
	offsetZ     int
	lights      []LightDTO
	skipModelID string // when validating hypothetical removal
}

func validateScenePlacements(pls []scenePlacementLights) error {
	for _, p := range pls {
		if p.skipModelID != "" && p.modelID == p.skipModelID {
			continue
		}
		if err := validateLightsContainment(p.lights, p.offsetX, p.offsetY, p.offsetZ); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) loadScenePlacementsWithLights(ctx context.Context, tx *sql.Tx, sceneID string) ([]scenePlacementLights, error) {
	var rows *sql.Rows
	var err error
	if tx != nil {
		rows, err = tx.QueryContext(ctx, `
			SELECT model_id, offset_x, offset_y, offset_z FROM scene_models WHERE scene_id = ? ORDER BY ordinal ASC, model_id ASC
		`, sceneID)
	} else {
		rows, err = s.db.QueryContext(ctx, `
			SELECT model_id, offset_x, offset_y, offset_z FROM scene_models WHERE scene_id = ? ORDER BY ordinal ASC, model_id ASC
		`, sceneID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type placementRow struct {
		modelID string
		offX    int
		offY    int
		offZ    int
	}
	var baseRows []placementRow
	for rows.Next() {
		var pr placementRow
		if err := rows.Scan(&pr.modelID, &pr.offX, &pr.offY, &pr.offZ); err != nil {
			return nil, err
		}
		baseRows = append(baseRows, pr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load model details after the placement cursor has been consumed to avoid
	// nested queries on the same connection while rows are still open.
	out := make([]scenePlacementLights, 0, len(baseRows))
	for _, pr := range baseRows {
		d, err := s.getDetailTx(ctx, tx, pr.modelID)
		if err != nil {
			return nil, err
		}
		out = append(out, scenePlacementLights{
			modelID: pr.modelID,
			offsetX: pr.offX,
			offsetY: pr.offY,
			offsetZ: pr.offZ,
			lights:  d.Lights,
		})
	}
	return out, nil
}

func (s *Store) getDetailTx(ctx context.Context, tx *sql.Tx, modelID string) (*Detail, error) {
	var queryRow func(context.Context, string, ...any) *sql.Row
	var query func(context.Context, string, ...any) (*sql.Rows, error)
	if tx != nil {
		queryRow = tx.QueryRowContext
		query = tx.QueryContext
	} else {
		queryRow = s.db.QueryRowContext
		query = s.db.QueryContext
	}
	row := queryRow(ctx, `SELECT id, name, created_at FROM models WHERE id = ?`, modelID)
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

	rows, err := query(ctx, `
		SELECT idx, x, y, z, "on", color, brightness_pct FROM lights WHERE model_id = ? ORDER BY idx ASC
	`, modelID)
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

func maxSceneXForPlacements(pls []scenePlacementLights) float64 {
	var mx float64
	for _, p := range pls {
		for _, L := range p.lights {
			sx := sceneSpaceCoord(L.X, p.offsetX)
			if sx > mx {
				mx = sx
			}
		}
	}
	return mx
}

func minFloat64(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	m := xs[0]
	for _, v := range xs[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

// OffsetsForFirstSceneModel returns minimal non-negative integer offsets so every light lies in the non-negative scene octant (architecture §3.12 create-time step).
func OffsetsForFirstSceneModel(lights []LightDTO) (ox, oy, oz int) {
	if len(lights) == 0 {
		return 0, 0, 0
	}
	xs := make([]float64, len(lights))
	ys := make([]float64, len(lights))
	zs := make([]float64, len(lights))
	for i, L := range lights {
		xs[i] = L.X
		ys[i] = L.Y
		zs[i] = L.Z
	}
	mMinX := minFloat64(xs)
	mMinY := minFloat64(ys)
	mMinZ := minFloat64(zs)
	if mMinX < 0 {
		ox = int(math.Ceil(-mMinX - sceneCoordEps))
	}
	if mMinY < 0 {
		oy = int(math.Ceil(-mMinY - sceneCoordEps))
	}
	if mMinZ < 0 {
		oz = int(math.Ceil(-mMinZ - sceneCoordEps))
	}
	return ox, oy, oz
}

// DefaultOffsetsForNewModel computes +X “to the right” placement (architecture §3.12).
func DefaultOffsetsForNewModel(existing []scenePlacementLights, newLights []LightDTO) (ox, oy, oz int) {
	Mx := maxSceneXForPlacements(existing)
	if len(newLights) == 0 {
		return int(math.Ceil(Mx - sceneCoordEps)), 0, 0
	}
	xs := make([]float64, len(newLights))
	ys := make([]float64, len(newLights))
	zs := make([]float64, len(newLights))
	for i, L := range newLights {
		xs[i] = L.X
		ys[i] = L.Y
		zs[i] = L.Z
	}
	mMinX := minFloat64(xs)
	mMinY := minFloat64(ys)
	mMinZ := minFloat64(zs)

	gap := 0.0
	ox = int(math.Ceil(Mx + gap - mMinX - sceneCoordEps))
	if ox < 0 {
		ox = 0
	}
	oy = 0
	if mMinY < 0 {
		oy = int(math.Ceil(-mMinY - sceneCoordEps))
	}
	oz = 0
	if mMinZ < 0 {
		oz = int(math.Ceil(-mMinZ - sceneCoordEps))
	}
	return ox, oy, oz
}

// ListScenes returns all scenes with model counts, newest first.
func (s *Store) ListScenes(ctx context.Context) ([]SceneSummary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.id, s.name, s.created_at, COUNT(sm.model_id) AS n
		FROM scenes s
		LEFT JOIN scene_models sm ON sm.scene_id = s.id
		GROUP BY s.id
		ORDER BY s.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []SceneSummary
	for rows.Next() {
		var sum SceneSummary
		var created string
		if err := rows.Scan(&sum.ID, &sum.Name, &created, &sum.ModelCount); err != nil {
			return nil, err
		}
		t, err := time.Parse(time.RFC3339Nano, created)
		if err != nil {
			t, err = time.Parse(time.RFC3339, created)
			if err != nil {
				return nil, fmt.Errorf("parse scene created_at: %w", err)
			}
		}
		sum.CreatedAt = t.UTC()
		out = append(out, sum)
	}
	return out, rows.Err()
}

// CreateScene inserts a scene with ≥1 models in request order. Integer offsets are computed server-side (architecture §3.12).
func (s *Store) CreateScene(ctx context.Context, name string, modelIDs []string) (*SceneSummary, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if len(modelIDs) == 0 {
		return nil, fmt.Errorf("at least one model is required")
	}

	var pls []scenePlacementLights
	seen := make(map[string]struct{})
	for _, raw := range modelIDs {
		mid := strings.TrimSpace(raw)
		if mid == "" {
			return nil, fmt.Errorf("model_id is required")
		}
		if _, dup := seen[mid]; dup {
			return nil, fmt.Errorf("duplicate model_id in request")
		}
		seen[mid] = struct{}{}

		var n int
		if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM models WHERE id = ?`, mid).Scan(&n); err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, ErrNotFound
		}

		d, err := s.getDetailTx(ctx, nil, mid)
		if err != nil {
			return nil, err
		}

		var ox, oy, oz int
		if len(pls) == 0 {
			ox, oy, oz = OffsetsForFirstSceneModel(d.Lights)
		} else {
			ox, oy, oz = DefaultOffsetsForNewModel(pls, d.Lights)
		}
		pls = append(pls, scenePlacementLights{
			modelID: mid, offsetX: ox, offsetY: oy, offsetZ: oz, lights: d.Lights,
		})
		if err := validateScenePlacements(pls); err != nil {
			return nil, err
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	id := uuid.NewString()
	created := time.Now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO scenes (id, name, created_at, boundary_margin_m) VALUES (?, ?, ?, ?)
	`, id, name, created, DefaultSceneBoundaryMarginM); err != nil {
		if isUniqueConstraint(err) {
			return nil, ErrDuplicateSceneName
		}
		return nil, err
	}

	for ord, p := range pls {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO scene_models (scene_id, model_id, offset_x, offset_y, offset_z, ordinal) VALUES (?, ?, ?, ?, ?, ?)
		`, id, p.modelID, p.offsetX, p.offsetY, p.offsetZ, ord); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	t, _ := time.Parse(time.RFC3339Nano, created)
	return &SceneSummary{
		ID:         id,
		Name:       name,
		CreatedAt:  t,
		ModelCount: len(pls),
	}, nil
}

// SceneExists reports whether a scene id exists (lightweight lookup for SSE and similar).
func (s *Store) SceneExists(ctx context.Context, id string) (bool, error) {
	var one int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM scenes WHERE id = ? LIMIT 1`, id).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// ListSceneIDsForModel returns every scene that references the model (REQ-029 revision fan-out).
func (s *Store) ListSceneIDsForModel(ctx context.Context, modelID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT scene_id FROM scene_models WHERE model_id = ? ORDER BY scene_id`, modelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var sid string
		if err := rows.Scan(&sid); err != nil {
			return nil, err
		}
		out = append(out, sid)
	}
	return out, rows.Err()
}

// ListModelIDsInScene returns model ids assigned to the scene in placement order.
func (s *Store) ListModelIDsInScene(ctx context.Context, sceneID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT model_id FROM scene_models WHERE scene_id = ? ORDER BY ordinal ASC, model_id ASC`, sceneID)
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

func scanSceneBoundaryMarginM(ns sql.NullFloat64) float64 {
	if !ns.Valid || math.IsNaN(ns.Float64) || math.IsInf(ns.Float64, 0) || ns.Float64 < 0 {
		return DefaultSceneBoundaryMarginM
	}
	return ns.Float64
}

func (s *Store) getSceneBoundaryMarginM(ctx context.Context, sceneID string) (float64, error) {
	row := s.db.QueryRowContext(ctx, `SELECT boundary_margin_m FROM scenes WHERE id = ?`, sceneID)
	var ns sql.NullFloat64
	if err := row.Scan(&ns); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrSceneNotFound
		}
		return 0, err
	}
	return scanSceneBoundaryMarginM(ns), nil
}

// PatchSceneBoundaryMarginM updates REQ-034 scene padding (SI meters).
func (s *Store) PatchSceneBoundaryMarginM(ctx context.Context, sceneID string, marginM float64) error {
	if math.IsNaN(marginM) || math.IsInf(marginM, 0) {
		return fmt.Errorf("boundary_margin_m must be a finite number")
	}
	if marginM < 0 {
		return fmt.Errorf("boundary_margin_m must be >= 0")
	}
	if marginM > MaxSceneBoundaryMarginM {
		return fmt.Errorf("boundary_margin_m must be <= %g", MaxSceneBoundaryMarginM)
	}
	res, err := s.db.ExecContext(ctx, `UPDATE scenes SET boundary_margin_m = ? WHERE id = ?`, marginM, sceneID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrSceneNotFound
	}
	return nil
}

// GetScene returns scene metadata and all models with scene-space lights.
func (s *Store) GetScene(ctx context.Context, sceneID string) (*SceneDetail, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, name, created_at, boundary_margin_m FROM scenes WHERE id = ?`, sceneID)
	var d SceneDetail
	var created string
	var bm sql.NullFloat64
	if err := row.Scan(&d.ID, &d.Name, &created, &bm); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSceneNotFound
		}
		return nil, err
	}
	t, err := time.Parse(time.RFC3339Nano, created)
	if err != nil {
		t, err = time.Parse(time.RFC3339, created)
		if err != nil {
			return nil, fmt.Errorf("parse scene created_at: %w", err)
		}
	}
	d.CreatedAt = t.UTC()
	d.BoundaryMarginM = scanSceneBoundaryMarginM(bm)

	rows, err := s.db.QueryContext(ctx, `
		SELECT sm.model_id, sm.offset_x, sm.offset_y, sm.offset_z, m.name
		FROM scene_models sm
		INNER JOIN models m ON m.id = sm.model_id
		WHERE sm.scene_id = ?
		ORDER BY sm.ordinal ASC, sm.model_id ASC
	`, sceneID)
	if err != nil {
		return nil, err
	}
	type rowItem struct {
		it SceneItemDetail
	}
	var pending []rowItem
	for rows.Next() {
		var r rowItem
		if err := rows.Scan(&r.it.ModelID, &r.it.OffsetX, &r.it.OffsetY, &r.it.OffsetZ, &r.it.Name); err != nil {
			_ = rows.Close()
			return nil, err
		}
		pending = append(pending, r)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	for _, r := range pending {
		it := r.it
		md, err := s.Get(ctx, it.ModelID)
		if err != nil {
			return nil, err
		}
		for _, L := range md.Lights {
			sx := sceneSpaceCoord(L.X, it.OffsetX)
			sy := sceneSpaceCoord(L.Y, it.OffsetY)
			sz := sceneSpaceCoord(L.Z, it.OffsetZ)
			it.Lights = append(it.Lights, SceneLightDTO{
				ID: L.ID, X: L.X, Y: L.Y, Z: L.Z,
				Sx: sx, Sy: sy, Sz: sz,
				On: L.On, Color: L.Color, BrightnessPct: L.BrightnessPct,
			})
		}
		d.Items = append(d.Items, it)
	}
	return &d, nil
}

// DeleteScene removes a scene and its scene_models rows.
func (s *Store) DeleteScene(ctx context.Context, sceneID string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM scenes WHERE id = ?`, sceneID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrSceneNotFound
	}
	return nil
}

// AddSceneModel adds a model to a scene; if ox, oy, oz are nil, default +X placement is used.
func (s *Store) AddSceneModel(ctx context.Context, sceneID, modelID string, ox, oy, oz *int) (*SceneModelPlacement, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var n int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM scenes WHERE id = ?`, sceneID).Scan(&n); err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, ErrSceneNotFound
	}
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM models WHERE id = ?`, modelID).Scan(&n); err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, ErrNotFound
	}
	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM scene_models WHERE scene_id = ? AND model_id = ?`, sceneID, modelID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, ErrModelAlreadyInScene
	}

	existing, err := s.loadScenePlacementsWithLights(ctx, tx, sceneID)
	if err != nil {
		return nil, err
	}
	newDetail, err := s.getDetailTx(ctx, tx, modelID)
	if err != nil {
		return nil, err
	}

	var offX, offY, offZ int
	if ox != nil && oy != nil && oz != nil {
		offX, offY, offZ = *ox, *oy, *oz
		if offX < 0 || offY < 0 || offZ < 0 {
			return nil, fmt.Errorf("offsets must be non-negative")
		}
	} else if ox != nil || oy != nil || oz != nil {
		return nil, fmt.Errorf("either provide all three offsets or omit all for default placement")
	} else {
		offX, offY, offZ = DefaultOffsetsForNewModel(existing, newDetail.Lights)
	}

	if err := validateLightsContainment(newDetail.Lights, offX, offY, offZ); err != nil {
		return nil, err
	}

	merged := append(slicesClonePlacements(existing), scenePlacementLights{
		modelID: modelID, offsetX: offX, offsetY: offY, offsetZ: offZ, lights: newDetail.Lights,
	})
	if err := validateScenePlacements(merged); err != nil {
		return nil, err
	}

	var nextOrd int
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(ordinal), -1) + 1 FROM scene_models WHERE scene_id = ?
	`, sceneID).Scan(&nextOrd); err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO scene_models (scene_id, model_id, offset_x, offset_y, offset_z, ordinal) VALUES (?, ?, ?, ?, ?, ?)
	`, sceneID, modelID, offX, offY, offZ, nextOrd); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &SceneModelPlacement{ModelID: modelID, OffsetX: offX, OffsetY: offY, OffsetZ: offZ}, nil
}

func slicesClonePlacements(p []scenePlacementLights) []scenePlacementLights {
	out := make([]scenePlacementLights, len(p))
	copy(out, p)
	return out
}

// PatchSceneModelOffsets updates offsets for one model in a scene.
func (s *Store) PatchSceneModelOffsets(ctx context.Context, sceneID, modelID string, offX, offY, offZ int) (*SceneModelPlacement, error) {
	if offX < 0 || offY < 0 || offZ < 0 {
		return nil, fmt.Errorf("offsets must be non-negative")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var n int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM scene_models WHERE scene_id = ? AND model_id = ?`, sceneID, modelID).Scan(&n); err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, ErrNotFound
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE scene_models SET offset_x = ?, offset_y = ?, offset_z = ? WHERE scene_id = ? AND model_id = ?
	`, offX, offY, offZ, sceneID, modelID); err != nil {
		return nil, err
	}

	pls, err := s.loadScenePlacementsWithLights(ctx, tx, sceneID)
	if err != nil {
		return nil, err
	}
	if err := validateScenePlacements(pls); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &SceneModelPlacement{ModelID: modelID, OffsetX: offX, OffsetY: offY, OffsetZ: offZ}, nil
}

// RemoveSceneModel removes a model from a scene. ErrSceneLastModel if it is the only model (no mutation).
func (s *Store) RemoveSceneModel(ctx context.Context, sceneID, modelID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM scene_models WHERE scene_id = ?`, sceneID).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		return ErrSceneNotFound
	}

	var member int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM scene_models WHERE scene_id = ? AND model_id = ?`, sceneID, modelID).Scan(&member); err != nil {
		return err
	}
	if member == 0 {
		return ErrNotFound
	}

	if count == 1 {
		return ErrSceneLastModel
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM scene_models WHERE scene_id = ? AND model_id = ?`, sceneID, modelID); err != nil {
		return err
	}
	return tx.Commit()
}

// GetSceneDimensions returns scene-space dimensions used for region queries.
// Extents are the axis-aligned bounding box of all lights in scene space (sx,sy,sz),
// expanded by the scene's boundary_margin_m on each axis (REQ-034, REQ-033 motion).
func (s *Store) GetSceneDimensions(ctx context.Context, sceneID string) (*SceneDimensions, error) {
	marginM, err := s.getSceneBoundaryMarginM(ctx, sceneID)
	if err != nil {
		return nil, err
	}
	lights, err := s.listSceneLightsTx(ctx, nil, sceneID)
	if err != nil {
		return nil, err
	}
	if len(lights) == 0 {
		m := marginM
		return &SceneDimensions{
			Origin:  ScenePoint{X: 0, Y: 0, Z: 0},
			Size:    SceneDimensionsSize{Width: m, Height: m, Depth: m},
			Max:     ScenePoint{X: m, Y: m, Z: m},
			MarginM: marginM,
		}, nil
	}
	mnX, mnY, mnZ := lights[0].Sx, lights[0].Sy, lights[0].Sz
	mx, my, mz := mnX, mnY, mnZ
	for _, L := range lights[1:] {
		if L.Sx < mnX {
			mnX = L.Sx
		}
		if L.Sy < mnY {
			mnY = L.Sy
		}
		if L.Sz < mnZ {
			mnZ = L.Sz
		}
		if L.Sx > mx {
			mx = L.Sx
		}
		if L.Sy > my {
			my = L.Sy
		}
		if L.Sz > mz {
			mz = L.Sz
		}
	}
	// Unclamped padded AABB: must match REQ-034 faint boundary cuboid and shape-animation
	// physics (SceneLightsCanvas uses min(sx,sy,sz)−m with no floor at 0).
	minX := mnX - marginM
	minY := mnY - marginM
	minZ := mnZ - marginM
	maxX := mx + marginM
	maxY := my + marginM
	maxZ := mz + marginM
	return &SceneDimensions{
		Origin: ScenePoint{X: minX, Y: minY, Z: minZ},
		Size: SceneDimensionsSize{
			Width:  maxX - minX,
			Height: maxY - minY,
			Depth:  maxZ - minZ,
		},
		Max:     ScenePoint{X: maxX, Y: maxY, Z: maxZ},
		MarginM: marginM,
	}, nil
}

// ListSceneLights returns all scene lights in a flattened scene-space form.
func (s *Store) ListSceneLights(ctx context.Context, sceneID string) ([]SceneLightFlat, error) {
	return s.listSceneLightsTx(ctx, nil, sceneID)
}

// QuerySceneLightsCuboid returns scene lights contained in the supplied cuboid (inclusive boundaries).
func (s *Store) QuerySceneLightsCuboid(ctx context.Context, sceneID string, cuboid SceneCuboid) ([]SceneLightFlat, error) {
	if err := validateSceneCuboid(cuboid); err != nil {
		return nil, err
	}
	all, err := s.listSceneLightsTx(ctx, nil, sceneID)
	if err != nil {
		return nil, err
	}
	out := make([]SceneLightFlat, 0)
	for _, L := range all {
		if cuboidContains(cuboid, L) {
			out = append(out, L)
		}
	}
	return out, nil
}

// QuerySceneLightsSphere returns scene lights contained in the supplied sphere (inclusive boundary).
func (s *Store) QuerySceneLightsSphere(ctx context.Context, sceneID string, sph SceneSphere) ([]SceneLightFlat, error) {
	if err := validateSceneSphere(sph); err != nil {
		return nil, err
	}
	all, err := s.listSceneLightsTx(ctx, nil, sceneID)
	if err != nil {
		return nil, err
	}
	out := make([]SceneLightFlat, 0)
	for _, L := range all {
		if sphereContains(sph, L) {
			out = append(out, L)
		}
	}
	return out, nil
}

func (s *Store) patchSceneLightsByRegion(ctx context.Context, sceneID string, patch LightStatePatch, include func(SceneLightFlat) bool) (*SceneBulkPatchResult, error) {
	if err := validateScenePatch(patch); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	out, err := s.patchSceneLightsByRegionTx(ctx, tx, sceneID, patch, include)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Store) patchSceneLightsByRegionTx(ctx context.Context, tx *sql.Tx, sceneID string, patch LightStatePatch, include func(SceneLightFlat) bool) (*SceneBulkPatchResult, error) {
	if err := validateScenePatch(patch); err != nil {
		return nil, err
	}
	all, err := s.listSceneLightsTx(ctx, tx, sceneID)
	if err != nil {
		return nil, err
	}

	var colorNorm *string
	if patch.Color != nil {
		c, err := ValidateColor(*patch.Color)
		if err != nil {
			return nil, err
		}
		colorNorm = &c
	}
	var brightnessNorm *float64
	if patch.BrightnessPct != nil {
		if err := ValidateBrightnessPct(*patch.BrightnessPct); err != nil {
			return nil, err
		}
		b := *patch.BrightnessPct
		brightnessNorm = &b
	}

	updated := make([]ScenePatchedState, 0)
	writeCount := 0
	allEquiv := true
	matchedAny := false
	for _, L := range all {
		if !include(L) {
			continue
		}
		matchedAny = true
		prevOn := L.On
		prevColor := strings.ToLower(strings.TrimSpace(L.Color))
		prevBr := L.BrightnessPct
		on := L.On
		color := L.Color
		brightness := L.BrightnessPct
		if patch.On != nil {
			on = *patch.On
		}
		if colorNorm != nil {
			color = *colorNorm
		}
		if brightnessNorm != nil {
			brightness = *brightnessNorm
		}
		rowUnchanged := EquivLightStateTriple(prevOn, prevColor, prevBr, on, color, brightness)
		if !rowUnchanged {
			allEquiv = false
			if _, err := tx.ExecContext(ctx, `
				UPDATE lights SET "on" = ?, color = ?, brightness_pct = ? WHERE model_id = ? AND idx = ?
			`, boolToInt(on), color, brightness, L.ModelID, L.LightID); err != nil {
				return nil, err
			}
			writeCount++
		}
		updated = append(updated, ScenePatchedState{
			ModelID:       L.ModelID,
			ID:            L.LightID,
			On:            on,
			Color:         color,
			BrightnessPct: brightness,
			Sx:            L.Sx,
			Sy:            L.Sy,
			Sz:            L.Sz,
		})
	}

	res := &SceneBulkPatchResult{
		UpdatedCount: writeCount,
		States:       updated,
	}
	if matchedAny && allEquiv {
		res.UnchangedAll = true
	}
	return res, nil
}

// PatchSceneLightsCuboid bulk-updates lights inside a cuboid (inclusive boundaries).
func (s *Store) PatchSceneLightsCuboid(ctx context.Context, sceneID string, cuboid SceneCuboid, patch LightStatePatch) (*SceneBulkPatchResult, error) {
	if err := validateSceneCuboid(cuboid); err != nil {
		return nil, err
	}
	return s.patchSceneLightsByRegion(ctx, sceneID, patch, func(L SceneLightFlat) bool {
		return cuboidContains(cuboid, L)
	})
}

// PatchSceneLightsSphere bulk-updates lights inside a sphere (inclusive boundary).
func (s *Store) PatchSceneLightsSphere(ctx context.Context, sceneID string, sph SceneSphere, patch LightStatePatch) (*SceneBulkPatchResult, error) {
	if err := validateSceneSphere(sph); err != nil {
		return nil, err
	}
	return s.patchSceneLightsByRegion(ctx, sceneID, patch, func(L SceneLightFlat) bool {
		return sphereContains(sph, L)
	})
}

// PatchSceneLightsScene applies the same state patch to every light in the scene (REQ-020 / architecture §3.15).
func (s *Store) PatchSceneLightsScene(ctx context.Context, sceneID string, patch LightStatePatch) (*SceneBulkPatchResult, error) {
	return s.patchSceneLightsByRegion(ctx, sceneID, patch, func(SceneLightFlat) bool { return true })
}

// ErrSceneLightNotInScene is returned when a batch update references a light not in the scene.
var ErrSceneLightNotInScene = errors.New("light is not part of this scene")

func validateSceneBatchUpdates(updates []SceneBatchLightUpdate) error {
	if len(updates) == 0 {
		return fmt.Errorf("updates must be non-empty")
	}
	seen := make(map[string]struct{}, len(updates))
	for _, u := range updates {
		if strings.TrimSpace(u.ModelID) == "" {
			return fmt.Errorf("model_id is required for each update")
		}
		if u.Patch.On == nil && u.Patch.Color == nil && u.Patch.BrightnessPct == nil {
			return fmt.Errorf("each update requires at least one of on, color, brightness_pct")
		}
		if err := validateScenePatch(u.Patch); err != nil {
			return err
		}
		key := u.ModelID + "\x00" + strconv.Itoa(u.LightID)
		if _, dup := seen[key]; dup {
			return fmt.Errorf("duplicate model_id and light_id in updates")
		}
		seen[key] = struct{}{}
	}
	return nil
}

// PatchSceneLightsBatch applies per-light patches in one transaction (REQ-020 / architecture §3.15).
func (s *Store) PatchSceneLightsBatch(ctx context.Context, sceneID string, updates []SceneBatchLightUpdate) (*SceneBulkPatchResult, error) {
	if err := validateSceneBatchUpdates(updates); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	out, err := s.patchSceneLightsBatchTx(ctx, tx, sceneID, updates)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Store) patchSceneLightsBatchTx(ctx context.Context, tx *sql.Tx, sceneID string, updates []SceneBatchLightUpdate) (*SceneBulkPatchResult, error) {
	all, err := s.listSceneLightsTx(ctx, tx, sceneID)
	if err != nil {
		return nil, err
	}
	member := make(map[string]SceneLightFlat, len(all))
	for _, L := range all {
		k := L.ModelID + "\x00" + strconv.Itoa(L.LightID)
		member[k] = L
	}

	updated := make([]ScenePatchedState, 0, len(updates))
	writeCount := 0
	allUnchanged := true
	for _, u := range updates {
		k := u.ModelID + "\x00" + strconv.Itoa(u.LightID)
		flat, ok := member[k]
		if !ok {
			return nil, ErrSceneLightNotInScene
		}

		row := tx.QueryRowContext(ctx, `
			SELECT "on", color, brightness_pct FROM lights WHERE model_id = ? AND idx = ?
		`, u.ModelID, u.LightID)
		var onInt int
		var color string
		var brightness float64
		if err := row.Scan(&onInt, &color, &brightness); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrSceneLightNotInScene
			}
			return nil, err
		}
		prevOn := onInt != 0
		on := prevOn
		prevColor := strings.ToLower(strings.TrimSpace(color))
		prevBr := brightness
		if u.Patch.On != nil {
			on = *u.Patch.On
		}
		if u.Patch.Color != nil {
			c, err := ValidateColor(*u.Patch.Color)
			if err != nil {
				return nil, err
			}
			color = c
		}
		if u.Patch.BrightnessPct != nil {
			if err := ValidateBrightnessPct(*u.Patch.BrightnessPct); err != nil {
				return nil, err
			}
			brightness = *u.Patch.BrightnessPct
		}

		rowUnchanged := EquivLightStateTriple(prevOn, prevColor, prevBr, on, color, brightness)
		if rowUnchanged {
			// no UPDATE
		} else {
			allUnchanged = false
			if _, err := tx.ExecContext(ctx, `
				UPDATE lights SET "on" = ?, color = ?, brightness_pct = ? WHERE model_id = ? AND idx = ?
			`, boolToInt(on), color, brightness, u.ModelID, u.LightID); err != nil {
				return nil, err
			}
			writeCount++
		}
		updated = append(updated, ScenePatchedState{
			ModelID:       u.ModelID,
			ID:            u.LightID,
			On:            on,
			Color:         color,
			BrightnessPct: brightness,
			Sx:            flat.Sx,
			Sy:            flat.Sy,
			Sz:            flat.Sz,
		})
	}

	return &SceneBulkPatchResult{
		UpdatedCount: writeCount,
		States:       updated,
		UnchangedAll: allUnchanged,
	}, nil
}
