// This is an F5 LTM exporter for getting data from the F5 Local Traffic Management Device
// Author: Thomas Elsgaard <thomas.elsgaard@trucecommerce.com>

package main

import (
	"github.com/truecommercedk/f5ltm_exporter/config"
	"github.com/truecommercedk/f5ltm_exporter/prober"
	"maragu.dev/env"
	"net/http"
)

func main() {

	_ = env.Load()

	envConfig := config.Config{
		F5User: env.GetStringOrDefault("F5_USER", ""),
		F5Pass: env.GetStringOrDefault("F5_PASS", ""),
	}

	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		prober.Handler(w, r, envConfig)
	})
	http.ListenAndServe(":9143", nil)

}
