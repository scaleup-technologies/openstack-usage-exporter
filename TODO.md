Neutron (number of Provider Network Floating IPs / Fixed IPs)
Neutron FIPS:

select project_id, COUNT(id) as total_fips from floatingips GROUP BY project_id;

Neutron Router:

select project_id, COUNT(id) as total_routers from routers GROUP BY project_id;

Designate Zones:

select tenant_id, COUNT(id) as total_zones from zones WHERE tenant_id != '00000000-0000-0000-0000-000000000000' GROUP BY t
enant_id;

Nova local storage:

SELECT project_id, SUM(vcpus) AS total_vcpus, SUM(memory_mb) AS total_ram_mb, SUM(root_gb) as total_root_gb from instances WHERE deleted = 0 GROUP BY project_id;

Cinder / Nova Snapshot / Backup Storage
Octavia (number of LBs)
Swift/S3 (Storage in GiB)
Manila (Storage in GiB)
