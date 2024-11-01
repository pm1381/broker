package router

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func AllRoutes()  {
	http.Handle("/metrics", promhttp.Handler())
	log.Println("/metrcis for prometheus is accessible from port :80")
}