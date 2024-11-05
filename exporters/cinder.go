package exporters

import (
	"database/sql"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

type CinderUsageExporter struct {
	db            *sql.DB
	volumes       *prometheus.Desc
	size          *prometheus.Desc
	snapshots     *prometheus.Desc // Add new metric for snapshots
}

func NewCinderUsageExporter(db *sql.DB) (*CinderUsageExporter, error) {
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
		snapshots: prometheus.NewDesc(
			"openstack_project_snapshots",
			"Total number of snapshots per OpenStack project",
			[]string{"project_id"}, nil,
		),
	}, nil
}

func (e *CinderUsageExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.volumes
	ch <- e.size
	ch <- e.snapshots // Add snapshots to the Describe method
}

func (e *CinderUsageExporter) Collect(ch chan<- prometheus.Metric) {
	e.collectMetrics(ch)
}

func (e *CinderUsageExporter) collectMetrics(ch chan<- prometheus.Metric) {
	// Combined query to get volumes and snapshots data
	rows, err := e.db.Query(`
		SELECT 
			v.project_id, 
			COUNT(DISTINCT v.id) AS total_volumes, 
			SUM(v.size) AS total_size_gb, 
			COALESCE(s.total_snapshots, 0) AS total_snapshots
		FROM 
			volumes v
		LEFT JOIN 
			(SELECT project_id, COUNT(id) AS total_snapshots FROM snapshots GROUP BY project_id) s 
			ON v.project_id = s.project_id
		WHERE 
			v.deleted = 0
		GROUP BY 
			v.project_id
	`)

	if err != nil {
		log.Println("Error querying Cinder and Snapshot databases:", err)
		return
	}
	defer rows.Close()

	// Iterate over the result rows
	for rows.Next() {
		var projectID string
		var totalVolumes float64
		var totalSizeGB float64
		var totalSnapshots float64 // Add variable for snapshots

		// Scan the row data
		if err := rows.Scan(&projectID, &totalVolumes, &totalSizeGB, &totalSnapshots); err != nil {
			log.Println("Error scanning Cinder row:", err)
			continue
		}

		// Export the volumes metric
		ch <- prometheus.MustNewConstMetric(
			e.volumes,
			prometheus.GaugeValue,
			totalVolumes,
			projectID,
		)

		// Export the size metric
		ch <- prometheus.MustNewConstMetric(
			e.size,
			prometheus.GaugeValue,
			totalSizeGB,
			projectID,
		)

		// Export the snapshots metric
		ch <- prometheus.MustNewConstMetric(
			e.snapshots,
			prometheus.GaugeValue,
			totalSnapshots,
			projectID,
		)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error in Cinder result set:", err)
	}
}
