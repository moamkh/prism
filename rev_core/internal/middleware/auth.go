package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"rev_core/internal/db"
	"rev_core/internal/metrics"
)

// Simple in-memory rate limiter: token hash -> []timestamps
type rateLimiter struct {
	mu      sync.RWMutex
	windows map[string][]time.Time
}

func newRateLimiter() *rateLimiter {
	rl := &rateLimiter{
		windows: make(map[string][]time.Time),
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		rl.mu.Lock()
		for key, times := range rl.windows {
			var filtered []time.Time
			for _, t := range times {
				if now.Sub(t) < time.Minute {
					filtered = append(filtered, t)
				}
			}
			if len(filtered) == 0 {
				delete(rl.windows, key)
			} else {
				rl.windows[key] = filtered
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *rateLimiter) allow(tokenHash string, rpm int) bool {
	if rpm <= 0 {
		return true
	}
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	var times []time.Time
	for _, t := range rl.windows[tokenHash] {
		if now.Sub(t) < time.Minute {
			times = append(times, t)
		}
	}
	if len(times) >= rpm {
		rl.windows[tokenHash] = times
		return false
	}
	rl.windows[tokenHash] = append(times, now)
	return true
}

type AuthMiddleware struct {
	db        *db.DB
	rateLimit *rateLimiter
}

func NewAuth(database *db.DB) *AuthMiddleware {
	return &AuthMiddleware{
		db:        database,
		rateLimit: newRateLimiter(),
	}
}

func (a *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			metrics.AuthFailuresTotal.WithLabelValues("missing_auth").Inc()
			http.Error(w, `{"error":"Missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			metrics.AuthFailuresTotal.WithLabelValues("invalid_format").Inc()
			http.Error(w, `{"error":"Invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}
		plainKey := parts[1]
		keyHash := sha256.Sum256([]byte(plainKey))
		keyHashStr := hex.EncodeToString(keyHash[:])

		token, err := a.db.GetTokenByHash(keyHashStr)
		if err != nil || token == nil {
			metrics.AuthFailuresTotal.WithLabelValues("invalid_token").Inc()
			http.Error(w, `{"error":"Invalid token"}`, http.StatusUnauthorized)
			return
		}

		// Rate limit check
		rpm := 0
		if token.RequestsPerMinute.Valid {
			rpm = int(token.RequestsPerMinute.Int32)
		}
		if rpm == 0 {
			defRpmStr, _ := a.db.GetConfig("default_rpm_limit")
			if defRpmStr != "" {
				fmt.Sscanf(defRpmStr, "%d", &rpm)
			}
		}
		if !a.rateLimit.allow(keyHashStr, rpm) {
			metrics.AuthFailuresTotal.WithLabelValues("rate_limited").Inc()
			http.Error(w, `{"error":"Rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		// Extract model from body and apply token limits in a single pass
		modelUUID := uuid.Nil
		var maxOutputTokens int
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				var bodyMap map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &bodyMap); err == nil {
					// Extract model name
					if m, ok := bodyMap["model"].(string); ok && m != "" {
						modelUUID = a.findModelUUID(m)
					}

					// Check token permissions and per-model limits
					if modelUUID != uuid.Nil {
						perms, err := a.db.GetTokenPermissions(token.ID)
						if err != nil {
							http.Error(w, `{"error":"Internal error"}`, http.StatusInternalServerError)
							return
						}
						var allowed bool
						for _, perm := range perms {
							if perm.ModelID == modelUUID {
								allowed = true
								if perm.MaxOutputTokens.Valid {
									maxOutputTokens = int(perm.MaxOutputTokens.Int32)
								}
								break
							}
						}
						if !allowed {
							metrics.AuthFailuresTotal.WithLabelValues("unauthorized_model").Inc()
							http.Error(w, `{"error":"Token not authorized for this model"}`, http.StatusForbidden)
							return
						}
					}

					// Apply max_output_tokens limit by capping max_tokens / max_completion_tokens
					if maxOutputTokens > 0 {
						modified := false
						if m, ok := bodyMap["max_tokens"].(float64); !ok || int(m) > maxOutputTokens {
							bodyMap["max_tokens"] = maxOutputTokens
							modified = true
						}
						if m, ok := bodyMap["max_completion_tokens"].(float64); !ok || int(m) > maxOutputTokens {
							bodyMap["max_completion_tokens"] = maxOutputTokens
							modified = true
						}
						if modified {
							newBody, _ := json.Marshal(bodyMap)
							bodyBytes = newBody
						}
					}
				}
				// Re-inject body so downstream can read it
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
				r.ContentLength = int64(len(bodyBytes))
				r.GetBody = func() (io.ReadCloser, error) {
					return io.NopCloser(bytes.NewReader(bodyBytes)), nil
				}
			}
		}

		// Pass token and model IDs downstream via headers
		r.Header.Set("X-Proxy-Token-ID", token.ID.String())
		if modelUUID != uuid.Nil {
			r.Header.Set("X-Proxy-Model-ID", modelUUID.String())
		}

		next.ServeHTTP(w, r)
	})
}

func (a *AuthMiddleware) findModelUUID(modelIDStr string) uuid.UUID {
	models, err := a.db.GetAllModels()
	if err != nil {
		return uuid.Nil
	}
	for _, m := range models {
		if m.ModelID == modelIDStr {
			return m.ID
		}
		if m.DisplayModelID != nil && *m.DisplayModelID == modelIDStr {
			return m.ID
		}
	}
	return uuid.Nil
}
