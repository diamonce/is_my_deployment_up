package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const version = "3.0.0"

type Service struct {
	ServiceId   string `json:"serviceId"`
	ServiceName string `json:"serviceName"`
	IPAddress   string `json:"ipAddress"`
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
}

type Config struct {
	Servers []Service `json:"servers"`
}

var configLoaded bool

func loadConfig() Config {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "./config.json"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Config file not found at %s, using defaults: %v", path, err)
		configLoaded = true
		return defaultConfig()
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Printf("Invalid config JSON, using defaults: %v", err)
		configLoaded = true
		return defaultConfig()
	}

	configLoaded = true
	return cfg
}

func defaultConfig() Config {
	return Config{
		Servers: []Service{
			{
				ServiceId:   "dc_depops_sp",
				ServiceName: "DevOps та Kubernetes 3.0 Status Page",
				IPAddress:   "34.116.191.131",
				Port:        80,
				Protocol:    "http",
			},
			{
				ServiceId:   "google",
				ServiceName: "Google",
				IPAddress:   "google.com",
				Port:        80,
				Protocol:    "http",
			},
			{
				ServiceId:   "olekluk",
				ServiceName: "OlekLUk",
				IPAddress:   "34.133.93.117",
				Port:        80,
				Protocol:    "http",
			},
		},
	}
}

func checkService(service Service) ([]byte, error) {
	url := fmt.Sprintf("%s://%s:%d", service.Protocol, service.IPAddress, service.Port)
	return checkServiceURL(service, url)
}

func checkServiceURL(service Service, url string) ([]byte, error) {
	client := http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}

	status := "down"
	if err == nil && resp.StatusCode == http.StatusOK {
		status = "up"
	}

	return json.Marshal(map[string]string{
		"service_name": service.ServiceName,
		"status":       status,
	})
}

func newMux(cfg Config) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("./html")))

	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"Version": version}
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
			return
		}
		w.Write(jsonResponse)
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !configLoaded {
			http.Error(w, `{"status":"not ready"}`, http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})

	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		serviceIds := make([]string, len(cfg.Servers))
		for i, service := range cfg.Servers {
			serviceIds[i] = service.ServiceId
		}
		serviceIdsJson, err := json.Marshal(serviceIds)
		if err != nil {
			http.Error(w, "Error marshaling service IDs", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(serviceIdsJson)
	})

	for _, service := range cfg.Servers {
		svc := service // capture loop variable
		mux.HandleFunc(fmt.Sprintf("/status/%s", svc.ServiceId), func(w http.ResponseWriter, r *http.Request) {
			status, err := checkService(svc)
			if err != nil {
				http.Error(w, "Error checking service status", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(status)
		})
	}

	return mux
}

func main() {
	cfg := loadConfig()
	mux := newMux(cfg)

	srv := &http.Server{
		Addr:    ":8088",
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("Server starting on :8088 (version %s)", version)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %s", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %s", err)
	}

	log.Println("Server stopped")
}
