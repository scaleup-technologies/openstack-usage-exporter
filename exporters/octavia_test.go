package exporters

import (
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestOctaviaUsageExporter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"project_id", "total_lbs"}).
		AddRow("c352b0ed-30ca-4634-9c2d-1947efc29096", 5).
		AddRow("6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096", 3)

	mock.ExpectQuery("SELECT project_id, COUNT").WillReturnRows(rows)

	exporter, err := NewOctaviaUsageExporter(db)
	if err != nil {
		t.Fatalf("Failed to create NewOctaviaUsageExporter: %v", err)
	}

	expectedMetrics := `
        # HELP openstack_project_load_balancers Total number of load balancers per OpenStack project
        # TYPE openstack_project_load_balancers gauge
        openstack_project_load_balancers{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 3
        openstack_project_load_balancers{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 5
	`

	if err := testutil.CollectAndCompare(exporter, strings.NewReader(expectedMetrics)); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
