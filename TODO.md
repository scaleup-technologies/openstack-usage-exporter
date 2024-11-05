Neutron (number of Provider Network Fixed IPs)

Designate Zones:

select tenant_id, COUNT(id) as total_zones from zones WHERE tenant_id != '00000000-0000-0000-0000-000000000000' GROUP BY t
enant_id;

Cinder / Nova Snapshot / Backup Storage
Octavia (number of LBs)
Manila (Storage in GiB)
