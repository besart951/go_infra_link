# Facility benchmark (5000000 FieldDevices)

Git: `da4128276ebd2fb1ae2207675929301a8b942f91`  
Generated: 2026-08-22T12:53:43Z

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

- `database_bytes`: `28913653439`
- `expected_field_devices`: `5000000`
- `field_device_cursor_values_index_bytes`: `9437429760`
- `field_device_cursor_values_rows`: `5000000`
- `field_device_cursor_values_table_bytes`: `560979968`
- `field_devices_bmk_null_rows`: `1000000`
- `field_devices_index_bytes`: `5703311360`
- `field_devices_rows`: `5000000`
- `field_devices_table_bytes`: `889430016`
- `project_field_devices_index_bytes`: `1732026368`
- `project_field_devices_rows`: `5000000`
- `project_field_devices_table_bytes`: `505741312`
- `schema_migrations`: `32`
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

`202604030001`, `202604030007`, `202604030008`, `202604030009`, `202604300001`, `202604300002`, `202604300003`, `202604300004`, `202604300005`, `202605010001`, `202605010002`, `202605010003`, `202605020001`, `202605030001`, `202605070001`, `202605080001`, `202605080002`, `202605080003`, `202605110001`, `202605110002`, `202608130001`, `202608160001`, `202608210001`, `202608210002`, `202608220001`, `202608220002`, `202608220003`, `202608220004`, `202608220005`, `202608220006`, `202608220007`, `202609050001`

| Scenario | Samples | Failures | Error rate | p50 ms | p95 ms | p99 ms | Gate |
|---|---:|---:|---:|---:|---:|---:|---|
| created_at_asc_first | 15 | 0 | 0.0000% | 12.67 | 17.75 | 17.75 | true |
| created_at_asc_next | 15 | 0 | 0.0000% | 13.26 | 16.37 | 16.37 | true |
| created_at_desc_first | 15 | 0 | 0.0000% | 14.58 | 30.16 | 30.16 | true |
| created_at_desc_next | 15 | 0 | 0.0000% | 15.72 | 19.32 | 19.32 | true |
| apparat_nr_asc_first | 15 | 0 | 0.0000% | 14.31 | 21.63 | 21.63 | true |
| apparat_nr_asc_next | 15 | 0 | 0.0000% | 12.85 | 16.39 | 16.39 | true |
| apparat_nr_desc_first | 15 | 0 | 0.0000% | 15.16 | 24.67 | 24.67 | true |
| apparat_nr_desc_next | 15 | 0 | 0.0000% | 19.20 | 23.65 | 23.65 | true |
| sps_system_type_asc_first | 15 | 0 | 0.0000% | 18.55 | 24.73 | 24.73 | true |
| sps_system_type_asc_next | 15 | 0 | 0.0000% | 17.86 | 23.64 | 23.64 | true |
| sps_system_type_desc_first | 15 | 0 | 0.0000% | 17.38 | 21.30 | 21.30 | true |
| sps_system_type_desc_next | 15 | 0 | 0.0000% | 17.24 | 20.82 | 20.82 | true |
| description_asc_first | 15 | 0 | 0.0000% | 15.49 | 22.05 | 22.05 | true |
| description_asc_next | 15 | 0 | 0.0000% | 15.23 | 22.43 | 22.43 | true |
| description_desc_first | 15 | 0 | 0.0000% | 13.27 | 15.86 | 15.86 | true |
| description_desc_next | 15 | 0 | 0.0000% | 17.91 | 23.75 | 23.75 | true |
| spec_supplier_asc_first | 15 | 0 | 0.0000% | 16.35 | 19.15 | 19.15 | true |
| spec_supplier_asc_next | 15 | 0 | 0.0000% | 12.94 | 18.07 | 18.07 | true |
| spec_supplier_desc_first | 15 | 0 | 0.0000% | 15.12 | 25.55 | 25.55 | true |
| spec_supplier_desc_next | 15 | 0 | 0.0000% | 14.78 | 20.05 | 20.05 | true |
| spec_size_asc_first | 15 | 0 | 0.0000% | 16.80 | 18.44 | 18.44 | true |
| spec_size_asc_next | 15 | 0 | 0.0000% | 16.34 | 20.10 | 20.10 | true |
| spec_size_desc_first | 15 | 0 | 0.0000% | 18.48 | 21.51 | 21.51 | true |
| spec_size_desc_next | 15 | 0 | 0.0000% | 15.93 | 19.48 | 19.48 | true |
| spec_acdc_asc_first | 15 | 0 | 0.0000% | 13.97 | 16.20 | 16.20 | true |
| spec_acdc_asc_next | 15 | 0 | 0.0000% | 14.14 | 20.35 | 20.35 | true |
| spec_acdc_desc_first | 15 | 0 | 0.0000% | 14.95 | 27.35 | 27.35 | true |
| spec_acdc_desc_next | 15 | 0 | 0.0000% | 13.32 | 15.84 | 15.84 | true |
| spec_power_asc_first | 15 | 0 | 0.0000% | 12.49 | 15.76 | 15.76 | true |
| spec_power_asc_next | 15 | 0 | 0.0000% | 12.31 | 17.69 | 17.69 | true |
| spec_power_desc_first | 15 | 0 | 0.0000% | 12.51 | 14.71 | 14.71 | true |
| spec_power_desc_next | 15 | 0 | 0.0000% | 12.19 | 15.25 | 15.25 | true |
| building_scope_first | 15 | 0 | 0.0000% | 332.26 | 423.37 | 423.37 | true |
| building_scope_previous | 15 | 0 | 0.0000% | 271.83 | 316.43 | 316.43 | true |
| building_scope_next | 15 | 0 | 0.0000% | 325.15 | 373.58 | 373.58 | true |
| cabinet_scope_first | 15 | 0 | 0.0000% | 29.90 | 35.43 | 35.43 | true |
| cabinet_scope_previous | 15 | 0 | 0.0000% | 21.01 | 26.80 | 26.80 | true |
| cabinet_scope_next | 15 | 0 | 0.0000% | 29.40 | 31.78 | 31.78 | true |
| controller_scope_first | 15 | 0 | 0.0000% | 14.28 | 17.29 | 17.29 | true |
| controller_scope_previous | 15 | 0 | 0.0000% | 15.18 | 17.85 | 17.85 | true |
| controller_scope_next | 15 | 0 | 0.0000% | 10.97 | 12.30 | 12.30 | true |
| project_scope_first | 15 | 0 | 0.0000% | 436.40 | 493.76 | 493.76 | true |
| project_scope_previous | 15 | 0 | 0.0000% | 95.36 | 112.19 | 112.19 | true |
| project_scope_next | 15 | 0 | 0.0000% | 408.11 | 459.44 | 459.44 | true |
| search_0_1_percent_first | 15 | 0 | 0.0000% | 63.24 | 83.69 | 83.69 | true |
| search_0_1_percent_previous | 15 | 0 | 0.0000% | 54.88 | 82.36 | 82.36 | true |
| search_0_1_percent_next | 15 | 0 | 0.0000% | 61.50 | 109.48 | 109.48 | true |
| search_1_percent_first | 15 | 0 | 0.0000% | 128.86 | 179.93 | 179.93 | true |
| search_1_percent_previous | 15 | 0 | 0.0000% | 60.92 | 85.55 | 85.55 | true |
| search_1_percent_next | 15 | 0 | 0.0000% | 136.55 | 164.98 | 164.98 | true |
| search_10_percent_first | 15 | 0 | 0.0000% | 17.49 | 37.27 | 37.27 | true |
| search_10_percent_previous | 15 | 0 | 0.0000% | 221.52 | 297.47 | 297.47 | true |
| search_10_percent_next | 15 | 0 | 0.0000% | 16.49 | 20.60 | 20.60 | true |
| combined_filter_first | 15 | 0 | 0.0000% | 17.13 | 18.94 | 18.94 | true |
| combined_filter_previous | 15 | 0 | 0.0000% | 10.35 | 11.96 | 11.96 | true |
| combined_filter_next | 15 | 0 | 0.0000% | 12.76 | 14.69 | 14.69 | true |
| created_at_asc_p25_next | 15 | 0 | 0.0000% | 131.29 | 170.98 | 170.98 | true |
| created_at_asc_p25_previous | 15 | 0 | 0.0000% | 411.02 | 462.93 | 462.93 | true |
| created_at_asc_p50_next | 15 | 0 | 0.0000% | 314.50 | 366.13 | 366.13 | true |
| created_at_asc_p50_previous | 15 | 0 | 0.0000% | 315.89 | 341.26 | 341.26 | true |
| created_at_asc_p75_next | 15 | 0 | 0.0000% | 428.41 | 487.83 | 487.83 | true |
| created_at_asc_p75_previous | 15 | 0 | 0.0000% | 162.02 | 179.67 | 179.67 | true |
| created_at_asc_p99_next | 15 | 0 | 0.0000% | 224.30 | 257.91 | 257.91 | true |
| created_at_asc_p99_previous | 15 | 0 | 0.0000% | 18.96 | 21.77 | 21.77 | true |
| created_at_desc_p25_next | 15 | 0 | 0.0000% | 151.02 | 173.45 | 173.45 | true |
| created_at_desc_p25_previous | 15 | 0 | 0.0000% | 398.05 | 459.40 | 459.40 | true |
| created_at_desc_p50_next | 15 | 0 | 0.0000% | 265.58 | 308.70 | 308.70 | true |
| created_at_desc_p50_previous | 15 | 0 | 0.0000% | 257.50 | 301.99 | 301.99 | true |
| created_at_desc_p75_next | 15 | 0 | 0.0000% | 410.73 | 497.91 | 497.91 | true |
| created_at_desc_p75_previous | 15 | 0 | 0.0000% | 162.64 | 196.26 | 196.26 | true |
| created_at_desc_p99_next | 15 | 0 | 0.0000% | 227.47 | 317.45 | 317.45 | true |
| created_at_desc_p99_previous | 15 | 0 | 0.0000% | 18.15 | 30.70 | 30.70 | true |
| apparat_nr_asc_p25_next | 15 | 0 | 0.0000% | 11.19 | 14.95 | 14.95 | true |
| apparat_nr_asc_p25_previous | 15 | 0 | 0.0000% | 407.96 | 428.91 | 428.91 | true |
| apparat_nr_asc_p50_next | 15 | 0 | 0.0000% | 15.67 | 68.63 | 68.63 | true |
| apparat_nr_asc_p50_previous | 15 | 0 | 0.0000% | 257.33 | 316.43 | 316.43 | true |
| apparat_nr_asc_p75_next | 15 | 0 | 0.0000% | 13.18 | 17.95 | 17.95 | true |
| apparat_nr_asc_p75_previous | 15 | 0 | 0.0000% | 145.90 | 178.08 | 178.08 | true |
| apparat_nr_asc_p99_next | 15 | 0 | 0.0000% | 15.05 | 16.77 | 16.77 | true |
| apparat_nr_asc_p99_previous | 15 | 0 | 0.0000% | 17.16 | 19.48 | 19.48 | true |
| apparat_nr_desc_p25_next | 15 | 0 | 0.0000% | 13.27 | 17.84 | 17.84 | true |
| apparat_nr_desc_p25_previous | 15 | 0 | 0.0000% | 398.05 | 439.36 | 439.36 | true |
| apparat_nr_desc_p50_next | 15 | 0 | 0.0000% | 13.73 | 19.28 | 19.28 | true |
| apparat_nr_desc_p50_previous | 15 | 0 | 0.0000% | 261.19 | 299.16 | 299.16 | true |
| apparat_nr_desc_p75_next | 15 | 0 | 0.0000% | 12.80 | 17.57 | 17.57 | true |
| apparat_nr_desc_p75_previous | 15 | 0 | 0.0000% | 143.21 | 160.54 | 160.54 | true |
| apparat_nr_desc_p99_next | 15 | 0 | 0.0000% | 19.27 | 23.45 | 23.45 | true |
| apparat_nr_desc_p99_previous | 15 | 0 | 0.0000% | 17.80 | 67.14 | 67.14 | true |
| sps_system_type_asc_p25_next | 15 | 0 | 0.0000% | 89.88 | 140.25 | 140.25 | true |
| sps_system_type_asc_p25_previous | 15 | 0 | 0.0000% | 231.99 | 281.15 | 281.15 | true |
| sps_system_type_asc_p50_next | 15 | 0 | 0.0000% | 160.57 | 220.34 | 220.34 | true |
| sps_system_type_asc_p50_previous | 15 | 0 | 0.0000% | 155.31 | 178.96 | 178.96 | true |
| sps_system_type_asc_p75_next | 15 | 0 | 0.0000% | 233.65 | 245.81 | 245.81 | true |
| sps_system_type_asc_p75_previous | 15 | 0 | 0.0000% | 91.36 | 124.56 | 124.56 | true |
| sps_system_type_asc_p99_next | 15 | 0 | 0.0000% | 344.83 | 368.54 | 368.54 | true |
| sps_system_type_asc_p99_previous | 15 | 0 | 0.0000% | 17.98 | 32.54 | 32.54 | true |
| sps_system_type_desc_p25_next | 15 | 0 | 0.0000% | 85.94 | 107.30 | 107.30 | true |
| sps_system_type_desc_p25_previous | 15 | 0 | 0.0000% | 232.65 | 263.16 | 263.16 | true |
| sps_system_type_desc_p50_next | 15 | 0 | 0.0000% | 165.76 | 218.33 | 218.33 | true |
| sps_system_type_desc_p50_previous | 15 | 0 | 0.0000% | 165.87 | 188.94 | 188.94 | true |
| sps_system_type_desc_p75_next | 15 | 0 | 0.0000% | 234.61 | 287.89 | 287.89 | true |
| sps_system_type_desc_p75_previous | 15 | 0 | 0.0000% | 86.08 | 103.10 | 103.10 | true |
| sps_system_type_desc_p99_next | 15 | 0 | 0.0000% | 376.78 | 428.56 | 428.56 | true |
| sps_system_type_desc_p99_previous | 15 | 0 | 0.0000% | 19.51 | 24.87 | 24.87 | true |
| description_asc_p25_next | 15 | 0 | 0.0000% | 11.93 | 16.15 | 16.15 | true |
| description_asc_p25_previous | 15 | 0 | 0.0000% | 904.74 | 1006.81 | 1006.81 | false |
| description_asc_p50_next | 15 | 0 | 0.0000% | 17.39 | 31.30 | 31.30 | true |
| description_asc_p50_previous | 15 | 0 | 0.0000% | 624.95 | 713.03 | 713.03 | true |
| description_asc_p75_next | 15 | 0 | 0.0000% | 16.39 | 23.85 | 23.85 | true |
| description_asc_p75_previous | 15 | 0 | 0.0000% | 278.41 | 312.78 | 312.78 | true |
| description_asc_p99_next | 15 | 0 | 0.0000% | 12.81 | 20.82 | 20.82 | true |
| description_asc_p99_previous | 15 | 0 | 0.0000% | 12.39 | 14.06 | 14.06 | true |
| description_desc_p25_next | 15 | 0 | 0.0000% | 11.75 | 15.25 | 15.25 | true |
| description_desc_p25_previous | 15 | 0 | 0.0000% | 839.88 | 878.88 | 878.88 | false |
| description_desc_p50_next | 15 | 0 | 0.0000% | 13.19 | 14.12 | 14.12 | true |
| description_desc_p50_previous | 15 | 0 | 0.0000% | 490.81 | 530.05 | 530.05 | true |
| description_desc_p75_next | 15 | 0 | 0.0000% | 12.50 | 18.32 | 18.32 | true |
| description_desc_p75_previous | 15 | 0 | 0.0000% | 171.30 | 235.88 | 235.88 | true |
| description_desc_p99_next | 15 | 0 | 0.0000% | 16.16 | 22.92 | 22.92 | true |
| description_desc_p99_previous | 15 | 0 | 0.0000% | 10.44 | 15.45 | 15.45 | true |
| spec_supplier_asc_p25_next | 15 | 0 | 0.0000% | 11.26 | 14.45 | 14.45 | true |
| spec_supplier_asc_p25_previous | 15 | 0 | 0.0000% | 772.63 | 847.84 | 847.84 | false |
| spec_supplier_asc_p50_next | 15 | 0 | 0.0000% | 15.63 | 25.23 | 25.23 | true |
| spec_supplier_asc_p50_previous | 15 | 0 | 0.0000% | 395.17 | 433.53 | 433.53 | true |
| spec_supplier_asc_p75_next | 15 | 0 | 0.0000% | 18.81 | 28.67 | 28.67 | true |
| spec_supplier_asc_p75_previous | 15 | 0 | 0.0000% | 81.59 | 101.14 | 101.14 | true |
| spec_supplier_asc_p99_next | 15 | 0 | 0.0000% | 13.83 | 15.29 | 15.29 | true |
| spec_supplier_asc_p99_previous | 15 | 0 | 0.0000% | 12.98 | 14.52 | 14.52 | true |
| spec_supplier_desc_p25_next | 15 | 0 | 0.0000% | 11.87 | 14.66 | 14.66 | true |
| spec_supplier_desc_p25_previous | 15 | 0 | 0.0000% | 791.97 | 856.43 | 856.43 | false |
| spec_supplier_desc_p50_next | 15 | 0 | 0.0000% | 15.60 | 27.28 | 27.28 | true |
| spec_supplier_desc_p50_previous | 15 | 0 | 0.0000% | 475.91 | 549.05 | 549.05 | true |
| spec_supplier_desc_p75_next | 15 | 0 | 0.0000% | 12.97 | 14.74 | 14.74 | true |
| spec_supplier_desc_p75_previous | 15 | 0 | 0.0000% | 119.09 | 139.18 | 139.18 | true |
| spec_supplier_desc_p99_next | 15 | 0 | 0.0000% | 14.25 | 17.33 | 17.33 | true |
| spec_supplier_desc_p99_previous | 15 | 0 | 0.0000% | 13.14 | 17.24 | 17.24 | true |
| spec_size_asc_p25_next | 15 | 0 | 0.0000% | 15.90 | 29.36 | 29.36 | true |
| spec_size_asc_p25_previous | 15 | 0 | 0.0000% | 197.53 | 224.32 | 224.32 | true |
| spec_size_asc_p50_next | 15 | 0 | 0.0000% | 15.67 | 18.66 | 18.66 | true |
| spec_size_asc_p50_previous | 15 | 0 | 0.0000% | 145.33 | 158.27 | 158.27 | true |
| spec_size_asc_p75_next | 15 | 0 | 0.0000% | 16.01 | 18.28 | 18.28 | true |
| spec_size_asc_p75_previous | 15 | 0 | 0.0000% | 73.18 | 103.00 | 103.00 | true |
| spec_size_asc_p99_next | 15 | 0 | 0.0000% | 16.04 | 18.27 | 18.27 | true |
| spec_size_asc_p99_previous | 15 | 0 | 0.0000% | 20.95 | 28.71 | 28.71 | true |
| spec_size_desc_p25_next | 15 | 0 | 0.0000% | 14.43 | 16.73 | 16.73 | true |
| spec_size_desc_p25_previous | 15 | 0 | 0.0000% | 208.23 | 226.60 | 226.60 | true |
| spec_size_desc_p50_next | 15 | 0 | 0.0000% | 18.55 | 28.44 | 28.44 | true |
| spec_size_desc_p50_previous | 15 | 0 | 0.0000% | 143.65 | 163.69 | 163.69 | true |
| spec_size_desc_p75_next | 15 | 0 | 0.0000% | 18.98 | 27.00 | 27.00 | true |
| spec_size_desc_p75_previous | 15 | 0 | 0.0000% | 75.83 | 98.63 | 98.63 | true |
| spec_size_desc_p99_next | 15 | 0 | 0.0000% | 15.68 | 17.49 | 17.49 | true |
| spec_size_desc_p99_previous | 15 | 0 | 0.0000% | 17.25 | 20.17 | 20.17 | true |
| spec_acdc_asc_p25_next | 15 | 0 | 0.0000% | 10.99 | 14.37 | 14.37 | true |
| spec_acdc_asc_p25_previous | 15 | 0 | 0.0000% | 13.11 | 17.23 | 17.23 | true |
| spec_acdc_asc_p50_next | 15 | 0 | 0.0000% | 12.44 | 13.64 | 13.64 | true |
| spec_acdc_asc_p50_previous | 15 | 0 | 0.0000% | 12.05 | 15.49 | 15.49 | true |
| spec_acdc_asc_p75_next | 15 | 0 | 0.0000% | 11.89 | 15.62 | 15.62 | true |
| spec_acdc_asc_p75_previous | 15 | 0 | 0.0000% | 12.67 | 16.81 | 16.81 | true |
| spec_acdc_asc_p99_next | 15 | 0 | 0.0000% | 12.91 | 15.78 | 15.78 | true |
| spec_acdc_asc_p99_previous | 15 | 0 | 0.0000% | 11.65 | 15.38 | 15.38 | true |
| spec_acdc_desc_p25_next | 15 | 0 | 0.0000% | 11.73 | 14.58 | 14.58 | true |
| spec_acdc_desc_p25_previous | 15 | 0 | 0.0000% | 12.54 | 13.40 | 13.40 | true |
| spec_acdc_desc_p50_next | 15 | 0 | 0.0000% | 12.21 | 15.04 | 15.04 | true |
| spec_acdc_desc_p50_previous | 15 | 0 | 0.0000% | 12.57 | 14.42 | 14.42 | true |
| spec_acdc_desc_p75_next | 15 | 0 | 0.0000% | 12.06 | 15.53 | 15.53 | true |
| spec_acdc_desc_p75_previous | 15 | 0 | 0.0000% | 12.25 | 14.70 | 14.70 | true |
| spec_acdc_desc_p99_next | 15 | 0 | 0.0000% | 12.68 | 14.28 | 14.28 | true |
| spec_acdc_desc_p99_previous | 15 | 0 | 0.0000% | 11.88 | 15.25 | 15.25 | true |
| spec_power_asc_p25_next | 15 | 0 | 0.0000% | 11.90 | 13.72 | 13.72 | true |
| spec_power_asc_p25_previous | 15 | 0 | 0.0000% | 11.87 | 15.75 | 15.75 | true |
| spec_power_asc_p50_next | 15 | 0 | 0.0000% | 11.74 | 15.46 | 15.46 | true |
| spec_power_asc_p50_previous | 15 | 0 | 0.0000% | 12.81 | 14.63 | 14.63 | true |
| spec_power_asc_p75_next | 15 | 0 | 0.0000% | 11.56 | 27.82 | 27.82 | true |
| spec_power_asc_p75_previous | 15 | 0 | 0.0000% | 14.03 | 17.98 | 17.98 | true |
| spec_power_asc_p99_next | 15 | 0 | 0.0000% | 13.16 | 16.93 | 16.93 | true |
| spec_power_asc_p99_previous | 15 | 0 | 0.0000% | 11.63 | 14.91 | 14.91 | true |
| spec_power_desc_p25_next | 15 | 0 | 0.0000% | 19.03 | 26.20 | 26.20 | true |
| spec_power_desc_p25_previous | 15 | 0 | 0.0000% | 16.61 | 21.01 | 21.01 | true |
| spec_power_desc_p50_next | 15 | 0 | 0.0000% | 15.29 | 21.00 | 21.00 | true |
| spec_power_desc_p50_previous | 15 | 0 | 0.0000% | 11.75 | 14.08 | 14.08 | true |
| spec_power_desc_p75_next | 15 | 0 | 0.0000% | 11.49 | 14.46 | 14.46 | true |
| spec_power_desc_p75_previous | 15 | 0 | 0.0000% | 13.27 | 17.24 | 17.24 | true |
| spec_power_desc_p99_next | 15 | 0 | 0.0000% | 12.33 | 15.36 | 15.36 | true |
| spec_power_desc_p99_previous | 15 | 0 | 0.0000% | 11.71 | 14.51 | 14.51 | true |
