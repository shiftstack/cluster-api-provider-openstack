package metrics

import (
	"context"
	"sync"
	"time"
)

type refresher interface {
	Refresh()
}

func refreshAll(targets []refresher) {
	var wg sync.WaitGroup
	for _, t := range targets {
		wg.Add(1)
		go func(t refresher) {
			t.Refresh()
			wg.Done()
		}(t)
	}
	wg.Wait()
}

// keepFresh calls the Refresh method of the targets every ttl, until ctx is
// cancelled.
func keepFresh(ctx context.Context, ttl time.Duration, targets ...refresher) {
	refreshAll(targets)

	ticker := time.NewTicker(ttl)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			refreshAll(targets)
		}
	}
}
