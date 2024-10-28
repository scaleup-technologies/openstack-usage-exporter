Neutron FIPS:

select project_id, COUNT(id) as total_fips from floatingips GROUP BY project_id;

Neutron Router:

select project_id, COUNT(id) as total_routers from routers GROUP BY project_id;

Designate Zones:

select tenant_id, COUNT(id) as total_zones from zones WHERE tenant_id != '00000000-0000-0000-0000-000000000000' GROUP BY t
enant_id;
