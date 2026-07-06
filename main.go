package main

import (
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"time"
)

var users = []string{"alice", "bob", "charlie", "diana", "eve", "frank"}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("homepage visited", "method", r.Method, "path", r.URL.Path, "remote", r.RemoteAddr)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Placeholder2 Service is running"))
	})

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("health check", "method", r.Method, "remote", r.RemoteAddr)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Placeholder2 Service is Healthy"))
	})

	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		user := users[rand.IntN(len(users))]
		delay := time.Duration(20+rand.IntN(100)) * time.Millisecond
		time.Sleep(delay)
		if rand.IntN(100) < 20 {
			slog.Warn("login failed", "user", user, "reason", "invalid credentials", "delay_ms", delay.Milliseconds())
			http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
			return
		}
		token := fmt.Sprintf("tok_%d", rand.IntN(999999))
		slog.Info("login successful", "user", user, "token", token, "delay_ms", delay.Milliseconds())
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"user":"%s","token":"%s"}`, user, token)
	})

	mux.HandleFunc("POST /api/logout", func(w http.ResponseWriter, r *http.Request) {
		user := users[rand.IntN(len(users))]
		slog.Info("user logged out", "user", user)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"user":"%s","status":"logged_out"}`, user)
	})

	mux.HandleFunc("GET /api/profile", func(w http.ResponseWriter, r *http.Request) {
		user := users[rand.IntN(len(users))]
		delay := time.Duration(10+rand.IntN(80)) * time.Millisecond
		time.Sleep(delay)
		if rand.IntN(100) < 5 {
			slog.Error("profile fetch failed", "user", user, "error", "database connection timeout")
			http.Error(w, `{"error":"database connection timeout"}`, http.StatusInternalServerError)
			return
		}
		slog.Info("profile fetched", "user", user, "delay_ms", delay.Milliseconds())
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"user":"%s","email":"%s@example.com"}`, user, user)
	})

	mux.HandleFunc("GET /api/version", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("version requested", "remote", r.RemoteAddr)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"service":"placeholder2-service","version":"0.0.2"}`)
	})

	mux.HandleFunc("POST /api/register", func(w http.ResponseWriter, r *http.Request) {
		delay := time.Duration(100+rand.IntN(400)) * time.Millisecond
		time.Sleep(delay)
		userID := rand.IntN(90000) + 10000
		user := fmt.Sprintf("user_%d", userID)
		if rand.IntN(100) < 10 {
			slog.Error("registration failed", "user", user, "error", "email already exists", "delay_ms", delay.Milliseconds())
			http.Error(w, `{"error":"email already exists"}`, http.StatusConflict)
			return
		}
		slog.Info("user registered", "user_id", userID, "user", user, "delay_ms", delay.Milliseconds())
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"user_id":%d,"user":"%s","status":"registered"}`, userID, user)
	})

	handler := loggingMiddleware(mux)

	// Start background traffic simulator if enabled
	if os.Getenv("ENABLE_TRAFFIC_SIMULATOR") == "true" {
		go simulateTraffic()
		slog.Info("traffic simulator enabled")
	} else {
		slog.Info("traffic simulator disabled, set ENABLE_TRAFFIC_SIMULATOR=true to enable")
	}

	slog.Info("Placeholder2 Service starting", "port", 8081)
	if err := http.ListenAndServe(":8081", handler); err != nil {
		slog.Error("Error starting service", "error", err)
	}
}

func simulateTraffic() {
	time.Sleep(3 * time.Second) // wait for server to start
	client := &http.Client{Timeout: 5 * time.Second}
	base := "http://localhost:8081"

	endpoints := []struct {
		method string
		path   string
		weight int
	}{
		{"GET", "/", 5},
		{"GET", "/health", 15},
		{"POST", "/api/login", 30},
		{"POST", "/api/logout", 10},
		{"GET", "/api/profile", 25},
		{"POST", "/api/register", 15},
	}

	totalWeight := 0
	for _, e := range endpoints {
		totalWeight += e.weight
	}

	slog.Info("traffic simulator started", "endpoints", len(endpoints))

	for {
		pick := rand.IntN(totalWeight)
		cumulative := 0
		for _, e := range endpoints {
			cumulative += e.weight
			if pick < cumulative {
				req, _ := http.NewRequest(e.method, base+e.path, nil)
				req.Header.Set("User-Agent", "TrafficSimulator/1.0")
				resp, err := client.Do(req)
				if err != nil {
					slog.Warn("simulator request failed", "path", e.path, "error", err.Error())
				} else {
					resp.Body.Close()
				}
				break
			}
		}
		// Random interval between 500ms and 3s
		time.Sleep(time.Duration(500+rand.IntN(2500)) * time.Millisecond)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		slog.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration", time.Since(start).String(),
			"remote", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
