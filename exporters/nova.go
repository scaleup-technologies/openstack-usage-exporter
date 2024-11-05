package exporters

import (
	"database/sql"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

type NovaUsageExporter struct {
	db     				*sql.DB
	vcpus  				*prometheus.Desc
	ram_mb 				*prometheus.Desc
	local_storage_gb	*prometheus.Desc
}

func NewNovaUsageExporter(db *sql.DB) (*NovaUsageExporter, error) {
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
		local_storage_gb: prometheus.NewDesc(
			"openstack_prject_local_storage_gb",
			"Total local storage in GB per Openstack project",
			[]string{"project_id"}, nil,
		),
	}, nil
}

func (e *NovaUsageExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.vcpus
	ch <- e.ram_mb
	ch <- e.local_storage_gb
}

func (e *NovaUsageExporter) Collect(ch chan<- prometheus.Metric) {
	e.collectMetrics(ch)
}

func (e *NovaUsageExporter) collectMetrics(ch chan<- prometheus.Metric) {
	rows, err := e.db.Query("SELECT project_id, SUM(vcpus) AS total_vcpus, SUM(memory_mb) AS total_ram_mb, SUM(root_gb) as total_root_gb FROM instances WHERE deleted = 0 GROUP BY project_id")
	if err != nil {
		log.Println("Error querying Nova database:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var projectID string
		var totalVcpus float64
		var totalRamMB float64
		var totalLocalStorageGB float64
		if err := rows.Scan(&projectID, &totalVcpus, &totalRamMB, &totalLocalStorageGB); err != nil {
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

		ch <- prometheus.MustNewConstMetric(
			e.local_storage_gb,
			prometheus.GaugeValue,
			totalLocalStorageGB,
			projectID,
		)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error in Nova result set:", err)
	}
}
