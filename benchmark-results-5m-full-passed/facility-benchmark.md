# Facility benchmark (5000000 FieldDevices)

Git: `da4128276ebd2fb1ae2207675929301a8b942f91`  
Generated: 2026-08-22T13:56:03Z

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

- `database_bytes`: `30370608831`
- `expected_field_devices`: `5000000`
- `field_device_cursor_values_index_bytes`: `9437429760`
- `field_device_cursor_values_rows`: `5000000`
- `field_device_cursor_values_table_bytes`: `560979968`
- `field_devices_bmk_null_rows`: `1000000`
- `field_devices_index_bytes`: `5703311360`
- `field_devices_rows`: `5000000`
- `field_devices_table_bytes`: `889430016`
- `project_field_devices_index_bytes`: `2642747392`
- `project_field_devices_rows`: `5000000`
- `project_field_devices_table_bytes`: `1051820032`
- `schema_migrations`: `33`
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

`202604030001`, `202604030007`, `202604030008`, `202604030009`, `202604300001`, `202604300002`, `202604300003`, `202604300004`, `202604300005`, `202605010001`, `202605010002`, `202605010003`, `202605020001`, `202605030001`, `202605070001`, `202605080001`, `202605080002`, `202605080003`, `202605110001`, `202605110002`, `202608130001`, `202608160001`, `202608210001`, `202608210002`, `202608220001`, `202608220002`, `202608220003`, `202608220004`, `202608220005`, `202608220006`, `202608220007`, `202608220008`, `202609050001`

| Scenario | Samples | Failures | Error rate | p50 ms | p95 ms | p99 ms | Gate |
|---|---:|---:|---:|---:|---:|---:|---|
| created_at_asc_first | 900 | 0 | 0.0000% | 29.95 | 57.32 | 75.17 | true |
| created_at_asc_next | 900 | 0 | 0.0000% | 31.05 | 61.10 | 84.44 | true |
| created_at_desc_first | 900 | 0 | 0.0000% | 30.35 | 59.32 | 81.11 | true |
| created_at_desc_next | 900 | 0 | 0.0000% | 28.88 | 56.18 | 89.81 | true |
| apparat_nr_asc_first | 900 | 0 | 0.0000% | 29.32 | 56.02 | 82.52 | true |
| apparat_nr_asc_next | 900 | 0 | 0.0000% | 30.85 | 64.96 | 92.84 | true |
| apparat_nr_desc_first | 900 | 0 | 0.0000% | 32.81 | 64.19 | 81.81 | true |
| apparat_nr_desc_next | 900 | 0 | 0.0000% | 30.64 | 73.63 | 110.03 | true |
| sps_system_type_asc_first | 900 | 0 | 0.0000% | 43.29 | 98.53 | 137.94 | true |
| sps_system_type_asc_next | 900 | 0 | 0.0000% | 41.27 | 82.79 | 158.98 | true |
| sps_system_type_desc_first | 900 | 0 | 0.0000% | 39.93 | 83.83 | 115.33 | true |
| sps_system_type_desc_next | 900 | 0 | 0.0000% | 40.56 | 91.09 | 134.47 | true |
| description_asc_first | 900 | 0 | 0.0000% | 34.19 | 62.42 | 90.79 | true |
| description_asc_next | 900 | 0 | 0.0000% | 31.77 | 63.98 | 96.14 | true |
| description_desc_first | 900 | 0 | 0.0000% | 34.72 | 65.44 | 98.49 | true |
| description_desc_next | 900 | 0 | 0.0000% | 39.76 | 69.21 | 99.89 | true |
| spec_supplier_asc_first | 900 | 0 | 0.0000% | 38.68 | 70.92 | 116.16 | true |
| spec_supplier_asc_next | 900 | 0 | 0.0000% | 39.08 | 74.46 | 128.28 | true |
| spec_supplier_desc_first | 900 | 0 | 0.0000% | 40.78 | 81.99 | 109.45 | true |
| spec_supplier_desc_next | 900 | 0 | 0.0000% | 41.86 | 86.24 | 158.70 | true |
| spec_size_asc_first | 900 | 0 | 0.0000% | 49.97 | 102.99 | 165.67 | true |
| spec_size_asc_next | 900 | 0 | 0.0000% | 54.82 | 104.34 | 157.72 | true |
| spec_size_desc_first | 900 | 0 | 0.0000% | 53.79 | 99.14 | 132.14 | true |
| spec_size_desc_next | 900 | 0 | 0.0000% | 51.01 | 91.03 | 136.21 | true |
| spec_acdc_asc_first | 900 | 0 | 0.0000% | 48.02 | 102.99 | 192.58 | true |
| spec_acdc_asc_next | 900 | 0 | 0.0000% | 35.13 | 73.79 | 322.99 | true |
| spec_acdc_desc_first | 900 | 0 | 0.0000% | 33.68 | 62.24 | 90.36 | true |
| spec_acdc_desc_next | 900 | 0 | 0.0000% | 34.12 | 70.31 | 102.50 | true |
| spec_power_asc_first | 900 | 0 | 0.0000% | 35.31 | 70.19 | 107.97 | true |
| spec_power_asc_next | 900 | 0 | 0.0000% | 32.87 | 69.58 | 86.53 | true |
| spec_power_desc_first | 900 | 0 | 0.0000% | 33.50 | 62.45 | 102.42 | true |
| spec_power_desc_next | 900 | 0 | 0.0000% | 32.22 | 64.57 | 97.24 | true |
| building_scope_first | 900 | 0 | 0.0000% | 1490.73 | 1834.63 | 2002.70 | false |
| building_scope_previous | 900 | 0 | 0.0000% | 42.63 | 107.11 | 148.84 | true |
| building_scope_next | 900 | 0 | 0.0000% | 1079.31 | 1353.19 | 1528.71 | false |
| cabinet_scope_first | 900 | 0 | 0.0000% | 117.59 | 179.02 | 219.38 | true |
| cabinet_scope_previous | 900 | 0 | 0.0000% | 48.99 | 90.66 | 126.79 | true |
| cabinet_scope_next | 900 | 0 | 0.0000% | 98.46 | 161.41 | 202.42 | true |
| controller_scope_first | 900 | 0 | 0.0000% | 49.04 | 88.22 | 118.53 | true |
| controller_scope_previous | 900 | 0 | 0.0000% | 46.82 | 83.38 | 109.94 | true |
| controller_scope_next | 900 | 0 | 0.0000% | 35.72 | 69.64 | 106.31 | true |
| project_scope_first | 900 | 0 | 0.0000% | 51.24 | 87.68 | 128.04 | true |
| project_scope_previous | 900 | 0 | 0.0000% | 51.51 | 93.11 | 149.47 | true |
| project_scope_next | 900 | 0 | 0.0000% | 50.52 | 93.96 | 122.95 | true |
| search_0_1_percent_first | 900 | 0 | 0.0000% | 265.94 | 383.67 | 444.49 | true |
| search_0_1_percent_previous | 900 | 0 | 0.0000% | 175.98 | 241.46 | 274.45 | true |
| search_0_1_percent_next | 900 | 0 | 0.0000% | 231.27 | 355.26 | 388.78 | true |
| search_1_percent_first | 900 | 0 | 0.0000% | 561.88 | 744.10 | 827.47 | true |
| search_1_percent_previous | 900 | 0 | 0.0000% | 217.08 | 310.59 | 369.75 | true |
| search_1_percent_next | 900 | 0 | 0.0000% | 574.31 | 722.73 | 801.78 | true |
| search_10_percent_first | 900 | 0 | 0.0000% | 43.95 | 76.44 | 118.42 | true |
| search_10_percent_previous | 900 | 0 | 0.0000% | 867.00 | 1068.92 | 1156.68 | false |
| search_10_percent_next | 900 | 0 | 0.0000% | 33.23 | 84.67 | 118.54 | true |
| combined_filter_first | 900 | 0 | 0.0000% | 23.85 | 68.89 | 99.61 | true |
| combined_filter_previous | 900 | 0 | 0.0000% | 23.57 | 60.44 | 86.08 | true |
| combined_filter_next | 900 | 0 | 0.0000% | 24.14 | 76.16 | 92.70 | true |
| created_at_asc_p25_next | 900 | 0 | 0.0000% | 26.23 | 53.26 | 82.67 | true |
| created_at_asc_p25_previous | 900 | 0 | 0.0000% | 25.08 | 53.92 | 92.88 | true |
| created_at_asc_p50_next | 900 | 0 | 0.0000% | 28.14 | 74.22 | 139.31 | true |
| created_at_asc_p50_previous | 900 | 0 | 0.0000% | 27.59 | 65.87 | 82.62 | true |
| created_at_asc_p75_next | 900 | 0 | 0.0000% | 29.07 | 60.01 | 85.21 | true |
| created_at_asc_p75_previous | 900 | 0 | 0.0000% | 30.73 | 66.30 | 89.42 | true |
| created_at_asc_p99_next | 900 | 0 | 0.0000% | 34.17 | 67.41 | 92.42 | true |
| created_at_asc_p99_previous | 900 | 0 | 0.0000% | 32.63 | 68.23 | 96.02 | true |
| created_at_desc_p25_next | 900 | 0 | 0.0000% | 29.51 | 66.93 | 95.75 | true |
| created_at_desc_p25_previous | 900 | 0 | 0.0000% | 32.86 | 67.45 | 96.27 | true |
| created_at_desc_p50_next | 900 | 0 | 0.0000% | 32.75 | 78.59 | 109.30 | true |
| created_at_desc_p50_previous | 900 | 0 | 0.0000% | 32.12 | 66.15 | 95.19 | true |
| created_at_desc_p75_next | 900 | 0 | 0.0000% | 33.48 | 64.09 | 93.27 | true |
| created_at_desc_p75_previous | 900 | 0 | 0.0000% | 34.80 | 71.60 | 148.14 | true |
| created_at_desc_p99_next | 900 | 0 | 0.0000% | 31.49 | 62.31 | 82.59 | true |
| created_at_desc_p99_previous | 900 | 0 | 0.0000% | 34.06 | 72.62 | 100.06 | true |
| apparat_nr_asc_p25_next | 900 | 0 | 0.0000% | 33.03 | 71.40 | 98.95 | true |
| apparat_nr_asc_p25_previous | 900 | 0 | 0.0000% | 31.54 | 64.29 | 90.51 | true |
| apparat_nr_asc_p50_next | 900 | 0 | 0.0000% | 35.92 | 74.23 | 101.48 | true |
| apparat_nr_asc_p50_previous | 900 | 0 | 0.0000% | 33.57 | 69.22 | 85.35 | true |
| apparat_nr_asc_p75_next | 900 | 0 | 0.0000% | 31.65 | 66.02 | 97.49 | true |
| apparat_nr_asc_p75_previous | 900 | 0 | 0.0000% | 38.14 | 70.91 | 100.35 | true |
| apparat_nr_asc_p99_next | 900 | 0 | 0.0000% | 32.74 | 80.12 | 109.67 | true |
| apparat_nr_asc_p99_previous | 900 | 0 | 0.0000% | 33.66 | 67.14 | 85.21 | true |
| apparat_nr_desc_p25_next | 900 | 0 | 0.0000% | 36.12 | 85.62 | 125.03 | true |
| apparat_nr_desc_p25_previous | 900 | 0 | 0.0000% | 32.58 | 70.91 | 98.41 | true |
| apparat_nr_desc_p50_next | 900 | 0 | 0.0000% | 32.72 | 70.86 | 90.80 | true |
| apparat_nr_desc_p50_previous | 900 | 0 | 0.0000% | 30.79 | 78.67 | 95.67 | true |
| apparat_nr_desc_p75_next | 900 | 0 | 0.0000% | 29.13 | 67.27 | 93.96 | true |
| apparat_nr_desc_p75_previous | 900 | 0 | 0.0000% | 35.35 | 68.80 | 89.19 | true |
| apparat_nr_desc_p99_next | 900 | 0 | 0.0000% | 32.08 | 70.69 | 157.61 | true |
| apparat_nr_desc_p99_previous | 900 | 0 | 0.0000% | 31.80 | 71.15 | 90.64 | true |
| sps_system_type_asc_p25_next | 900 | 0 | 0.0000% | 42.98 | 99.46 | 134.68 | true |
| sps_system_type_asc_p25_previous | 900 | 0 | 0.0000% | 40.69 | 92.54 | 134.88 | true |
| sps_system_type_asc_p50_next | 900 | 0 | 0.0000% | 41.14 | 87.99 | 249.03 | true |
| sps_system_type_asc_p50_previous | 900 | 0 | 0.0000% | 39.94 | 85.97 | 143.21 | true |
| sps_system_type_asc_p75_next | 900 | 0 | 0.0000% | 41.26 | 96.33 | 130.82 | true |
| sps_system_type_asc_p75_previous | 900 | 0 | 0.0000% | 38.70 | 82.67 | 109.10 | true |
| sps_system_type_asc_p99_next | 900 | 0 | 0.0000% | 1253.74 | 1658.86 | 1933.90 | false |
| sps_system_type_asc_p99_previous | 900 | 0 | 0.0000% | 43.93 | 110.11 | 162.81 | true |
| sps_system_type_desc_p25_next | 900 | 0 | 0.0000% | 125.27 | 237.06 | 287.95 | true |
| sps_system_type_desc_p25_previous | 900 | 0 | 0.0000% | 312.60 | 516.64 | 598.14 | true |
| sps_system_type_desc_p50_next | 900 | 0 | 0.0000% | 224.82 | 385.18 | 459.15 | true |
| sps_system_type_desc_p50_previous | 900 | 0 | 0.0000% | 209.78 | 384.64 | 446.47 | true |
| sps_system_type_desc_p75_next | 900 | 0 | 0.0000% | 314.27 | 535.44 | 583.70 | true |
| sps_system_type_desc_p75_previous | 900 | 0 | 0.0000% | 115.86 | 248.82 | 284.09 | true |
| sps_system_type_desc_p99_next | 900 | 0 | 0.0000% | 405.22 | 613.53 | 682.36 | true |
| sps_system_type_desc_p99_previous | 900 | 0 | 0.0000% | 48.77 | 99.33 | 130.71 | true |
| description_asc_p25_next | 900 | 0 | 0.0000% | 30.60 | 69.81 | 91.82 | true |
| description_asc_p25_previous | 900 | 0 | 0.0000% | 29.48 | 62.53 | 102.11 | true |
| description_asc_p50_next | 900 | 0 | 0.0000% | 34.74 | 72.53 | 168.20 | true |
| description_asc_p50_previous | 900 | 0 | 0.0000% | 29.88 | 68.43 | 110.14 | true |
| description_asc_p75_next | 900 | 0 | 0.0000% | 34.32 | 66.33 | 81.33 | true |
| description_asc_p75_previous | 900 | 0 | 0.0000% | 33.35 | 77.86 | 101.87 | true |
| description_asc_p99_next | 900 | 0 | 0.0000% | 31.88 | 63.34 | 97.56 | true |
| description_asc_p99_previous | 900 | 0 | 0.0000% | 34.19 | 74.81 | 97.03 | true |
| description_desc_p25_next | 900 | 0 | 0.0000% | 34.97 | 64.43 | 91.34 | true |
| description_desc_p25_previous | 900 | 0 | 0.0000% | 30.20 | 75.21 | 103.85 | true |
| description_desc_p50_next | 900 | 0 | 0.0000% | 30.75 | 65.58 | 94.93 | true |
| description_desc_p50_previous | 900 | 0 | 0.0000% | 29.16 | 71.39 | 100.77 | true |
| description_desc_p75_next | 900 | 0 | 0.0000% | 30.28 | 74.74 | 95.25 | true |
| description_desc_p75_previous | 900 | 0 | 0.0000% | 33.97 | 74.40 | 149.51 | true |
| description_desc_p99_next | 900 | 0 | 0.0000% | 34.52 | 75.81 | 104.33 | true |
| description_desc_p99_previous | 900 | 0 | 0.0000% | 32.60 | 70.00 | 105.03 | true |
| spec_supplier_asc_p25_next | 900 | 0 | 0.0000% | 31.43 | 62.47 | 86.48 | true |
| spec_supplier_asc_p25_previous | 900 | 0 | 0.0000% | 42.43 | 94.19 | 127.67 | true |
| spec_supplier_asc_p50_next | 900 | 0 | 0.0000% | 35.87 | 71.14 | 115.69 | true |
| spec_supplier_asc_p50_previous | 900 | 0 | 0.0000% | 32.25 | 74.00 | 109.74 | true |
| spec_supplier_asc_p75_next | 900 | 0 | 0.0000% | 32.01 | 77.01 | 101.61 | true |
| spec_supplier_asc_p75_previous | 900 | 0 | 0.0000% | 36.23 | 68.71 | 96.01 | true |
| spec_supplier_asc_p99_next | 900 | 0 | 0.0000% | 30.37 | 64.95 | 98.42 | true |
| spec_supplier_asc_p99_previous | 900 | 0 | 0.0000% | 32.78 | 69.46 | 94.78 | true |
| spec_supplier_desc_p25_next | 900 | 0 | 0.0000% | 35.70 | 75.13 | 98.57 | true |
| spec_supplier_desc_p25_previous | 900 | 0 | 0.0000% | 34.66 | 72.90 | 128.06 | true |
| spec_supplier_desc_p50_next | 900 | 0 | 0.0000% | 33.34 | 69.51 | 94.96 | true |
| spec_supplier_desc_p50_previous | 900 | 0 | 0.0000% | 34.09 | 80.18 | 98.80 | true |
| spec_supplier_desc_p75_next | 900 | 0 | 0.0000% | 37.33 | 82.44 | 127.33 | true |
| spec_supplier_desc_p75_previous | 900 | 0 | 0.0000% | 32.31 | 72.12 | 100.19 | true |
| spec_supplier_desc_p99_next | 900 | 0 | 0.0000% | 31.13 | 67.12 | 95.90 | true |
| spec_supplier_desc_p99_previous | 900 | 0 | 0.0000% | 33.00 | 77.08 | 147.50 | true |
| spec_size_asc_p25_next | 900 | 0 | 0.0000% | 42.07 | 98.50 | 138.27 | true |
| spec_size_asc_p25_previous | 900 | 0 | 0.0000% | 44.60 | 99.15 | 134.81 | true |
| spec_size_asc_p50_next | 900 | 0 | 0.0000% | 40.10 | 96.65 | 118.29 | true |
| spec_size_asc_p50_previous | 900 | 0 | 0.0000% | 43.27 | 95.53 | 140.07 | true |
| spec_size_asc_p75_next | 900 | 0 | 0.0000% | 41.83 | 96.89 | 149.98 | true |
| spec_size_asc_p75_previous | 900 | 0 | 0.0000% | 43.19 | 98.17 | 119.97 | true |
| spec_size_asc_p99_next | 900 | 0 | 0.0000% | 38.75 | 89.45 | 136.64 | true |
| spec_size_asc_p99_previous | 900 | 0 | 0.0000% | 41.41 | 91.19 | 130.54 | true |
| spec_size_desc_p25_next | 900 | 0 | 0.0000% | 42.21 | 94.72 | 137.72 | true |
| spec_size_desc_p25_previous | 900 | 0 | 0.0000% | 40.85 | 99.13 | 138.56 | true |
| spec_size_desc_p50_next | 900 | 0 | 0.0000% | 43.29 | 92.59 | 128.97 | true |
| spec_size_desc_p50_previous | 900 | 0 | 0.0000% | 42.66 | 90.34 | 116.36 | true |
| spec_size_desc_p75_next | 900 | 0 | 0.0000% | 44.75 | 93.00 | 130.37 | true |
| spec_size_desc_p75_previous | 900 | 0 | 0.0000% | 49.76 | 85.64 | 120.94 | true |
| spec_size_desc_p99_next | 900 | 0 | 0.0000% | 43.71 | 98.07 | 122.53 | true |
| spec_size_desc_p99_previous | 900 | 0 | 0.0000% | 40.25 | 96.89 | 155.74 | true |
| spec_acdc_asc_p25_next | 900 | 0 | 0.0000% | 31.13 | 61.67 | 87.79 | true |
| spec_acdc_asc_p25_previous | 900 | 0 | 0.0000% | 31.67 | 66.99 | 92.50 | true |
| spec_acdc_asc_p50_next | 900 | 0 | 0.0000% | 30.11 | 71.32 | 98.45 | true |
| spec_acdc_asc_p50_previous | 900 | 0 | 0.0000% | 30.95 | 60.20 | 84.61 | true |
| spec_acdc_asc_p75_next | 900 | 0 | 0.0000% | 34.66 | 70.19 | 92.58 | true |
| spec_acdc_asc_p75_previous | 900 | 0 | 0.0000% | 31.84 | 75.00 | 98.97 | true |
| spec_acdc_asc_p99_next | 900 | 0 | 0.0000% | 32.14 | 72.51 | 98.79 | true |
| spec_acdc_asc_p99_previous | 900 | 0 | 0.0000% | 34.32 | 74.79 | 105.32 | true |
| spec_acdc_desc_p25_next | 900 | 0 | 0.0000% | 31.09 | 65.81 | 100.75 | true |
| spec_acdc_desc_p25_previous | 900 | 0 | 0.0000% | 32.09 | 70.87 | 125.12 | true |
| spec_acdc_desc_p50_next | 900 | 0 | 0.0000% | 32.98 | 72.14 | 97.80 | true |
| spec_acdc_desc_p50_previous | 900 | 0 | 0.0000% | 33.53 | 77.27 | 103.30 | true |
| spec_acdc_desc_p75_next | 900 | 0 | 0.0000% | 28.64 | 67.98 | 98.69 | true |
| spec_acdc_desc_p75_previous | 900 | 0 | 0.0000% | 32.92 | 75.66 | 102.25 | true |
| spec_acdc_desc_p99_next | 900 | 0 | 0.0000% | 32.06 | 66.11 | 95.14 | true |
| spec_acdc_desc_p99_previous | 900 | 0 | 0.0000% | 27.86 | 58.89 | 82.14 | true |
| spec_power_asc_p25_next | 900 | 0 | 0.0000% | 34.63 | 66.13 | 87.11 | true |
| spec_power_asc_p25_previous | 900 | 0 | 0.0000% | 28.89 | 72.10 | 202.37 | true |
| spec_power_asc_p50_next | 900 | 0 | 0.0000% | 29.08 | 69.25 | 103.31 | true |
| spec_power_asc_p50_previous | 900 | 0 | 0.0000% | 33.54 | 66.10 | 109.72 | true |
| spec_power_asc_p75_next | 900 | 0 | 0.0000% | 32.49 | 80.11 | 161.00 | true |
| spec_power_asc_p75_previous | 900 | 0 | 0.0000% | 31.16 | 74.62 | 110.80 | true |
| spec_power_asc_p99_next | 900 | 0 | 0.0000% | 30.47 | 61.80 | 94.03 | true |
| spec_power_asc_p99_previous | 900 | 0 | 0.0000% | 32.04 | 76.44 | 96.66 | true |
| spec_power_desc_p25_next | 900 | 0 | 0.0000% | 34.28 | 62.85 | 86.69 | true |
| spec_power_desc_p25_previous | 900 | 0 | 0.0000% | 33.51 | 65.91 | 86.97 | true |
| spec_power_desc_p50_next | 900 | 0 | 0.0000% | 37.38 | 72.81 | 109.80 | true |
| spec_power_desc_p50_previous | 900 | 0 | 0.0000% | 32.95 | 59.64 | 92.15 | true |
| spec_power_desc_p75_next | 900 | 0 | 0.0000% | 27.29 | 71.66 | 105.28 | true |
| spec_power_desc_p75_previous | 900 | 0 | 0.0000% | 27.45 | 59.87 | 100.41 | true |
| spec_power_desc_p99_next | 900 | 0 | 0.0000% | 31.29 | 60.52 | 91.37 | true |
| spec_power_desc_p99_previous | 900 | 0 | 0.0000% | 32.35 | 68.79 | 119.26 | true |
