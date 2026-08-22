# Facility benchmark (5000000 FieldDevices)

Git: `da4128276ebd2fb1ae2207675929301a8b942f91`  
Generated: 2026-08-22T15:00:44Z

## PostgreSQL

- `effective_cache_size`: `6GB`
- `fsync`: `on`
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

- `database_bytes`: `34588210879`
- `expected_field_devices`: `5000000`
- `field_device_cursor_values_index_bytes`: `11846475776`
- `field_device_cursor_values_rows`: `5000000`
- `field_device_cursor_values_table_bytes`: `991911936`
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
| created_at_asc_first | 15 | 0 | 0.0000% | 12.76 | 33.02 | 33.02 | true |
| created_at_asc_next | 15 | 0 | 0.0000% | 13.04 | 18.98 | 18.98 | true |
| created_at_desc_first | 15 | 0 | 0.0000% | 13.18 | 20.11 | 20.11 | true |
| created_at_desc_next | 15 | 0 | 0.0000% | 13.89 | 17.66 | 17.66 | true |
| apparat_nr_asc_first | 15 | 0 | 0.0000% | 11.78 | 19.55 | 19.55 | true |
| apparat_nr_asc_next | 15 | 0 | 0.0000% | 13.53 | 14.96 | 14.96 | true |
| apparat_nr_desc_first | 15 | 0 | 0.0000% | 13.80 | 24.03 | 24.03 | true |
| apparat_nr_desc_next | 15 | 0 | 0.0000% | 13.05 | 15.65 | 15.65 | true |
| sps_system_type_asc_first | 15 | 0 | 0.0000% | 19.25 | 25.22 | 25.22 | true |
| sps_system_type_asc_next | 15 | 0 | 0.0000% | 17.76 | 20.56 | 20.56 | true |
| sps_system_type_desc_first | 15 | 0 | 0.0000% | 21.92 | 62.80 | 62.80 | true |
| sps_system_type_desc_next | 15 | 0 | 0.0000% | 18.55 | 22.05 | 22.05 | true |
| description_asc_first | 15 | 0 | 0.0000% | 14.53 | 18.66 | 18.66 | true |
| description_asc_next | 15 | 0 | 0.0000% | 13.96 | 17.00 | 17.00 | true |
| description_desc_first | 15 | 0 | 0.0000% | 14.58 | 22.18 | 22.18 | true |
| description_desc_next | 15 | 0 | 0.0000% | 15.81 | 19.34 | 19.34 | true |
| spec_supplier_asc_first | 15 | 0 | 0.0000% | 14.67 | 25.65 | 25.65 | true |
| spec_supplier_asc_next | 15 | 0 | 0.0000% | 13.96 | 17.90 | 17.90 | true |
| spec_supplier_desc_first | 15 | 0 | 0.0000% | 16.63 | 29.26 | 29.26 | true |
| spec_supplier_desc_next | 15 | 0 | 0.0000% | 17.51 | 35.40 | 35.40 | true |
| spec_size_asc_first | 15 | 0 | 0.0000% | 18.17 | 22.76 | 22.76 | true |
| spec_size_asc_next | 15 | 0 | 0.0000% | 18.86 | 38.60 | 38.60 | true |
| spec_size_desc_first | 15 | 0 | 0.0000% | 17.79 | 19.47 | 19.47 | true |
| spec_size_desc_next | 15 | 0 | 0.0000% | 18.06 | 20.62 | 20.62 | true |
| spec_acdc_asc_first | 15 | 0 | 0.0000% | 25.18 | 27.85 | 27.85 | true |
| spec_acdc_asc_next | 15 | 0 | 0.0000% | 15.28 | 21.24 | 21.24 | true |
| spec_acdc_desc_first | 15 | 0 | 0.0000% | 15.23 | 20.80 | 20.80 | true |
| spec_acdc_desc_next | 15 | 0 | 0.0000% | 13.62 | 15.21 | 15.21 | true |
| spec_power_asc_first | 15 | 0 | 0.0000% | 15.54 | 22.60 | 22.60 | true |
| spec_power_asc_next | 15 | 0 | 0.0000% | 12.68 | 16.03 | 16.03 | true |
| spec_power_desc_first | 15 | 0 | 0.0000% | 13.26 | 14.05 | 14.05 | true |
| spec_power_desc_next | 15 | 0 | 0.0000% | 13.85 | 16.29 | 16.29 | true |
| building_scope_first | 15 | 0 | 0.0000% | 17.82 | 19.76 | 19.76 | true |
| building_scope_previous | 15 | 0 | 0.0000% | 18.18 | 25.16 | 25.16 | true |
| building_scope_next | 15 | 0 | 0.0000% | 16.99 | 18.41 | 18.41 | true |
| cabinet_scope_first | 15 | 0 | 0.0000% | 33.52 | 41.34 | 41.34 | true |
| cabinet_scope_previous | 15 | 0 | 0.0000% | 16.80 | 21.03 | 21.03 | true |
| cabinet_scope_next | 15 | 0 | 0.0000% | 24.76 | 31.94 | 31.94 | true |
| controller_scope_first | 15 | 0 | 0.0000% | 16.02 | 19.32 | 19.32 | true |
| controller_scope_previous | 15 | 0 | 0.0000% | 15.74 | 17.87 | 17.87 | true |
| controller_scope_next | 15 | 0 | 0.0000% | 14.07 | 16.71 | 16.71 | true |
| project_scope_first | 15 | 0 | 0.0000% | 16.94 | 40.95 | 40.95 | true |
| project_scope_previous | 15 | 0 | 0.0000% | 18.90 | 23.57 | 23.57 | true |
| project_scope_next | 15 | 0 | 0.0000% | 17.50 | 21.30 | 21.30 | true |
| search_0_1_percent_first | 15 | 0 | 0.0000% | 60.50 | 88.60 | 88.60 | true |
| search_0_1_percent_previous | 15 | 0 | 0.0000% | 40.14 | 93.14 | 93.14 | true |
| search_0_1_percent_next | 15 | 0 | 0.0000% | 53.11 | 69.43 | 69.43 | true |
| search_1_percent_first | 15 | 0 | 0.0000% | 422.21 | 465.01 | 465.01 | true |
| search_1_percent_previous | 15 | 0 | 0.0000% | 43.09 | 52.58 | 52.58 | true |
| search_1_percent_next | 15 | 0 | 0.0000% | 426.72 | 454.05 | 454.05 | true |
| search_10_percent_first | 15 | 0 | 0.0000% | 22.90 | 27.53 | 27.53 | true |
| search_10_percent_previous | 15 | 0 | 0.0000% | 134.69 | 153.54 | 153.54 | true |
| search_10_percent_next | 15 | 0 | 0.0000% | 21.15 | 22.93 | 22.93 | true |
| combined_filter_first | 15 | 0 | 0.0000% | 10.82 | 14.27 | 14.27 | true |
| combined_filter_previous | 15 | 0 | 0.0000% | 13.69 | 15.73 | 15.73 | true |
| combined_filter_next | 15 | 0 | 0.0000% | 9.92 | 11.32 | 11.32 | true |
| created_at_asc_p25_next | 15 | 0 | 0.0000% | 11.88 | 13.75 | 13.75 | true |
| created_at_asc_p25_previous | 15 | 0 | 0.0000% | 10.04 | 34.13 | 34.13 | true |
| created_at_asc_p50_next | 15 | 0 | 0.0000% | 16.22 | 63.92 | 63.92 | true |
| created_at_asc_p50_previous | 15 | 0 | 0.0000% | 28.48 | 56.12 | 56.12 | true |
| created_at_asc_p75_next | 15 | 0 | 0.0000% | 11.32 | 14.11 | 14.11 | true |
| created_at_asc_p75_previous | 15 | 0 | 0.0000% | 11.90 | 14.72 | 14.72 | true |
| created_at_asc_p99_next | 15 | 0 | 0.0000% | 14.85 | 18.16 | 18.16 | true |
| created_at_asc_p99_previous | 15 | 0 | 0.0000% | 12.76 | 15.39 | 15.39 | true |
| created_at_desc_p25_next | 15 | 0 | 0.0000% | 12.32 | 15.83 | 15.83 | true |
| created_at_desc_p25_previous | 15 | 0 | 0.0000% | 11.46 | 14.20 | 14.20 | true |
| created_at_desc_p50_next | 15 | 0 | 0.0000% | 12.57 | 15.24 | 15.24 | true |
| created_at_desc_p50_previous | 15 | 0 | 0.0000% | 11.79 | 15.14 | 15.14 | true |
| created_at_desc_p75_next | 15 | 0 | 0.0000% | 11.45 | 13.82 | 13.82 | true |
| created_at_desc_p75_previous | 15 | 0 | 0.0000% | 11.32 | 13.96 | 13.96 | true |
| created_at_desc_p99_next | 15 | 0 | 0.0000% | 12.71 | 14.78 | 14.78 | true |
| created_at_desc_p99_previous | 15 | 0 | 0.0000% | 15.53 | 18.59 | 18.59 | true |
| apparat_nr_asc_p25_next | 15 | 0 | 0.0000% | 11.70 | 14.33 | 14.33 | true |
| apparat_nr_asc_p25_previous | 15 | 0 | 0.0000% | 13.39 | 18.06 | 18.06 | true |
| apparat_nr_asc_p50_next | 15 | 0 | 0.0000% | 10.91 | 14.60 | 14.60 | true |
| apparat_nr_asc_p50_previous | 15 | 0 | 0.0000% | 10.72 | 322.21 | 322.21 | true |
| apparat_nr_asc_p75_next | 15 | 0 | 0.0000% | 11.26 | 14.77 | 14.77 | true |
| apparat_nr_asc_p75_previous | 15 | 0 | 0.0000% | 12.28 | 15.50 | 15.50 | true |
| apparat_nr_asc_p99_next | 15 | 0 | 0.0000% | 14.59 | 23.21 | 23.21 | true |
| apparat_nr_asc_p99_previous | 15 | 0 | 0.0000% | 15.82 | 58.37 | 58.37 | true |
| apparat_nr_desc_p25_next | 15 | 0 | 0.0000% | 11.13 | 13.22 | 13.22 | true |
| apparat_nr_desc_p25_previous | 15 | 0 | 0.0000% | 10.79 | 13.63 | 13.63 | true |
| apparat_nr_desc_p50_next | 15 | 0 | 0.0000% | 11.17 | 15.00 | 15.00 | true |
| apparat_nr_desc_p50_previous | 15 | 0 | 0.0000% | 11.22 | 14.81 | 14.81 | true |
| apparat_nr_desc_p75_next | 15 | 0 | 0.0000% | 11.27 | 13.70 | 13.70 | true |
| apparat_nr_desc_p75_previous | 15 | 0 | 0.0000% | 12.44 | 49.26 | 49.26 | true |
| apparat_nr_desc_p99_next | 15 | 0 | 0.0000% | 13.04 | 14.21 | 14.21 | true |
| apparat_nr_desc_p99_previous | 15 | 0 | 0.0000% | 11.93 | 14.66 | 14.66 | true |
| sps_system_type_asc_p25_next | 15 | 0 | 0.0000% | 15.84 | 18.48 | 18.48 | true |
| sps_system_type_asc_p25_previous | 15 | 0 | 0.0000% | 15.72 | 33.81 | 33.81 | true |
| sps_system_type_asc_p50_next | 15 | 0 | 0.0000% | 16.60 | 20.83 | 20.83 | true |
| sps_system_type_asc_p50_previous | 15 | 0 | 0.0000% | 27.21 | 32.83 | 32.83 | true |
| sps_system_type_asc_p75_next | 15 | 0 | 0.0000% | 15.19 | 19.30 | 19.30 | true |
| sps_system_type_asc_p75_previous | 15 | 0 | 0.0000% | 16.88 | 18.09 | 18.09 | true |
| sps_system_type_asc_p99_next | 15 | 0 | 0.0000% | 14.97 | 22.93 | 22.93 | true |
| sps_system_type_asc_p99_previous | 15 | 0 | 0.0000% | 18.52 | 22.88 | 22.88 | true |
| sps_system_type_desc_p25_next | 15 | 0 | 0.0000% | 47.09 | 63.88 | 63.88 | true |
| sps_system_type_desc_p25_previous | 15 | 0 | 0.0000% | 95.18 | 119.56 | 119.56 | true |
| sps_system_type_desc_p50_next | 15 | 0 | 0.0000% | 71.34 | 80.08 | 80.08 | true |
| sps_system_type_desc_p50_previous | 15 | 0 | 0.0000% | 74.44 | 97.30 | 97.30 | true |
| sps_system_type_desc_p75_next | 15 | 0 | 0.0000% | 106.69 | 149.73 | 149.73 | true |
| sps_system_type_desc_p75_previous | 15 | 0 | 0.0000% | 43.56 | 52.61 | 52.61 | true |
| sps_system_type_desc_p99_next | 15 | 0 | 0.0000% | 125.67 | 167.14 | 167.14 | true |
| sps_system_type_desc_p99_previous | 15 | 0 | 0.0000% | 17.25 | 20.62 | 20.62 | true |
| description_asc_p25_next | 15 | 0 | 0.0000% | 19.78 | 31.27 | 31.27 | true |
| description_asc_p25_previous | 15 | 0 | 0.0000% | 15.55 | 24.34 | 24.34 | true |
| description_asc_p50_next | 15 | 0 | 0.0000% | 19.19 | 25.86 | 25.86 | true |
| description_asc_p50_previous | 15 | 0 | 0.0000% | 12.78 | 14.14 | 14.14 | true |
| description_asc_p75_next | 15 | 0 | 0.0000% | 11.85 | 17.32 | 17.32 | true |
| description_asc_p75_previous | 15 | 0 | 0.0000% | 11.70 | 14.09 | 14.09 | true |
| description_asc_p99_next | 15 | 0 | 0.0000% | 11.60 | 13.90 | 13.90 | true |
| description_asc_p99_previous | 15 | 0 | 0.0000% | 12.06 | 14.82 | 14.82 | true |
| description_desc_p25_next | 15 | 0 | 0.0000% | 12.96 | 16.53 | 16.53 | true |
| description_desc_p25_previous | 15 | 0 | 0.0000% | 20.77 | 26.76 | 26.76 | true |
| description_desc_p50_next | 15 | 0 | 0.0000% | 11.46 | 14.32 | 14.32 | true |
| description_desc_p50_previous | 15 | 0 | 0.0000% | 15.09 | 23.42 | 23.42 | true |
| description_desc_p75_next | 15 | 0 | 0.0000% | 16.91 | 22.98 | 22.98 | true |
| description_desc_p75_previous | 15 | 0 | 0.0000% | 11.01 | 14.62 | 14.62 | true |
| description_desc_p99_next | 15 | 0 | 0.0000% | 13.99 | 21.36 | 21.36 | true |
| description_desc_p99_previous | 15 | 0 | 0.0000% | 13.87 | 15.86 | 15.86 | true |
| spec_supplier_asc_p25_next | 15 | 0 | 0.0000% | 13.06 | 32.16 | 32.16 | true |
| spec_supplier_asc_p25_previous | 15 | 0 | 0.0000% | 12.39 | 17.32 | 17.32 | true |
| spec_supplier_asc_p50_next | 15 | 0 | 0.0000% | 12.17 | 14.57 | 14.57 | true |
| spec_supplier_asc_p50_previous | 15 | 0 | 0.0000% | 11.82 | 16.55 | 16.55 | true |
| spec_supplier_asc_p75_next | 15 | 0 | 0.0000% | 12.49 | 14.78 | 14.78 | true |
| spec_supplier_asc_p75_previous | 15 | 0 | 0.0000% | 12.63 | 15.26 | 15.26 | true |
| spec_supplier_asc_p99_next | 15 | 0 | 0.0000% | 13.32 | 18.01 | 18.01 | true |
| spec_supplier_asc_p99_previous | 15 | 0 | 0.0000% | 12.79 | 15.96 | 15.96 | true |
| spec_supplier_desc_p25_next | 15 | 0 | 0.0000% | 11.81 | 16.64 | 16.64 | true |
| spec_supplier_desc_p25_previous | 15 | 0 | 0.0000% | 11.83 | 14.73 | 14.73 | true |
| spec_supplier_desc_p50_next | 15 | 0 | 0.0000% | 13.41 | 16.15 | 16.15 | true |
| spec_supplier_desc_p50_previous | 15 | 0 | 0.0000% | 12.69 | 13.69 | 13.69 | true |
| spec_supplier_desc_p75_next | 15 | 0 | 0.0000% | 13.55 | 15.81 | 15.81 | true |
| spec_supplier_desc_p75_previous | 15 | 0 | 0.0000% | 12.64 | 14.99 | 14.99 | true |
| spec_supplier_desc_p99_next | 15 | 0 | 0.0000% | 13.66 | 15.58 | 15.58 | true |
| spec_supplier_desc_p99_previous | 15 | 0 | 0.0000% | 13.68 | 24.23 | 24.23 | true |
| spec_size_asc_p25_next | 15 | 0 | 0.0000% | 16.72 | 22.14 | 22.14 | true |
| spec_size_asc_p25_previous | 15 | 0 | 0.0000% | 16.90 | 18.77 | 18.77 | true |
| spec_size_asc_p50_next | 15 | 0 | 0.0000% | 19.60 | 22.57 | 22.57 | true |
| spec_size_asc_p50_previous | 15 | 0 | 0.0000% | 17.38 | 20.74 | 20.74 | true |
| spec_size_asc_p75_next | 15 | 0 | 0.0000% | 20.14 | 27.29 | 27.29 | true |
| spec_size_asc_p75_previous | 15 | 0 | 0.0000% | 15.00 | 42.87 | 42.87 | true |
| spec_size_asc_p99_next | 15 | 0 | 0.0000% | 16.94 | 21.71 | 21.71 | true |
| spec_size_asc_p99_previous | 15 | 0 | 0.0000% | 15.09 | 18.91 | 18.91 | true |
| spec_size_desc_p25_next | 15 | 0 | 0.0000% | 17.07 | 47.88 | 47.88 | true |
| spec_size_desc_p25_previous | 15 | 0 | 0.0000% | 15.91 | 18.02 | 18.02 | true |
| spec_size_desc_p50_next | 15 | 0 | 0.0000% | 16.33 | 21.06 | 21.06 | true |
| spec_size_desc_p50_previous | 15 | 0 | 0.0000% | 15.82 | 19.50 | 19.50 | true |
| spec_size_desc_p75_next | 15 | 0 | 0.0000% | 16.30 | 21.88 | 21.88 | true |
| spec_size_desc_p75_previous | 15 | 0 | 0.0000% | 15.11 | 20.53 | 20.53 | true |
| spec_size_desc_p99_next | 15 | 0 | 0.0000% | 16.94 | 18.41 | 18.41 | true |
| spec_size_desc_p99_previous | 15 | 0 | 0.0000% | 18.83 | 24.62 | 24.62 | true |
| spec_acdc_asc_p25_next | 15 | 0 | 0.0000% | 12.73 | 16.00 | 16.00 | true |
| spec_acdc_asc_p25_previous | 15 | 0 | 0.0000% | 12.78 | 17.00 | 17.00 | true |
| spec_acdc_asc_p50_next | 15 | 0 | 0.0000% | 14.43 | 21.25 | 21.25 | true |
| spec_acdc_asc_p50_previous | 15 | 0 | 0.0000% | 12.63 | 17.73 | 17.73 | true |
| spec_acdc_asc_p75_next | 15 | 0 | 0.0000% | 12.67 | 15.28 | 15.28 | true |
| spec_acdc_asc_p75_previous | 15 | 0 | 0.0000% | 11.78 | 94.83 | 94.83 | true |
| spec_acdc_asc_p99_next | 15 | 0 | 0.0000% | 17.36 | 43.21 | 43.21 | true |
| spec_acdc_asc_p99_previous | 15 | 0 | 0.0000% | 12.07 | 17.02 | 17.02 | true |
| spec_acdc_desc_p25_next | 15 | 0 | 0.0000% | 12.04 | 15.44 | 15.44 | true |
| spec_acdc_desc_p25_previous | 15 | 0 | 0.0000% | 12.81 | 18.64 | 18.64 | true |
| spec_acdc_desc_p50_next | 15 | 0 | 0.0000% | 12.91 | 14.60 | 14.60 | true |
| spec_acdc_desc_p50_previous | 15 | 0 | 0.0000% | 14.08 | 17.91 | 17.91 | true |
| spec_acdc_desc_p75_next | 15 | 0 | 0.0000% | 12.30 | 16.48 | 16.48 | true |
| spec_acdc_desc_p75_previous | 15 | 0 | 0.0000% | 14.75 | 18.68 | 18.68 | true |
| spec_acdc_desc_p99_next | 15 | 0 | 0.0000% | 11.82 | 16.12 | 16.12 | true |
| spec_acdc_desc_p99_previous | 15 | 0 | 0.0000% | 12.86 | 15.29 | 15.29 | true |
| spec_power_asc_p25_next | 15 | 0 | 0.0000% | 12.34 | 21.65 | 21.65 | true |
| spec_power_asc_p25_previous | 15 | 0 | 0.0000% | 12.00 | 14.62 | 14.62 | true |
| spec_power_asc_p50_next | 15 | 0 | 0.0000% | 11.42 | 14.80 | 14.80 | true |
| spec_power_asc_p50_previous | 15 | 0 | 0.0000% | 12.44 | 15.48 | 15.48 | true |
| spec_power_asc_p75_next | 15 | 0 | 0.0000% | 12.06 | 15.47 | 15.47 | true |
| spec_power_asc_p75_previous | 15 | 0 | 0.0000% | 14.57 | 24.36 | 24.36 | true |
| spec_power_asc_p99_next | 15 | 0 | 0.0000% | 12.40 | 15.31 | 15.31 | true |
| spec_power_asc_p99_previous | 15 | 0 | 0.0000% | 11.60 | 13.48 | 13.48 | true |
| spec_power_desc_p25_next | 15 | 0 | 0.0000% | 17.73 | 20.75 | 20.75 | true |
| spec_power_desc_p25_previous | 15 | 0 | 0.0000% | 13.74 | 19.53 | 19.53 | true |
| spec_power_desc_p50_next | 15 | 0 | 0.0000% | 12.39 | 15.32 | 15.32 | true |
| spec_power_desc_p50_previous | 15 | 0 | 0.0000% | 23.47 | 27.47 | 27.47 | true |
| spec_power_desc_p75_next | 15 | 0 | 0.0000% | 12.92 | 20.98 | 20.98 | true |
| spec_power_desc_p75_previous | 15 | 0 | 0.0000% | 11.39 | 14.63 | 14.63 | true |
| spec_power_desc_p99_next | 15 | 0 | 0.0000% | 12.25 | 14.21 | 14.21 | true |
| spec_power_desc_p99_previous | 15 | 0 | 0.0000% | 11.69 | 13.87 | 13.87 | true |
