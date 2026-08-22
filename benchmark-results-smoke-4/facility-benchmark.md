# Facility benchmark (50000 FieldDevices)

Git: `da4128276ebd2fb1ae2207675929301a8b942f91`  
Generated: 2026-08-22T09:12:49Z

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

- `go_arch`: `amd64`
- `go_os`: `windows`
- `logical_cpus`: `8`

## Database

- `database_bytes`: `359364287`
- `expected_field_devices`: `50000`
- `field_device_cursor_values_rows`: `50000`
- `field_devices_rows`: `50000`
- `project_field_devices_rows`: `50000`
- `schema_migrations`: `30`
- `specifications_rows`: `50000`

| Scenario | Samples | Failures | p50 ms | p95 ms | p99 ms | Gate |
|---|---:|---:|---:|---:|---:|---|
| created_at_asc_first | 15 | 0 | 14.66 | 24.45 | 24.45 | true |
| created_at_asc_next | 15 | 0 | 9.79 | 17.56 | 17.56 | true |
| created_at_desc_first | 15 | 0 | 8.61 | 10.07 | 10.07 | true |
| created_at_desc_next | 15 | 0 | 9.87 | 13.63 | 13.63 | true |
| apparat_nr_asc_first | 15 | 0 | 9.68 | 15.97 | 15.97 | true |
| apparat_nr_asc_next | 15 | 0 | 10.48 | 15.83 | 15.83 | true |
| apparat_nr_desc_first | 15 | 0 | 8.99 | 14.66 | 14.66 | true |
| apparat_nr_desc_next | 15 | 0 | 10.10 | 13.97 | 13.97 | true |
| sps_system_type_asc_first | 15 | 0 | 10.83 | 13.94 | 13.94 | true |
| sps_system_type_asc_next | 15 | 0 | 10.28 | 12.99 | 12.99 | true |
| sps_system_type_desc_first | 15 | 0 | 12.97 | 18.19 | 18.19 | true |
| sps_system_type_desc_next | 15 | 0 | 12.51 | 16.20 | 16.20 | true |
| description_asc_first | 15 | 0 | 10.31 | 12.78 | 12.78 | true |
| description_asc_next | 15 | 0 | 10.80 | 20.94 | 20.94 | true |
| description_desc_first | 15 | 0 | 16.83 | 23.50 | 23.50 | true |
| description_desc_next | 15 | 0 | 12.41 | 21.37 | 21.37 | true |
| spec_supplier_asc_first | 15 | 0 | 10.85 | 17.40 | 17.40 | true |
| spec_supplier_asc_next | 15 | 0 | 11.85 | 36.84 | 36.84 | true |
| spec_supplier_desc_first | 15 | 0 | 11.98 | 19.97 | 19.97 | true |
| spec_supplier_desc_next | 15 | 0 | 11.07 | 19.54 | 19.54 | true |
| building_scope_first | 15 | 0 | 12.61 | 18.11 | 18.11 | true |
| building_scope_first_previous | 15 | 0 | 14.69 | 19.26 | 19.26 | true |
| building_scope_first_previous_next | 15 | 0 | 13.61 | 19.11 | 19.11 | true |
| cabinet_scope_first | 15 | 0 | 7.68 | 9.80 | 9.80 | true |
| cabinet_scope_first_previous | 15 | 0 | 4.30 | 9.11 | 9.11 | true |
| cabinet_scope_first_previous_next | 15 | 0 | 3.48 | 5.22 | 5.22 | true |
| controller_scope_first | 15 | 0 | 4.18 | 5.55 | 5.55 | true |
| controller_scope_first_previous | 15 | 0 | 1.27 | 1.71 | 1.71 | true |
| controller_scope_first_previous_next | 15 | 0 | 1.68 | 2.39 | 2.39 | true |
| project_scope_first | 15 | 0 | 13.04 | 16.45 | 16.45 | true |
| project_scope_first_previous | 15 | 0 | 11.67 | 15.18 | 15.18 | true |
| project_scope_first_previous_next | 15 | 0 | 13.75 | 19.13 | 19.13 | true |
| search_0_1_percent_first | 15 | 0 | 4.54 | 11.84 | 11.84 | true |
| search_0_1_percent_first_previous | 15 | 0 | 3.24 | 3.78 | 3.78 | true |
| search_0_1_percent_first_previous_next | 15 | 0 | 3.21 | 4.29 | 4.29 | true |
| search_1_percent_first | 15 | 0 | 26.22 | 31.31 | 31.31 | true |
| search_1_percent_first_previous | 15 | 0 | 12.45 | 17.19 | 17.19 | true |
| search_1_percent_first_previous_next | 15 | 0 | 11.95 | 21.90 | 21.90 | true |
| search_10_percent_first | 15 | 0 | 22.50 | 28.30 | 28.30 | true |
| search_10_percent_first_previous | 15 | 0 | 11.69 | 14.01 | 14.01 | true |
| search_10_percent_first_previous_next | 15 | 0 | 11.47 | 17.44 | 17.44 | true |
| combined_filter_first | 15 | 0 | 1.66 | 2.11 | 2.11 | true |
| combined_filter_first_previous | 15 | 0 | 0.98 | 1.40 | 1.40 | true |
| combined_filter_first_previous_next | 15 | 0 | 1.60 | 2.91 | 2.91 | true |
| created_at_asc_p25_next | 15 | 0 | 8.92 | 12.96 | 12.96 | true |
| created_at_asc_p25_previous | 15 | 0 | 14.47 | 18.50 | 18.50 | true |
| created_at_asc_p50_next | 15 | 0 | 12.23 | 21.47 | 21.47 | true |
| created_at_asc_p50_previous | 15 | 0 | 9.60 | 13.79 | 13.79 | true |
| created_at_asc_p75_next | 15 | 0 | 12.49 | 17.61 | 17.61 | true |
| created_at_asc_p75_previous | 15 | 0 | 12.29 | 18.73 | 18.73 | true |
| created_at_asc_p99_next | 15 | 0 | 12.24 | 15.34 | 15.34 | true |
| created_at_asc_p99_previous | 15 | 0 | 8.61 | 12.50 | 12.50 | true |
| created_at_desc_p25_next | 15 | 0 | 14.18 | 22.39 | 22.39 | true |
| created_at_desc_p25_previous | 15 | 0 | 13.75 | 20.13 | 20.13 | true |
| created_at_desc_p50_next | 15 | 0 | 15.36 | 313.73 | 313.73 | true |
| created_at_desc_p50_previous | 15 | 0 | 10.87 | 14.64 | 14.64 | true |
| created_at_desc_p75_next | 15 | 0 | 16.37 | 40.59 | 40.59 | true |
| created_at_desc_p75_previous | 15 | 0 | 13.16 | 60.60 | 60.60 | true |
| created_at_desc_p99_next | 15 | 0 | 9.23 | 13.58 | 13.58 | true |
| created_at_desc_p99_previous | 15 | 0 | 8.35 | 11.59 | 11.59 | true |
| apparat_nr_asc_p25_next | 15 | 0 | 11.67 | 20.69 | 20.69 | true |
| apparat_nr_asc_p25_previous | 15 | 0 | 17.87 | 28.71 | 28.71 | true |
| apparat_nr_asc_p50_next | 15 | 0 | 15.90 | 20.11 | 20.11 | true |
| apparat_nr_asc_p50_previous | 15 | 0 | 14.01 | 33.08 | 33.08 | true |
| apparat_nr_asc_p75_next | 15 | 0 | 13.97 | 55.50 | 55.50 | true |
| apparat_nr_asc_p75_previous | 15 | 0 | 10.41 | 12.22 | 12.22 | true |
| apparat_nr_asc_p99_next | 15 | 0 | 12.66 | 20.47 | 20.47 | true |
| apparat_nr_asc_p99_previous | 15 | 0 | 12.40 | 17.16 | 17.16 | true |
| apparat_nr_desc_p25_next | 15 | 0 | 19.32 | 25.32 | 25.32 | true |
| apparat_nr_desc_p25_previous | 15 | 0 | 20.73 | 27.30 | 27.30 | true |
| apparat_nr_desc_p50_next | 15 | 0 | 15.41 | 31.98 | 31.98 | true |
| apparat_nr_desc_p50_previous | 15 | 0 | 11.18 | 16.34 | 16.34 | true |
| apparat_nr_desc_p75_next | 15 | 0 | 11.79 | 13.82 | 13.82 | true |
| apparat_nr_desc_p75_previous | 15 | 0 | 9.44 | 14.50 | 14.50 | true |
| apparat_nr_desc_p99_next | 15 | 0 | 9.79 | 308.50 | 308.50 | true |
| apparat_nr_desc_p99_previous | 15 | 0 | 7.68 | 11.56 | 11.56 | true |
| sps_system_type_asc_p25_next | 15 | 0 | 11.48 | 14.88 | 14.88 | true |
| sps_system_type_asc_p25_previous | 15 | 0 | 17.46 | 24.54 | 24.54 | true |
| sps_system_type_asc_p50_next | 15 | 0 | 14.77 | 18.53 | 18.53 | true |
| sps_system_type_asc_p50_previous | 15 | 0 | 14.84 | 18.57 | 18.57 | true |
| sps_system_type_asc_p75_next | 15 | 0 | 19.16 | 28.84 | 28.84 | true |
| sps_system_type_asc_p75_previous | 15 | 0 | 11.71 | 15.07 | 15.07 | true |
| sps_system_type_asc_p99_next | 15 | 0 | 20.14 | 24.19 | 24.19 | true |
| sps_system_type_asc_p99_previous | 15 | 0 | 9.88 | 18.31 | 18.31 | true |
| sps_system_type_desc_p25_next | 15 | 0 | 14.73 | 24.26 | 24.26 | true |
| sps_system_type_desc_p25_previous | 15 | 0 | 18.82 | 24.16 | 24.16 | true |
| sps_system_type_desc_p50_next | 15 | 0 | 15.68 | 18.97 | 18.97 | true |
| sps_system_type_desc_p50_previous | 15 | 0 | 15.74 | 59.87 | 59.87 | true |
| sps_system_type_desc_p75_next | 15 | 0 | 16.83 | 28.74 | 28.74 | true |
| sps_system_type_desc_p75_previous | 15 | 0 | 11.30 | 16.01 | 16.01 | true |
| sps_system_type_desc_p99_next | 15 | 0 | 19.25 | 21.90 | 21.90 | true |
| sps_system_type_desc_p99_previous | 15 | 0 | 10.08 | 21.32 | 21.32 | true |
| description_asc_p25_next | 15 | 0 | 21.46 | 39.31 | 39.31 | true |
| description_asc_p25_previous | 15 | 0 | 18.32 | 39.76 | 39.76 | true |
| description_asc_p50_next | 15 | 0 | 14.75 | 26.10 | 26.10 | true |
| description_asc_p50_previous | 15 | 0 | 14.77 | 19.23 | 19.23 | true |
| description_asc_p75_next | 15 | 0 | 17.85 | 20.26 | 20.26 | true |
| description_asc_p75_previous | 15 | 0 | 9.99 | 31.74 | 31.74 | true |
| description_asc_p99_next | 15 | 0 | 9.59 | 10.69 | 10.69 | true |
| description_asc_p99_previous | 15 | 0 | 12.51 | 19.96 | 19.96 | true |
| description_desc_p25_next | 15 | 0 | 15.41 | 23.93 | 23.93 | true |
| description_desc_p25_previous | 15 | 0 | 18.26 | 23.33 | 23.33 | true |
| description_desc_p50_next | 15 | 0 | 16.07 | 23.48 | 23.48 | true |
| description_desc_p50_previous | 15 | 0 | 16.09 | 18.78 | 18.78 | true |
| description_desc_p75_next | 15 | 0 | 21.52 | 24.07 | 24.07 | true |
| description_desc_p75_previous | 15 | 0 | 11.25 | 20.53 | 20.53 | true |
| description_desc_p99_next | 15 | 0 | 11.54 | 18.60 | 18.60 | true |
| description_desc_p99_previous | 15 | 0 | 13.61 | 17.02 | 17.02 | true |
| spec_supplier_asc_p25_next | 15 | 0 | 13.27 | 16.83 | 16.83 | true |
| spec_supplier_asc_p25_previous | 15 | 0 | 15.66 | 19.46 | 19.46 | true |
| spec_supplier_asc_p50_next | 15 | 0 | 15.31 | 19.46 | 19.46 | true |
| spec_supplier_asc_p50_previous | 15 | 0 | 13.80 | 21.02 | 21.02 | true |
| spec_supplier_asc_p75_next | 15 | 0 | 18.05 | 20.34 | 20.34 | true |
| spec_supplier_asc_p75_previous | 15 | 0 | 11.08 | 11.87 | 11.87 | true |
| spec_supplier_asc_p99_next | 15 | 0 | 17.61 | 21.30 | 21.30 | true |
| spec_supplier_asc_p99_previous | 15 | 0 | 12.52 | 21.94 | 21.94 | true |
| spec_supplier_desc_p25_next | 15 | 0 | 12.28 | 15.32 | 15.32 | true |
| spec_supplier_desc_p25_previous | 15 | 0 | 17.91 | 31.69 | 31.69 | true |
| spec_supplier_desc_p50_next | 15 | 0 | 15.04 | 19.90 | 19.90 | true |
| spec_supplier_desc_p50_previous | 15 | 0 | 15.74 | 26.11 | 26.11 | true |
| spec_supplier_desc_p75_next | 15 | 0 | 21.75 | 30.79 | 30.79 | true |
| spec_supplier_desc_p75_previous | 15 | 0 | 12.02 | 320.58 | 320.58 | true |
| spec_supplier_desc_p99_next | 15 | 0 | 23.37 | 33.03 | 33.03 | true |
| spec_supplier_desc_p99_previous | 15 | 0 | 11.47 | 20.30 | 20.30 | true |
