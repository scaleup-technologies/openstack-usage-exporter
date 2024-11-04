package exporters

import (
	"database/sql"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

type NeutronUsageExporter struct {
	db     *sql.DB
	floatingIPs  *prometheus.Desc
}

func NewNeutronUsageExporter(db *sql.DB) (*NeutronUsageExporter, error) {
	return &NeutronUsageExporter{
		db: db,
		floatingIPs: prometheus.NewDesc(
			"openstack_project_floating_ips",
			"Total number of floating IPs per OpenStack project",
			[]string{"project_id"}, nil,
		),
	}, nil
}

func (e *NeutronUsageExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.floatingIPs
}

func (e *NeutronUsageExporter) Collect(ch chan<- prometheus.Metric) {
	e.collectMetrics(ch)
}

func (e *NeutronUsageExporter) collectMetrics(ch chan<- prometheus.Metric) {
	rows, err := e.db.Query("select project_id, COUNT(id) as total_fips from floatingips GROUP BY project_id")
	if err != nil {
		log.Println("Error querying Neutron database:", err)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var projectID string
		var totalFloatingIPs float64
		if err := rows.Scan(&projectID, &totalFloatingIPs); err != nil {
			log.Println("Error scanning Neutron row:", err)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			e.floatingIPs,
			prometheus.GaugeValue,
			totalFloatingIPs,
			projectID,
		)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error in Neutron result set:", err)
	}
}
