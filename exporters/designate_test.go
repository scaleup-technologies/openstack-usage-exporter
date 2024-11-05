package exporters

import (
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestDesignateUsageExporter(t *testing.T) {
	// Create a mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Define the rows returned by the mock query
	// Simulate 2 projects with a certain number of zones
	rows := sqlmock.NewRows([]string{"tenant_id", "total_zones"}).
		AddRow("c352b0ed-30ca-4634-9c2d-1947efc29096", 5).
		AddRow("6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096", 3)

	// Set the expected query and return the rows
	mock.ExpectQuery("SELECT tenant_id, COUNT").WillReturnRows(rows)

	// Create the exporter
	exporter, err := NewDesignateUsageExporter(db)
	if err != nil {
		t.Fatalf("Failed to create NewDesignateUsageExporter: %v", err)
	}

	// Define the expected metrics in the format Prometheus expects
	expectedMetrics := `
        # HELP openstack_project_zones Total number of zones per OpenStack project
        # TYPE openstack_project_zones gauge
        openstack_project_zones{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 3
        openstack_project_zones{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 5
	`

	// Compare the expected metrics with the actual ones collected by the exporter
	if err := testutil.CollectAndCompare(exporter, strings.NewReader(expectedMetrics)); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
