package main

import (
	"encoding/json"
	_ "embed"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"rev_core/internal/db"
	"rev_core/internal/limiter"
	"rev_core/internal/logger"
	"rev_core/internal/metrics"
	"rev_core/internal/middleware"
	"rev_core/internal/proxy"
	"rev_core/internal/usage"
)

//go:embed docs/swagger.json
var swaggerJSON []byte

//go:embed docs/swagger-ui.html
var swaggerUI []byte

//go:embed docs/status.html
var statusHTML []byte

//go:embed docs/logs.html
var logsHTML []byte

var startTime = time.Now()

func main() {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgresql://postgres:postgres@localhost:5432/reverse_proxy_manager_db"
	}

	database, err := db.New(databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Conn.Close()

	usageLogger := usage.New(database)
	defer usageLogger.Stop()

	lim := limiter.New(database)
	auth := middleware.NewAuth(database)
	proxyHandler := proxy.New(database, usageLogger, lim)

	// Public routes (no auth)
	publicMux := http.NewServeMux()
	publicMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	publicMux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(swaggerUI)
	})
	publicMux.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(swaggerJSON)
	})
	publicMux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// Update DB gauge metrics before scraping
		stats := database.Conn.Stats()
		metrics.DBConnectionsInUse.Set(float64(stats.InUse))
		metrics.DBConnectionsIdle.Set(float64(stats.Idle))
		promhttp.Handler().ServeHTTP(w, r)
	})
	publicMux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(statusHTML)
	})
	publicMux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		dbStats := database.Conn.Stats()
		bufLen, bufCap, dropped := usageLogger.BufferStats()
		limStatus := lim.Status()

		type providerStatus struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Max       int64  `json:"max"`
			Inflight  int64  `json:"inflight"`
		}

		providers, _ := database.GetActiveProviders()
		providerMap := make(map[string]string)
		for _, p := range providers {
			providerMap[p.ID.String()] = p.Name
		}

		var provList []providerStatus
		for id, st := range limStatus {
			provList = append(provList, providerStatus{
				ID:       id.String(),
				Name:     providerMap[id.String()],
				Max:      st.Max,
				Inflight: st.Inflight,
			})
		}

		total := atomic.LoadInt64(&metrics.RequestCount)
		latencySum := atomic.LoadInt64(&metrics.RequestLatencySum)
		avgLatency := int64(0)
		if total > 0 {
			avgLatency = latencySum / total
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"started_at": startTime.UTC().Format(time.RFC3339),
			"db": map[string]interface{}{
				"in_use": dbStats.InUse,
				"idle":   dbStats.Idle,
			},
			"usage": map[string]interface{}{
				"buffer_len": bufLen,
				"buffer_cap": bufCap,
				"dropped":    dropped,
			},
			"providers": provList,
			"requests": map[string]interface{}{
				"total":         total,
				"avg_latency_ms": avgLatency,
			},
		})
	})

	// Log ring buffer (captures last 1000 lines)
	ringLog := logger.NewRing(1000)
	log.SetOutput(io.MultiWriter(os.Stderr, ringLog))

	publicMux.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(logsHTML)
	})
	publicMux.HandleFunc("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodDelete {
			ringLog.Clear()
			w.WriteHeader(http.StatusNoContent)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"lines": ringLog.Lines(),
		})
	})

	// Protected routes (auth + limiter + proxy)
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			proxyHandler.HandleModelsList(w, r)
			return
		}
		proxyHandler.Handler(w, r)
	})
	protectedMux.HandleFunc("/v1/models/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			modelID := strings.TrimPrefix(r.URL.Path, "/v1/models/")
			if modelID != "" {
				proxyHandler.HandleModelGet(w, r, modelID)
				return
			}
		}
		proxyHandler.Handler(w, r)
	})
	protectedMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxyHandler.Handler(w, r)
	})

	var protectedHandler http.Handler = protectedMux
	protectedHandler = auth.Handler(protectedHandler)

	// Combine: public first, then protected fallback
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/health" || path == "/docs" || path == "/swagger/doc.json" ||
			path == "/metrics" || path == "/status" || path == "/api/status" ||
			path == "/logs" || path == "/api/logs" {
			publicMux.ServeHTTP(w, r)
			return
		}
		protectedHandler.ServeHTTP(w, r)
	})

	// Metrics instrumentation
	handler = metrics.InstrumentHandler(handler)

	// CORS wrapper (configurable via env)
	corsOrigin := os.Getenv("CORS_ALLOW_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "*"
	}
	corsMethods := os.Getenv("CORS_ALLOW_METHODS")
	if corsMethods == "" {
		corsMethods = "GET, POST, PUT, DELETE, OPTIONS"
	}
	corsHeaders := os.Getenv("CORS_ALLOW_HEADERS")
	if corsHeaders == "" {
		corsHeaders = "Authorization, Content-Type, X-Requested-With"
	}
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		w.Header().Set("Access-Control-Allow-Methods", corsMethods)
		w.Header().Set("Access-Control-Allow-Headers", corsHeaders)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		handler.ServeHTTP(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Reverse proxy server starting on :%s", port)
	log.Printf("Docs available at http://localhost:%s/docs", port)
	log.Printf("Status panel available at http://localhost:%s/status", port)
	log.Printf("Logs viewer available at http://localhost:%s/logs", port)
	log.Printf("Metrics available at http://localhost:%s/metrics", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
