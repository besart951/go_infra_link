# Facility benchmark (5000000 FieldDevices)

Git: `da4128276ebd2fb1ae2207675929301a8b942f91`  
Generated: 2026-08-22T12:41:25Z

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
| created_at_asc_first | 15 | 0 | 0.0000% | 12.99 | 19.78 | 19.78 | true |
| created_at_asc_next | 15 | 0 | 0.0000% | 28.48 | 44.82 | 44.82 | true |
| created_at_desc_first | 15 | 0 | 0.0000% | 15.25 | 20.64 | 20.64 | true |
| created_at_desc_next | 15 | 0 | 0.0000% | 15.50 | 20.22 | 20.22 | true |
| apparat_nr_asc_first | 15 | 0 | 0.0000% | 12.92 | 15.37 | 15.37 | true |
| apparat_nr_asc_next | 15 | 0 | 0.0000% | 18.28 | 35.91 | 35.91 | true |
| apparat_nr_desc_first | 15 | 0 | 0.0000% | 12.44 | 318.96 | 318.96 | true |
| apparat_nr_desc_next | 15 | 0 | 0.0000% | 14.89 | 24.62 | 24.62 | true |
| sps_system_type_asc_first | 15 | 0 | 0.0000% | 18.84 | 21.85 | 21.85 | true |
| sps_system_type_asc_next | 15 | 0 | 0.0000% | 17.93 | 21.36 | 21.36 | true |
| sps_system_type_desc_first | 15 | 0 | 0.0000% | 19.43 | 29.44 | 29.44 | true |
| sps_system_type_desc_next | 15 | 0 | 0.0000% | 17.69 | 21.12 | 21.12 | true |
| description_asc_first | 15 | 0 | 0.0000% | 13.94 | 16.50 | 16.50 | true |
| description_asc_next | 15 | 0 | 0.0000% | 15.01 | 16.91 | 16.91 | true |
| description_desc_first | 15 | 0 | 0.0000% | 15.39 | 17.86 | 17.86 | true |
| description_desc_next | 15 | 0 | 0.0000% | 15.81 | 19.10 | 19.10 | true |
| spec_supplier_asc_first | 15 | 0 | 0.0000% | 13.98 | 15.47 | 15.47 | true |
| spec_supplier_asc_next | 15 | 0 | 0.0000% | 12.87 | 16.09 | 16.09 | true |
| spec_supplier_desc_first | 15 | 0 | 0.0000% | 13.50 | 35.91 | 35.91 | true |
| spec_supplier_desc_next | 15 | 0 | 0.0000% | 16.75 | 22.53 | 22.53 | true |
| spec_size_asc_first | 15 | 0 | 0.0000% | 17.84 | 24.43 | 24.43 | true |
| spec_size_asc_next | 15 | 0 | 0.0000% | 17.51 | 23.89 | 23.89 | true |
| spec_size_desc_first | 15 | 0 | 0.0000% | 16.67 | 20.43 | 20.43 | true |
| spec_size_desc_next | 15 | 0 | 0.0000% | 17.04 | 19.13 | 19.13 | true |
| spec_acdc_asc_first | 15 | 0 | 0.0000% | 13.49 | 15.36 | 15.36 | true |
| spec_acdc_asc_next | 15 | 0 | 0.0000% | 14.32 | 16.51 | 16.51 | true |
| spec_acdc_desc_first | 15 | 0 | 0.0000% | 14.23 | 16.97 | 16.97 | true |
| spec_acdc_desc_next | 15 | 0 | 0.0000% | 12.92 | 16.50 | 16.50 | true |
| spec_power_asc_first | 15 | 0 | 0.0000% | 13.12 | 15.50 | 15.50 | true |
| spec_power_asc_next | 15 | 0 | 0.0000% | 13.20 | 16.37 | 16.37 | true |
| spec_power_desc_first | 15 | 0 | 0.0000% | 12.37 | 33.06 | 33.06 | true |
| spec_power_desc_next | 15 | 0 | 0.0000% | 25.06 | 47.98 | 47.98 | true |
| building_scope_first | 15 | 0 | 0.0000% | 341.75 | 426.50 | 426.50 | true |
| building_scope_previous | 15 | 0 | 0.0000% | 282.62 | 337.79 | 337.79 | true |
| building_scope_next | 15 | 0 | 0.0000% | 345.24 | 372.26 | 372.26 | true |
| cabinet_scope_first | 15 | 0 | 0.0000% | 38.33 | 48.52 | 48.52 | true |
| cabinet_scope_previous | 15 | 0 | 0.0000% | 24.54 | 29.34 | 29.34 | true |
| cabinet_scope_next | 15 | 0 | 0.0000% | 29.59 | 34.45 | 34.45 | true |
| controller_scope_first | 15 | 0 | 0.0000% | 14.68 | 21.35 | 21.35 | true |
| controller_scope_previous | 15 | 0 | 0.0000% | 16.24 | 23.53 | 23.53 | true |
| controller_scope_next | 15 | 0 | 0.0000% | 12.66 | 15.53 | 15.53 | true |
| project_scope_first | 15 | 0 | 0.0000% | 419.11 | 474.54 | 474.54 | true |
| project_scope_previous | 15 | 0 | 0.0000% | 101.46 | 117.69 | 117.69 | true |
| project_scope_next | 15 | 0 | 0.0000% | 430.65 | 491.54 | 491.54 | true |
| search_0_1_percent_first | 15 | 0 | 0.0000% | 66.21 | 69.03 | 69.03 | true |
| search_0_1_percent_previous | 15 | 0 | 0.0000% | 49.43 | 54.21 | 54.21 | true |
| search_0_1_percent_next | 15 | 0 | 0.0000% | 66.87 | 85.68 | 85.68 | true |
| search_1_percent_first | 15 | 0 | 0.0000% | 150.19 | 218.65 | 218.65 | true |
| search_1_percent_previous | 15 | 0 | 0.0000% | 64.99 | 73.86 | 73.86 | true |
| search_1_percent_next | 15 | 0 | 0.0000% | 154.03 | 187.30 | 187.30 | true |
| search_10_percent_first | 15 | 0 | 0.0000% | 13.54 | 16.66 | 16.66 | true |
| search_10_percent_previous | 15 | 0 | 0.0000% | 253.89 | 271.88 | 271.88 | true |
| search_10_percent_next | 15 | 0 | 0.0000% | 16.52 | 24.72 | 24.72 | true |
| combined_filter_first | 15 | 0 | 0.0000% | 14.67 | 32.70 | 32.70 | true |
| combined_filter_previous | 15 | 0 | 0.0000% | 10.34 | 17.39 | 17.39 | true |
| combined_filter_next | 15 | 0 | 0.0000% | 10.67 | 13.22 | 13.22 | true |
| created_at_asc_p25_next | 15 | 0 | 0.0000% | 141.30 | 212.94 | 212.94 | true |
| created_at_asc_p25_previous | 15 | 0 | 0.0000% | 385.88 | 441.34 | 441.34 | true |
| created_at_asc_p50_next | 15 | 0 | 0.0000% | 278.97 | 339.98 | 339.98 | true |
| created_at_asc_p50_previous | 15 | 0 | 0.0000% | 259.49 | 378.28 | 378.28 | true |
| created_at_asc_p75_next | 15 | 0 | 0.0000% | 414.75 | 499.85 | 499.85 | true |
| created_at_asc_p75_previous | 15 | 0 | 0.0000% | 139.88 | 182.74 | 182.74 | true |
| created_at_asc_p99_next | 15 | 0 | 0.0000% | 232.82 | 289.49 | 289.49 | true |
| created_at_asc_p99_previous | 15 | 0 | 0.0000% | 18.55 | 20.93 | 20.93 | true |
| created_at_desc_p25_next | 15 | 0 | 0.0000% | 144.93 | 180.02 | 180.02 | true |
| created_at_desc_p25_previous | 15 | 0 | 0.0000% | 424.49 | 456.64 | 456.64 | true |
| created_at_desc_p50_next | 15 | 0 | 0.0000% | 254.08 | 306.18 | 306.18 | true |
| created_at_desc_p50_previous | 15 | 0 | 0.0000% | 301.21 | 353.94 | 353.94 | true |
| created_at_desc_p75_next | 15 | 0 | 0.0000% | 378.05 | 476.00 | 476.00 | true |
| created_at_desc_p75_previous | 15 | 0 | 0.0000% | 148.56 | 182.10 | 182.10 | true |
| created_at_desc_p99_next | 15 | 0 | 0.0000% | 244.74 | 268.03 | 268.03 | true |
| created_at_desc_p99_previous | 15 | 0 | 0.0000% | 18.75 | 22.41 | 22.41 | true |
| apparat_nr_asc_p25_next | 15 | 0 | 0.0000% | 168.18 | 228.22 | 228.22 | true |
| apparat_nr_asc_p25_previous | 15 | 0 | 0.0000% | 388.13 | 423.94 | 423.94 | true |
| apparat_nr_asc_p50_next | 15 | 0 | 0.0000% | 288.50 | 333.81 | 333.81 | true |
| apparat_nr_asc_p50_previous | 15 | 0 | 0.0000% | 268.13 | 328.58 | 328.58 | true |
| apparat_nr_asc_p75_next | 15 | 0 | 0.0000% | 442.15 | 504.82 | 504.82 | true |
| apparat_nr_asc_p75_previous | 15 | 0 | 0.0000% | 135.89 | 164.88 | 164.88 | true |
| apparat_nr_asc_p99_next | 15 | 0 | 0.0000% | 131.48 | 146.20 | 146.20 | true |
| apparat_nr_asc_p99_previous | 15 | 0 | 0.0000% | 15.91 | 20.41 | 20.41 | true |
| apparat_nr_desc_p25_next | 15 | 0 | 0.0000% | 162.19 | 188.71 | 188.71 | true |
| apparat_nr_desc_p25_previous | 15 | 0 | 0.0000% | 414.59 | 458.17 | 458.17 | true |
| apparat_nr_desc_p50_next | 15 | 0 | 0.0000% | 303.13 | 329.16 | 329.16 | true |
| apparat_nr_desc_p50_previous | 15 | 0 | 0.0000% | 309.29 | 397.79 | 397.79 | true |
| apparat_nr_desc_p75_next | 15 | 0 | 0.0000% | 458.88 | 529.56 | 529.56 | true |
| apparat_nr_desc_p75_previous | 15 | 0 | 0.0000% | 160.37 | 198.53 | 198.53 | true |
| apparat_nr_desc_p99_next | 15 | 0 | 0.0000% | 152.13 | 175.31 | 175.31 | true |
| apparat_nr_desc_p99_previous | 15 | 0 | 0.0000% | 18.81 | 31.68 | 31.68 | true |
| sps_system_type_asc_p25_next | 15 | 0 | 0.0000% | 103.08 | 158.12 | 158.12 | true |
| sps_system_type_asc_p25_previous | 15 | 0 | 0.0000% | 254.17 | 266.90 | 266.90 | true |
| sps_system_type_asc_p50_next | 15 | 0 | 0.0000% | 172.63 | 198.84 | 198.84 | true |
| sps_system_type_asc_p50_previous | 15 | 0 | 0.0000% | 156.91 | 168.07 | 168.07 | true |
| sps_system_type_asc_p75_next | 15 | 0 | 0.0000% | 258.12 | 299.58 | 299.58 | true |
| sps_system_type_asc_p75_previous | 15 | 0 | 0.0000% | 82.39 | 139.23 | 139.23 | true |
| sps_system_type_asc_p99_next | 15 | 0 | 0.0000% | 407.48 | 432.60 | 432.60 | true |
| sps_system_type_asc_p99_previous | 15 | 0 | 0.0000% | 20.33 | 25.68 | 25.68 | true |
| sps_system_type_desc_p25_next | 15 | 0 | 0.0000% | 89.16 | 110.96 | 110.96 | true |
| sps_system_type_desc_p25_previous | 15 | 0 | 0.0000% | 234.36 | 286.41 | 286.41 | true |
| sps_system_type_desc_p50_next | 15 | 0 | 0.0000% | 186.98 | 209.81 | 209.81 | true |
| sps_system_type_desc_p50_previous | 15 | 0 | 0.0000% | 169.00 | 249.18 | 249.18 | true |
| sps_system_type_desc_p75_next | 15 | 0 | 0.0000% | 263.16 | 296.59 | 296.59 | true |
| sps_system_type_desc_p75_previous | 15 | 0 | 0.0000% | 87.24 | 107.96 | 107.96 | true |
| sps_system_type_desc_p99_next | 15 | 0 | 0.0000% | 356.20 | 377.44 | 377.44 | true |
| sps_system_type_desc_p99_previous | 15 | 0 | 0.0000% | 21.00 | 23.66 | 23.66 | true |
| description_asc_p25_next | 15 | 0 | 0.0000% | 360.20 | 423.92 | 423.92 | true |
| description_asc_p25_previous | 15 | 0 | 0.0000% | 899.88 | 943.49 | 943.49 | false |
| description_asc_p50_next | 15 | 0 | 0.0000% | 713.70 | 775.83 | 775.83 | false |
| description_asc_p50_previous | 15 | 0 | 0.0000% | 601.73 | 680.94 | 680.94 | true |
| description_asc_p75_next | 15 | 0 | 0.0000% | 1061.27 | 1133.09 | 1133.09 | false |
| description_asc_p75_previous | 15 | 0 | 0.0000% | 272.53 | 352.85 | 352.85 | true |
| description_asc_p99_next | 15 | 0 | 0.0000% | 14.68 | 25.76 | 25.76 | true |
| description_asc_p99_previous | 15 | 0 | 0.0000% | 11.30 | 25.49 | 25.49 | true |
| description_desc_p25_next | 15 | 0 | 0.0000% | 454.10 | 484.79 | 484.79 | true |
| description_desc_p25_previous | 15 | 0 | 0.0000% | 826.87 | 900.23 | 900.23 | false |
| description_desc_p50_next | 15 | 0 | 0.0000% | 813.87 | 873.51 | 873.51 | false |
| description_desc_p50_previous | 15 | 0 | 0.0000% | 495.05 | 557.93 | 557.93 | true |
| description_desc_p75_next | 15 | 0 | 0.0000% | 1121.83 | 1318.16 | 1318.16 | false |
| description_desc_p75_previous | 15 | 0 | 0.0000% | 197.76 | 202.85 | 202.85 | true |
| description_desc_p99_next | 15 | 0 | 0.0000% | 15.71 | 19.96 | 19.96 | true |
| description_desc_p99_previous | 15 | 0 | 0.0000% | 13.42 | 17.59 | 17.59 | true |
| spec_supplier_asc_p25_next | 15 | 0 | 0.0000% | 329.38 | 405.85 | 405.85 | true |
| spec_supplier_asc_p25_previous | 15 | 0 | 0.0000% | 753.36 | 790.40 | 790.40 | false |
| spec_supplier_asc_p50_next | 15 | 0 | 0.0000% | 683.75 | 739.07 | 739.07 | true |
| spec_supplier_asc_p50_previous | 15 | 0 | 0.0000% | 432.40 | 518.31 | 518.31 | true |
| spec_supplier_asc_p75_next | 15 | 0 | 0.0000% | 1000.45 | 1076.80 | 1076.80 | false |
| spec_supplier_asc_p75_previous | 15 | 0 | 0.0000% | 89.53 | 102.28 | 102.28 | true |
| spec_supplier_asc_p99_next | 15 | 0 | 0.0000% | 16.73 | 26.95 | 26.95 | true |
| spec_supplier_asc_p99_previous | 15 | 0 | 0.0000% | 17.70 | 320.35 | 320.35 | true |
| spec_supplier_desc_p25_next | 15 | 0 | 0.0000% | 319.93 | 434.98 | 434.98 | true |
| spec_supplier_desc_p25_previous | 15 | 0 | 0.0000% | 780.81 | 823.19 | 823.19 | false |
| spec_supplier_desc_p50_next | 15 | 0 | 0.0000% | 653.44 | 677.88 | 677.88 | true |
| spec_supplier_desc_p50_previous | 15 | 0 | 0.0000% | 457.68 | 505.37 | 505.37 | true |
| spec_supplier_desc_p75_next | 15 | 0 | 0.0000% | 976.87 | 1002.76 | 1002.76 | false |
| spec_supplier_desc_p75_previous | 15 | 0 | 0.0000% | 115.86 | 130.86 | 130.86 | true |
| spec_supplier_desc_p99_next | 15 | 0 | 0.0000% | 12.33 | 14.11 | 14.11 | true |
| spec_supplier_desc_p99_previous | 15 | 0 | 0.0000% | 12.24 | 14.16 | 14.16 | true |
| spec_size_asc_p25_next | 15 | 0 | 0.0000% | 80.84 | 96.86 | 96.86 | true |
| spec_size_asc_p25_previous | 15 | 0 | 0.0000% | 194.76 | 237.44 | 237.44 | true |
| spec_size_asc_p50_next | 15 | 0 | 0.0000% | 162.66 | 227.75 | 227.75 | true |
| spec_size_asc_p50_previous | 15 | 0 | 0.0000% | 129.82 | 169.69 | 169.69 | true |
| spec_size_asc_p75_next | 15 | 0 | 0.0000% | 217.31 | 275.36 | 275.36 | true |
| spec_size_asc_p75_previous | 15 | 0 | 0.0000% | 74.30 | 96.48 | 96.48 | true |
| spec_size_asc_p99_next | 15 | 0 | 0.0000% | 284.06 | 363.95 | 363.95 | true |
| spec_size_asc_p99_previous | 15 | 0 | 0.0000% | 19.18 | 22.90 | 22.90 | true |
| spec_size_desc_p25_next | 15 | 0 | 0.0000% | 92.27 | 148.50 | 148.50 | true |
| spec_size_desc_p25_previous | 15 | 0 | 0.0000% | 205.13 | 241.29 | 241.29 | true |
| spec_size_desc_p50_next | 15 | 0 | 0.0000% | 150.15 | 185.14 | 185.14 | true |
| spec_size_desc_p50_previous | 15 | 0 | 0.0000% | 149.50 | 176.80 | 176.80 | true |
| spec_size_desc_p75_next | 15 | 0 | 0.0000% | 227.62 | 257.66 | 257.66 | true |
| spec_size_desc_p75_previous | 15 | 0 | 0.0000% | 77.96 | 91.62 | 91.62 | true |
| spec_size_desc_p99_next | 15 | 0 | 0.0000% | 297.33 | 305.11 | 305.11 | true |
| spec_size_desc_p99_previous | 15 | 0 | 0.0000% | 18.63 | 20.95 | 20.95 | true |
| spec_acdc_asc_p25_next | 15 | 0 | 0.0000% | 12.58 | 25.50 | 25.50 | true |
| spec_acdc_asc_p25_previous | 15 | 0 | 0.0000% | 11.79 | 14.86 | 14.86 | true |
| spec_acdc_asc_p50_next | 15 | 0 | 0.0000% | 19.55 | 29.41 | 29.41 | true |
| spec_acdc_asc_p50_previous | 15 | 0 | 0.0000% | 12.02 | 16.45 | 16.45 | true |
| spec_acdc_asc_p75_next | 15 | 0 | 0.0000% | 11.88 | 15.14 | 15.14 | true |
| spec_acdc_asc_p75_previous | 15 | 0 | 0.0000% | 11.76 | 15.53 | 15.53 | true |
| spec_acdc_asc_p99_next | 15 | 0 | 0.0000% | 13.84 | 16.93 | 16.93 | true |
| spec_acdc_asc_p99_previous | 15 | 0 | 0.0000% | 12.19 | 13.95 | 13.95 | true |
| spec_acdc_desc_p25_next | 15 | 0 | 0.0000% | 12.24 | 14.90 | 14.90 | true |
| spec_acdc_desc_p25_previous | 15 | 0 | 0.0000% | 13.08 | 14.28 | 14.28 | true |
| spec_acdc_desc_p50_next | 15 | 0 | 0.0000% | 11.13 | 14.17 | 14.17 | true |
| spec_acdc_desc_p50_previous | 15 | 0 | 0.0000% | 11.75 | 13.93 | 13.93 | true |
| spec_acdc_desc_p75_next | 15 | 0 | 0.0000% | 13.16 | 15.98 | 15.98 | true |
| spec_acdc_desc_p75_previous | 15 | 0 | 0.0000% | 11.91 | 15.09 | 15.09 | true |
| spec_acdc_desc_p99_next | 15 | 0 | 0.0000% | 11.43 | 13.64 | 13.64 | true |
| spec_acdc_desc_p99_previous | 15 | 0 | 0.0000% | 10.99 | 12.19 | 12.19 | true |
| spec_power_asc_p25_next | 15 | 0 | 0.0000% | 13.39 | 20.35 | 20.35 | true |
| spec_power_asc_p25_previous | 15 | 0 | 0.0000% | 12.12 | 14.80 | 14.80 | true |
| spec_power_asc_p50_next | 15 | 0 | 0.0000% | 12.70 | 15.92 | 15.92 | true |
| spec_power_asc_p50_previous | 15 | 0 | 0.0000% | 20.83 | 26.50 | 26.50 | true |
| spec_power_asc_p75_next | 15 | 0 | 0.0000% | 12.83 | 16.58 | 16.58 | true |
| spec_power_asc_p75_previous | 15 | 0 | 0.0000% | 12.66 | 16.64 | 16.64 | true |
| spec_power_asc_p99_next | 15 | 0 | 0.0000% | 12.26 | 14.69 | 14.69 | true |
| spec_power_asc_p99_previous | 15 | 0 | 0.0000% | 11.37 | 15.27 | 15.27 | true |
| spec_power_desc_p25_next | 15 | 0 | 0.0000% | 13.67 | 18.00 | 18.00 | true |
| spec_power_desc_p25_previous | 15 | 0 | 0.0000% | 11.91 | 14.80 | 14.80 | true |
| spec_power_desc_p50_next | 15 | 0 | 0.0000% | 11.66 | 15.12 | 15.12 | true |
| spec_power_desc_p50_previous | 15 | 0 | 0.0000% | 12.44 | 13.91 | 13.91 | true |
| spec_power_desc_p75_next | 15 | 0 | 0.0000% | 12.30 | 14.57 | 14.57 | true |
| spec_power_desc_p75_previous | 15 | 0 | 0.0000% | 13.69 | 15.70 | 15.70 | true |
| spec_power_desc_p99_next | 15 | 0 | 0.0000% | 12.32 | 311.52 | 311.52 | true |
| spec_power_desc_p99_previous | 15 | 0 | 0.0000% | 16.01 | 20.53 | 20.53 | true |
