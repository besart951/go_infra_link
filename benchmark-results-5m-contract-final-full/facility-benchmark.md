# Facility benchmark (5000000 FieldDevices)

Git: `da4128276ebd2fb1ae2207675929301a8b942f91`  
Generated: 2026-08-22T16:43:14Z

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
| created_at_asc_first | 900 | 0 | 0.0000% | 19.94 | 48.55 | 67.97 | true |
| created_at_asc_next | 900 | 0 | 0.0000% | 21.10 | 49.32 | 86.86 | true |
| created_at_desc_first | 900 | 0 | 0.0000% | 22.44 | 54.01 | 77.07 | true |
| created_at_desc_next | 900 | 0 | 0.0000% | 21.94 | 49.87 | 62.31 | true |
| apparat_nr_asc_first | 900 | 0 | 0.0000% | 21.62 | 53.29 | 71.01 | true |
| apparat_nr_asc_next | 900 | 0 | 0.0000% | 19.51 | 44.02 | 66.06 | true |
| apparat_nr_desc_first | 900 | 0 | 0.0000% | 20.02 | 46.58 | 81.94 | true |
| apparat_nr_desc_next | 900 | 0 | 0.0000% | 21.16 | 51.00 | 79.56 | true |
| sps_system_type_asc_first | 900 | 0 | 0.0000% | 24.45 | 53.38 | 80.10 | true |
| sps_system_type_asc_next | 900 | 0 | 0.0000% | 27.17 | 59.67 | 92.81 | true |
| sps_system_type_desc_first | 900 | 0 | 0.0000% | 27.32 | 63.33 | 152.87 | true |
| sps_system_type_desc_next | 900 | 0 | 0.0000% | 26.84 | 53.95 | 75.37 | true |
| description_asc_first | 900 | 0 | 0.0000% | 22.26 | 50.10 | 73.78 | true |
| description_asc_next | 900 | 0 | 0.0000% | 20.70 | 48.28 | 65.70 | true |
| description_desc_first | 900 | 0 | 0.0000% | 24.58 | 54.62 | 69.14 | true |
| description_desc_next | 900 | 0 | 0.0000% | 22.35 | 49.91 | 78.12 | true |
| spec_supplier_asc_first | 900 | 0 | 0.0000% | 22.23 | 47.82 | 74.60 | true |
| spec_supplier_asc_next | 900 | 0 | 0.0000% | 23.41 | 54.65 | 85.79 | true |
| spec_supplier_desc_first | 900 | 0 | 0.0000% | 27.65 | 62.32 | 94.22 | true |
| spec_supplier_desc_next | 900 | 0 | 0.0000% | 26.11 | 57.98 | 83.85 | true |
| spec_size_asc_first | 900 | 0 | 0.0000% | 27.45 | 54.35 | 83.71 | true |
| spec_size_asc_next | 900 | 0 | 0.0000% | 27.99 | 58.47 | 94.78 | true |
| spec_size_desc_first | 900 | 0 | 0.0000% | 25.76 | 56.76 | 104.59 | true |
| spec_size_desc_next | 900 | 0 | 0.0000% | 29.15 | 62.61 | 90.80 | true |
| spec_acdc_asc_first | 900 | 0 | 0.0000% | 25.12 | 64.96 | 104.55 | true |
| spec_acdc_asc_next | 900 | 0 | 0.0000% | 24.87 | 66.90 | 129.37 | true |
| spec_acdc_desc_first | 900 | 0 | 0.0000% | 22.23 | 41.94 | 70.56 | true |
| spec_acdc_desc_next | 900 | 0 | 0.0000% | 21.54 | 53.10 | 76.41 | true |
| spec_power_asc_first | 900 | 0 | 0.0000% | 22.80 | 51.96 | 74.18 | true |
| spec_power_asc_next | 900 | 0 | 0.0000% | 23.94 | 53.29 | 68.56 | true |
| spec_power_desc_first | 900 | 0 | 0.0000% | 19.09 | 50.10 | 81.34 | true |
| spec_power_desc_next | 900 | 0 | 0.0000% | 20.86 | 49.86 | 70.01 | true |
| building_scope_first | 900 | 0 | 0.0000% | 26.69 | 53.30 | 146.15 | true |
| building_scope_previous | 900 | 0 | 0.0000% | 26.20 | 58.99 | 91.54 | true |
| building_scope_next | 900 | 0 | 0.0000% | 26.87 | 55.29 | 92.09 | true |
| cabinet_scope_first | 900 | 0 | 0.0000% | 37.93 | 85.69 | 111.92 | true |
| cabinet_scope_previous | 900 | 0 | 0.0000% | 27.78 | 63.75 | 101.48 | true |
| cabinet_scope_next | 900 | 0 | 0.0000% | 30.82 | 66.39 | 86.41 | true |
| controller_scope_first | 900 | 0 | 0.0000% | 25.75 | 52.96 | 76.53 | true |
| controller_scope_previous | 900 | 0 | 0.0000% | 24.75 | 57.22 | 80.63 | true |
| controller_scope_next | 900 | 0 | 0.0000% | 19.24 | 42.96 | 64.78 | true |
| project_scope_first | 900 | 0 | 0.0000% | 28.00 | 58.22 | 96.85 | true |
| project_scope_previous | 900 | 0 | 0.0000% | 28.02 | 54.65 | 79.25 | true |
| project_scope_next | 900 | 0 | 0.0000% | 28.21 | 56.48 | 79.40 | true |
| search_0_1_percent_first | 900 | 0 | 0.0000% | 196.50 | 335.08 | 381.36 | true |
| search_0_1_percent_previous | 900 | 0 | 0.0000% | 230.46 | 416.24 | 482.85 | true |
| search_0_1_percent_next | 900 | 0 | 0.0000% | 202.97 | 378.76 | 445.33 | true |
| search_1_percent_first | 900 | 0 | 0.0000% | 116.89 | 247.16 | 299.19 | true |
| search_1_percent_previous | 900 | 0 | 0.0000% | 198.16 | 373.15 | 418.62 | true |
| search_1_percent_next | 900 | 0 | 0.0000% | 116.36 | 270.91 | 303.38 | true |
| search_10_percent_first | 900 | 0 | 0.0000% | 111.66 | 242.34 | 279.22 | true |
| search_10_percent_previous | 900 | 0 | 0.0000% | 25.76 | 51.11 | 143.50 | true |
| search_10_percent_next | 900 | 0 | 0.0000% | 108.99 | 251.40 | 293.58 | true |
| combined_filter_first | 900 | 0 | 0.0000% | 22.20 | 51.08 | 68.60 | true |
| combined_filter_previous | 900 | 0 | 0.0000% | 23.73 | 57.29 | 88.87 | true |
| combined_filter_next | 900 | 0 | 0.0000% | 21.08 | 45.40 | 62.14 | true |
| created_at_asc_p25_next | 900 | 0 | 0.0000% | 18.13 | 38.50 | 50.29 | true |
| created_at_asc_p25_previous | 900 | 0 | 0.0000% | 17.75 | 33.41 | 55.55 | true |
| created_at_asc_p50_next | 900 | 0 | 0.0000% | 17.43 | 37.27 | 58.06 | true |
| created_at_asc_p50_previous | 900 | 0 | 0.0000% | 18.05 | 35.27 | 54.08 | true |
| created_at_asc_p75_next | 900 | 0 | 0.0000% | 19.30 | 45.40 | 70.39 | true |
| created_at_asc_p75_previous | 900 | 0 | 0.0000% | 18.45 | 42.54 | 86.40 | true |
| created_at_asc_p99_next | 900 | 0 | 0.0000% | 18.07 | 34.69 | 58.94 | true |
| created_at_asc_p99_previous | 900 | 0 | 0.0000% | 20.34 | 46.16 | 69.71 | true |
| created_at_desc_p25_next | 900 | 0 | 0.0000% | 20.34 | 43.29 | 71.51 | true |
| created_at_desc_p25_previous | 900 | 0 | 0.0000% | 19.32 | 37.96 | 52.96 | true |
| created_at_desc_p50_next | 900 | 0 | 0.0000% | 20.42 | 46.36 | 82.25 | true |
| created_at_desc_p50_previous | 900 | 0 | 0.0000% | 19.82 | 39.34 | 65.88 | true |
| created_at_desc_p75_next | 900 | 0 | 0.0000% | 18.84 | 49.06 | 86.07 | true |
| created_at_desc_p75_previous | 900 | 0 | 0.0000% | 20.12 | 44.49 | 77.16 | true |
| created_at_desc_p99_next | 900 | 0 | 0.0000% | 19.85 | 44.81 | 63.20 | true |
| created_at_desc_p99_previous | 900 | 0 | 0.0000% | 24.63 | 59.47 | 109.47 | true |
| apparat_nr_asc_p25_next | 900 | 0 | 0.0000% | 20.79 | 40.29 | 61.48 | true |
| apparat_nr_asc_p25_previous | 900 | 0 | 0.0000% | 21.16 | 51.72 | 77.03 | true |
| apparat_nr_asc_p50_next | 900 | 0 | 0.0000% | 19.71 | 43.72 | 62.79 | true |
| apparat_nr_asc_p50_previous | 900 | 0 | 0.0000% | 19.42 | 43.02 | 63.98 | true |
| apparat_nr_asc_p75_next | 900 | 0 | 0.0000% | 23.30 | 52.58 | 83.21 | true |
| apparat_nr_asc_p75_previous | 900 | 0 | 0.0000% | 23.28 | 44.38 | 77.20 | true |
| apparat_nr_asc_p99_next | 900 | 0 | 0.0000% | 21.44 | 51.88 | 67.89 | true |
| apparat_nr_asc_p99_previous | 900 | 0 | 0.0000% | 21.43 | 49.47 | 65.18 | true |
| apparat_nr_desc_p25_next | 900 | 0 | 0.0000% | 20.79 | 46.56 | 63.15 | true |
| apparat_nr_desc_p25_previous | 900 | 0 | 0.0000% | 19.79 | 45.01 | 70.60 | true |
| apparat_nr_desc_p50_next | 900 | 0 | 0.0000% | 20.89 | 42.30 | 65.01 | true |
| apparat_nr_desc_p50_previous | 900 | 0 | 0.0000% | 19.89 | 43.76 | 87.10 | true |
| apparat_nr_desc_p75_next | 900 | 0 | 0.0000% | 20.13 | 48.30 | 62.85 | true |
| apparat_nr_desc_p75_previous | 900 | 0 | 0.0000% | 20.54 | 46.59 | 59.82 | true |
| apparat_nr_desc_p99_next | 900 | 0 | 0.0000% | 19.24 | 40.78 | 64.36 | true |
| apparat_nr_desc_p99_previous | 900 | 0 | 0.0000% | 17.69 | 43.84 | 171.88 | true |
| sps_system_type_asc_p25_next | 900 | 0 | 0.0000% | 29.07 | 62.13 | 86.76 | true |
| sps_system_type_asc_p25_previous | 900 | 0 | 0.0000% | 26.92 | 54.53 | 82.28 | true |
| sps_system_type_asc_p50_next | 900 | 0 | 0.0000% | 27.58 | 55.51 | 89.74 | true |
| sps_system_type_asc_p50_previous | 900 | 0 | 0.0000% | 27.38 | 54.58 | 84.58 | true |
| sps_system_type_asc_p75_next | 900 | 0 | 0.0000% | 24.95 | 51.52 | 79.52 | true |
| sps_system_type_asc_p75_previous | 900 | 0 | 0.0000% | 25.58 | 53.08 | 91.85 | true |
| sps_system_type_asc_p99_next | 900 | 0 | 0.0000% | 26.21 | 58.59 | 93.35 | true |
| sps_system_type_asc_p99_previous | 900 | 0 | 0.0000% | 37.50 | 90.29 | 117.09 | true |
| sps_system_type_desc_p25_next | 900 | 0 | 0.0000% | 134.90 | 247.61 | 296.42 | true |
| sps_system_type_desc_p25_previous | 900 | 0 | 0.0000% | 353.18 | 502.65 | 576.52 | true |
| sps_system_type_desc_p50_next | 900 | 0 | 0.0000% | 258.01 | 401.96 | 457.08 | true |
| sps_system_type_desc_p50_previous | 900 | 0 | 0.0000% | 243.11 | 390.20 | 457.38 | true |
| sps_system_type_desc_p75_next | 900 | 0 | 0.0000% | 383.97 | 562.51 | 608.16 | true |
| sps_system_type_desc_p75_previous | 900 | 0 | 0.0000% | 117.79 | 240.74 | 281.48 | true |
| sps_system_type_desc_p99_next | 900 | 0 | 0.0000% | 478.05 | 679.22 | 775.82 | true |
| sps_system_type_desc_p99_previous | 900 | 0 | 0.0000% | 22.94 | 52.45 | 88.24 | true |
| description_asc_p25_next | 900 | 0 | 0.0000% | 18.20 | 43.01 | 78.50 | true |
| description_asc_p25_previous | 900 | 0 | 0.0000% | 17.02 | 34.83 | 63.55 | true |
| description_asc_p50_next | 900 | 0 | 0.0000% | 17.93 | 39.54 | 57.73 | true |
| description_asc_p50_previous | 900 | 0 | 0.0000% | 18.21 | 40.41 | 58.24 | true |
| description_asc_p75_next | 900 | 0 | 0.0000% | 17.21 | 35.09 | 62.41 | true |
| description_asc_p75_previous | 900 | 0 | 0.0000% | 17.59 | 38.17 | 66.68 | true |
| description_asc_p99_next | 900 | 0 | 0.0000% | 17.42 | 38.59 | 86.60 | true |
| description_asc_p99_previous | 900 | 0 | 0.0000% | 18.80 | 41.84 | 62.63 | true |
| description_desc_p25_next | 900 | 0 | 0.0000% | 18.41 | 44.74 | 79.34 | true |
| description_desc_p25_previous | 900 | 0 | 0.0000% | 17.83 | 38.95 | 56.78 | true |
| description_desc_p50_next | 900 | 0 | 0.0000% | 19.29 | 36.48 | 58.54 | true |
| description_desc_p50_previous | 900 | 0 | 0.0000% | 17.51 | 34.52 | 57.53 | true |
| description_desc_p75_next | 900 | 0 | 0.0000% | 20.44 | 43.54 | 59.63 | true |
| description_desc_p75_previous | 900 | 0 | 0.0000% | 20.57 | 37.34 | 61.40 | true |
| description_desc_p99_next | 900 | 0 | 0.0000% | 18.13 | 40.37 | 63.99 | true |
| description_desc_p99_previous | 900 | 0 | 0.0000% | 17.29 | 41.13 | 76.79 | true |
| spec_supplier_asc_p25_next | 900 | 0 | 0.0000% | 21.07 | 44.62 | 73.96 | true |
| spec_supplier_asc_p25_previous | 900 | 0 | 0.0000% | 20.86 | 42.67 | 66.06 | true |
| spec_supplier_asc_p50_next | 900 | 0 | 0.0000% | 20.38 | 40.78 | 65.51 | true |
| spec_supplier_asc_p50_previous | 900 | 0 | 0.0000% | 20.12 | 37.20 | 64.06 | true |
| spec_supplier_asc_p75_next | 900 | 0 | 0.0000% | 21.40 | 47.57 | 82.46 | true |
| spec_supplier_asc_p75_previous | 900 | 0 | 0.0000% | 19.99 | 38.86 | 59.36 | true |
| spec_supplier_asc_p99_next | 900 | 0 | 0.0000% | 21.68 | 44.35 | 67.16 | true |
| spec_supplier_asc_p99_previous | 900 | 0 | 0.0000% | 23.92 | 51.93 | 77.83 | true |
| spec_supplier_desc_p25_next | 900 | 0 | 0.0000% | 24.99 | 53.71 | 138.56 | true |
| spec_supplier_desc_p25_previous | 900 | 0 | 0.0000% | 23.30 | 54.68 | 89.43 | true |
| spec_supplier_desc_p50_next | 900 | 0 | 0.0000% | 22.62 | 50.80 | 73.66 | true |
| spec_supplier_desc_p50_previous | 900 | 0 | 0.0000% | 21.64 | 50.88 | 79.78 | true |
| spec_supplier_desc_p75_next | 900 | 0 | 0.0000% | 21.62 | 49.29 | 66.36 | true |
| spec_supplier_desc_p75_previous | 900 | 0 | 0.0000% | 19.42 | 42.85 | 68.20 | true |
| spec_supplier_desc_p99_next | 900 | 0 | 0.0000% | 21.69 | 44.50 | 65.64 | true |
| spec_supplier_desc_p99_previous | 900 | 0 | 0.0000% | 22.34 | 42.64 | 65.17 | true |
| spec_size_asc_p25_next | 900 | 0 | 0.0000% | 25.52 | 55.36 | 96.83 | true |
| spec_size_asc_p25_previous | 900 | 0 | 0.0000% | 25.64 | 57.11 | 86.56 | true |
| spec_size_asc_p50_next | 900 | 0 | 0.0000% | 25.48 | 54.08 | 86.00 | true |
| spec_size_asc_p50_previous | 900 | 0 | 0.0000% | 25.19 | 52.42 | 87.08 | true |
| spec_size_asc_p75_next | 900 | 0 | 0.0000% | 26.02 | 52.37 | 77.04 | true |
| spec_size_asc_p75_previous | 900 | 0 | 0.0000% | 25.54 | 55.11 | 84.24 | true |
| spec_size_asc_p99_next | 900 | 0 | 0.0000% | 26.48 | 55.44 | 86.41 | true |
| spec_size_asc_p99_previous | 900 | 0 | 0.0000% | 27.16 | 59.67 | 92.23 | true |
| spec_size_desc_p25_next | 900 | 0 | 0.0000% | 24.66 | 52.92 | 70.59 | true |
| spec_size_desc_p25_previous | 900 | 0 | 0.0000% | 31.61 | 62.93 | 94.06 | true |
| spec_size_desc_p50_next | 900 | 0 | 0.0000% | 28.30 | 62.38 | 87.78 | true |
| spec_size_desc_p50_previous | 900 | 0 | 0.0000% | 27.78 | 55.64 | 82.88 | true |
| spec_size_desc_p75_next | 900 | 0 | 0.0000% | 28.05 | 64.08 | 101.54 | true |
| spec_size_desc_p75_previous | 900 | 0 | 0.0000% | 25.67 | 55.30 | 96.44 | true |
| spec_size_desc_p99_next | 900 | 0 | 0.0000% | 24.25 | 50.06 | 81.63 | true |
| spec_size_desc_p99_previous | 900 | 0 | 0.0000% | 26.70 | 57.44 | 79.38 | true |
| spec_acdc_asc_p25_next | 900 | 0 | 0.0000% | 20.36 | 43.64 | 69.66 | true |
| spec_acdc_asc_p25_previous | 900 | 0 | 0.0000% | 20.98 | 43.81 | 77.10 | true |
| spec_acdc_asc_p50_next | 900 | 0 | 0.0000% | 19.22 | 35.71 | 72.55 | true |
| spec_acdc_asc_p50_previous | 900 | 0 | 0.0000% | 22.03 | 48.44 | 66.64 | true |
| spec_acdc_asc_p75_next | 900 | 0 | 0.0000% | 20.52 | 46.94 | 74.29 | true |
| spec_acdc_asc_p75_previous | 900 | 0 | 0.0000% | 17.80 | 42.52 | 72.25 | true |
| spec_acdc_asc_p99_next | 900 | 0 | 0.0000% | 20.98 | 59.53 | 146.35 | true |
| spec_acdc_asc_p99_previous | 900 | 0 | 0.0000% | 20.80 | 45.31 | 67.52 | true |
| spec_acdc_desc_p25_next | 900 | 0 | 0.0000% | 19.80 | 42.67 | 63.82 | true |
| spec_acdc_desc_p25_previous | 900 | 0 | 0.0000% | 19.70 | 54.22 | 82.70 | true |
| spec_acdc_desc_p50_next | 900 | 0 | 0.0000% | 21.19 | 43.51 | 79.98 | true |
| spec_acdc_desc_p50_previous | 900 | 0 | 0.0000% | 19.83 | 42.45 | 62.70 | true |
| spec_acdc_desc_p75_next | 900 | 0 | 0.0000% | 18.60 | 44.05 | 91.10 | true |
| spec_acdc_desc_p75_previous | 900 | 0 | 0.0000% | 20.22 | 43.56 | 71.35 | true |
| spec_acdc_desc_p99_next | 900 | 0 | 0.0000% | 20.00 | 40.09 | 63.89 | true |
| spec_acdc_desc_p99_previous | 900 | 0 | 0.0000% | 21.50 | 48.43 | 67.37 | true |
| spec_power_asc_p25_next | 900 | 0 | 0.0000% | 21.04 | 40.77 | 62.54 | true |
| spec_power_asc_p25_previous | 900 | 0 | 0.0000% | 20.45 | 41.80 | 68.16 | true |
| spec_power_asc_p50_next | 900 | 0 | 0.0000% | 20.85 | 45.91 | 63.85 | true |
| spec_power_asc_p50_previous | 900 | 0 | 0.0000% | 18.97 | 40.18 | 63.46 | true |
| spec_power_asc_p75_next | 900 | 0 | 0.0000% | 21.27 | 44.31 | 67.63 | true |
| spec_power_asc_p75_previous | 900 | 0 | 0.0000% | 20.55 | 42.39 | 62.28 | true |
| spec_power_asc_p99_next | 900 | 0 | 0.0000% | 22.27 | 48.76 | 69.54 | true |
| spec_power_asc_p99_previous | 900 | 0 | 0.0000% | 18.34 | 41.81 | 75.82 | true |
| spec_power_desc_p25_next | 900 | 0 | 0.0000% | 17.11 | 42.44 | 105.10 | true |
| spec_power_desc_p25_previous | 900 | 0 | 0.0000% | 21.79 | 43.10 | 68.39 | true |
| spec_power_desc_p50_next | 900 | 0 | 0.0000% | 22.25 | 46.21 | 77.03 | true |
| spec_power_desc_p50_previous | 900 | 0 | 0.0000% | 20.15 | 41.21 | 71.28 | true |
| spec_power_desc_p75_next | 900 | 0 | 0.0000% | 16.42 | 37.42 | 314.32 | true |
| spec_power_desc_p75_previous | 900 | 0 | 0.0000% | 17.48 | 42.63 | 119.38 | true |
| spec_power_desc_p99_next | 900 | 0 | 0.0000% | 19.44 | 38.77 | 75.01 | true |
| spec_power_desc_p99_previous | 900 | 0 | 0.0000% | 20.31 | 41.17 | 58.24 | true |
