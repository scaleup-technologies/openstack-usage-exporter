package exporters

import (
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCinderUsageExporter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	volumeRows := sqlmock.NewRows([]string{
		"project_id", "total_volumes", "volumes_size_gb"}).
		AddRow("6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096", 2, 10).
		AddRow("c352b0ed-30ca-4634-9c2d-1947efc29096", 12, 43)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT project_id, COUNT(id) AS total_volumes, SUM(size) AS volumes_size_gb FROM volumes WHERE deleted = 0 GROUP BY project_id")).
		WillReturnRows(volumeRows)

	snapshotRows := sqlmock.NewRows([]string{
		"project_id", "total_snapshots", "snapshot_size_gb"}).
		AddRow("6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096", 2, 8).
		AddRow("c352b0ed-30ca-4634-9c2d-1947efc29096", 5, 15)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT project_id, COUNT(id) AS total_snapshots, SUM(volume_size) AS snapshot_size_gb FROM snapshots WHERE deleted = 0 GROUP BY project_id")).
		WillReturnRows(snapshotRows)

	backupRows := sqlmock.NewRows([]string{
		"project_id", "total_backups", "total_backups_size_gb"}).
		AddRow("6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096", 1, 5).
		AddRow("c352b0ed-30ca-4634-9c2d-1947efc29096", 3, 10)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT project_id, COUNT(id) AS total_backups, SUM(size) AS total_backups_size_gb FROM backups WHERE deleted = 0 GROUP BY project_id")).
		WillReturnRows(backupRows)

	exporter, err := NewCinderUsageExporter(db)
	if err != nil {
		t.Fatalf("Failed to create NewCinderUsageExporter: %v", err)
	}

	expectedMetrics := `
        # HELP openstack_project_backups Total number of backups per OpenStack project
        # TYPE openstack_project_backups gauge
        openstack_project_backups{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 1
        openstack_project_backups{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 3
        # HELP openstack_project_backups_size_gb Total size of backups in GB per OpenStack project
        # TYPE openstack_project_backups_size_gb gauge
        openstack_project_backups_size_gb{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 5
        openstack_project_backups_size_gb{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 10
        # HELP openstack_project_snapshots Total number of snapshots per OpenStack project
        # TYPE openstack_project_snapshots gauge
        openstack_project_snapshots{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 2
        openstack_project_snapshots{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 5
        # HELP openstack_project_snapshots_size_gb Total size of snapshots in GB per OpenStack project
        # TYPE openstack_project_snapshots_size_gb gauge
        openstack_project_snapshots_size_gb{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 8
        openstack_project_snapshots_size_gb{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 15
        # HELP openstack_project_volume_size_gb Total volume size in GB per OpenStack project
        # TYPE openstack_project_volume_size_gb gauge
        openstack_project_volume_size_gb{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 10
        openstack_project_volume_size_gb{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 43
        # HELP openstack_project_volumes Total number of volumes per OpenStack project
        # TYPE openstack_project_volumes gauge
        openstack_project_volumes{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 2
        openstack_project_volumes{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 12
	`

	if err := testutil.CollectAndCompare(exporter, strings.NewReader(expectedMetrics)); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
