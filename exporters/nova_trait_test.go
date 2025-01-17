package exporters

import (
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNovaTraitUsageExporter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"project_id", "total_instances", "total_vcpus"}).
		AddRow("c352b0ed-30ca-4634-9c2d-1947efc29096", 0, 0).
		AddRow("6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096", 4, 5)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	exporter, err := NewNovaTraitUsageExporter(db, "CUSTOM_TRAIT")
	if err != nil {
		t.Fatalf("Failed to create NewNovaTraitUsageExporter: %v", err)
	}

	expectedMetrics := `
        # HELP openstack_project_instances_trait__custom_trait Total number of instances per OpenStack project with image trait CUSTOM_TRAIT
        # TYPE openstack_project_instances_trait__custom_trait gauge
        openstack_project_instances_trait__custom_trait{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 4
        openstack_project_instances_trait__custom_trait{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 0
        # HELP openstack_project_vcpus_trait__custom_trait Total number of vcpus per OpenStack project for instances with image trait CUSTOM_TRAIT
        # TYPE openstack_project_vcpus_trait__custom_trait gauge
        openstack_project_vcpus_trait__custom_trait{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 5
        openstack_project_vcpus_trait__custom_trait{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 0

	`

	if err := testutil.CollectAndCompare(exporter, strings.NewReader(expectedMetrics)); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
