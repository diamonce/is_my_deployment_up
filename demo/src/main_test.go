package main

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestCheckServiceUp(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	svc := Service{
		ServiceId:   "test",
		ServiceName: "Test Service",
		IPAddress:   ts.Listener.Addr().String(),
		Port:        0, // port is already in the address
		Protocol:    "http",
	}

	// Override URL format: httptest includes port in address
	result, err := checkServiceURL(svc, ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]string
	if err := json.Unmarshal(result, &data); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if data["status"] != "up" {
		t.Errorf("expected status 'up', got %q", data["status"])
	}
	if data["service_name"] != "Test Service" {
		t.Errorf("expected service_name 'Test Service', got %q", data["service_name"])
	}
}

func TestCheckServiceDown(t *testing.T) {
	svc := Service{
		ServiceId:   "down",
		ServiceName: "Down Service",
		IPAddress:   "127.0.0.1",
		Port:        1, // nothing listening
		Protocol:    "http",
	}

	result, err := checkService(svc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]string
	if err := json.Unmarshal(result, &data); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if data["status"] != "down" {
		t.Errorf("expected status 'down', got %q", data["status"])
	}
}

func TestVersionEndpoint(t *testing.T) {
	cfg := Config{Servers: []Service{}}
	mux := newMux(cfg)

	req := httptest.NewRequest("GET", "/version", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var data map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &data); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if data["Version"] != "3.0.0" {
		t.Errorf("expected version '3.0.0', got %q", data["Version"])
	}
}

func TestStatusEndpoint(t *testing.T) {
	cfg := Config{
		Servers: []Service{
			{ServiceId: "svc1", ServiceName: "Service 1"},
			{ServiceId: "svc2", ServiceName: "Service 2"},
		},
	}
	mux := newMux(cfg)

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var ids []string
	if err := json.Unmarshal(w.Body.Bytes(), &ids); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(ids) != 2 {
		t.Fatalf("expected 2 service IDs, got %d", len(ids))
	}
	if ids[0] != "svc1" || ids[1] != "svc2" {
		t.Errorf("unexpected IDs: %v", ids)
	}
}

func TestStatusServiceEndpoint(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	host, portStr, _ := net.SplitHostPort(ts.Listener.Addr().String())
	port, _ := strconv.Atoi(portStr)

	cfg := Config{
		Servers: []Service{
			{
				ServiceId:   "test_svc",
				ServiceName: "Test",
				IPAddress:   host,
				Port:        port,
				Protocol:    "http",
			},
		},
	}
	mux := newMux(cfg)

	req := httptest.NewRequest("GET", "/status/test_svc", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var data map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &data); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if data["service_name"] != "Test" {
		t.Errorf("expected service_name 'Test', got %q", data["service_name"])
	}
}

func TestHealthzEndpoint(t *testing.T) {
	cfg := Config{Servers: []Service{}}
	mux := newMux(cfg)

	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestReadyzEndpoint(t *testing.T) {
	cfg := Config{Servers: []Service{}}

	// configLoaded is set by loadConfig; set it manually for test
	configLoaded = true
	mux := newMux(cfg)

	req := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestReadyzNotReady(t *testing.T) {
	cfg := Config{Servers: []Service{}}

	configLoaded = false
	mux := newMux(cfg)

	req := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}

	// Reset for other tests
	configLoaded = true
}

func TestLoadConfigValidFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	data := `{"servers":[{"serviceId":"test","serviceName":"Test","ipAddress":"1.2.3.4","port":80,"protocol":"http"}]}`
	if err := os.WriteFile(cfgPath, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CONFIG_PATH", cfgPath)
	cfg := loadConfig()

	if len(cfg.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(cfg.Servers))
	}
	if cfg.Servers[0].ServiceId != "test" {
		t.Errorf("expected serviceId 'test', got %q", cfg.Servers[0].ServiceId)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	t.Setenv("CONFIG_PATH", "/nonexistent/config.json")
	cfg := loadConfig()

	if len(cfg.Servers) != 3 {
		t.Errorf("expected 3 default servers, got %d", len(cfg.Servers))
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	if err := os.WriteFile(cfgPath, []byte("not json"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CONFIG_PATH", cfgPath)
	cfg := loadConfig()

	if len(cfg.Servers) != 3 {
		t.Errorf("expected 3 default servers, got %d", len(cfg.Servers))
	}
}
