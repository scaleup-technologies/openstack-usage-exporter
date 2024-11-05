package exporters

import (
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestDesignateUsageExporter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create sqlmock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"tenant_id", "total_zones"}).
		AddRow("c352b0ed-30ca-4634-9c2d-1947efc29096", 5).
		AddRow("6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096", 3)

	mock.ExpectQuery("SELECT tenant_id, COUNT").WillReturnRows(rows)

	exporter, err := NewDesignateUsageExporter(db)
	if err != nil {
		t.Fatalf("Failed to create NewDesignateUsageExporter: %v", err)
	}

	expectedMetrics := `
        # HELP openstack_project_dns_zones Total number of dns zones per OpenStack project
        # TYPE openstack_project_dns_zones gauge
        openstack_project_dns_zones{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 3
        openstack_project_dns_zones{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 5
	`

	if err := testutil.CollectAndCompare(exporter, strings.NewReader(expectedMetrics)); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
