package exporters

import (
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestManilaUsageExporter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	sharesRows := sqlmock.NewRows([]string{
		"project_id", "shares_size"}).
		AddRow("6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096", 2).
		AddRow("c352b0ed-30ca-4634-9c2d-1947efc29096", 12)
	mock.ExpectQuery("SELECT project_id, SUM\\(size\\) AS shares_size FROM shares").WillReturnRows(sharesRows)

	shareSnapshotsRows := sqlmock.NewRows([]string{
		"project_id", "shares_size"}).
		AddRow("6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096", 4).
		AddRow("c352b0ed-30ca-4634-9c2d-1947efc29096", 24)
	mock.ExpectQuery("SELECT project_id, SUM\\(size\\) AS share_snapshots_size FROM share_snapshots").WillReturnRows(shareSnapshotsRows)

	shareBackupsRows := sqlmock.NewRows([]string{
		"project_id", "shares_size"}).
		AddRow("6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096", 1).
		AddRow("c352b0ed-30ca-4634-9c2d-1947efc29096", 0)
	mock.ExpectQuery("SELECT project_id, SUM\\(size\\) AS share_backups_size FROM share_backups").WillReturnRows(shareBackupsRows)

	exporter, err := NewManilaUsageExporter(db)
	if err != nil {
		t.Fatalf("Failed to create ManilaUsageExporter: %v", err)
	}

	expectedMetrics := `
        # HELP openstack_project_share_backups_size_gb Total share backup size in GB per OpenStack project
        # TYPE openstack_project_share_backups_size_gb gauge
        openstack_project_share_backups_size_gb{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 1
        openstack_project_share_backups_size_gb{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 0
        # HELP openstack_project_share_snapshots_size_gb Total share snapshot size in GB per OpenStack project
        # TYPE openstack_project_share_snapshots_size_gb gauge
        openstack_project_share_snapshots_size_gb{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 4
        openstack_project_share_snapshots_size_gb{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 24
        # HELP openstack_project_shares_size_gb Total share size in GB per OpenStack project
        # TYPE openstack_project_shares_size_gb gauge
        openstack_project_shares_size_gb{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 2
        openstack_project_shares_size_gb{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 12
	`

	if err := testutil.CollectAndCompare(exporter, strings.NewReader(expectedMetrics)); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
