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

	floatingIPRows := sqlmock.NewRows([]string{"project_id", "total_fips"}).
		AddRow("c352b0ed-30ca-4634-9c2d-1947efc29096", 2).
		AddRow("6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096", 3)
	mock.ExpectQuery("SELECT project_id, COUNT\\(id\\) AS total_fips FROM floatingips GROUP BY project_id").WillReturnRows(floatingIPRows)

	routerRows := sqlmock.NewRows([]string{"project_id", "total_routers"}).
		AddRow("c352b0ed-30ca-4634-9c2d-1947efc29096", 1).
		AddRow("6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096", 2)
	mock.ExpectQuery("SELECT project_id, COUNT\\(id\\) AS total_routers FROM routers GROUP BY project_id").WillReturnRows(routerRows)

	exporter, err := NewNeutronUsageExporter(db)
	if err != nil {
		t.Fatalf("Failed to create NeutronUsageExporter: %v", err)
	}

	expectedMetrics := `
        # HELP openstack_project_floating_ips Total number of floating IPs per OpenStack project
        # TYPE openstack_project_floating_ips gauge
        openstack_project_floating_ips{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 3
        openstack_project_floating_ips{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 2
        # HELP openstack_project_routers Total number of routers per OpenStack project
        # TYPE openstack_project_routers gauge
        openstack_project_routers{project_id="6ee08ba2-2ca1-4c91-b139-4bf0dbaa4096"} 2
        openstack_project_routers{project_id="c352b0ed-30ca-4634-9c2d-1947efc29096"} 1
	`

	if err := testutil.CollectAndCompare(exporter, strings.NewReader(expectedMetrics)); err != nil {
		t.Errorf("unexpected collecting result:\n%s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
