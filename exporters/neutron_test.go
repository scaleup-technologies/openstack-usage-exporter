package exporters

import (
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNeutronUsageExporter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"project_id", "floatingIPs"}).
		AddRow("c352b0ed-30ca-4634-9c2d-1947efc29096", 2).
		AddRow("6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096", 3)
	mock.ExpectQuery("SELECT project_id, SUM").WillReturnRows(rows)

	exporter, err := NewNeutronUsageExporter(db)
	if err != nil {
		t.Fatalf("Failed to create NewNeutronUsageExporter: %v", err)
	}

	expectedMetrics := `
		# HELP openstack_project_total_fips Total number of floating IPs per OpenStack project
		# TYPE openstack_project_total_fips gauge
		openstack_project_fips{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 2
		openstack_project_fips{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 3
	`

	if err := testutil.CollectAndCompare(exporter, strings.NewReader(expectedMetrics)); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
