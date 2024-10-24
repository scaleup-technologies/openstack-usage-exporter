package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type UsageExporter struct {
	novaDB   *sql.DB
	cinderDB *sql.DB
	vcpus    *prometheus.Desc
	volumes  *prometheus.Desc
	size     *prometheus.Desc
}

func NewUsageExporter(novaDSN, cinderDSN string) (*UsageExporter, error) {
	novaDB, err := sql.Open("mysql", novaDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Nova DB: %v", err)
	}

	cinderDB, err := sql.Open("mysql", cinderDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Cinder DB: %v", err)
	}

	return &UsageExporter{
		novaDB:   novaDB,
		cinderDB: cinderDB,
		vcpus: prometheus.NewDesc(
			"openstack_project_vcpus",
			"Total vCPUs used per OpenStack project",
			[]string{"project_id"}, nil,
		),
		volumes: prometheus.NewDesc(
			"openstack_project_volumes",
			"Total number of volumes per OpenStack project",
			[]string{"project_id"}, nil,
		),
		size: prometheus.NewDesc(
			"openstack_project_volume_size_gb",
			"Total volume size in GB per OpenStack project",
			[]string{"project_id"}, nil,
		),
	}, nil
}

func (e *UsageExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.vcpus
	ch <- e.volumes
	ch <- e.size
}

func (e *UsageExporter) Collect(ch chan<- prometheus.Metric) {
	e.collectNovaMetrics(ch)
	e.collectCinderMetrics(ch)
}

func (e *UsageExporter) collectNovaMetrics(ch chan<- prometheus.Metric) {
	rows, err := e.novaDB.Query("SELECT project_id, SUM(vcpus) as total_vcpus FROM instances WHERE deleted = 0 GROUP BY project_id")
	if err != nil {
		log.Println("Error querying Nova database:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var projectID string
		var totalVCPUs float64
		if err := rows.Scan(&projectID, &totalVCPUs); err != nil {
			log.Println("Error scanning Nova row:", err)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			e.vcpus,
			prometheus.GaugeValue,
			totalVCPUs,
			projectID,
		)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error in Nova result set:", err)
	}
}

func (e *UsageExporter) collectCinderMetrics(ch chan<- prometheus.Metric) {
	rows, err := e.cinderDB.Query("SELECT project_id, COUNT(id) as total_volumes, SUM(size) as total_size_gb FROM volumes WHERE deleted = 0 GROUP BY project_id")
	if err != nil {
		log.Println("Error querying Cinder database:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var projectID string
		var totalVolumes float64
		var totalSizeGB float64
		if err := rows.Scan(&projectID, &totalVolumes, &totalSizeGB); err != nil {
			log.Println("Error scanning Cinder row:", err)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			e.volumes,
			prometheus.GaugeValue,
			totalVolumes,
			projectID,
		)

		ch <- prometheus.MustNewConstMetric(
			e.size,
			prometheus.GaugeValue,
			totalSizeGB,
			projectID,
		)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error in Cinder result set:", err)
	}
}

func main() {
	baseDSN := os.Getenv("BASE_DSN")

	if baseDSN == "" {
		log.Fatalf("BASE_DSN not set")
	}

	novaDSN := baseDSN + "/nova"
	cinderDSN := baseDSN + "/cinder"

	exporter, err := NewUsageExporter(novaDSN, cinderDSN)

	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}

	prometheus.MustRegister(exporter)

	http.Handle("/metrics", promhttp.Handler())
	fmt.Println("Starting OpenStack Nova exporter on :8080/metrics")
	log.Fatal(http.ListenAndServe(":9143", nil))
}
