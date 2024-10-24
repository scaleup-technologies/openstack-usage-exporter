package exporters

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

type CinderUsageExporter struct {
	db      *sql.DB
	volumes *prometheus.Desc
	size    *prometheus.Desc
}

func NewCinderUsageExporter(dsn string) (*CinderUsageExporter, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Cinder DB: %v", err)
	}

	return &CinderUsageExporter{
		db: db,
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

func (e *CinderUsageExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.volumes
	ch <- e.size
}

func (e *CinderUsageExporter) Collect(ch chan<- prometheus.Metric) {
	e.collectMetrics(ch)
}

func (e *CinderUsageExporter) collectMetrics(ch chan<- prometheus.Metric) {
	rows, err := e.db.Query("SELECT project_id, COUNT(id) as total_volumes, SUM(size) as total_size_gb FROM volumes WHERE deleted = 0 GROUP BY project_id")
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
