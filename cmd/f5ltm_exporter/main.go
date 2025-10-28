package main

import (
	"f5ltm_exporter/internal/f5api"
	"f5ltm_exporter/prober"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	flagHost   = flag.String("host", "0.0.0.0", "Host address to bind the exporter (e.g., 0.0.0.0)")
	flagPort   = flag.Int("port", 9143, "Port number to bind the exporter (e.g., 9143)")
	flagF5User = flag.String("f5-user", "", "Username for F5 LTM authentication (required)")
	flagF5Pass = flag.String("f5-pass", "", "Password for F5 LTM authentication (required)")
)

func main() {
	flag.Parse()
	logger := createLogger()

	if *flagF5User == "" || *flagF5Pass == "" {
		fmt.Fprintln(os.Stderr, "Error: --f5-user and --f5-pass are required")
		flag.Usage()
		os.Exit(1)
	}

	cache := &targetCache{
		user:   *flagF5User,
		pass:   *flagF5Pass,
		models: make(map[string]*f5api.Model),
	}

	address := net.JoinHostPort(*flagHost, strconv.Itoa(*flagPort))
	startServer(address, cache, logger)
}

func createLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func startServer(address string, cache *targetCache, logger *slog.Logger) {
	mux := http.NewServeMux()

	mux.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "Target parameter is missing", http.StatusBadRequest)
			logger.Error("Missing target parameter")
			return
		}

		f5 := cache.getOrCreate(target)
		prober.Handler(w, r, f5, logger)
	})

	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	logger.Info("F5 Exporter starting", slog.String("bind_address", address))

	if err := http.ListenAndServe(address, mux); err != nil {
		logger.Error("Failed to start HTTP server", slog.Any("error", err))
		os.Exit(1)
	}
}

// -----------------------------
// Target cache management
// -----------------------------

type targetCache struct {
	mu     sync.Mutex
	models map[string]*f5api.Model
	user   string
	pass   string
}

func (c *targetCache) getOrCreate(host string) *f5api.Model {
	c.mu.Lock()
	defer c.mu.Unlock()

	if m, ok := c.models[host]; ok {
		return m
	}

	m := &f5api.Model{
		User:       c.user,
		Pass:       c.pass,
		Host:       host,
		Port:       "443",
		MaxRetries: 3,
		RetryDelay: 500 * time.Millisecond,
	}

	c.models[host] = m
	return m
}
