package exporters

import (
	"database/sql"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

type DesignateUsageExporter struct {
	db        *sql.DB
	zones     *prometheus.Desc
}

func NewDesignateUsageExporter(db *sql.DB) (*DesignateUsageExporter, error) {
	return &DesignateUsageExporter{
		db: db,
		zones: prometheus.NewDesc(
			"openstack_project_dns_zones",
			"Total number of dns zones per OpenStack project",
			[]string{"project_id"}, nil,
		),
	}, nil
}

func (e *DesignateUsageExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.zones
}

func (e *DesignateUsageExporter) Collect(ch chan<- prometheus.Metric) {
	e.collectMetrics(ch)
}

func (e *DesignateUsageExporter) collectMetrics(ch chan<- prometheus.Metric) {
	rows, err := e.db.Query(`
		SELECT tenant_id, COUNT(id) AS total_zones
		FROM zones
		WHERE tenant_id != '00000000-0000-0000-0000-000000000000'
		GROUP BY tenant_id
	`)

	if err != nil {
		log.Println("Error querying Designate database:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var projectID string
		var totalZones float64
		if err := rows.Scan(&projectID, &totalZones); err != nil {
			log.Println("Error scanning Designate row:", err)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			e.zones,
			prometheus.GaugeValue,
			totalZones,
			projectID,
		)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error in Designate result set:", err)
	}
}
