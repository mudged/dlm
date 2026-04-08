package lightstate

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"
	"sync"
)

// Default triple for new lights (REQ-014 / architecture §3.9).
const (
	DefaultOn            = false
	DefaultColor         = "#ffffff"
	DefaultBrightnessPct = 100.0
)

var hexColorRe = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// ErrNotFound is returned when the model has no state slot (unknown model to this store).
var ErrNotFound = errors.New("model not found in light state store")

// ErrInvalidLightIndex is returned when light id is out of range for the model.
var ErrInvalidLightIndex = errors.New("light index out of range")

// DTO matches API JSON for one light's operational state.
type DTO struct {
	ID            int     `json:"id"`
	On            bool    `json:"on"`
	Color         string  `json:"color"`
	BrightnessPct float64 `json:"brightness_pct"`
}

// Patch is a partial update; nil fields are unchanged.
type Patch struct {
	On            *bool
	Color         *string
	BrightnessPct *float64
}

// Store holds authoritative in-memory per-light triples (REQ-039).
type Store struct {
	mu      sync.RWMutex
	byModel map[string][]lightTriple // index = light id
}

type lightTriple struct {
	on            bool
	color         string
	brightnessPct float64
}

// New returns an empty store.
func New() *Store {
	return &Store{byModel: make(map[string][]lightTriple)}
}

// ValidateColor normalizes hex color.
func ValidateColor(s string) (string, error) {
	s = strings.TrimSpace(s)
	if !hexColorRe.MatchString(s) {
		return "", fmt.Errorf("color must be #RRGGBB with six hex digits")
	}
	return strings.ToLower(s), nil
}

// ValidateBrightnessPct checks range [0, 100].
func ValidateBrightnessPct(v float64) error {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return fmt.Errorf("brightness_pct must be a finite number")
	}
	if v < 0 || v > 100 {
		return fmt.Errorf("brightness_pct must be between 0 and 100")
	}
	return nil
}

// EnsureModel allocates or resizes state for n lights (0..n-1) with defaults. If n changes, resets.
func (s *Store) EnsureModel(modelID string, n int) {
	if n < 0 {
		n = 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	sl := make([]lightTriple, n)
	for i := range sl {
		sl[i] = lightTriple{on: DefaultOn, color: DefaultColor, brightnessPct: DefaultBrightnessPct}
	}
	s.byModel[modelID] = sl
}

// RemoveModel drops all state for a model.
func (s *Store) RemoveModel(modelID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byModel, modelID)
}

// Clear removes all models (e.g. factory reset before re-seed).
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byModel = make(map[string][]lightTriple)
}

func (s *Store) getSlice(modelID string) ([]lightTriple, bool) {
	sl, ok := s.byModel[modelID]
	return sl, ok
}

// List returns state for all lights in order of id.
func (s *Store) List(modelID string) ([]DTO, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sl, ok := s.getSlice(modelID)
	if !ok {
		return nil, ErrNotFound
	}
	out := make([]DTO, len(sl))
	for i := range sl {
		t := sl[i]
		out[i] = DTO{ID: i, On: t.on, Color: t.color, BrightnessPct: t.brightnessPct}
	}
	return out, nil
}

// Get returns one light's state.
func (s *Store) Get(modelID string, lightID int) (*DTO, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sl, ok := s.getSlice(modelID)
	if !ok {
		return nil, ErrNotFound
	}
	if lightID < 0 || lightID >= len(sl) {
		return nil, ErrInvalidLightIndex
	}
	t := sl[lightID]
	return &DTO{ID: lightID, On: t.on, Color: t.color, BrightnessPct: t.brightnessPct}, nil
}

// Patch merges patch into one light; returns merged DTO and whether effective triple unchanged (REQ-031).
func (s *Store) Patch(modelID string, lightID int, patch Patch) (*DTO, bool, error) {
	if patch.On == nil && patch.Color == nil && patch.BrightnessPct == nil {
		return nil, false, fmt.Errorf("at least one of on, color, brightness_pct is required")
	}
	if patch.Color != nil {
		if _, err := ValidateColor(*patch.Color); err != nil {
			return nil, false, err
		}
	}
	if patch.BrightnessPct != nil {
		if err := ValidateBrightnessPct(*patch.BrightnessPct); err != nil {
			return nil, false, err
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	sl, ok := s.getSlice(modelID)
	if !ok {
		return nil, false, ErrNotFound
	}
	if lightID < 0 || lightID >= len(sl) {
		return nil, false, ErrInvalidLightIndex
	}
	t := sl[lightID]
	prevOn, prevColor, prevBr := t.on, t.color, t.brightnessPct
	on, color, br := prevOn, prevColor, prevBr
	if patch.On != nil {
		on = *patch.On
	}
	if patch.Color != nil {
		c, _ := ValidateColor(*patch.Color)
		color = c
	}
	if patch.BrightnessPct != nil {
		br = *patch.BrightnessPct
	}
	unchanged := EquivLightStateTriple(prevOn, prevColor, prevBr, on, color, br)
	if !unchanged {
		sl[lightID] = lightTriple{on: on, color: color, brightnessPct: br}
	}
	return &DTO{ID: lightID, On: on, Color: color, BrightnessPct: br}, unchanged, nil
}

// BatchPatch applies the same patch to many ids (sorted ascending in result). ids must be unique.
func (s *Store) BatchPatch(modelID string, ids []int, patch Patch) ([]DTO, bool, error) {
	if len(ids) == 0 {
		return nil, false, fmt.Errorf("batch ids must be non-empty")
	}
	if patch.On == nil && patch.Color == nil && patch.BrightnessPct == nil {
		return nil, false, fmt.Errorf("batch patch requires at least one of on, color, brightness_pct")
	}
	if patch.Color != nil {
		if _, err := ValidateColor(*patch.Color); err != nil {
			return nil, false, err
		}
	}
	if patch.BrightnessPct != nil {
		if err := ValidateBrightnessPct(*patch.BrightnessPct); err != nil {
			return nil, false, err
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	sl, ok := s.getSlice(modelID)
	if !ok {
		return nil, false, ErrNotFound
	}
	seen := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		if _, dup := seen[id]; dup {
			return nil, false, fmt.Errorf("duplicate light ids in batch")
		}
		seen[id] = struct{}{}
		if id < 0 || id >= len(sl) {
			return nil, false, ErrInvalidLightIndex
		}
	}

	allUnchanged := true
	out := make([]DTO, 0, len(ids))
	sorted := slices.Clone(ids)
	slices.Sort(sorted)
	for _, lightID := range sorted {
		t := sl[lightID]
		prevOn, prevColor, prevBr := t.on, t.color, t.brightnessPct
		on, color, br := prevOn, prevColor, prevBr
		if patch.On != nil {
			on = *patch.On
		}
		if patch.Color != nil {
			c, _ := ValidateColor(*patch.Color)
			color = c
		}
		if patch.BrightnessPct != nil {
			br = *patch.BrightnessPct
		}
		rowUnchanged := EquivLightStateTriple(prevOn, prevColor, prevBr, on, color, br)
		if !rowUnchanged {
			allUnchanged = false
			sl[lightID] = lightTriple{on: on, color: color, brightnessPct: br}
		}
		out = append(out, DTO{ID: lightID, On: on, Color: color, BrightnessPct: br})
	}
	return out, allUnchanged, nil
}

// ResetAll sets every light to defaults; returns all states and whether nothing changed.
func (s *Store) ResetAll(modelID string) ([]DTO, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sl, ok := s.getSlice(modelID)
	if !ok {
		return nil, false, ErrNotFound
	}
	allDefault := true
	for i := range sl {
		t := sl[i]
		if !EquivLightStateTriple(t.on, t.color, t.brightnessPct, DefaultOn, DefaultColor, DefaultBrightnessPct) {
			allDefault = false
			break
		}
	}
	if !allDefault {
		for i := range sl {
			sl[i] = lightTriple{on: DefaultOn, color: DefaultColor, brightnessPct: DefaultBrightnessPct}
		}
	}
	out := make([]DTO, len(sl))
	for i := range sl {
		t := sl[i]
		out[i] = DTO{ID: i, On: t.on, Color: t.color, BrightnessPct: t.brightnessPct}
	}
	return out, allDefault, nil
}

// SetTriple sets absolute state if different from current; returns whether unchanged (equiv).
func (s *Store) SetTriple(modelID string, lightID int, on bool, color string, brightness float64) (unchanged bool, err error) {
	c, err := ValidateColor(color)
	if err != nil {
		return false, err
	}
	if err := ValidateBrightnessPct(brightness); err != nil {
		return false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	sl, ok := s.getSlice(modelID)
	if !ok {
		return false, ErrNotFound
	}
	if lightID < 0 || lightID >= len(sl) {
		return false, ErrInvalidLightIndex
	}
	t := sl[lightID]
	if EquivLightStateTriple(t.on, t.color, t.brightnessPct, on, c, brightness) {
		return true, nil
	}
	sl[lightID] = lightTriple{on: on, color: c, brightnessPct: brightness}
	return false, nil
}
