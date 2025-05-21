package exporters

import (
	"database/sql"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

type NeutronUsageExporter struct {
	db                *sql.DB
	externalNetworkId string
	floatingIPs       *prometheus.Desc
	routers           *prometheus.Desc
}

func NewNeutronUsageExporter(db *sql.DB, externalNetworkId string) (*NeutronUsageExporter, error) {
	return &NeutronUsageExporter{
		db:                db,
		externalNetworkId: externalNetworkId,
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
	floatingIPsCounts := make(map[string]float64)
	rows, err := e.db.Query("SELECT project_id, COUNT(id) AS total_fips FROM floatingips GROUP BY project_id")
	if err != nil {
		log.Println("Error querying floating IP counts:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var projectID string
		var totalFloatingIPs float64
		if err := rows.Scan(&projectID, &totalFloatingIPs); err != nil {
			log.Println("Error scanning floating IP row:", err)
			continue
		}
		floatingIPsCounts[projectID] = totalFloatingIPs
	}

	if err := rows.Err(); err != nil {
		log.Println("Error in floating IPs result set:", err)
		return
	}

	routerCounts := make(map[string]float64)
	rows, err = e.db.Query("SELECT r.project_id, COUNT(r.id) AS total_routers FROM routers r INNER JOIN ports p ON r.gw_port_id = p.id WHERE p.network_id = ? GROUP BY r.project_id", e.externalNetworkId)
	if err != nil {
		log.Println("Error querying router counts:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var projectID string
		var totalRouters float64
		if err := rows.Scan(&projectID, &totalRouters); err != nil {
			log.Println("Error scanning router row:", err)
			continue
		}
		routerCounts[projectID] = totalRouters
	}

	if err := rows.Err(); err != nil {
		log.Println("Error in routers result set:", err)
		return
	}

	projectIDs := make(map[string]bool)
	for projectID := range floatingIPsCounts {
		projectIDs[projectID] = true
	}
	for projectID := range routerCounts {
		projectIDs[projectID] = true
	}

	for projectID := range projectIDs {
		totalFloatingIPs := floatingIPsCounts[projectID]
		totalRouters := routerCounts[projectID]

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
}
