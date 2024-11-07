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
	rows, err := e.db.Query("SELECT project_id, COUNT(id) AS total_volumes, SUM(size) AS volumes_size_gb FROM volumes WHERE deleted = 0 GROUP BY project_id")
	if err != nil {
		log.Println("Error querying Volumes:", err)
		return
	}
	defer rows.Close()

	volumesData := make(map[string]struct {
		totalVolumes    float64
		volumesSizeGB   float64
	})

	for rows.Next() {
		var projectID string
		var totalVolumes, volumesSize float64

		if err := rows.Scan(&projectID, &totalVolumes, &volumesSize); err != nil {
			log.Println("Error scanning Volumes row:", err)
			continue
		}

		volumesData[projectID] = struct {
			totalVolumes  float64
			volumesSizeGB float64
		}{
			totalVolumes:  totalVolumes,
			volumesSizeGB: volumesSize,
		}
	}

	if err := rows.Err(); err != nil {
		log.Println("Error in Volumes result set:", err)
	}

	rows, err = e.db.Query("SELECT project_id, COUNT(id) AS total_snapshots FROM snapshots WHERE deleted = 0 GROUP BY project_id")
	if err != nil {
		log.Println("Error querying Snapshots:", err)
		return
	}
	defer rows.Close()

	snapshotsData := make(map[string]float64)

	for rows.Next() {
		var projectID string
		var totalSnapshots float64

		if err := rows.Scan(&projectID, &totalSnapshots); err != nil {
			log.Println("Error scanning Snapshots row:", err)
			continue
		}

		snapshotsData[projectID] = totalSnapshots
	}

	if err := rows.Err(); err != nil {
		log.Println("Error in Snapshots result set:", err)
	}

	rows, err = e.db.Query("SELECT project_id, COUNT(id) AS total_backups, SUM(size) AS total_backups_size_gb FROM backups WHERE deleted = 0 GROUP BY project_id")
	if err != nil {
		log.Println("Error querying Backups:", err)
		return
	}
	defer rows.Close()

	backupsData := make(map[string]struct {
		totalBackups      float64
		totalBackupsSizeGB float64
	})

	for rows.Next() {
		var projectID string
		var totalBackups, totalBackupsSizeGB float64

		if err := rows.Scan(&projectID, &totalBackups, &totalBackupsSizeGB); err != nil {
			log.Println("Error scanning Backups row:", err)
			continue
		}

		backupsData[projectID] = struct {
			totalBackups      float64
			totalBackupsSizeGB float64
		}{
			totalBackups:      totalBackups,
			totalBackupsSizeGB: totalBackupsSizeGB,
		}
	}

	if err := rows.Err(); err != nil {
		log.Println("Error in Backups result set:", err)
	}

	for projectID, volumes := range volumesData {
		totalSnapshots := snapshotsData[projectID]
		backups := backupsData[projectID]

		ch <- prometheus.MustNewConstMetric(
			e.volumes,
			prometheus.GaugeValue,
			volumes.totalVolumes,
			projectID,
		)

		ch <- prometheus.MustNewConstMetric(
			e.volumesSize,
			prometheus.GaugeValue,
			volumes.volumesSizeGB,
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
			backups.totalBackups,
			projectID,
		)

		ch <- prometheus.MustNewConstMetric(
			e.backupsSize,
			prometheus.GaugeValue,
			backups.totalBackupsSizeGB,
			projectID,
		)
	}
}
