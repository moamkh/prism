package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"
	"time"

	"rev_core/internal/models"
)

// TestBuildTransport_HTTPProxy verifies HTTP proxy transport is built correctly
func TestBuildTransport_HTTPProxy(t *testing.T) {
	proxyURL := "http://127.0.0.1:8080"
	provider := &models.Provider{
		EnableProxy: true,
		HTTPProxy:   &proxyURL,
	}

	transport := buildTransport(provider)
	if transport == nil {
		t.Fatal("expected transport, got nil")
	}

	// Verify Proxy function is set
	if transport.Proxy == nil {
		t.Fatal("expected Proxy function to be set")
	}

	req, _ := http.NewRequest("GET", "https://api.openai.com/v1/models", nil)
	proxyURLResult, err := transport.Proxy(req)
	if err != nil {
		t.Fatalf("Proxy function failed: %v", err)
	}
	if proxyURLResult == nil {
		t.Fatal("expected proxy URL, got nil")
	}
	if proxyURLResult.Host != "127.0.0.1:8080" {
		t.Fatalf("expected proxy host 127.0.0.1:8080, got %s", proxyURLResult.Host)
	}
}

// TestBuildTransport_SOCKS5Proxy verifies SOCKS5 proxy transport is built correctly
func TestBuildTransport_SOCKS5Proxy(t *testing.T) {
	proxyAddr := "127.0.0.1:1080"
	provider := &models.Provider{
		EnableProxy: true,
		Socks5Proxy: &proxyAddr,
	}

	transport := buildTransport(provider)
	if transport == nil {
		t.Fatal("expected transport, got nil")
	}

	// Verify DialContext is set
	if transport.DialContext == nil {
		t.Fatal("expected DialContext to be set")
	}
}

// TestBuildTransport_NoProxy returns nil when no proxy configured
func TestBuildTransport_NoProxy(t *testing.T) {
	provider := &models.Provider{
		EnableProxy: true,
	}

	transport := buildTransport(provider)
	if transport != nil {
		t.Fatal("expected nil transport when no proxy configured")
	}
}

// TestBuildTransport_HTTPProxyWithoutScheme adds http:// if missing
func TestBuildTransport_HTTPProxyWithoutScheme(t *testing.T) {
	proxyAddr := "127.0.0.1:8080"
	provider := &models.Provider{
		EnableProxy: true,
		HTTPProxy:   &proxyAddr,
	}

	transport := buildTransport(provider)
	if transport == nil {
		t.Fatal("expected transport")
	}

	req, _ := http.NewRequest("GET", "https://api.openai.com/v1/models", nil)
	proxyURLResult, err := transport.Proxy(req)
	if err != nil {
		t.Fatalf("Proxy function failed: %v", err)
	}
	if proxyURLResult == nil {
		t.Fatal("expected proxy URL")
	}
	if proxyURLResult.Scheme != "http" {
		t.Fatalf("expected scheme http, got %s", proxyURLResult.Scheme)
	}
}

// TestHTTPProxyIntegration proves requests actually go through the HTTP proxy
func TestHTTPProxyIntegration(t *testing.T) {
	// 1. Start a mock upstream server (simulates OpenAI API)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"models":[]}`))
	}))
	defer upstream.Close()

	// 2. Start a mock HTTP proxy server that records requests
	var proxiedRequests []*http.Request
	proxyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxiedRequests = append(proxiedRequests, r)
		// Forward to upstream
		resp, err := http.Get(upstream.URL + r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
	}))
	defer proxyServer.Close()

	// 3. Create provider config pointing at upstream with HTTP proxy enabled
	proxyURL := proxyServer.URL
	provider := &models.Provider{
		Name:        "TestProvider",
		BaseURL:     upstream.URL,
		APIToken:    "test-token",
		EnableProxy: true,
		HTTPProxy:   &proxyURL,
	}

	// 4. Build reverse proxy with custom transport
	targetURL, _ := url.Parse(provider.BaseURL)
	revProxy := httputil.NewSingleHostReverseProxy(targetURL)
	revProxy.Director = func(req *http.Request) {
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.Host = targetURL.Host
		req.URL.Path = "/v1/models"
	}

	transport := buildTransport(provider)
	if transport == nil {
		t.Fatal("expected transport")
	}
	revProxy.Transport = transport

	// 5. Send request through reverse proxy
	req := httptest.NewRequest("GET", "/v1/models", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	revProxy.ServeHTTP(w, req)

	// 6. Verify response
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// 7. Verify request went through the proxy
	if len(proxiedRequests) == 0 {
		t.Fatal("expected request to go through proxy, but proxiedRequests is empty")
	}

	t.Logf("SUCCESS: Request went through proxy. Proxied requests: %d", len(proxiedRequests))
}

// TestSOCKS5ProxyIntegration proves requests go through SOCKS5 proxy
func TestSOCKS5ProxyIntegration(t *testing.T) {
	// 1. Start a mock upstream server
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"models":[]}`))
	}))
	defer upstream.Close()

	// 2. Start a mock SOCKS5 proxy server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	proxyAddr := listener.Addr().String()
	var proxied bool

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handleSOCKS5(conn, upstream.URL, &proxied)
		}
	}()

	// 3. Create provider with SOCKS5 proxy
	provider := &models.Provider{
		Name:        "TestProvider",
		BaseURL:     upstream.URL,
		APIToken:    "test-token",
		EnableProxy: true,
		Socks5Proxy: &proxyAddr,
	}

	// 4. Build transport and make request
	transport := buildTransport(provider)

	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}
	resp, err := client.Get(upstream.URL + "/v1/models")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
	}

	if !proxied {
		t.Fatal("expected request to go through SOCKS5 proxy")
	}

	t.Log("SUCCESS: Request went through SOCKS5 proxy")
}

// handleSOCKS5 is a very basic SOCKS5 server for testing
func handleSOCKS5(clientConn net.Conn, upstreamURL string, proxied *bool) {
	defer clientConn.Close()

	// Read SOCKS5 greeting
	buf := make([]byte, 2)
	clientConn.Read(buf)
	nmethods := int(buf[1])
	methods := make([]byte, nmethods)
	clientConn.Read(methods)

	// Send no-auth method
	clientConn.Write([]byte{0x05, 0x00})

	// Read request
	req := make([]byte, 4)
	clientConn.Read(req)

	if req[1] != 0x01 { // CONNECT
		clientConn.Write([]byte{0x05, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}

	addr := ""
	if req[3] == 0x01 { // IPv4
		ip := make([]byte, 4)
		clientConn.Read(ip)
		port := make([]byte, 2)
		clientConn.Read(port)
		addr = fmt.Sprintf("%d.%d.%d.%d:%d", ip[0], ip[1], ip[2], ip[3], int(port[0])<<8+int(port[1]))
	} else if req[3] == 0x03 { // Domain
		lenBuf := make([]byte, 1)
		clientConn.Read(lenBuf)
		domain := make([]byte, lenBuf[0])
		clientConn.Read(domain)
		port := make([]byte, 2)
		clientConn.Read(port)
		addr = fmt.Sprintf("%s:%d", string(domain), int(port[0])<<8+int(port[1]))
	}
	_ = addr

	// Send success response
	clientConn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})

	// Parse upstream URL to get host:port
	u, _ := url.Parse(upstreamURL)
	host := u.Host
	if !strings.Contains(host, ":") {
		if u.Scheme == "https" {
			host = host + ":443"
		} else {
			host = host + ":80"
		}
	}

	serverConn, err := net.Dial("tcp", host)
	if err != nil {
		return
	}
	defer serverConn.Close()

	*proxied = true

	// Relay traffic
	go io.Copy(serverConn, clientConn)
	io.Copy(clientConn, serverConn)
}
