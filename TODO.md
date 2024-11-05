Neutron (number of Provider Network Fixed IPs)

Designate Zones:

select tenant_id, COUNT(id) as total_zones from zones WHERE tenant_id != '00000000-0000-0000-0000-000000000000' GROUP BY t
enant_id;

Nova local storage:

SELECT project_id, SUM(vcpus) AS total_vcpus, SUM(memory_mb) AS total_ram_mb, SUM(root_gb) as total_root_gb from instances WHERE deleted = 0 GROUP BY project_id;

Cinder / Nova Snapshot / Backup Storage
Octavia (number of LBs)
Manila (Storage in GiB)
