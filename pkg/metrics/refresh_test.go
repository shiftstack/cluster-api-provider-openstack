package metrics

import (
	"context"
	"testing"
	"time"
)

type oneself struct {
	done int
}

func (os *oneself) Refresh() {
	os.done++
}

func TestKeepRefresh(t *testing.T) {
	me, you, them := new(oneself), new(oneself), new(oneself)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	keepFresh(ctx, time.Millisecond*30, me, you, them)

	for _, person := range [...]*oneself{me, you, them} {
		if want, have := 34, person.done; want != have {
			t.Errorf("expected Refresh to have been called %d times, found %d", want, have)
		}
	}
}
