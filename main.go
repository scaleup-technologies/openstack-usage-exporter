package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ScaleUp-Technologies/openstack-usage-exporter/exporters"
	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	baseDSN := os.Getenv("BASE_DSN")

	if baseDSN == "" {
		log.Fatalf("BASE_DSN not set")
	}

	novaDSN := baseDSN + "/nova"
	cinderDSN := baseDSN + "/cinder"

	novaExporter, err := exporters.NewNovaUsageExporter(novaDSN)
	cinderExporter, err := exporters.NewCinderUsageExporter(cinderDSN)

	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}

	prometheus.MustRegister(novaExporter)
	prometheus.MustRegister(cinderExporter)

	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("Starting OpenStack Nova exporter on :8080/metrics")
	log.Fatal(http.ListenAndServe(":9143", nil))
}
