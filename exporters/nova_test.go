package exporters

import (
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNovaUsageExporter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"project_id", "total_vcpus", "total_ram_mb"}).
		AddRow("c352b0ed-30ca-4634-9c2d-1947efc29096", 2, 1024).
		AddRow("6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096", 8, 2048)
	mock.ExpectQuery("SELECT project_id, SUM").WillReturnRows(rows)

	exporter, err := NewNovaUsageExporter(db)
	if err != nil {
		t.Fatalf("Failed to create NewNovaUsageExporter: %v", err)
	}

	expectedMetrics := `
		# HELP openstack_project_vcpus Total number of vcpus per OpenStack project
		# TYPE openstack_project_vcpus gauge
		openstack_project_vcpus{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 2
		openstack_project_vcpus{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 8
		# HELP openstack_project_ram_mb Total ram usage in MB per OpenStack project
		# TYPE openstack_project_ram_mb gauge
		openstack_project_ram_mb{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 1024
		openstack_project_ram_mb{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 2048
	`

	if err := testutil.CollectAndCompare(exporter, strings.NewReader(expectedMetrics)); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
