package main

import (
	"f5ltm_exporter/internal/f5api"
	"f5ltm_exporter/internal/logging"
	"f5ltm_exporter/prober"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	flagHost          = flag.String("host", "127.0.0.1", "Host address to bind the exporter (e.g., 0.0.0.0)")
	flagPort          = flag.Int("port", 9143, "Port number to bind the exporter (e.g., 9143)")
	flagF5User        = flag.String("f5-user", "", "Username for F5 LTM authentication (required)")
	flagF5Pass        = flag.String("f5-pass", "", "Password for F5 LTM authentication (required)")
	flagTLSSkipVerify = flag.Bool("tls-skip-verify", false, "Skip TLS certificate verification (use only for testing)")
	flagLogFormat     = flag.String("log-format", "json", "Log format: json or text")
	flagLogLevel      = flag.String("log-level", "info", "Log level: debug, info, warn, error")
)

func main() {
	flag.Parse()
	logger := logging.NewWithOptions(*flagLogFormat, *flagLogLevel)
	slog.SetDefault(logger)

	if *flagF5User == "" || *flagF5Pass == "" {
		fmt.Fprintln(os.Stderr, "Error: --f5-user and --f5-pass are required")
		flag.Usage()
		os.Exit(1)
	}

	address := net.JoinHostPort(*flagHost, strconv.Itoa(*flagPort))
	startServer(address, logger)
}

func startServer(address string, logger *slog.Logger) {
	mux := http.NewServeMux()

	mux.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "Target parameter is missing", http.StatusBadRequest)
			logger.Error("Missing target parameter")
			return
		}

		// Create a fresh F5 API client for this target/scrape
		f5 := &f5api.Model{
			User:            *flagF5User,
			Pass:            *flagF5Pass,
			Host:            target,
			Port:            "443",
			MaxRetries:      3,
			RetryDelay:      500 * time.Millisecond,
			InsecureSkipTLS: *flagTLSSkipVerify,
			Logger:          logger,
		}

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
