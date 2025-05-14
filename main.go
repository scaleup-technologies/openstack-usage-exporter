package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/scaleup-technologies/openstack-usage-exporter/exporters"
)

type Exporter interface {
	prometheus.Collector
}

func GetBoolEnv(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists || strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return strings.EqualFold(value, "true") || value == "1"
}

func main() {
	baseDSN := os.Getenv("BASE_DSN")

	if baseDSN == "" {
		log.Fatalf("BASE_DSN not set")
	}

	enabledExporters := map[string]bool{
		"cinder":     GetBoolEnv("CINDER_ENABLED", true),
		"nova":       GetBoolEnv("NOVA_ENABLED", true),
		"nova-trait": GetBoolEnv("NOVA_TRAIT_ENABLED", false),
		"neutron":    GetBoolEnv("NEUTRON_ENABLED", true),
		"designate":  GetBoolEnv("DESIGNATE_ENABLED", true),
		"octavia":    GetBoolEnv("OCTAVIA_ENABLED", true),
		"manila":     GetBoolEnv("MANILA_ENABLED", false),
	}

	for name, enabled := range enabledExporters {
		if !enabled {
			continue
		}

		dsn := baseDSN + "/" + strings.Split(name, "-")[0]

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
		case "nova-trait":
			trait, exists := os.LookupEnv("NOVA_TRAIT")
			if !exists {
				log.Fatalf("NOVA_TRAIT not set")
			}
			exporter, err = exporters.NewNovaTraitUsageExporter(db, trait)
		case "neutron":
			externalNetworkId, exists := os.LookupEnv("NEUTRON_ROUTER_EXTERNAL_NETWORK_ID")
			if !exists {
				log.Fatalf("NEUTRON_ROUTER_EXTERNAL_NETWORK_ID not set")
			}
			exporter, err = exporters.NewNeutronUsageExporter(db, externalNetworkId)
		case "designate":
			exporter, err = exporters.NewDesignateUsageExporter(db)
		case "octavia":
			exporter, err = exporters.NewOctaviaUsageExporter(db)
		case "manila":
			exporter, err = exporters.NewManilaUsageExporter(db)
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
