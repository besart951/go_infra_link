# Facility benchmark (5000000 FieldDevices)

Git: `da4128276ebd2fb1ae2207675929301a8b942f91`  
Generated: 2026-08-22T13:00:47Z

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
| created_at_asc_first | 15 | 0 | 0.0000% | 12.96 | 21.07 | 21.07 | true |
| created_at_asc_next | 15 | 0 | 0.0000% | 13.68 | 15.62 | 15.62 | true |
| created_at_desc_first | 15 | 0 | 0.0000% | 12.71 | 15.75 | 15.75 | true |
| created_at_desc_next | 15 | 0 | 0.0000% | 12.58 | 15.21 | 15.21 | true |
| apparat_nr_asc_first | 15 | 0 | 0.0000% | 12.15 | 14.94 | 14.94 | true |
| apparat_nr_asc_next | 15 | 0 | 0.0000% | 11.62 | 15.17 | 15.17 | true |
| apparat_nr_desc_first | 15 | 0 | 0.0000% | 11.82 | 14.43 | 14.43 | true |
| apparat_nr_desc_next | 15 | 0 | 0.0000% | 13.65 | 37.28 | 37.28 | true |
| sps_system_type_asc_first | 15 | 0 | 0.0000% | 16.30 | 27.34 | 27.34 | true |
| sps_system_type_asc_next | 15 | 0 | 0.0000% | 15.72 | 19.92 | 19.92 | true |
| sps_system_type_desc_first | 15 | 0 | 0.0000% | 17.33 | 19.30 | 19.30 | true |
| sps_system_type_desc_next | 15 | 0 | 0.0000% | 17.09 | 22.56 | 22.56 | true |
| description_asc_first | 15 | 0 | 0.0000% | 12.94 | 15.29 | 15.29 | true |
| description_asc_next | 15 | 0 | 0.0000% | 14.02 | 19.23 | 19.23 | true |
| description_desc_first | 15 | 0 | 0.0000% | 13.05 | 14.52 | 14.52 | true |
| description_desc_next | 15 | 0 | 0.0000% | 13.94 | 16.60 | 16.60 | true |
| spec_supplier_asc_first | 15 | 0 | 0.0000% | 13.91 | 17.09 | 17.09 | true |
| spec_supplier_asc_next | 15 | 0 | 0.0000% | 14.07 | 15.83 | 15.83 | true |
| spec_supplier_desc_first | 15 | 0 | 0.0000% | 15.83 | 24.18 | 24.18 | true |
| spec_supplier_desc_next | 15 | 0 | 0.0000% | 14.15 | 16.83 | 16.83 | true |
| spec_size_asc_first | 15 | 0 | 0.0000% | 17.22 | 19.16 | 19.16 | true |
| spec_size_asc_next | 15 | 0 | 0.0000% | 16.44 | 19.95 | 19.95 | true |
| spec_size_desc_first | 15 | 0 | 0.0000% | 17.52 | 19.10 | 19.10 | true |
| spec_size_desc_next | 15 | 0 | 0.0000% | 16.91 | 18.75 | 18.75 | true |
| spec_acdc_asc_first | 15 | 0 | 0.0000% | 12.11 | 15.58 | 15.58 | true |
| spec_acdc_asc_next | 15 | 0 | 0.0000% | 12.52 | 15.34 | 15.34 | true |
| spec_acdc_desc_first | 15 | 0 | 0.0000% | 13.14 | 15.29 | 15.29 | true |
| spec_acdc_desc_next | 15 | 0 | 0.0000% | 14.01 | 30.33 | 30.33 | true |
| spec_power_asc_first | 15 | 0 | 0.0000% | 16.20 | 110.62 | 110.62 | true |
| spec_power_asc_next | 15 | 0 | 0.0000% | 14.10 | 17.04 | 17.04 | true |
| spec_power_desc_first | 15 | 0 | 0.0000% | 13.70 | 17.09 | 17.09 | true |
| spec_power_desc_next | 15 | 0 | 0.0000% | 14.36 | 15.65 | 15.65 | true |
| building_scope_first | 15 | 0 | 0.0000% | 321.38 | 371.33 | 371.33 | true |
| building_scope_previous | 15 | 0 | 0.0000% | 18.83 | 24.02 | 24.02 | true |
| building_scope_next | 15 | 0 | 0.0000% | 246.90 | 294.81 | 294.81 | true |
| cabinet_scope_first | 15 | 0 | 0.0000% | 27.59 | 74.53 | 74.53 | true |
| cabinet_scope_previous | 15 | 0 | 0.0000% | 14.87 | 18.82 | 18.82 | true |
| cabinet_scope_next | 15 | 0 | 0.0000% | 28.69 | 35.91 | 35.91 | true |
| controller_scope_first | 15 | 0 | 0.0000% | 13.93 | 16.47 | 16.47 | true |
| controller_scope_previous | 15 | 0 | 0.0000% | 13.99 | 15.93 | 15.93 | true |
| controller_scope_next | 15 | 0 | 0.0000% | 11.07 | 12.44 | 12.44 | true |
| project_scope_first | 15 | 0 | 0.0000% | 401.95 | 466.38 | 466.38 | true |
| project_scope_previous | 15 | 0 | 0.0000% | 103.77 | 127.48 | 127.48 | true |
| project_scope_next | 15 | 0 | 0.0000% | 385.12 | 417.21 | 417.21 | true |
| search_0_1_percent_first | 15 | 0 | 0.0000% | 64.94 | 81.60 | 81.60 | true |
| search_0_1_percent_previous | 15 | 0 | 0.0000% | 50.02 | 58.50 | 58.50 | true |
| search_0_1_percent_next | 15 | 0 | 0.0000% | 67.27 | 105.25 | 105.25 | true |
| search_1_percent_first | 15 | 0 | 0.0000% | 137.95 | 165.99 | 165.99 | true |
| search_1_percent_previous | 15 | 0 | 0.0000% | 61.90 | 95.69 | 95.69 | true |
| search_1_percent_next | 15 | 0 | 0.0000% | 155.13 | 174.23 | 174.23 | true |
| search_10_percent_first | 15 | 0 | 0.0000% | 14.31 | 16.77 | 16.77 | true |
| search_10_percent_previous | 15 | 0 | 0.0000% | 253.55 | 310.94 | 310.94 | true |
| search_10_percent_next | 15 | 0 | 0.0000% | 15.03 | 17.48 | 17.48 | true |
| combined_filter_first | 15 | 0 | 0.0000% | 10.74 | 13.19 | 13.19 | true |
| combined_filter_previous | 15 | 0 | 0.0000% | 9.87 | 11.66 | 11.66 | true |
| combined_filter_next | 15 | 0 | 0.0000% | 10.48 | 17.79 | 17.79 | true |
| created_at_asc_p25_next | 15 | 0 | 0.0000% | 11.11 | 16.20 | 16.20 | true |
| created_at_asc_p25_previous | 15 | 0 | 0.0000% | 12.64 | 18.32 | 18.32 | true |
| created_at_asc_p50_next | 15 | 0 | 0.0000% | 12.20 | 15.54 | 15.54 | true |
| created_at_asc_p50_previous | 15 | 0 | 0.0000% | 11.56 | 14.72 | 14.72 | true |
| created_at_asc_p75_next | 15 | 0 | 0.0000% | 11.87 | 14.96 | 14.96 | true |
| created_at_asc_p75_previous | 15 | 0 | 0.0000% | 11.81 | 15.49 | 15.49 | true |
| created_at_asc_p99_next | 15 | 0 | 0.0000% | 13.36 | 21.96 | 21.96 | true |
| created_at_asc_p99_previous | 15 | 0 | 0.0000% | 11.50 | 15.02 | 15.02 | true |
| created_at_desc_p25_next | 15 | 0 | 0.0000% | 12.18 | 13.41 | 13.41 | true |
| created_at_desc_p25_previous | 15 | 0 | 0.0000% | 14.55 | 17.55 | 17.55 | true |
| created_at_desc_p50_next | 15 | 0 | 0.0000% | 11.89 | 14.96 | 14.96 | true |
| created_at_desc_p50_previous | 15 | 0 | 0.0000% | 13.48 | 22.28 | 22.28 | true |
| created_at_desc_p75_next | 15 | 0 | 0.0000% | 21.44 | 25.44 | 25.44 | true |
| created_at_desc_p75_previous | 15 | 0 | 0.0000% | 11.31 | 15.77 | 15.77 | true |
| created_at_desc_p99_next | 15 | 0 | 0.0000% | 13.83 | 17.32 | 17.32 | true |
| created_at_desc_p99_previous | 15 | 0 | 0.0000% | 12.99 | 14.82 | 14.82 | true |
| apparat_nr_asc_p25_next | 15 | 0 | 0.0000% | 10.83 | 14.90 | 14.90 | true |
| apparat_nr_asc_p25_previous | 15 | 0 | 0.0000% | 11.06 | 14.32 | 14.32 | true |
| apparat_nr_asc_p50_next | 15 | 0 | 0.0000% | 10.34 | 13.88 | 13.88 | true |
| apparat_nr_asc_p50_previous | 15 | 0 | 0.0000% | 10.49 | 12.86 | 12.86 | true |
| apparat_nr_asc_p75_next | 15 | 0 | 0.0000% | 11.12 | 13.64 | 13.64 | true |
| apparat_nr_asc_p75_previous | 15 | 0 | 0.0000% | 11.45 | 13.65 | 13.65 | true |
| apparat_nr_asc_p99_next | 15 | 0 | 0.0000% | 11.54 | 14.26 | 14.26 | true |
| apparat_nr_asc_p99_previous | 15 | 0 | 0.0000% | 13.72 | 24.07 | 24.07 | true |
| apparat_nr_desc_p25_next | 15 | 0 | 0.0000% | 14.70 | 15.27 | 15.27 | true |
| apparat_nr_desc_p25_previous | 15 | 0 | 0.0000% | 11.54 | 15.71 | 15.71 | true |
| apparat_nr_desc_p50_next | 15 | 0 | 0.0000% | 20.24 | 30.56 | 30.56 | true |
| apparat_nr_desc_p50_previous | 15 | 0 | 0.0000% | 12.70 | 16.38 | 16.38 | true |
| apparat_nr_desc_p75_next | 15 | 0 | 0.0000% | 11.79 | 15.06 | 15.06 | true |
| apparat_nr_desc_p75_previous | 15 | 0 | 0.0000% | 11.60 | 13.31 | 13.31 | true |
| apparat_nr_desc_p99_next | 15 | 0 | 0.0000% | 11.88 | 13.73 | 13.73 | true |
| apparat_nr_desc_p99_previous | 15 | 0 | 0.0000% | 12.38 | 14.65 | 14.65 | true |
| sps_system_type_asc_p25_next | 15 | 0 | 0.0000% | 86.31 | 99.88 | 99.88 | true |
| sps_system_type_asc_p25_previous | 15 | 0 | 0.0000% | 240.16 | 285.40 | 285.40 | true |
| sps_system_type_asc_p50_next | 15 | 0 | 0.0000% | 155.76 | 179.14 | 179.14 | true |
| sps_system_type_asc_p50_previous | 15 | 0 | 0.0000% | 157.66 | 199.37 | 199.37 | true |
| sps_system_type_asc_p75_next | 15 | 0 | 0.0000% | 238.21 | 287.21 | 287.21 | true |
| sps_system_type_asc_p75_previous | 15 | 0 | 0.0000% | 86.01 | 105.58 | 105.58 | true |
| sps_system_type_asc_p99_next | 15 | 0 | 0.0000% | 354.95 | 362.03 | 362.03 | true |
| sps_system_type_asc_p99_previous | 15 | 0 | 0.0000% | 19.59 | 22.20 | 22.20 | true |
| sps_system_type_desc_p25_next | 15 | 0 | 0.0000% | 94.83 | 117.12 | 117.12 | true |
| sps_system_type_desc_p25_previous | 15 | 0 | 0.0000% | 242.55 | 331.42 | 331.42 | true |
| sps_system_type_desc_p50_next | 15 | 0 | 0.0000% | 154.79 | 183.78 | 183.78 | true |
| sps_system_type_desc_p50_previous | 15 | 0 | 0.0000% | 160.08 | 172.80 | 172.80 | true |
| sps_system_type_desc_p75_next | 15 | 0 | 0.0000% | 236.12 | 254.83 | 254.83 | true |
| sps_system_type_desc_p75_previous | 15 | 0 | 0.0000% | 85.44 | 99.95 | 99.95 | true |
| sps_system_type_desc_p99_next | 15 | 0 | 0.0000% | 354.06 | 365.63 | 365.63 | true |
| sps_system_type_desc_p99_previous | 15 | 0 | 0.0000% | 18.55 | 20.72 | 20.72 | true |
| description_asc_p25_next | 15 | 0 | 0.0000% | 24.91 | 32.61 | 32.61 | true |
| description_asc_p25_previous | 15 | 0 | 0.0000% | 11.90 | 14.01 | 14.01 | true |
| description_asc_p50_next | 15 | 0 | 0.0000% | 13.78 | 17.72 | 17.72 | true |
| description_asc_p50_previous | 15 | 0 | 0.0000% | 13.85 | 15.72 | 15.72 | true |
| description_asc_p75_next | 15 | 0 | 0.0000% | 12.67 | 20.74 | 20.74 | true |
| description_asc_p75_previous | 15 | 0 | 0.0000% | 18.72 | 29.45 | 29.45 | true |
| description_asc_p99_next | 15 | 0 | 0.0000% | 11.14 | 14.06 | 14.06 | true |
| description_asc_p99_previous | 15 | 0 | 0.0000% | 12.52 | 18.62 | 18.62 | true |
| description_desc_p25_next | 15 | 0 | 0.0000% | 11.29 | 13.94 | 13.94 | true |
| description_desc_p25_previous | 15 | 0 | 0.0000% | 12.11 | 12.92 | 12.92 | true |
| description_desc_p50_next | 15 | 0 | 0.0000% | 10.78 | 12.98 | 12.98 | true |
| description_desc_p50_previous | 15 | 0 | 0.0000% | 12.57 | 21.89 | 21.89 | true |
| description_desc_p75_next | 15 | 0 | 0.0000% | 12.19 | 13.68 | 13.68 | true |
| description_desc_p75_previous | 15 | 0 | 0.0000% | 11.21 | 31.99 | 31.99 | true |
| description_desc_p99_next | 15 | 0 | 0.0000% | 16.73 | 34.62 | 34.62 | true |
| description_desc_p99_previous | 15 | 0 | 0.0000% | 11.80 | 14.84 | 14.84 | true |
| spec_supplier_asc_p25_next | 15 | 0 | 0.0000% | 12.58 | 14.09 | 14.09 | true |
| spec_supplier_asc_p25_previous | 15 | 0 | 0.0000% | 11.69 | 13.91 | 13.91 | true |
| spec_supplier_asc_p50_next | 15 | 0 | 0.0000% | 14.70 | 19.39 | 19.39 | true |
| spec_supplier_asc_p50_previous | 15 | 0 | 0.0000% | 12.28 | 14.29 | 14.29 | true |
| spec_supplier_asc_p75_next | 15 | 0 | 0.0000% | 12.41 | 33.15 | 33.15 | true |
| spec_supplier_asc_p75_previous | 15 | 0 | 0.0000% | 13.81 | 17.02 | 17.02 | true |
| spec_supplier_asc_p99_next | 15 | 0 | 0.0000% | 11.23 | 15.94 | 15.94 | true |
| spec_supplier_asc_p99_previous | 15 | 0 | 0.0000% | 14.37 | 19.96 | 19.96 | true |
| spec_supplier_desc_p25_next | 15 | 0 | 0.0000% | 12.55 | 31.78 | 31.78 | true |
| spec_supplier_desc_p25_previous | 15 | 0 | 0.0000% | 14.38 | 23.09 | 23.09 | true |
| spec_supplier_desc_p50_next | 15 | 0 | 0.0000% | 11.63 | 55.67 | 55.67 | true |
| spec_supplier_desc_p50_previous | 15 | 0 | 0.0000% | 12.70 | 30.10 | 30.10 | true |
| spec_supplier_desc_p75_next | 15 | 0 | 0.0000% | 11.55 | 15.34 | 15.34 | true |
| spec_supplier_desc_p75_previous | 15 | 0 | 0.0000% | 14.80 | 21.81 | 21.81 | true |
| spec_supplier_desc_p99_next | 15 | 0 | 0.0000% | 13.38 | 17.97 | 17.97 | true |
| spec_supplier_desc_p99_previous | 15 | 0 | 0.0000% | 13.57 | 16.41 | 16.41 | true |
| spec_size_asc_p25_next | 15 | 0 | 0.0000% | 15.67 | 18.66 | 18.66 | true |
| spec_size_asc_p25_previous | 15 | 0 | 0.0000% | 16.95 | 20.97 | 20.97 | true |
| spec_size_asc_p50_next | 15 | 0 | 0.0000% | 15.50 | 18.38 | 18.38 | true |
| spec_size_asc_p50_previous | 15 | 0 | 0.0000% | 14.52 | 17.87 | 17.87 | true |
| spec_size_asc_p75_next | 15 | 0 | 0.0000% | 15.65 | 18.84 | 18.84 | true |
| spec_size_asc_p75_previous | 15 | 0 | 0.0000% | 14.27 | 16.50 | 16.50 | true |
| spec_size_asc_p99_next | 15 | 0 | 0.0000% | 15.27 | 19.74 | 19.74 | true |
| spec_size_asc_p99_previous | 15 | 0 | 0.0000% | 14.82 | 17.36 | 17.36 | true |
| spec_size_desc_p25_next | 15 | 0 | 0.0000% | 15.30 | 17.95 | 17.95 | true |
| spec_size_desc_p25_previous | 15 | 0 | 0.0000% | 15.27 | 18.89 | 18.89 | true |
| spec_size_desc_p50_next | 15 | 0 | 0.0000% | 15.58 | 17.42 | 17.42 | true |
| spec_size_desc_p50_previous | 15 | 0 | 0.0000% | 16.25 | 17.76 | 17.76 | true |
| spec_size_desc_p75_next | 15 | 0 | 0.0000% | 15.44 | 17.97 | 17.97 | true |
| spec_size_desc_p75_previous | 15 | 0 | 0.0000% | 13.80 | 18.04 | 18.04 | true |
| spec_size_desc_p99_next | 15 | 0 | 0.0000% | 14.55 | 17.68 | 17.68 | true |
| spec_size_desc_p99_previous | 15 | 0 | 0.0000% | 15.31 | 17.95 | 17.95 | true |
| spec_acdc_asc_p25_next | 15 | 0 | 0.0000% | 11.32 | 14.50 | 14.50 | true |
| spec_acdc_asc_p25_previous | 15 | 0 | 0.0000% | 12.44 | 15.22 | 15.22 | true |
| spec_acdc_asc_p50_next | 15 | 0 | 0.0000% | 12.08 | 14.88 | 14.88 | true |
| spec_acdc_asc_p50_previous | 15 | 0 | 0.0000% | 13.06 | 15.11 | 15.11 | true |
| spec_acdc_asc_p75_next | 15 | 0 | 0.0000% | 12.34 | 34.00 | 34.00 | true |
| spec_acdc_asc_p75_previous | 15 | 0 | 0.0000% | 11.53 | 14.45 | 14.45 | true |
| spec_acdc_asc_p99_next | 15 | 0 | 0.0000% | 12.41 | 13.85 | 13.85 | true |
| spec_acdc_asc_p99_previous | 15 | 0 | 0.0000% | 12.17 | 14.88 | 14.88 | true |
| spec_acdc_desc_p25_next | 15 | 0 | 0.0000% | 13.20 | 25.42 | 25.42 | true |
| spec_acdc_desc_p25_previous | 15 | 0 | 0.0000% | 12.21 | 13.90 | 13.90 | true |
| spec_acdc_desc_p50_next | 15 | 0 | 0.0000% | 12.15 | 16.66 | 16.66 | true |
| spec_acdc_desc_p50_previous | 15 | 0 | 0.0000% | 13.11 | 15.64 | 15.64 | true |
| spec_acdc_desc_p75_next | 15 | 0 | 0.0000% | 13.28 | 15.97 | 15.97 | true |
| spec_acdc_desc_p75_previous | 15 | 0 | 0.0000% | 12.12 | 69.22 | 69.22 | true |
| spec_acdc_desc_p99_next | 15 | 0 | 0.0000% | 11.37 | 14.01 | 14.01 | true |
| spec_acdc_desc_p99_previous | 15 | 0 | 0.0000% | 12.96 | 15.97 | 15.97 | true |
| spec_power_asc_p25_next | 15 | 0 | 0.0000% | 11.43 | 13.63 | 13.63 | true |
| spec_power_asc_p25_previous | 15 | 0 | 0.0000% | 12.17 | 18.14 | 18.14 | true |
| spec_power_asc_p50_next | 15 | 0 | 0.0000% | 10.29 | 13.55 | 13.55 | true |
| spec_power_asc_p50_previous | 15 | 0 | 0.0000% | 11.67 | 13.96 | 13.96 | true |
| spec_power_asc_p75_next | 15 | 0 | 0.0000% | 12.35 | 14.65 | 14.65 | true |
| spec_power_asc_p75_previous | 15 | 0 | 0.0000% | 12.82 | 15.16 | 15.16 | true |
| spec_power_asc_p99_next | 15 | 0 | 0.0000% | 12.19 | 25.36 | 25.36 | true |
| spec_power_asc_p99_previous | 15 | 0 | 0.0000% | 13.39 | 21.06 | 21.06 | true |
| spec_power_desc_p25_next | 15 | 0 | 0.0000% | 11.66 | 14.71 | 14.71 | true |
| spec_power_desc_p25_previous | 15 | 0 | 0.0000% | 15.39 | 47.09 | 47.09 | true |
| spec_power_desc_p50_next | 15 | 0 | 0.0000% | 11.83 | 13.73 | 13.73 | true |
| spec_power_desc_p50_previous | 15 | 0 | 0.0000% | 11.04 | 14.15 | 14.15 | true |
| spec_power_desc_p75_next | 15 | 0 | 0.0000% | 12.01 | 21.24 | 21.24 | true |
| spec_power_desc_p75_previous | 15 | 0 | 0.0000% | 10.88 | 44.13 | 44.13 | true |
| spec_power_desc_p99_next | 15 | 0 | 0.0000% | 13.61 | 19.44 | 19.44 | true |
| spec_power_desc_p99_previous | 15 | 0 | 0.0000% | 11.86 | 14.91 | 14.91 | true |
