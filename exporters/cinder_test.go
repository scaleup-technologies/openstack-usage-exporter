package exporters

import (
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

	rows := sqlmock.NewRows([]string{
		"project_id", "total_volumes", "volumes_size_gb", "total_snapshots", "total_backups", "total_backups_size_gb",
	}).
		AddRow("c352b0ed-30ca-4634-9c2d-1947efc29096", 12, 43, 5, 3, 10).
		AddRow("6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096", 2, 10, 2, 1, 5)

	mock.ExpectQuery("SELECT project_id, COUNT").WillReturnRows(rows)

	exporter, err := NewCinderUsageExporter(db)
	if err != nil {
		t.Fatalf("Failed to create NewCinderUsageExporter: %v", err)
	}

	expectedMetrics := `
        # HELP openstack_project_volume_size_gb Total volume size in GB per OpenStack project
        # TYPE openstack_project_volume_size_gb gauge
        openstack_project_volume_size_gb{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 10
        openstack_project_volume_size_gb{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 43
        # HELP openstack_project_volumes Total number of volumes per OpenStack project
        # TYPE openstack_project_volumes gauge
        openstack_project_volumes{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 2
        openstack_project_volumes{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 12
        # HELP openstack_project_snapshots Total number of snapshots per OpenStack project
        # TYPE openstack_project_snapshots gauge
        openstack_project_snapshots{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 2
        openstack_project_snapshots{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 5
        # HELP openstack_project_backups Total number of backups per OpenStack project
        # TYPE openstack_project_backups gauge
        openstack_project_backups{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 1
        openstack_project_backups{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 3
        # HELP openstack_project_backups_size_gb Total size of backups in GB per OpenStack project
        # TYPE openstack_project_backups_size_gb gauge
        openstack_project_backups_size_gb{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 5
        openstack_project_backups_size_gb{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 10
	`

	if err := testutil.CollectAndCompare(exporter, strings.NewReader(expectedMetrics)); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
