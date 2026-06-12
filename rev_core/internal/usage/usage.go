package usage

import (
	"database/sql"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"rev_core/internal/db"
	"rev_core/internal/metrics"
	"rev_core/internal/models"
)

func nowEpoch() int64 {
	return time.Now().UTC().Unix()
}

type Batcher struct {
	db            *db.DB
	buffer        chan models.UsageLog
	batchSize     int
	flushInterval time.Duration
	done          chan struct{}
	droppedTotal  int64
}

func New(database *db.DB) *Batcher {
	b := &Batcher{
		db:        database,
		buffer:    make(chan models.UsageLog, 1000),
		batchSize: 100,
		flushInterval: 1 * time.Second,
		done:      make(chan struct{}),
	}
	go b.loop()
	return b
}

func (b *Batcher) loop() {
	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()
	batch := make([]models.UsageLog, 0, b.batchSize)

	for {
		select {
		case entry := <-b.buffer:
			batch = append(batch, entry)
			if len(batch) >= b.batchSize {
				b.flush(batch)
				batch = make([]models.UsageLog, 0, b.batchSize)
			}
		case <-ticker.C:
			if len(batch) > 0 {
				b.flush(batch)
				batch = make([]models.UsageLog, 0, b.batchSize)
			}
		case <-b.done:
			if len(batch) > 0 {
				b.flush(batch)
			}
			return
		}
	}
}

func (b *Batcher) flush(batch []models.UsageLog) {
	if err := b.db.InsertUsageLogBatch(batch); err != nil {
		// In production, send to a dead-letter log or retry queue
	}
}

func (b *Batcher) Log(tokenID, providerID, modelID uuid.UUID, modelName, providerName, path string, inputTokens, outputTokens, latencyMs, statusCode int, isSuccessful bool, errMsg string) {
	select {
	case b.buffer <- models.UsageLog{
		ID:           uuid.New(),
		TokenID:      uuidToNull(tokenID),
		ProviderID:   uuidToNull(providerID),
		ModelID:      uuidToNull(modelID),
		ModelName:    stringToNull(modelName),
		ProviderName: stringToNull(providerName),
		RequestPath:  path,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  inputTokens + outputTokens,
		LatencyMs:    intToNullInt32(latencyMs),
		StatusCode:   intToNullInt32(statusCode),
		IsSuccessful: isSuccessful,
		ErrorMessage: stringToNull(errMsg),
		CreatedAt:    nowEpoch(),
	}:
		metrics.UsageBufferLength.Set(float64(len(b.buffer)))
	default:
		// Buffer full, drop log to avoid blocking. In production, alert on this.
		atomic.AddInt64(&b.droppedTotal, 1)
		metrics.UsageDroppedTotal.Inc()
		metrics.UsageBufferLength.Set(float64(len(b.buffer)))
	}
}

// BufferStats returns current buffer and drop statistics.
func (b *Batcher) BufferStats() (length int, capacity int, dropped int64) {
	return len(b.buffer), cap(b.buffer), atomic.LoadInt64(&b.droppedTotal)
}

func (b *Batcher) Stop() {
	close(b.done)
}

func uuidToNull(id uuid.UUID) uuid.NullUUID {
	if id == uuid.Nil {
		return uuid.NullUUID{Valid: false}
	}
	return uuid.NullUUID{UUID: id, Valid: true}
}

func intToNullInt32(v int) sql.NullInt32 {
	if v < 0 {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{Int32: int32(v), Valid: true}
}

func stringToNull(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
