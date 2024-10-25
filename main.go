package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ScaleUp-Technologies/openstack-usage-exporter/exporters"
	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Exporter interface {
	prometheus.Collector
}

func main() {
	baseDSN := os.Getenv("BASE_DSN")

	if baseDSN == "" {
		log.Fatalf("BASE_DSN not set")
	}

	enabledExporters := map[string]bool{
		"cinder": true,
		"nova":   false,
	}

	for name, enabled := range enabledExporters {
		if !enabled {
			continue
		}

		dsn := baseDSN + "/" + name

		db, err := sql.Open("mysql", dsn)
		if err != nil {
			log.Fatalf("failed to connect to database: %s", err)
		}

		var exporter prometheus.Collector

		switch name {
		case "cinder":
			exporter, err = exporters.NewCinderUsageExporter(db)
		case "nova":
			exporter, err = exporters.NewNovaUsageExporter(db)
		default:
			log.Fatalf("unknown exporter type: %s", name)
		}

		if err != nil {
			log.Fatalf("failed to initialize exporter: %s", err)
		}

		prometheus.MustRegister(exporter)

	}

	HTTP_BIND := ":9143"
	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("Starting OpenStack Usage exporter on /metrics")
	log.Fatal(http.ListenAndServe(HTTP_BIND, nil))
}
