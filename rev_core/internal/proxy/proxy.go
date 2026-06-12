package proxy

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/proxy"
	"rev_core/internal/db"
	"rev_core/internal/limiter"
	"rev_core/internal/models"
	"rev_core/internal/usage"
)

type Proxy struct {
	db              *db.DB
	usage           *usage.Batcher
	limiter         *limiter.Limiter
	modelLimiter    *limiter.ModelLimiter
	transportCache  map[uuid.UUID]*http.Transport
	transportMu     sync.RWMutex
}

func New(database *db.DB, usageLogger *usage.Batcher, lim *limiter.Limiter, modelLim *limiter.ModelLimiter) *Proxy {
	return &Proxy{
		db:             database,
		usage:          usageLogger,
		limiter:        lim,
		modelLimiter:   modelLim,
		transportCache: make(map[uuid.UUID]*http.Transport),
	}
}

// getTransport returns a cached *http.Transport for a provider, creating one if needed.
// Transports are reused so connections can be pooled through the SOCKS5/HTTP proxy.
func (p *Proxy) getTransport(provider *models.Provider) *http.Transport {
	p.transportMu.RLock()
	if t, ok := p.transportCache[provider.ID]; ok {
		p.transportMu.RUnlock()
		return t
	}
	p.transportMu.RUnlock()

	p.transportMu.Lock()
	defer p.transportMu.Unlock()
	// Double-check after acquiring write lock
	if t, ok := p.transportCache[provider.ID]; ok {
		return t
	}

	t := buildTransport(provider)
	if t != nil {
		p.transportCache[provider.ID] = t
	}
	return t
}

// retryRoundTripper wraps an http.RoundTripper and retries on transient errors.
type retryRoundTripper struct {
	base       http.RoundTripper
	maxRetries int
	baseDelay  time.Duration
}

func (rt *retryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var lastErr error
	for attempt := 0; attempt <= rt.maxRetries; attempt++ {
		if attempt > 0 {
			delay := rt.baseDelay * time.Duration(1<<(attempt-1))
			if delay > 5*time.Second {
				delay = 5 * time.Second
			}
			// Add jitter (±50%) to prevent synchronized retry storms
			jitter := time.Duration(rand.Int63n(int64(delay))) - delay/2
			delay = delay + jitter
			if delay < 50*time.Millisecond {
				delay = 50 * time.Millisecond
			}
			log.Printf("[PROXY] Retry attempt %d/%d for %s %s after %v", attempt, rt.maxRetries, req.Method, req.URL.Path, delay)
			time.Sleep(delay)
			// Clone request body for retry
			if req.Body != nil && req.GetBody != nil {
				body, err := req.GetBody()
				if err == nil {
					req.Body = body
				}
			}
		}
		resp, err := rt.base.RoundTrip(req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		errStr := err.Error()
		if !isRetryableError(errStr) {
			return nil, err
		}
		log.Printf("[PROXY] Transient error on attempt %d: %v", attempt, err)
	}
	return nil, lastErr
}

func isRetryableError(errStr string) bool {
	lower := strings.ToLower(errStr)
	retryable := []string{
		"remote end closed connection",
		"connection reset by peer",
		"connection refused",
		"broken pipe",
		"eof",
		"timeout",
		"temporary",
		"no route to host",
	}
	for _, s := range retryable {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}

func (p *Proxy) Handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var (
		tokenUUID       uuid.UUID
		providerID      uuid.UUID
		modelUUID       uuid.UUID
		providerModelID string
		providerName    string
		statusCode      int
		errMsg          string
		inputTokens     int
		outputTokens    int
	)

	// Always log usage — successes and failures alike.
	defer func() {
		isSuccessful := statusCode >= 200 && statusCode < 300
		p.usage.Log(
			tokenUUID, providerID, modelUUID, providerModelID, providerName,
			r.URL.Path, inputTokens, outputTokens,
			int(time.Since(start).Milliseconds()), statusCode, isSuccessful, errMsg,
		)
	}()

	modelIDStr := r.Header.Get("X-Proxy-Model-ID")
	if modelIDStr != "" {
		modelUUID = uuid.MustParse(modelIDStr)
	}

	provider, pModelID, err := p.resolveProvider(modelUUID)
	if err != nil {
		statusCode = http.StatusBadRequest
		errMsg = err.Error()
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	providerID = provider.ID
	providerName = provider.Name
	providerModelID = pModelID

	if p.modelLimiter != nil && modelUUID != uuid.Nil {
		if err := p.modelLimiter.Acquire(r.Context(), modelUUID); err != nil {
			statusCode = http.StatusTooManyRequests
			errMsg = "Model concurrency limit reached or queue full"
			http.Error(w, `{"error":"Model concurrency limit reached or queue full"}`, http.StatusTooManyRequests)
			return
		}
		defer p.modelLimiter.Release(modelUUID)
	}

	if p.limiter != nil {
		if err := p.limiter.Acquire(r.Context(), provider.ID); err != nil {
			statusCode = http.StatusTooManyRequests
			errMsg = "Too many concurrent requests for this provider, queue timeout"
			http.Error(w, `{"error":"Too many concurrent requests for this provider, queue timeout"}`, http.StatusTooManyRequests)
			return
		}
		defer p.limiter.Release(provider.ID)
	}

	tokenIDStr := r.Header.Get("X-Proxy-Token-ID")
	if tokenIDStr != "" {
		tokenUUID = uuid.MustParse(tokenIDStr)
	}

	targetURL, err := url.Parse(provider.BaseURL)
	if err != nil {
		statusCode = http.StatusInternalServerError
		errMsg = "Invalid provider base URL"
		http.Error(w, `{"error":"Invalid provider base URL"}`, http.StatusInternalServerError)
		return
	}

	r.Header.Set("Authorization", "Bearer "+provider.APIToken)
	r.Header.Del("X-Proxy-Token-ID")
	r.Header.Del("X-Proxy-Model-ID")

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.Host = targetURL.Host
		req.URL.Path = r.URL.Path
		req.URL.RawQuery = r.URL.RawQuery
	}
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Del("Access-Control-Allow-Origin")
		resp.Header.Del("Access-Control-Allow-Methods")
		resp.Header.Del("Access-Control-Allow-Headers")
		resp.Header.Del("Access-Control-Expose-Headers")
		return nil
	}

	// Configure proxy transport (only if provider has enable_proxy=true)
	log.Printf("[PROXY] Provider=%s EnableProxy=%v HTTPProxy=%v SOCKS5Proxy=%v",
		provider.Name, provider.EnableProxy,
		func() string { if provider.HTTPProxy != nil { return *provider.HTTPProxy }; return "<nil>" }(),
		func() string { if provider.Socks5Proxy != nil { return *provider.Socks5Proxy }; return "<nil>" }(),
	)
	if provider.EnableProxy {
		transport := p.getTransport(provider)
		if transport != nil {
			// Wrap with retry logic for transient upstream errors
			proxy.Transport = &retryRoundTripper{
				base:       transport,
				maxRetries: 1,
				baseDelay:  500 * time.Millisecond,
			}
			log.Printf("[PROXY] Using cached transport with retry for provider %s", provider.Name)
		} else {
			log.Printf("[PROXY] No proxy URL configured for provider %s", provider.Name)
		}
	} else {
		log.Printf("[PROXY] Proxy disabled for provider %s", provider.Name)
	}

	isStream := false
	if r.Method == http.MethodPost {
		bodyBytes, _ := io.ReadAll(r.Body)
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &bodyMap); err == nil {
			if s, ok := bodyMap["stream"].(bool); ok {
				isStream = s
			}
			// Rewrite model field to actual provider model_id
			if providerModelID != "" {
				if _, ok := bodyMap["model"].(string); ok {
					bodyMap["model"] = providerModelID
					newBody, _ := json.Marshal(bodyMap)
					bodyBytes = newBody
				}
			}
		}
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		r.ContentLength = int64(len(bodyBytes))
		r.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(bodyBytes)), nil
		}
	}

	if isStream {
		recorder := &streamRecorder{
			ResponseWriter: w,
			modelID:        providerModelID,
		}
		proxy.ServeHTTP(recorder, r)
		inputTokens = recorder.inputTokens
		outputTokens = recorder.outputTokens
		statusCode = http.StatusOK
		if recorder.statusCode != 0 {
			statusCode = recorder.statusCode
		}
		if statusCode >= 400 {
			errMsg = "Upstream error"
		}
	} else {
		recorder := &responseRecorder{ResponseWriter: w}
		proxy.ServeHTTP(recorder, r)
		statusCode = recorder.statusCode
		if len(recorder.body) > 0 {
			inputTokens, outputTokens = p.extractUsage(recorder.body, recorder.header().Get("Content-Encoding"))
		}
		if statusCode >= 400 {
			errMsg = "Upstream error"
		}
		for k, v := range recorder.header() {
			w.Header()[k] = v
		}
		w.WriteHeader(recorder.statusCode)
		w.Write(recorder.body)
	}
}

// HandleModelsList aggregates /v1/models from all active providers and filters
// by the caller's token permissions. Replaces visible model IDs with display_model_id
// when configured.
func (p *Proxy) HandleModelsList(w http.ResponseWriter, r *http.Request) {
	tokenIDStr := r.Header.Get("X-Proxy-Token-ID")
	var tokenUUID uuid.UUID
	if tokenIDStr != "" {
		tokenUUID = uuid.MustParse(tokenIDStr)
	}

	allModels, err := p.db.GetAllModels()
	if err != nil {
		http.Error(w, `{"error":"failed to load models"}`, http.StatusInternalServerError)
		return
	}

	// Build set of allowed model UUIDs for this token
	allowedModelUUIDs := make(map[uuid.UUID]bool)
	if tokenUUID != uuid.Nil {
		perms, err := p.db.GetTokenPermissions(tokenUUID)
		if err != nil {
			http.Error(w, `{"error":"failed to load token permissions"}`, http.StatusInternalServerError)
			return
		}
		for _, perm := range perms {
			allowedModelUUIDs[perm.ModelID] = true
		}
	}

	// Map actual model_id -> our Model record
	modelByActualID := make(map[string]models.Model)
	for _, m := range allModels {
		modelByActualID[m.ModelID] = m
	}

	providers, err := p.db.GetActiveProviders()
	if err != nil || len(providers) == 0 {
		http.Error(w, `{"error":"no active providers"}`, http.StatusServiceUnavailable)
		return
	}

	// Fetch from all providers concurrently
	type result struct {
		items []map[string]interface{}
		err   error
	}
	results := make(chan result, len(providers))
	var wg sync.WaitGroup

	for i := range providers {
		wg.Add(1)
		go func(prv models.Provider) {
			defer wg.Done()
			items, err := p.fetchModelsFromProvider(&prv)
			results <- result{items: items, err: err}
		}(providers[i])
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	// Merge and deduplicate by display ID
	seen := make(map[string]bool)
	var merged []map[string]interface{}
	for res := range results {
		if res.err != nil {
			continue
		}
		for _, item := range res.items {
			actualID, _ := item["id"].(string)
			if actualID == "" {
				continue
			}
			ourModel, hasModel := modelByActualID[actualID]
			if !hasModel {
				continue // skip models not registered in our DB
			}
			// Filter by token permissions
			if len(allowedModelUUIDs) > 0 && !allowedModelUUIDs[ourModel.ID] {
				continue
			}
			displayID := actualID
			if ourModel.DisplayModelID != nil && *ourModel.DisplayModelID != "" {
				displayID = *ourModel.DisplayModelID
			}
			if seen[displayID] {
				continue
			}
			seen[displayID] = true
			item["id"] = displayID
			merged = append(merged, item)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"object": "list",
		"data":   merged,
	})
}

// HandleModelGet proxies GET /v1/models/{model_id} to the provider that owns the model.
// Accepts either actual model_id or display_model_id in the URL.
func (p *Proxy) HandleModelGet(w http.ResponseWriter, r *http.Request, modelID string) {
	tokenIDStr := r.Header.Get("X-Proxy-Token-ID")
	var tokenUUID uuid.UUID
	if tokenIDStr != "" {
		tokenUUID = uuid.MustParse(tokenIDStr)
	}

	// Find model in DB by actual or display model_id
	allModels, err := p.db.GetAllModels()
	if err != nil {
		http.Error(w, `{"error":"failed to load models"}`, http.StatusInternalServerError)
		return
	}
	var targetModel *models.Model
	for i := range allModels {
		if allModels[i].ModelID == modelID {
			targetModel = &allModels[i]
			break
		}
		if allModels[i].DisplayModelID != nil && *allModels[i].DisplayModelID == modelID {
			targetModel = &allModels[i]
			break
		}
	}
	if targetModel == nil {
		http.Error(w, `{"error":"model not found"}`, http.StatusNotFound)
		return
	}

	// Check token permission
	if tokenUUID != uuid.Nil {
		perms, err := p.db.GetTokenPermissions(tokenUUID)
		if err != nil {
			http.Error(w, `{"error":"failed to load token permissions"}`, http.StatusInternalServerError)
			return
		}
		allowed := false
		for _, perm := range perms {
			if perm.ModelID == targetModel.ID {
				allowed = true
				break
			}
		}
		if !allowed {
			http.Error(w, `{"error":"Token not authorized for this model"}`, http.StatusForbidden)
			return
		}
	}

	// Find provider
	providers, err := p.db.GetActiveProviders()
	if err != nil {
		http.Error(w, `{"error":"failed to load providers"}`, http.StatusInternalServerError)
		return
	}
	var provider *models.Provider
	for i := range providers {
		if providers[i].ID == targetModel.ProviderID {
			provider = &providers[i]
			break
		}
	}
	if provider == nil {
		http.Error(w, `{"error":"provider for model not available"}`, http.StatusServiceUnavailable)
		return
	}

	// Proxy the request using actual model_id
	base := strings.TrimSuffix(provider.BaseURL, "/")
	urlStr := base + "/models/" + targetModel.ModelID
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, urlStr, nil)
	if err != nil {
		http.Error(w, `{"error":"failed to build request"}`, http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Bearer "+provider.APIToken)

	client := &http.Client{Timeout: 30 * time.Second}
	if provider.EnableProxy {
		transport := p.getTransport(provider)
		if transport != nil {
			client.Transport = transport
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"upstream request failed: %s"}`, err.Error()), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	// Rewrite response id to display_model_id if set
	var respMap map[string]interface{}
	if err := json.Unmarshal(body, &respMap); err == nil {
		if id, ok := respMap["id"].(string); ok && id == targetModel.ModelID {
			displayID := targetModel.ModelID
			if targetModel.DisplayModelID != nil && *targetModel.DisplayModelID != "" {
				displayID = *targetModel.DisplayModelID
			}
			respMap["id"] = displayID
		}
		body, _ = json.Marshal(respMap)
	}
	w.Header().Set("Content-Type", "application/json")
	// Strip upstream CORS headers — the global CORS wrapper handles them
	for k := range w.Header() {
		if strings.EqualFold(k, "Access-Control-Allow-Origin") ||
			strings.EqualFold(k, "Access-Control-Allow-Methods") ||
			strings.EqualFold(k, "Access-Control-Allow-Headers") ||
			strings.EqualFold(k, "Access-Control-Expose-Headers") {
			w.Header().Del(k)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func (p *Proxy) fetchModelsFromProvider(provider *models.Provider) ([]map[string]interface{}, error) {
	base := strings.TrimSuffix(provider.BaseURL, "/")
	urlStr := base + "/models"
	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+provider.APIToken)

	client := &http.Client{Timeout: 30 * time.Second}
	if provider.EnableProxy {
		transport := p.getTransport(provider)
		if transport != nil {
			client.Transport = transport
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream status %d", resp.StatusCode)
	}

	var payload struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return payload.Data, nil
}

func buildTransport(provider *models.Provider) *http.Transport {
	transport := &http.Transport{
		MaxIdleConns:        256,
		MaxIdleConnsPerHost: 128,
		IdleConnTimeout:     90 * time.Second,
		MaxConnsPerHost:     12, // limit concurrent new connections to avoid overwhelming SOCKS5 proxy
		DisableKeepAlives:   false,
	}

	if provider.Socks5Proxy != nil && *provider.Socks5Proxy != "" {
		addr := *provider.Socks5Proxy
		// Strip socks5:// prefix if present
		addr = strings.TrimPrefix(addr, "socks5://")
		addr = strings.TrimPrefix(addr, "socks5://")
		dialer, err := proxy.SOCKS5("tcp", addr, nil, proxy.Direct)
		if err != nil {
			log.Printf("[PROXY] Failed to create SOCKS5 dialer for %s: %v", addr, err)
			return nil
		}
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}
		log.Printf("[PROXY] SOCKS5 proxy configured: %s (pool: idle=%d perHost=%d)", addr, transport.MaxIdleConns, transport.MaxIdleConnsPerHost)
		return transport
	}

	if provider.HTTPProxy != nil && *provider.HTTPProxy != "" {
		proxyURLStr := *provider.HTTPProxy
		// Add http:// scheme if missing
		if !strings.Contains(proxyURLStr, "://") {
			proxyURLStr = "http://" + proxyURLStr
		}
		proxyURL, err := url.Parse(proxyURLStr)
		if err != nil {
			log.Printf("[PROXY] Failed to parse HTTP proxy URL %s: %v", proxyURLStr, err)
			return nil
		}
		transport.Proxy = http.ProxyURL(proxyURL)
		log.Printf("[PROXY] HTTP proxy configured: %s", proxyURL.String())
		return transport
	}

	return nil
}

func (p *Proxy) resolveProvider(modelUUID uuid.UUID) (*models.Provider, string, error) {
	if modelUUID == uuid.Nil {
		providers, err := p.db.GetActiveProviders()
		if err != nil || len(providers) == 0 {
			return nil, "", fmt.Errorf("no active providers available")
		}
		return &providers[0], "", nil
	}

	allModels, err := p.db.GetAllModels()
	if err != nil {
		return nil, "", fmt.Errorf("failed to load models")
	}
	var targetModel *models.Model
	for i := range allModels {
		if allModels[i].ID == modelUUID {
			targetModel = &allModels[i]
			break
		}
	}
	if targetModel == nil {
		return nil, "", fmt.Errorf("model not found")
	}

	providers, err := p.db.GetActiveProviders()
	if err != nil {
		return nil, "", fmt.Errorf("failed to load providers")
	}
	for i := range providers {
		if providers[i].ID == targetModel.ProviderID {
			return &providers[i], targetModel.ModelID, nil
		}
	}
	return nil, "", fmt.Errorf("provider for model not available")
}

func (p *Proxy) extractUsage(body []byte, encoding string) (int, int) {
	var data []byte
	if encoding == "gzip" {
		reader, err := gzip.NewReader(bytes.NewReader(body))
		if err == nil {
			data, _ = io.ReadAll(reader)
			reader.Close()
		}
	} else {
		data = body
	}

	var resp struct {
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(data, &resp); err == nil {
		return resp.Usage.PromptTokens, resp.Usage.CompletionTokens
	}
	return 0, 0
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode  int
	body        []byte
	wroteHeader bool
}

func (rr *responseRecorder) header() http.Header {
	return rr.ResponseWriter.Header()
}

func (rr *responseRecorder) WriteHeader(code int) {
	if rr.wroteHeader {
		return
	}
	rr.statusCode = code
	rr.wroteHeader = true
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	if !rr.wroteHeader {
		rr.WriteHeader(http.StatusOK)
	}
	rr.body = append(rr.body, b...)
	return len(b), nil
}

func (rr *responseRecorder) Header() http.Header {
	return rr.ResponseWriter.Header()
}

type streamRecorder struct {
	http.ResponseWriter
	modelID       string
	inputTokens   int
	outputTokens  int
	statusCode    int
	headerWritten bool
}

func (sr *streamRecorder) Header() http.Header {
	return sr.ResponseWriter.Header()
}

func (sr *streamRecorder) WriteHeader(code int) {
	if sr.headerWritten {
		return
	}
	sr.statusCode = code
	sr.ResponseWriter.WriteHeader(code)
	sr.headerWritten = true
}

func (sr *streamRecorder) Write(b []byte) (int, error) {
	if !sr.headerWritten {
		sr.WriteHeader(http.StatusOK)
	}
	chunks := strings.Split(string(b), "\n\n")
	for _, chunk := range chunks {
		lines := strings.Split(chunk, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					continue
				}
				var event struct {
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
					} `json:"choices"`
					Usage struct {
						CompletionTokens int `json:"completion_tokens"`
						PromptTokens     int `json:"prompt_tokens"`
					} `json:"usage"`
				}
				if err := json.Unmarshal([]byte(data), &event); err == nil {
					if event.Usage.PromptTokens > 0 {
						sr.inputTokens = event.Usage.PromptTokens
					}
					if event.Usage.CompletionTokens > 0 {
						sr.outputTokens = event.Usage.CompletionTokens
					} else if len(event.Choices) > 0 && event.Choices[0].Delta.Content != "" {
						sr.outputTokens += estimateTokens(event.Choices[0].Delta.Content)
					}
				}
			}
		}
	}
	return sr.ResponseWriter.Write(b)
}

func estimateTokens(text string) int {
	if len(text) == 0 {
		return 0
	}
	return (len(text) + 3) / 4
}
