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

// NewDesignateUsageExporter initializes the DesignateUsageExporter with a given database connection.
func NewDesignateUsageExporter(db *sql.DB) (*DesignateUsageExporter, error) {
	return &DesignateUsageExporter{
		db: db,
		zones: prometheus.NewDesc(
			"openstack_project_zones",
			"Total number of zones per OpenStack project",
			[]string{"project_id"}, nil,
		),
	}, nil
}

// Describe sends the descriptor for the metric(s) to the Prometheus channel.
func (e *DesignateUsageExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.zones
}

// Collect fetches the metrics from the database and sends them to the Prometheus channel.
func (e *DesignateUsageExporter) Collect(ch chan<- prometheus.Metric) {
	e.collectMetrics(ch)
}

// collectMetrics executes the query to get the number of zones used by each project
// and sends the metrics to the Prometheus channel.
func (e *DesignateUsageExporter) collectMetrics(ch chan<- prometheus.Metric) {
	// Query to get the total number of zones for each project, excluding the empty tenant_id
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

	// Iterate over the rows returned by the query
	for rows.Next() {
		var projectID string
		var totalZones float64
		if err := rows.Scan(&projectID, &totalZones); err != nil {
			log.Println("Error scanning Designate row:", err)
			continue
		}

		// Send the total zones as a Prometheus metric
		ch <- prometheus.MustNewConstMetric(
			e.zones,
			prometheus.GaugeValue,
			totalZones,
			projectID,
		)
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		log.Println("Error in Designate result set:", err)
	}
}
