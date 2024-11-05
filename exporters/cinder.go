package exporters

import (
	"database/sql"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

type CinderUsageExporter struct {
	db                  *sql.DB
	volumes             *prometheus.Desc
	volumesSize         *prometheus.Desc
	snapshots           *prometheus.Desc
	backupsSize         *prometheus.Desc
	totalBackups        *prometheus.Desc
}

func NewCinderUsageExporter(db *sql.DB) (*CinderUsageExporter, error) {
	return &CinderUsageExporter{
		db: db,
		volumes: prometheus.NewDesc(
			"openstack_project_volumes",
			"Total number of volumes per OpenStack project",
			[]string{"project_id"}, nil,
		),
		volumesSize: prometheus.NewDesc(
			"openstack_project_volume_size_gb",
			"Total volume size in GB per OpenStack project",
			[]string{"project_id"}, nil,
		),
		snapshots: prometheus.NewDesc(
			"openstack_project_snapshots",
			"Total number of snapshots per OpenStack project",
			[]string{"project_id"}, nil,
		),
		backupsSize: prometheus.NewDesc(
			"openstack_project_backups_size_gb",
			"Total size of backups in GB per OpenStack project",
			[]string{"project_id"}, nil,
		),
		totalBackups: prometheus.NewDesc(
			"openstack_project_backups",
			"Total number of backups per OpenStack project",
			[]string{"project_id"}, nil,
		),
	}, nil
}

func (e *CinderUsageExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.volumes
	ch <- e.volumesSize
	ch <- e.snapshots
	ch <- e.backupsSize
	ch <- e.totalBackups
}

func (e *CinderUsageExporter) Collect(ch chan<- prometheus.Metric) {
	e.collectMetrics(ch)
}

func (e *CinderUsageExporter) collectMetrics(ch chan<- prometheus.Metric) {
	rows, err := e.db.Query(`
		SELECT 
			v.project_id, 
			COUNT(DISTINCT v.id) AS total_volumes, 
			SUM(v.size) AS volumes_size_gb,
			COALESCE(s.total_snapshots, 0) AS total_snapshots,
			COALESCE(b.total_backups, 0) AS total_backups,
			COALESCE(b.total_backups_size_gb, 0) AS total_backups_size_gb
		FROM 
			volumes v
		LEFT JOIN 
			(SELECT project_id, COUNT(id) AS total_snapshots 
			 FROM snapshots 
			 WHERE deleted = 0 
			 GROUP BY project_id) s 
		ON v.project_id = s.project_id
		LEFT JOIN 
			(SELECT project_id, COUNT(id) AS total_backups, SUM(size) AS total_backups_size_gb
			 FROM backups 
			 WHERE deleted = 0 
			 GROUP BY project_id) b
		ON v.project_id = b.project_id
		WHERE 
			v.deleted = 0
		GROUP BY 
			v.project_id
	`)

	if err != nil {
		log.Println("Error querying Cinder, Snapshots, and Backups databases:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var projectID string
		var totalVolumes float64
		var volumesSize float64
		var totalSnapshots float64
		var totalBackups float64
		var totalBackupsSizeGB float64

		if err := rows.Scan(&projectID, &totalVolumes, &volumesSize, &totalSnapshots, &totalBackups, &totalBackupsSizeGB); err != nil {
			log.Println("Error scanning Cinder row:", err)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			e.volumes,
			prometheus.GaugeValue,
			totalVolumes,
			projectID,
		)

		ch <- prometheus.MustNewConstMetric(
			e.volumesSize,
			prometheus.GaugeValue,
			volumesSize,
			projectID,
		)

		ch <- prometheus.MustNewConstMetric(
			e.snapshots,
			prometheus.GaugeValue,
			totalSnapshots,
			projectID,
		)

		ch <- prometheus.MustNewConstMetric(
			e.totalBackups,
			prometheus.GaugeValue,
			totalBackups,
			projectID,
		)

		ch <- prometheus.MustNewConstMetric(
			e.backupsSize,
			prometheus.GaugeValue,
			totalBackupsSizeGB,
			projectID,
		)
	}

	if err := rows.Err(); err != nil {
		log.Println("Error in Cinder result set:", err)
	}
}
