package limiter

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestModelLimiterAcquireRelease(t *testing.T) {
	ml := &modelLimiter{
		maxConcur: 2,
		maxQueue:  1,
	}

	ctx := context.Background()

	// Acquire 2 slots (max)
	if err := ml.Acquire(ctx); err != nil {
		t.Fatalf("expected first acquire to succeed, got %v", err)
	}
	if err := ml.Acquire(ctx); err != nil {
		t.Fatalf("expected second acquire to succeed, got %v", err)
	}

	_, active, _ := ml.snapshot()
	if active != 2 {
		t.Fatalf("expected active=2, got %d", active)
	}

	// Third acquire should queue (waiter)
	done := make(chan error, 1)
	go func() {
		done <- ml.Acquire(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	_, active, queue := ml.snapshot()
	if active != 2 {
		t.Fatalf("expected active=2, got %d", active)
	}
	if queue != 1 {
		t.Fatalf("expected queue=1, got %d", queue)
	}

	// Fourth acquire should fail immediately (queue full)
	if err := ml.Acquire(ctx); err == nil {
		t.Fatal("expected fourth acquire to fail with queue full")
	}

	// Release one slot -> queued request should proceed
	ml.Release()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected queued acquire to succeed after release, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for queued acquire")
	}

	_, active, queue = ml.snapshot()
	if active != 2 {
		t.Fatalf("expected active=2 after queued acquire proceeds, got %d", active)
	}
	if queue != 0 {
		t.Fatalf("expected queue=0, got %d", queue)
	}

	// Release all
	ml.Release()
	ml.Release()

	_, active, queue = ml.snapshot()
	if active != 0 {
		t.Fatalf("expected active=0, got %d", active)
	}
	if queue != 0 {
		t.Fatalf("expected queue=0, got %d", queue)
	}
}

func TestModelLimiterContextCancel(t *testing.T) {
	ml := &modelLimiter{
		maxConcur: 1,
		maxQueue:  1,
	}

	ctx := context.Background()
	if err := ml.Acquire(ctx); err != nil {
		t.Fatalf("expected acquire to succeed, got %v", err)
	}

	// Queue a waiter with a cancellable context
	ctx2, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- ml.Acquire(ctx2)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected cancelled acquire to fail")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for cancelled acquire")
	}

	_, _, queue := ml.snapshot()
	if queue != 0 {
		t.Fatalf("expected queue=0 after cancel, got %d", queue)
	}

	ml.Release()
}

func TestModelLimiterNoLimit(t *testing.T) {
	ml := &modelLimiter{
		maxConcur: 0,
		maxQueue:  0,
	}

	ctx := context.Background()
	// With maxConcur=0, all acquires should queue if we follow strict logic,
	// but in ModelLimiter.Acquire we skip models with no limit configured.
	// Here we test the limiter directly: maxConcur=0 means acquire will queue
	// or fail depending on maxQueue.
	if err := ml.Acquire(ctx); err == nil {
		t.Fatal("expected acquire to fail when maxConcur=0 and maxQueue=0")
	}
}

func TestModelLimiterStatus(t *testing.T) {
	m := &ModelLimiter{
		limiters: make(map[uuid.UUID]*modelLimiter),
	}
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	m.mu.Lock()
	m.limiters[id] = &modelLimiter{
		maxConcur: 5,
		maxQueue:  3,
		active:    2,
		waiters:   make([]chan struct{}, 1),
	}
	m.mu.Unlock()

	st := m.Status()
	info, ok := st[id]
	if !ok {
		t.Fatal("expected model in status")
	}
	if info.Max != 5 {
		t.Fatalf("expected max=5, got %d", info.Max)
	}
	if info.Inflight != 2 {
		t.Fatalf("expected inflight=2, got %d", info.Inflight)
	}
	if info.Queue != 1 {
		t.Fatalf("expected queue=1, got %d", info.Queue)
	}
}
