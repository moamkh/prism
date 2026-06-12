package limiter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"rev_core/internal/db"
	"rev_core/internal/metrics"
)

type modelLimiter struct {
	maxConcur int
	maxQueue  int
	active    int
	waiters   []chan struct{}
	mu        sync.Mutex
}

// Acquire attempts to enter the model's concurrency pool.
// If a slot is available, it proceeds immediately.
// If the pool is full but queue has space, it waits for a slot or context cancellation.
// If the queue is also full, it returns an error immediately.
func (ml *modelLimiter) Acquire(ctx context.Context) error {
	ml.mu.Lock()
	if ml.active < ml.maxConcur {
		ml.active++
		ml.mu.Unlock()
		return nil
	}
	if len(ml.waiters) >= ml.maxQueue {
		ml.mu.Unlock()
		return fmt.Errorf("model queue full")
	}
	ch := make(chan struct{}, 1)
	ml.waiters = append(ml.waiters, ch)
	ml.mu.Unlock()

	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		ml.mu.Lock()
		for i, w := range ml.waiters {
			if w == ch {
				ml.waiters = append(ml.waiters[:i], ml.waiters[i+1:]...)
				break
			}
		}
		ml.mu.Unlock()
		return ctx.Err()
	}
}

func (ml *modelLimiter) Release() {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	if len(ml.waiters) > 0 {
		ch := ml.waiters[0]
		ml.waiters = ml.waiters[1:]
		ch <- struct{}{}
	} else if ml.active > 0 {
		ml.active--
	}
}

func (ml *modelLimiter) snapshot() (max, inflight, queue int) {
	ml.mu.Lock()
	defer ml.mu.Unlock()
	return ml.maxConcur, ml.active, len(ml.waiters)
}

// ModelLimiter manages per-model concurrency and queue limits.
type ModelLimiter struct {
	limiters map[uuid.UUID]*modelLimiter
	mu       sync.RWMutex
	db       *db.DB
}

func NewModelLimiter(database *db.DB) *ModelLimiter {
	ml := &ModelLimiter{
		db:       database,
		limiters: make(map[uuid.UUID]*modelLimiter),
	}
	ml.refreshModels()
	go ml.modelPoller()
	return ml
}

func (ml *ModelLimiter) refreshModels() {
	models, err := ml.db.GetActiveModels()
	if err != nil {
		return
	}

	ml.mu.Lock()
	defer ml.mu.Unlock()

	newLimiters := make(map[uuid.UUID]*modelLimiter)
	for _, m := range models {
		maxConcur := m.MaxConcurrentRequests
		maxQueue := m.QueueSize
		if maxConcur <= 0 && maxQueue <= 0 {
			// No limits configured for this model.
			continue
		}
		if maxConcur < 0 {
			maxConcur = 0
		}
		if maxQueue < 0 {
			maxQueue = 0
		}

		if existing, ok := ml.limiters[m.ID]; ok && existing != nil {
			existing.mu.Lock()
			existing.maxConcur = maxConcur
			existing.maxQueue = maxQueue
			existing.mu.Unlock()
			newLimiters[m.ID] = existing
		} else {
			newLimiters[m.ID] = &modelLimiter{
				maxConcur: maxConcur,
				maxQueue:  maxQueue,
			}
		}
		metrics.ModelLimiterMax.WithLabelValues(m.ID.String()).Set(float64(maxConcur))
	}

	// For models that were removed or had limits cleared, keep their limiters
	// in the old map until they drain (active == 0 and no waiters).
	for id, l := range ml.limiters {
		if _, ok := newLimiters[id]; ok {
			continue
		}
		_, inflight, queue := l.snapshot()
		if inflight == 0 && queue == 0 {
			// Fully drained; drop it.
			continue
		}
		// Mark as closed so new requests are rejected.
		l.mu.Lock()
		l.maxConcur = 0
		l.maxQueue = 0
		l.mu.Unlock()
		newLimiters[id] = l
		metrics.ModelLimiterMax.WithLabelValues(id.String()).Set(0)
	}

	ml.limiters = newLimiters
}

func (ml *ModelLimiter) modelPoller() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		ml.refreshModels()
	}
}

// Acquire tries to acquire a slot for the given model.
// If the model has no limit configured, it returns nil immediately.
func (ml *ModelLimiter) Acquire(ctx context.Context, modelID uuid.UUID) error {
	ml.mu.RLock()
	l, ok := ml.limiters[modelID]
	ml.mu.RUnlock()

	if !ok || l == nil {
		return nil // no limit configured
	}

	err := l.Acquire(ctx)
	if err == nil {
		_, inflight, queue := l.snapshot()
		metrics.ModelLimiterInflight.WithLabelValues(modelID.String()).Set(float64(inflight))
		metrics.ModelLimiterQueue.WithLabelValues(modelID.String()).Set(float64(queue))
	}
	return err
}

func (ml *ModelLimiter) Release(modelID uuid.UUID) {
	ml.mu.RLock()
	l, ok := ml.limiters[modelID]
	ml.mu.RUnlock()

	if ok && l != nil {
		l.Release()
		_, inflight, queue := l.snapshot()
		metrics.ModelLimiterInflight.WithLabelValues(modelID.String()).Set(float64(inflight))
		metrics.ModelLimiterQueue.WithLabelValues(modelID.String()).Set(float64(queue))
	}
}

// Status returns a snapshot of per-model limiter state.
func (ml *ModelLimiter) Status() map[uuid.UUID]struct{ Max, Inflight, Queue int } {
	ml.mu.RLock()
	defer ml.mu.RUnlock()
	out := make(map[uuid.UUID]struct{ Max, Inflight, Queue int })
	for id, l := range ml.limiters {
		max, inflight, queue := l.snapshot()
		out[id] = struct{ Max, Inflight, Queue int }{
			Max:       max,
			Inflight:  inflight,
			Queue:     queue,
		}
	}
	return out
}
