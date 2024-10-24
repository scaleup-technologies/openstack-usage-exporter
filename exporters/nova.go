package exporters

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

type NovaUsageExporter struct {
	db     *sql.DB
	vcpus  *prometheus.Desc
	ram_mb *prometheus.Desc
}

func NewNovaUsageExporter(dsn string) (*NovaUsageExporter, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Nova DB: %v", err)
	}

	return &NovaUsageExporter{
		db: db,
		vcpus: prometheus.NewDesc(
			"openstack_project_vcpus",
			"Total number of vcpus per OpenStack project",
			[]string{"project_id"}, nil,
		),
		ram_mb: prometheus.NewDesc(
			"openstack_project_ram_mb",
			"Total ram usage in MB per OpenStack project",
			[]string{"project_id"}, nil,
		),
	}, nil
}

func (e *NovaUsageExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.vcpus
	ch <- e.ram_mb
}

func (e *NovaUsageExporter) Collect(ch chan<- prometheus.Metric) {
	e.collectMetrics(ch)
}

func (e *NovaUsageExporter) collectMetrics(ch chan<- prometheus.Metric) {
	rows, err := e.db.Query("SELECT project_id, SUM(vcpus) AS total_vcpus, SUM(memory_mb) AS total_ram_mb from instances WHERE deleted = 0 GROUP BY project_id")
	if err != nil {
		log.Println("Error querying Nova database:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var projectID string
		var totalVcpus float64
		var totalRamMB float64
		if err := rows.Scan(&projectID, &totalVcpus, &totalRamMB); err != nil {
			log.Println("Error scanning Nova row:", err)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			e.vcpus,
			prometheus.GaugeValue,
			totalVcpus,
			projectID,
		)

		ch <- prometheus.MustNewConstMetric(
			e.ram_mb,
			prometheus.GaugeValue,
			totalRamMB,
			projectID,
		)

	}
	if err := rows.Err(); err != nil {
		log.Println("Error in Nova result set:", err)
	}
}
