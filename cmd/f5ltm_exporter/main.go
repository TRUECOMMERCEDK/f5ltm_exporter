package main

import (
	"context"
	"encoding/json"
	"errors"
	"f5ltm_exporter/internal/f5api"
	"f5ltm_exporter/internal/logging"
	"f5ltm_exporter/prober"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	flagHost          = flag.String("host", "127.0.0.1", "Host address to bind the exporter")
	flagPort          = flag.Int("port", 9143, "Port number to bind the exporter")
	flagF5User        = flag.String("f5-user", "", "Username for F5 LTM authentication (required)")
	flagF5Pass        = flag.String("f5-pass", "", "Password for F5 LTM authentication (required)")
	flagTLSSkipVerify = flag.Bool("tls-skip-verify", false, "Skip TLS certificate verification (use only for testing)")
	flagLogFormat     = flag.String("log-format", "json", "Log format: json or text")
	flagLogLevel      = flag.String("log-level", "info", "Log level: debug, info, warn, error")
)

var release = "dev"

func main() {
	flag.Parse()
	logger := logging.NewWithOptions(*flagLogFormat, *flagLogLevel)
	slog.SetDefault(logger)

	logger.Info("Starting F5 LTM Exporter",
		slog.String("version", release),
		slog.String("log_format", *flagLogFormat),
		slog.String("log_level", *flagLogLevel))

	if *flagF5User == "" || *flagF5Pass == "" {
		_, err := fmt.Fprintln(os.Stderr, "Error: --f5-user and --f5-pass are required")
		if err != nil {
			return
		}
		flag.Usage()
		os.Exit(1)
	}

	address := net.JoinHostPort(*flagHost, strconv.Itoa(*flagPort))
	startServer(address, logger)
}

func startServer(address string, logger *slog.Logger) {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /probe", func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("target")
		if target == "" {
			http.Error(w, "Target parameter is missing", http.StatusBadRequest)
			logger.Error("Missing target parameter")
			return
		}

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

	mux.Handle("GET /metrics", promhttp.Handler())

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"version": release,
		})
		if err != nil {
			return
		}
	})

	server := &http.Server{
		Addr:         address,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// --- Graceful shutdown handling ---
	idleConnsClosed := make(chan struct{})
	go func() {
		// Wait for interrupt or SIGTERM
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		logger.Info("Shutdown signal received, waiting for ongoing scrapes to complete...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Graceful shutdown failed", slog.Any("error", err))
		} else {
			logger.Info("Exporter shut down cleanly")
		}

		close(idleConnsClosed)
	}()

	logger.Info("HTTP server listening",
		slog.String("bind_address", address))

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("HTTP server error", slog.Any("error", err))
		os.Exit(1)
	}

	<-idleConnsClosed
}
