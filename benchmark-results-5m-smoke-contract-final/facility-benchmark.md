# Facility benchmark (5000000 FieldDevices)

Git: `da4128276ebd2fb1ae2207675929301a8b942f91`  
Generated: 2026-08-22T16:40:47Z

## PostgreSQL

- `effective_cache_size`: `6GB`
- `fsync`: `on`
- `jit`: `off`
- `random_page_cost`: `1.1`
- `server_version`: `18.3 (Debian 18.3-1.pgdg13+1)`
- `shared_buffers`: `2GB`
- `synchronous_commit`: `on`
- `track_io_timing`: `on`
- `work_mem`: `32MB`

## Hardware

- `database_profile_memory`: `8 GiB`
- `database_profile_vcpus`: `4`
- `database_storage`: `persistent SSD volume`
- `go_arch`: `amd64`
- `go_os`: `windows`
- `runner_logical_cpus`: `8`

## Database

- `database_bytes`: `34157164223`
- `expected_field_devices`: `5000000`
- `field_device_building_cursor_values_index_bytes`: `871538688`
- `field_device_building_cursor_values_rows`: `5000000`
- `field_device_building_cursor_values_table_bytes`: `341336064`
- `field_device_cursor_values_index_bytes`: `11846475776`
- `field_device_cursor_values_rows`: `5000000`
- `field_device_cursor_values_table_bytes`: `560979968`
- `field_devices_bmk_null_rows`: `1000000`
- `field_devices_index_bytes`: `5867806720`
- `field_devices_rows`: `5000000`
- `field_devices_table_bytes`: `889430016`
- `project_field_devices_index_bytes`: `2642747392`
- `project_field_devices_rows`: `5000000`
- `project_field_devices_table_bytes`: `1051820032`
- `schema_migrations`: `35`
- `search_0_1_percent_rows`: `5000`
- `search_10_percent_rows`: `500000`
- `search_1_percent_rows`: `50000`
- `specification_supplier_null_rows`: `833334`
- `specifications_index_bytes`: `9304580096`
- `specifications_rows`: `5000000`
- `specifications_table_bytes`: `713703424`

## Cold cache

record separately; warm-cache p95 is the release gate

## Migrations

`202604030001`, `202604030007`, `202604030008`, `202604030009`, `202604300001`, `202604300002`, `202604300003`, `202604300004`, `202604300005`, `202605010001`, `202605010002`, `202605010003`, `202605020001`, `202605030001`, `202605070001`, `202605080001`, `202605080002`, `202605080003`, `202605110001`, `202605110002`, `202608130001`, `202608160001`, `202608210001`, `202608210002`, `202608220001`, `202608220002`, `202608220003`, `202608220004`, `202608220005`, `202608220006`, `202608220007`, `202608220008`, `202608220009`, `202608220010`, `202609050001`

| Scenario | Samples | Failures | Error rate | p50 ms | p95 ms | p99 ms | Gate |
|---|---:|---:|---:|---:|---:|---:|---|
| created_at_asc_first | 15 | 0 | 0.0000% | 10.54 | 16.74 | 16.74 | true |
| created_at_asc_next | 15 | 0 | 0.0000% | 11.32 | 14.20 | 14.20 | true |
| created_at_desc_first | 15 | 0 | 0.0000% | 11.97 | 20.92 | 20.92 | true |
| created_at_desc_next | 15 | 0 | 0.0000% | 10.17 | 14.55 | 14.55 | true |
| apparat_nr_asc_first | 15 | 0 | 0.0000% | 9.38 | 14.26 | 14.26 | true |
| apparat_nr_asc_next | 15 | 0 | 0.0000% | 9.53 | 12.24 | 12.24 | true |
| apparat_nr_desc_first | 15 | 0 | 0.0000% | 8.78 | 15.27 | 15.27 | true |
| apparat_nr_desc_next | 15 | 0 | 0.0000% | 11.89 | 15.02 | 15.02 | true |
| sps_system_type_asc_first | 15 | 0 | 0.0000% | 12.70 | 16.11 | 16.11 | true |
| sps_system_type_asc_next | 15 | 0 | 0.0000% | 12.75 | 15.35 | 15.35 | true |
| sps_system_type_desc_first | 15 | 0 | 0.0000% | 20.40 | 37.01 | 37.01 | true |
| sps_system_type_desc_next | 15 | 0 | 0.0000% | 12.76 | 14.21 | 14.21 | true |
| description_asc_first | 15 | 0 | 0.0000% | 8.66 | 11.79 | 11.79 | true |
| description_asc_next | 15 | 0 | 0.0000% | 10.03 | 12.66 | 12.66 | true |
| description_desc_first | 15 | 0 | 0.0000% | 9.47 | 12.13 | 12.13 | true |
| description_desc_next | 15 | 0 | 0.0000% | 10.71 | 13.75 | 13.75 | true |
| spec_supplier_asc_first | 15 | 0 | 0.0000% | 10.26 | 12.55 | 12.55 | true |
| spec_supplier_asc_next | 15 | 0 | 0.0000% | 9.68 | 11.80 | 11.80 | true |
| spec_supplier_desc_first | 15 | 0 | 0.0000% | 9.69 | 11.94 | 11.94 | true |
| spec_supplier_desc_next | 15 | 0 | 0.0000% | 9.44 | 13.07 | 13.07 | true |
| spec_size_asc_first | 15 | 0 | 0.0000% | 11.34 | 14.21 | 14.21 | true |
| spec_size_asc_next | 15 | 0 | 0.0000% | 10.39 | 14.54 | 14.54 | true |
| spec_size_desc_first | 15 | 0 | 0.0000% | 11.20 | 16.22 | 16.22 | true |
| spec_size_desc_next | 15 | 0 | 0.0000% | 13.29 | 16.30 | 16.30 | true |
| spec_acdc_asc_first | 15 | 0 | 0.0000% | 16.13 | 20.39 | 20.39 | true |
| spec_acdc_asc_next | 15 | 0 | 0.0000% | 11.32 | 23.36 | 23.36 | true |
| spec_acdc_desc_first | 15 | 0 | 0.0000% | 8.96 | 12.78 | 12.78 | true |
| spec_acdc_desc_next | 15 | 0 | 0.0000% | 9.11 | 12.54 | 12.54 | true |
| spec_power_asc_first | 15 | 0 | 0.0000% | 10.37 | 19.39 | 19.39 | true |
| spec_power_asc_next | 15 | 0 | 0.0000% | 11.31 | 13.60 | 13.60 | true |
| spec_power_desc_first | 15 | 0 | 0.0000% | 10.18 | 12.76 | 12.76 | true |
| spec_power_desc_next | 15 | 0 | 0.0000% | 10.49 | 17.93 | 17.93 | true |
| building_scope_first | 15 | 0 | 0.0000% | 9.19 | 15.17 | 15.17 | true |
| building_scope_previous | 15 | 0 | 0.0000% | 11.91 | 15.29 | 15.29 | true |
| building_scope_next | 15 | 0 | 0.0000% | 9.99 | 11.20 | 11.20 | true |
| cabinet_scope_first | 15 | 0 | 0.0000% | 22.54 | 34.23 | 34.23 | true |
| cabinet_scope_previous | 15 | 0 | 0.0000% | 11.06 | 13.47 | 13.47 | true |
| cabinet_scope_next | 15 | 0 | 0.0000% | 15.59 | 30.19 | 30.19 | true |
| controller_scope_first | 15 | 0 | 0.0000% | 11.85 | 14.51 | 14.51 | true |
| controller_scope_previous | 15 | 0 | 0.0000% | 12.68 | 29.50 | 29.50 | true |
| controller_scope_next | 15 | 0 | 0.0000% | 10.47 | 17.85 | 17.85 | true |
| project_scope_first | 15 | 0 | 0.0000% | 13.08 | 14.90 | 14.90 | true |
| project_scope_previous | 15 | 0 | 0.0000% | 11.21 | 15.77 | 15.77 | true |
| project_scope_next | 15 | 0 | 0.0000% | 10.02 | 12.77 | 12.77 | true |
| search_0_1_percent_first | 15 | 0 | 0.0000% | 54.84 | 69.84 | 69.84 | true |
| search_0_1_percent_previous | 15 | 0 | 0.0000% | 57.52 | 62.64 | 62.64 | true |
| search_0_1_percent_next | 15 | 0 | 0.0000% | 54.07 | 67.37 | 67.37 | true |
| search_1_percent_first | 15 | 0 | 0.0000% | 37.07 | 55.08 | 55.08 | true |
| search_1_percent_previous | 15 | 0 | 0.0000% | 53.20 | 59.23 | 59.23 | true |
| search_1_percent_next | 15 | 0 | 0.0000% | 35.36 | 40.08 | 40.08 | true |
| search_10_percent_first | 15 | 0 | 0.0000% | 33.31 | 36.17 | 36.17 | true |
| search_10_percent_previous | 15 | 0 | 0.0000% | 14.37 | 16.78 | 16.78 | true |
| search_10_percent_next | 15 | 0 | 0.0000% | 34.19 | 51.27 | 51.27 | true |
| combined_filter_first | 15 | 0 | 0.0000% | 11.59 | 12.91 | 12.91 | true |
| combined_filter_previous | 15 | 0 | 0.0000% | 12.34 | 23.02 | 23.02 | true |
| combined_filter_next | 15 | 0 | 0.0000% | 10.99 | 11.64 | 11.64 | true |
| created_at_asc_p25_next | 15 | 0 | 0.0000% | 8.52 | 11.59 | 11.59 | true |
| created_at_asc_p25_previous | 15 | 0 | 0.0000% | 10.26 | 12.59 | 12.59 | true |
| created_at_asc_p50_next | 15 | 0 | 0.0000% | 8.69 | 11.99 | 11.99 | true |
| created_at_asc_p50_previous | 15 | 0 | 0.0000% | 9.38 | 56.18 | 56.18 | true |
| created_at_asc_p75_next | 15 | 0 | 0.0000% | 8.28 | 12.62 | 12.62 | true |
| created_at_asc_p75_previous | 15 | 0 | 0.0000% | 10.01 | 24.85 | 24.85 | true |
| created_at_asc_p99_next | 15 | 0 | 0.0000% | 8.99 | 10.61 | 10.61 | true |
| created_at_asc_p99_previous | 15 | 0 | 0.0000% | 8.06 | 10.52 | 10.52 | true |
| created_at_desc_p25_next | 15 | 0 | 0.0000% | 8.13 | 9.83 | 9.83 | true |
| created_at_desc_p25_previous | 15 | 0 | 0.0000% | 8.78 | 10.53 | 10.53 | true |
| created_at_desc_p50_next | 15 | 0 | 0.0000% | 9.43 | 12.14 | 12.14 | true |
| created_at_desc_p50_previous | 15 | 0 | 0.0000% | 8.07 | 11.31 | 11.31 | true |
| created_at_desc_p75_next | 15 | 0 | 0.0000% | 8.24 | 21.58 | 21.58 | true |
| created_at_desc_p75_previous | 15 | 0 | 0.0000% | 14.33 | 17.71 | 17.71 | true |
| created_at_desc_p99_next | 15 | 0 | 0.0000% | 9.67 | 11.22 | 11.22 | true |
| created_at_desc_p99_previous | 15 | 0 | 0.0000% | 9.28 | 11.24 | 11.24 | true |
| apparat_nr_asc_p25_next | 15 | 0 | 0.0000% | 7.64 | 10.47 | 10.47 | true |
| apparat_nr_asc_p25_previous | 15 | 0 | 0.0000% | 8.68 | 10.47 | 10.47 | true |
| apparat_nr_asc_p50_next | 15 | 0 | 0.0000% | 10.69 | 15.94 | 15.94 | true |
| apparat_nr_asc_p50_previous | 15 | 0 | 0.0000% | 8.31 | 10.86 | 10.86 | true |
| apparat_nr_asc_p75_next | 15 | 0 | 0.0000% | 8.14 | 11.04 | 11.04 | true |
| apparat_nr_asc_p75_previous | 15 | 0 | 0.0000% | 8.51 | 10.96 | 10.96 | true |
| apparat_nr_asc_p99_next | 15 | 0 | 0.0000% | 15.42 | 21.31 | 21.31 | true |
| apparat_nr_asc_p99_previous | 15 | 0 | 0.0000% | 10.73 | 12.97 | 12.97 | true |
| apparat_nr_desc_p25_next | 15 | 0 | 0.0000% | 8.93 | 15.64 | 15.64 | true |
| apparat_nr_desc_p25_previous | 15 | 0 | 0.0000% | 9.09 | 10.85 | 10.85 | true |
| apparat_nr_desc_p50_next | 15 | 0 | 0.0000% | 7.97 | 12.56 | 12.56 | true |
| apparat_nr_desc_p50_previous | 15 | 0 | 0.0000% | 8.74 | 11.37 | 11.37 | true |
| apparat_nr_desc_p75_next | 15 | 0 | 0.0000% | 9.91 | 12.64 | 12.64 | true |
| apparat_nr_desc_p75_previous | 15 | 0 | 0.0000% | 8.75 | 10.53 | 10.53 | true |
| apparat_nr_desc_p99_next | 15 | 0 | 0.0000% | 8.55 | 10.86 | 10.86 | true |
| apparat_nr_desc_p99_previous | 15 | 0 | 0.0000% | 8.70 | 12.65 | 12.65 | true |
| sps_system_type_asc_p25_next | 15 | 0 | 0.0000% | 12.75 | 24.40 | 24.40 | true |
| sps_system_type_asc_p25_previous | 15 | 0 | 0.0000% | 12.62 | 14.31 | 14.31 | true |
| sps_system_type_asc_p50_next | 15 | 0 | 0.0000% | 11.44 | 14.03 | 14.03 | true |
| sps_system_type_asc_p50_previous | 15 | 0 | 0.0000% | 14.04 | 24.76 | 24.76 | true |
| sps_system_type_asc_p75_next | 15 | 0 | 0.0000% | 11.33 | 13.40 | 13.40 | true |
| sps_system_type_asc_p75_previous | 15 | 0 | 0.0000% | 12.34 | 16.48 | 16.48 | true |
| sps_system_type_asc_p99_next | 15 | 0 | 0.0000% | 12.11 | 18.61 | 18.61 | true |
| sps_system_type_asc_p99_previous | 15 | 0 | 0.0000% | 13.46 | 16.10 | 16.10 | true |
| sps_system_type_desc_p25_next | 15 | 0 | 0.0000% | 36.20 | 39.83 | 39.83 | true |
| sps_system_type_desc_p25_previous | 15 | 0 | 0.0000% | 81.82 | 98.09 | 98.09 | true |
| sps_system_type_desc_p50_next | 15 | 0 | 0.0000% | 58.49 | 78.90 | 78.90 | true |
| sps_system_type_desc_p50_previous | 15 | 0 | 0.0000% | 55.91 | 66.75 | 66.75 | true |
| sps_system_type_desc_p75_next | 15 | 0 | 0.0000% | 80.93 | 101.47 | 101.47 | true |
| sps_system_type_desc_p75_previous | 15 | 0 | 0.0000% | 43.22 | 57.82 | 57.82 | true |
| sps_system_type_desc_p99_next | 15 | 0 | 0.0000% | 102.86 | 106.64 | 106.64 | true |
| sps_system_type_desc_p99_previous | 15 | 0 | 0.0000% | 13.89 | 17.17 | 17.17 | true |
| description_asc_p25_next | 15 | 0 | 0.0000% | 9.50 | 13.98 | 13.98 | true |
| description_asc_p25_previous | 15 | 0 | 0.0000% | 9.47 | 21.01 | 21.01 | true |
| description_asc_p50_next | 15 | 0 | 0.0000% | 11.98 | 16.96 | 16.96 | true |
| description_asc_p50_previous | 15 | 0 | 0.0000% | 8.70 | 10.01 | 10.01 | true |
| description_asc_p75_next | 15 | 0 | 0.0000% | 8.38 | 10.09 | 10.09 | true |
| description_asc_p75_previous | 15 | 0 | 0.0000% | 7.66 | 9.20 | 9.20 | true |
| description_asc_p99_next | 15 | 0 | 0.0000% | 9.29 | 10.16 | 10.16 | true |
| description_asc_p99_previous | 15 | 0 | 0.0000% | 8.71 | 11.65 | 11.65 | true |
| description_desc_p25_next | 15 | 0 | 0.0000% | 8.61 | 9.85 | 9.85 | true |
| description_desc_p25_previous | 15 | 0 | 0.0000% | 8.87 | 11.61 | 11.61 | true |
| description_desc_p50_next | 15 | 0 | 0.0000% | 8.78 | 13.97 | 13.97 | true |
| description_desc_p50_previous | 15 | 0 | 0.0000% | 9.26 | 13.12 | 13.12 | true |
| description_desc_p75_next | 15 | 0 | 0.0000% | 9.95 | 18.43 | 18.43 | true |
| description_desc_p75_previous | 15 | 0 | 0.0000% | 14.64 | 21.41 | 21.41 | true |
| description_desc_p99_next | 15 | 0 | 0.0000% | 12.54 | 21.64 | 21.64 | true |
| description_desc_p99_previous | 15 | 0 | 0.0000% | 8.88 | 11.07 | 11.07 | true |
| spec_supplier_asc_p25_next | 15 | 0 | 0.0000% | 8.94 | 11.78 | 11.78 | true |
| spec_supplier_asc_p25_previous | 15 | 0 | 0.0000% | 9.87 | 13.80 | 13.80 | true |
| spec_supplier_asc_p50_next | 15 | 0 | 0.0000% | 8.34 | 12.03 | 12.03 | true |
| spec_supplier_asc_p50_previous | 15 | 0 | 0.0000% | 8.40 | 12.14 | 12.14 | true |
| spec_supplier_asc_p75_next | 15 | 0 | 0.0000% | 8.85 | 9.98 | 9.98 | true |
| spec_supplier_asc_p75_previous | 15 | 0 | 0.0000% | 9.93 | 12.86 | 12.86 | true |
| spec_supplier_asc_p99_next | 15 | 0 | 0.0000% | 9.04 | 11.20 | 11.20 | true |
| spec_supplier_asc_p99_previous | 15 | 0 | 0.0000% | 10.05 | 11.36 | 11.36 | true |
| spec_supplier_desc_p25_next | 15 | 0 | 0.0000% | 9.36 | 14.03 | 14.03 | true |
| spec_supplier_desc_p25_previous | 15 | 0 | 0.0000% | 9.53 | 12.12 | 12.12 | true |
| spec_supplier_desc_p50_next | 15 | 0 | 0.0000% | 9.27 | 11.86 | 11.86 | true |
| spec_supplier_desc_p50_previous | 15 | 0 | 0.0000% | 8.92 | 12.05 | 12.05 | true |
| spec_supplier_desc_p75_next | 15 | 0 | 0.0000% | 8.31 | 11.69 | 11.69 | true |
| spec_supplier_desc_p75_previous | 15 | 0 | 0.0000% | 9.39 | 10.81 | 10.81 | true |
| spec_supplier_desc_p99_next | 15 | 0 | 0.0000% | 9.40 | 12.82 | 12.82 | true |
| spec_supplier_desc_p99_previous | 15 | 0 | 0.0000% | 9.34 | 12.62 | 12.62 | true |
| spec_size_asc_p25_next | 15 | 0 | 0.0000% | 9.99 | 12.89 | 12.89 | true |
| spec_size_asc_p25_previous | 15 | 0 | 0.0000% | 9.52 | 15.91 | 15.91 | true |
| spec_size_asc_p50_next | 15 | 0 | 0.0000% | 10.23 | 13.36 | 13.36 | true |
| spec_size_asc_p50_previous | 15 | 0 | 0.0000% | 10.70 | 12.06 | 12.06 | true |
| spec_size_asc_p75_next | 15 | 0 | 0.0000% | 10.50 | 13.23 | 13.23 | true |
| spec_size_asc_p75_previous | 15 | 0 | 0.0000% | 9.87 | 12.23 | 12.23 | true |
| spec_size_asc_p99_next | 15 | 0 | 0.0000% | 9.86 | 13.33 | 13.33 | true |
| spec_size_asc_p99_previous | 15 | 0 | 0.0000% | 11.66 | 17.99 | 17.99 | true |
| spec_size_desc_p25_next | 15 | 0 | 0.0000% | 9.78 | 13.96 | 13.96 | true |
| spec_size_desc_p25_previous | 15 | 0 | 0.0000% | 10.04 | 24.59 | 24.59 | true |
| spec_size_desc_p50_next | 15 | 0 | 0.0000% | 11.19 | 37.86 | 37.86 | true |
| spec_size_desc_p50_previous | 15 | 0 | 0.0000% | 12.67 | 14.69 | 14.69 | true |
| spec_size_desc_p75_next | 15 | 0 | 0.0000% | 9.72 | 13.67 | 13.67 | true |
| spec_size_desc_p75_previous | 15 | 0 | 0.0000% | 11.18 | 13.68 | 13.68 | true |
| spec_size_desc_p99_next | 15 | 0 | 0.0000% | 10.25 | 13.31 | 13.31 | true |
| spec_size_desc_p99_previous | 15 | 0 | 0.0000% | 10.38 | 13.76 | 13.76 | true |
| spec_acdc_asc_p25_next | 15 | 0 | 0.0000% | 9.73 | 12.51 | 12.51 | true |
| spec_acdc_asc_p25_previous | 15 | 0 | 0.0000% | 8.39 | 10.23 | 10.23 | true |
| spec_acdc_asc_p50_next | 15 | 0 | 0.0000% | 10.43 | 16.62 | 16.62 | true |
| spec_acdc_asc_p50_previous | 15 | 0 | 0.0000% | 11.91 | 14.49 | 14.49 | true |
| spec_acdc_asc_p75_next | 15 | 0 | 0.0000% | 9.25 | 12.46 | 12.46 | true |
| spec_acdc_asc_p75_previous | 15 | 0 | 0.0000% | 8.95 | 11.49 | 11.49 | true |
| spec_acdc_asc_p99_next | 15 | 0 | 0.0000% | 9.86 | 12.24 | 12.24 | true |
| spec_acdc_asc_p99_previous | 15 | 0 | 0.0000% | 8.35 | 10.77 | 10.77 | true |
| spec_acdc_desc_p25_next | 15 | 0 | 0.0000% | 9.35 | 11.99 | 11.99 | true |
| spec_acdc_desc_p25_previous | 15 | 0 | 0.0000% | 9.82 | 12.87 | 12.87 | true |
| spec_acdc_desc_p50_next | 15 | 0 | 0.0000% | 8.48 | 10.25 | 10.25 | true |
| spec_acdc_desc_p50_previous | 15 | 0 | 0.0000% | 8.19 | 10.44 | 10.44 | true |
| spec_acdc_desc_p75_next | 15 | 0 | 0.0000% | 8.38 | 11.59 | 11.59 | true |
| spec_acdc_desc_p75_previous | 15 | 0 | 0.0000% | 9.21 | 14.08 | 14.08 | true |
| spec_acdc_desc_p99_next | 15 | 0 | 0.0000% | 16.70 | 24.75 | 24.75 | true |
| spec_acdc_desc_p99_previous | 15 | 0 | 0.0000% | 13.65 | 16.30 | 16.30 | true |
| spec_power_asc_p25_next | 15 | 0 | 0.0000% | 9.09 | 13.06 | 13.06 | true |
| spec_power_asc_p25_previous | 15 | 0 | 0.0000% | 9.62 | 12.00 | 12.00 | true |
| spec_power_asc_p50_next | 15 | 0 | 0.0000% | 9.84 | 13.54 | 13.54 | true |
| spec_power_asc_p50_previous | 15 | 0 | 0.0000% | 9.00 | 12.23 | 12.23 | true |
| spec_power_asc_p75_next | 15 | 0 | 0.0000% | 8.08 | 9.68 | 9.68 | true |
| spec_power_asc_p75_previous | 15 | 0 | 0.0000% | 9.67 | 12.23 | 12.23 | true |
| spec_power_asc_p99_next | 15 | 0 | 0.0000% | 8.77 | 10.81 | 10.81 | true |
| spec_power_asc_p99_previous | 15 | 0 | 0.0000% | 8.87 | 11.40 | 11.40 | true |
| spec_power_desc_p25_next | 15 | 0 | 0.0000% | 8.52 | 11.55 | 11.55 | true |
| spec_power_desc_p25_previous | 15 | 0 | 0.0000% | 8.70 | 10.87 | 10.87 | true |
| spec_power_desc_p50_next | 15 | 0 | 0.0000% | 9.16 | 11.18 | 11.18 | true |
| spec_power_desc_p50_previous | 15 | 0 | 0.0000% | 8.26 | 12.69 | 12.69 | true |
| spec_power_desc_p75_next | 15 | 0 | 0.0000% | 8.50 | 12.04 | 12.04 | true |
| spec_power_desc_p75_previous | 15 | 0 | 0.0000% | 8.08 | 11.34 | 11.34 | true |
| spec_power_desc_p99_next | 15 | 0 | 0.0000% | 9.17 | 11.10 | 11.10 | true |
| spec_power_desc_p99_previous | 15 | 0 | 0.0000% | 8.62 | 18.88 | 18.88 | true |
