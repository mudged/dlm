package httpapi

import (
	"context"
	"log/slog"
	"time"

	"example.com/dlm/backend/internal/store"
)

// StartRoutineScheduler runs a 1s ticker that advances active routine runs until ctx is cancelled.
func StartRoutineScheduler(ctx context.Context, log *slog.Logger, st *store.Store) {
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
			if err := st.TickRoutineRuns(context.Background()); err != nil && log != nil {
				log.Error("routine tick", "err", err)
			}
		}
	}
}
