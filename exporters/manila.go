package exporters

import (
	"database/sql"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

type ManilaUsageExporter struct {
	db                 *sql.DB
	sharesSize         *prometheus.Desc
	shareSnapshotsSize *prometheus.Desc
	shareBackupsSize   *prometheus.Desc
}

func NewManilaUsageExporter(db *sql.DB) (*ManilaUsageExporter, error) {
	return &ManilaUsageExporter{
		db: db,
		sharesSize: prometheus.NewDesc(
			"openstack_project_shares_size_gb",
			"Total share size in GB per OpenStack project",
			[]string{"project_id"}, nil,
		),
		shareSnapshotsSize: prometheus.NewDesc(
			"openstack_project_share_snapshots_size_gb",
			"Total share snapshot size in GB per OpenStack project",
			[]string{"project_id"}, nil,
		),
		shareBackupsSize: prometheus.NewDesc(
			"openstack_project_share_backups_size_gb",
			"Total share backup size in GB per OpenStack project",
			[]string{"project_id"}, nil,
		),
	}, nil
}

func (e *ManilaUsageExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.sharesSize
	ch <- e.shareSnapshotsSize
	ch <- e.shareBackupsSize
}

func (e *ManilaUsageExporter) Collect(ch chan<- prometheus.Metric) {
	e.collectShareSize(ch)
	e.collectShareSnapshotSize(ch)
	e.collectShareBackupSize(ch)
}

func (e *ManilaUsageExporter) collectShareSize(ch chan<- prometheus.Metric) {
	rows, err := e.db.Query("SELECT project_id, SUM(size) AS shares_size FROM shares WHERE deleted='False' GROUP BY project_id")
	if err != nil {
		log.Println("Error querying Manila database:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var projectID string
		var sharesSize float64
		if err := rows.Scan(&projectID, &sharesSize); err != nil {
			log.Println("Error scanning Manila row:", err)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			e.sharesSize,
			prometheus.GaugeValue,
			sharesSize,
			projectID,
		)

	}
	if err := rows.Err(); err != nil {
		log.Println("Error in Manila result set:", err)
	}
}

func (e *ManilaUsageExporter) collectShareSnapshotSize(ch chan<- prometheus.Metric) {
	rows, err := e.db.Query("SELECT project_id, SUM(size) AS share_snapshots_size FROM share_snapshots WHERE deleted='False' GROUP BY project_id")
	if err != nil {
		log.Println("Error querying Manila database:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var projectID string
		var shareSnapshotsSize float64
		if err := rows.Scan(&projectID, &shareSnapshotsSize); err != nil {
			log.Println("Error scanning Manila row:", err)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			e.shareSnapshotsSize,
			prometheus.GaugeValue,
			shareSnapshotsSize,
			projectID,
		)

	}
	if err := rows.Err(); err != nil {
		log.Println("Error in Manila result set:", err)
	}
}

func (e *ManilaUsageExporter) collectShareBackupSize(ch chan<- prometheus.Metric) {
	rows, err := e.db.Query("SELECT project_id, SUM(size) AS share_backups_size FROM share_backups WHERE deleted='False' GROUP BY project_id")
	if err != nil {
		log.Println("Error querying Manila database:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var projectID string
		var shareBackupsSize float64
		if err := rows.Scan(&projectID, &shareBackupsSize); err != nil {
			log.Println("Error scanning Manila row:", err)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			e.shareBackupsSize,
			prometheus.GaugeValue,
			shareBackupsSize,
			projectID,
		)

	}
	if err := rows.Err(); err != nil {
		log.Println("Error in Manila result set:", err)
	}
}
