package exporters

import (
	"database/sql"
	"log"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type NovaTraitUsageExporter struct {
	db        *sql.DB
	trait     string
	vcpus     *prometheus.Desc
	instances *prometheus.Desc
}

func NewNovaTraitUsageExporter(db *sql.DB, trait string) (*NovaTraitUsageExporter, error) {
	return &NovaTraitUsageExporter{
		db:    db,
		trait: trait,
		vcpus: prometheus.NewDesc(
			"openstack_project_vcpus_trait__"+strings.ToLower(trait),
			"Total number of vcpus per OpenStack project for instances with image trait "+trait,
			[]string{"project_id"}, nil,
		),
		instances: prometheus.NewDesc(
			"openstack_project_instances_trait__"+strings.ToLower(trait),
			"Total number of instances per OpenStack project with image trait "+trait,
			[]string{"project_id"}, nil,
		),
	}, nil
}

func (e *NovaTraitUsageExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.vcpus
	ch <- e.instances
}

func (e *NovaTraitUsageExporter) Collect(ch chan<- prometheus.Metric) {
	e.collectMetrics(ch)
}

func (e *NovaTraitUsageExporter) collectMetrics(ch chan<- prometheus.Metric) {
	rows, err := e.db.Query("SELECT i.project_id AS project_id, COUNT(i.id) AS total_instances, SUM(vcpus) AS total_vcpus FROM instances i INNER JOIN instance_system_metadata m on i.uuid = m.instance_uuid WHERE i.deleted = 0 AND m.key = ? and m.value = 'required' GROUP BY project_id", "image_trait:"+e.trait)
	if err != nil {
		log.Println("Error querying Nova database:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var projectID string
		var totalVcpus float64
		var totalInstances float64
		if err := rows.Scan(&projectID, &totalInstances, &totalVcpus); err != nil {
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
			e.instances,
			prometheus.GaugeValue,
			totalInstances,
			projectID,
		)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error in Nova result set:", err)
	}
}
