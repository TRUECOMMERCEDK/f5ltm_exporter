// This is an F5 LTM exporter for getting data from the F5 Local Traffic Management Device
// Author: Thomas Elsgaard <thomas.elsgaard@trucecommerce.com>

package main

import (
	"f5ltm_exporter/config"
	"f5ltm_exporter/prober"
	"log/slog"
	"maragu.dev/env"
	"net"
	"net/http"
	"os"
	"strconv"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	_ = env.Load()

	host := env.GetStringOrDefault("HOST", "0.0.0.0")
	port := env.GetIntOrDefault("PORT", 9143)

	address := net.JoinHostPort(host, strconv.Itoa(port))

	envConfig := config.Config{
		F5User: env.GetStringOrDefault("F5_USER", ""),
		F5Pass: env.GetStringOrDefault("F5_PASS", ""),
	}

	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		prober.Handler(w, r, envConfig, logger)
	})

	logger.Info("F5 Exporter Starting", "binding_address", address)

	if err := http.ListenAndServe(address, nil); err != nil {
		logger.Error("F5 Exporter Start failed", "binding_address", address)
	}

}
