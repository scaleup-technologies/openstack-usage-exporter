package exporters

import (
	"database/sql"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

type ManilaUsageExporter struct {
	db         *sql.DB
	sharesSize *prometheus.Desc
}

func NewManilaUsageExporter(db *sql.DB) (*ManilaUsageExporter, error) {
	return &ManilaUsageExporter{
		db: db,
		sharesSize: prometheus.NewDesc(
			"openstack_project_shares_size_gb",
			"Total share size in GB per OpenStack project",
			[]string{"project_id"}, nil,
		),
	}, nil
}

func (e *ManilaUsageExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.sharesSize
}

func (e *ManilaUsageExporter) Collect(ch chan<- prometheus.Metric) {
	e.collectMetrics(ch)
}

func (e *ManilaUsageExporter) collectMetrics(ch chan<- prometheus.Metric) {
	rows, err := e.db.Query("SELECT project_id, SUM(size) AS shares_size FROM shares WHERE deleted='False' GROUP BY project_id")
	if err != nil {
		log.Println("Error querying Manila database:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var projectID string
		var sharesSize float64
		if err := rows.Scan(&projectID, &sharesSize); err != nil {
			log.Println("Error scanning Manila row:", err)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			e.sharesSize,
			prometheus.GaugeValue,
			sharesSize,
			projectID,
		)

	}
	if err := rows.Err(); err != nil {
		log.Println("Error in Manila result set:", err)
	}
}
