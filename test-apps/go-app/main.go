package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

type Response struct {
	Message     string            `json:"message"`
	Timestamp   string            `json:"timestamp"`
	Environment string            `json:"environment"`
	Port        string            `json:"port"`
}

type HealthResponse struct {
	Status string `json:"status"`
	Uptime string `json:"uptime"`
	Memory string `json:"memory"`
}

type EnvResponse struct {
	Environment  map[string]string `json:"environment"`
	GoVersion    string            `json:"go_version"`
	Platform     string            `json:"platform"`
}

var startTime = time.Now()

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/env", envHandler)

	fmt.Printf("🚀 Server running on port %s\n", port)
	fmt.Printf("📊 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("🌍 Environment: http://localhost:%s/env\n", port)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development"
	}

	response := Response{
		Message:     "Hello from Flow Test Go App!",
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Environment: env,
		Port:        os.Getenv("PORT"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	response := HealthResponse{
		Status: "healthy",
		Uptime: time.Since(startTime).String(),
		Memory: fmt.Sprintf("%d KB", m.Alloc/1024),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func envHandler(w http.ResponseWriter, r *http.Request) {
	response := EnvResponse{
		Environment: make(map[string]string),
		GoVersion:   runtime.Version(),
		Platform:    runtime.GOOS + "/" + runtime.GOARCH,
	}

	for _, env := range os.Environ() {
		response.Environment[env] = os.Getenv(env)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
