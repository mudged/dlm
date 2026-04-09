package httpapi

import (
	"context"

	"example.com/dlm/backend/internal/store"
)

func (a *apiDeps) notifyModelLightsChanged(ctx context.Context, modelID string, deltas []LightsSSEDelta) {
	if a == nil || modelID == "" || len(deltas) == 0 {
		return
	}
	a.rev.NotifyModelLightsChanged(ctx, a.store, modelID, deltas)
	if a.pusher != nil {
		_ = a.pusher.PushModel(ctx, modelID)
	}
}

func (a *apiDeps) notifyAfterSceneLightPatch(ctx context.Context, sceneID string, states []store.ScenePatchedState) {
	if a == nil {
		return
	}
	a.rev.NotifyAfterSceneLightPatch(ctx, a.store, sceneID, states)
	if a.pusher == nil || len(states) == 0 {
		return
	}
	seen := make(map[string]struct{})
	for _, ps := range states {
		mid := ps.ModelID
		if mid == "" {
			continue
		}
		if _, ok := seen[mid]; ok {
			continue
		}
		seen[mid] = struct{}{}
		_ = a.pusher.PushModel(ctx, mid)
	}
}

// devicePusher is satisfied by *devices.Pusher.
type devicePusher interface {
	PushModel(ctx context.Context, modelID string) error
}
