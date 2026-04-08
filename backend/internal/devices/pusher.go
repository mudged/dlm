package devices

import (
	"context"
	"errors"
	"net/http"
	"time"

	"example.com/dlm/backend/internal/store"
)

// Pusher applies logical model light state to assigned WLED hardware (REQ-035, REQ-038, REQ-039).
// Push errors are non-fatal for API handlers — connectivity is best-effort.
type Pusher struct {
	Store  *store.Store
	Client *http.Client
}

// NewPusher returns a pusher with a default HTTP client when hc is nil.
func NewPusher(st *store.Store, hc *http.Client) *Pusher {
	if hc == nil {
		hc = &http.Client{Timeout: 5 * time.Second}
	}
	return &Pusher{Store: st, Client: hc}
}

// PushModel sends current logical state for modelID to its assigned WLED device, if any.
func (p *Pusher) PushModel(ctx context.Context, modelID string) error {
	if p == nil || p.Store == nil || modelID == "" {
		return nil
	}
	d, err := p.Store.GetDeviceForModel(ctx, modelID)
	if err != nil {
		if errors.Is(err, store.ErrDeviceNotFound) {
			return nil
		}
		return err
	}
	if d.Type != store.DeviceTypeWLED {
		return nil
	}
	states, err := p.Store.ListLightStates(ctx, modelID)
	if err != nil {
		return err
	}
	payload := buildWLEDState(states)
	return postJSONState(ctx, p.Client, d.BaseURL, d.WLEDPassword, payload)
}

// SyncAllAssignedModels pushes every model that has a device (process startup / §3.21).
func (p *Pusher) SyncAllAssignedModels(ctx context.Context) error {
	if p == nil || p.Store == nil {
		return nil
	}
	ids, err := p.Store.ListModelIDsWithDevices(ctx)
	if err != nil {
		return err
	}
	for _, mid := range ids {
		_ = p.PushModel(ctx, mid)
	}
	return nil
}
