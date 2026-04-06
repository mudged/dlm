package httpapi

import (
	"context"
	"log/slog"
	"time"

	"example.com/dlm/backend/internal/store"
)

// StartRoutineScheduler runs a 1s ticker that advances active routine runs until ctx is cancelled.
// rev receives revision bumps after each server-side routine tick that mutates lights (REQ-029).
func StartRoutineScheduler(ctx context.Context, log *slog.Logger, st *store.Store, rev *RevisionHub) {
	if st == nil {
		return
	}
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			touched, err := st.TickRoutineRuns(context.Background())
			if err != nil && log != nil {
				log.Error("routine tick", "err", err)
				continue
			}
			if rev == nil || len(touched) == 0 {
				continue
			}
			bg := context.Background()
			for _, sceneID := range touched {
				mids, lerr := st.ListModelIDsInScene(bg, sceneID)
				if lerr != nil {
					continue
				}
				for _, mid := range mids {
					rev.NotifyModelLightsChanged(bg, st, mid)
				}
			}
		}
	}
}
