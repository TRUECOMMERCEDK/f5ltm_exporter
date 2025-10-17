package main

import (
	"f5ltm_exporter/config"
	"f5ltm_exporter/prober"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
)

// CLI flags
var (
	flagHost   = flag.String("host", "0.0.0.0", "Host address to bind the exporter (e.g., 0.0.0.0)")
	flagPort   = flag.Int("port", 9143, "Port number to bind the exporter (e.g., 9143)")
	flagF5User = flag.String("f5-user", "", "Username for F5 LTM authentication (required)")
	flagF5Pass = flag.String("f5-pass", "", "Password for F5 LTM authentication (required)")
)

func main() {
	flag.Parse()

	logger := createLogger()

	// Validate required flags
	if *flagF5User == "" || *flagF5Pass == "" {
		fmt.Fprintln(os.Stderr, "Error: --f5-user and --f5-pass are required")
		flag.Usage()
		os.Exit(1)
	}

	cfg := config.Config{
		F5User: *flagF5User,
		F5Pass: *flagF5Pass,
	}

	address := net.JoinHostPort(*flagHost, strconv.Itoa(*flagPort))

	startServer(address, cfg, logger)
}

func createLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func startServer(address string, cfg config.Config, logger *slog.Logger) {
	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		prober.Handler(w, r, cfg, logger)
	})

	logger.Info("F5 Exporter starting", slog.String("bind_address", address))

	if err := http.ListenAndServe(address, nil); err != nil {
		logger.Error("Failed to start HTTP server", slog.String("bind_address", address), slog.Any("error", err))
		os.Exit(1)
	}
}
