package exporters

import (
	"database/sql"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

type OctaviaUsageExporter struct {
	db          *sql.DB
	loadBalancers *prometheus.Desc
}

func NewOctaviaUsageExporter(db *sql.DB) (*OctaviaUsageExporter, error) {
	return &OctaviaUsageExporter{
		db: db,
		loadBalancers: prometheus.NewDesc(
			"openstack_project_load_balancers",
			"Total number of load balancers per OpenStack project",
			[]string{"project_id"}, nil,
		),
	}, nil
}

func (e *OctaviaUsageExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.loadBalancers
}

func (e *OctaviaUsageExporter) Collect(ch chan<- prometheus.Metric) {
	e.collectMetrics(ch)
}

func (e *OctaviaUsageExporter) collectMetrics(ch chan<- prometheus.Metric) {
	rows, err := e.db.Query(`
		SELECT project_id, COUNT(id) as total_lbs 
		FROM load_balancer 
		WHERE provisioning_status != "DELETED" 
		GROUP BY project_id
	`)

	if err != nil {
		log.Println("Error querying Octavia database:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var projectID string
		var totalLoadBalancers float64
		if err := rows.Scan(&projectID, &totalLoadBalancers); err != nil {
			log.Println("Error scanning Octavia row:", err)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			e.loadBalancers,
			prometheus.GaugeValue,
			totalLoadBalancers,
			projectID,
		)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error in Octavia result set:", err)
	}
}
