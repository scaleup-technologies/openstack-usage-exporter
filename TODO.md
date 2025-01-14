Volume types

```
select vl.project_id, vt.name, sum(vl.size) from volumes vl left join volume_types vt on vl.volume_type_id = vt.id group by project_id, volume_type_id;
openstack_project_volume_size_gb{volume_type="SSD"}
```
