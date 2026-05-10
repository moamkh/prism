package limiter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
	"github.com/google/uuid"
	"rev_core/internal/db"
	"rev_core/internal/metrics"
)

type providerSem struct {
	sem      *semaphore.Weighted
	maxWeight int64
}

type Limiter struct {
	sems     map[uuid.UUID]*providerSem
	inflight map[uuid.UUID]int64
	mu       sync.RWMutex
	db       *db.DB
	queueTimeout time.Duration
}

func New(database *db.DB) *Limiter {
	l := &Limiter{
		db:       database,
		sems:     make(map[uuid.UUID]*providerSem),
		inflight: make(map[uuid.UUID]int64),
	}
	l.refreshProviders()
	go l.providerPoller()
	return l
}

func (l *Limiter) refreshProviders() {
	providers, err := l.db.GetActiveProviders()
	if err != nil {
		return
	}

	qtVal, _ := l.db.GetConfig("queue_timeout_seconds")
	qt := 30
	if v, err := fmt.Sscanf(qtVal, "%d", &qt); err != nil || v != 1 || qt <= 0 {
		qt = 30
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Reuse existing semaphores so in-flight requests don't panic on Release.
	newSems := make(map[uuid.UUID]*providerSem)
	for _, p := range providers {
		maxReq := int64(p.MaxConcurrentRequests)
		if maxReq <= 0 {
			maxReq = 100
		}
		if existing, ok := l.sems[p.ID]; ok && existing != nil && existing.sem != nil {
			newSems[p.ID] = &providerSem{
				sem:       existing.sem,
				maxWeight: maxReq,
			}
		} else {
			newSems[p.ID] = &providerSem{
				sem:       semaphore.NewWeighted(maxReq),
				maxWeight: maxReq,
			}
		}
		metrics.LimiterMax.WithLabelValues(p.ID.String()).Set(float64(maxReq))
	}

	// Drop inflight counters for providers that were removed.
	newInflight := make(map[uuid.UUID]int64)
	for id, v := range l.inflight {
		if _, ok := newSems[id]; ok {
			newInflight[id] = v
		}
	}

	l.sems = newSems
	l.inflight = newInflight
	l.queueTimeout = time.Duration(qt) * time.Second
}

func (l *Limiter) providerPoller() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		l.refreshProviders()
	}
}

func (l *Limiter) Acquire(ctx context.Context, providerID uuid.UUID) error {
	l.mu.RLock()
	ps, ok := l.sems[providerID]
	timeout := l.queueTimeout
	l.mu.RUnlock()

	if !ok || ps == nil {
		return fmt.Errorf("provider limiter not found")
	}

	queueCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := ps.sem.Acquire(queueCtx, 1)
	if err == nil {
		l.mu.Lock()
		l.inflight[providerID]++
		metrics.LimiterInflight.WithLabelValues(providerID.String()).Set(float64(l.inflight[providerID]))
		l.mu.Unlock()
	}
	return err
}

func (l *Limiter) Release(providerID uuid.UUID) {
	defer func() {
		if rec := recover(); rec != nil {
			// Defensive: semaphore panic should never crash the server.
		}
	}()
	l.mu.RLock()
	ps, ok := l.sems[providerID]
	l.mu.RUnlock()
	if ok && ps != nil && ps.sem != nil {
		ps.sem.Release(1)
		l.mu.Lock()
		if l.inflight[providerID] > 0 {
			l.inflight[providerID]--
		}
		metrics.LimiterInflight.WithLabelValues(providerID.String()).Set(float64(l.inflight[providerID]))
		l.mu.Unlock()
	}
}

// Status returns a snapshot of per-provider limiter state.
func (l *Limiter) Status() map[uuid.UUID]struct{ Max, Inflight int64 } {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make(map[uuid.UUID]struct{ Max, Inflight int64 })
	for id, ps := range l.sems {
		out[id] = struct{ Max, Inflight int64 }{
			Max:       ps.maxWeight,
			Inflight:  l.inflight[id],
		}
	}
	return out
}
