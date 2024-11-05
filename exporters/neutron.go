package exporters

import (
	"database/sql"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

type NeutronUsageExporter struct {
	db          *sql.DB
	floatingIPs *prometheus.Desc
	routers		*prometheus.Desc
}

func NewNeutronUsageExporter(db *sql.DB) (*NeutronUsageExporter, error) {
	return &NeutronUsageExporter{
		db: db,
		floatingIPs: prometheus.NewDesc(
			"openstack_project_floating_ips",
			"Total number of floating IPs per OpenStack project",
			[]string{"project_id"}, nil,
		),
		routers: prometheus.NewDesc(
			"openstack_project_routers",
			"Total number of routers per OpenStack project",
			[]string{"project_id"}, nil,
		),
	}, nil
}

func (e *NeutronUsageExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.floatingIPs
	ch <- e.routers
}

func (e *NeutronUsageExporter) Collect(ch chan<- prometheus.Metric) {
	e.collectMetrics(ch)
}

func (e *NeutronUsageExporter) collectMetrics(ch chan<- prometheus.Metric) {
	rows, err := e.db.Query("SELECT f.project_id, COUNT(f.id) AS total_fips, COALESCE(r.total_routers, 0) AS total_routers FROM floatingips f LEFT JOIN (SELECT project_id, COUNT(id) AS total_routers FROM routers GROUP BY project_id) r ON f.project_id = r.project_id GROUP BY f.project_id")

	if err != nil {
		log.Println("Error querying Neutron database:", err)
		return
	}

	defer rows.Close()

	for rows.Next() {
		var projectID string
		var totalFloatingIPs float64
		var totalRouters float64
		if err := rows.Scan(&projectID, &totalFloatingIPs, &totalRouters); err != nil {
			log.Println("Error scanning Neutron row:", err)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			e.floatingIPs,
			prometheus.GaugeValue,
			totalFloatingIPs,
			projectID,
		)

		ch <- prometheus.MustNewConstMetric(
			e.routers,
			prometheus.GaugeValue,
			totalRouters,
			projectID,
		)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error in Neutron result set:", err)
	}
}
