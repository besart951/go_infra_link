# Field Device Multi-Create Feature

## Overview

The multi-create endpoint allows you to create multiple field devices in a single API request. Each field device is validated independently, and the endpoint continues processing even if individual devices fail validation.

## Endpoint

```
POST /api/v1/facility/field-devices/multi-create
```

## Features

1. **Independent Validation**: Each field device is validated independently
2. **Continue on Failure**: If one device fails validation, the endpoint continues processing the remaining devices
3. **Detailed Error Reporting**: Returns specific error information for each failed device
4. **Apparat Number Uniqueness**: Validates that `apparat_nr` is unique for each combination of:
   - `sps_controller_system_type_id`
   - `apparat_id`
   - `system_part_id`

## Request Format

```json
{
  "field_devices": [
    {
      "bmk": "BMK1",
      "description": "First device",
      "apparat_nr": 1,
      "sps_controller_system_type_id": "uuid",
      "system_part_id": "uuid",
      "apparat_id": "uuid",
      "object_data_id": "uuid (optional)",
      "bacnet_objects": []
    },
    {
      "bmk": "BMK2",
      "description": "Second device",
      "apparat_nr": 2,
      "sps_controller_system_type_id": "uuid",
      "system_part_id": "uuid",
      "apparat_id": "uuid"
    }
  ]
}
```

## Response Format

```json
{
  "results": [
    {
      "index": 0,
      "success": true,
      "field_device": {
        "id": "uuid",
        "bmk": "BMK1",
        "description": "First device",
        "apparat_nr": 1,
        "sps_controller_system_type_id": "uuid",
        "system_part_id": "uuid",
        "apparat_id": "uuid",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      },
      "error": "",
      "error_field": ""
    },
    {
      "index": 1,
      "success": false,
      "field_device": null,
      "error": "apparatnummer ist bereits vergeben",
      "error_field": "fielddevice.apparat_nr"
    }
  ],
  "total_requests": 2,
  "success_count": 1,
  "failure_count": 1
}
```

## Validation Rules

### Per-Device Validation

1. **Required Fields**:
   - `apparat_nr`: Must be between 1 and 99
   - `sps_controller_system_type_id`: Must reference an existing SPS controller system type
   - `apparat_id`: Must reference an existing apparat
   - `system_part_id`: Must reference an existing system part

2. **Uniqueness**:
   - `apparat_nr` must be unique for the combination of:
     - `sps_controller_system_type_id`
     - `apparat_id`  
     - `system_part_id`

3. **Parent Entity Validation**:
   - All referenced entities (SPS controller system type, apparat, system part) must exist and not be soft-deleted

4. **Mutually Exclusive Options**:
   - Cannot specify both `object_data_id` and `bacnet_objects` in the same request

## Error Handling

### Common Error Fields

- `fielddevice`: General field device errors
- `fielddevice.apparat_nr`: Apparat number validation errors
- `fielddevice.sps_controller_system_type_id`: SPS controller system type reference errors
- `fielddevice.apparat_id`: Apparat reference errors
- `fielddevice.system_part_id`: System part reference errors

### Error Messages

- `"apparat_nr is required"`: The apparat_nr field is missing or zero
- `"apparat_nr must be between 1 and 99"`: The apparat_nr is outside the valid range
- `"apparatnummer ist bereits vergeben"`: The apparat_nr is already taken for this scope
- `"one or more parent entities (SPS controller, apparat, system part) not found"`: Referenced entities don't exist

## Usage Examples

### Example 1: Create Multiple Devices

```bash
curl -X POST http://localhost:8080/api/v1/facility/field-devices/multi-create \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: your-token" \
  -d '{
    "field_devices": [
      {
        "bmk": "BMK1",
        "description": "Device 1",
        "apparat_nr": 1,
        "sps_controller_system_type_id": "uuid1",
        "system_part_id": "uuid2",
        "apparat_id": "uuid3"
      },
      {
        "bmk": "BMK2",
        "description": "Device 2",
        "apparat_nr": 2,
        "sps_controller_system_type_id": "uuid1",
        "system_part_id": "uuid2",
        "apparat_id": "uuid3"
      }
    ]
  }'
```

### Example 2: Linking to a Project

After successfully creating field devices, you can link them to a project using the existing project field device endpoint:

```bash
# For each successfully created field device
curl -X POST http://localhost:8080/api/v1/projects/{project_id}/field-devices \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: your-token" \
  -d '{
    "field_device_id": "uuid-of-created-device"
  }'
```

## Bruno API Tests

The implementation includes Bruno API test collections:

1. **`multi_create.bru`**: Basic multi-create test with 3 valid devices
2. **`multi_create_with_errors.bru`**: Demonstrates error handling with duplicate apparat_nr

Run these tests in Bruno to validate the functionality.

## Architecture

The implementation follows the hexagonal architecture pattern:

- **Domain Layer** (`internal/domain/facility/field_device.go`):
  - `FieldDeviceCreateItem`: Input model for multi-create
  - `FieldDeviceCreateResult`: Result for individual device creation
  - `FieldDeviceMultiCreateResult`: Aggregate result

- **Service Layer** (`internal/service/facility/field_device_service.go`):
  - `MultiCreate`: Business logic for batch creation with validation

- **Handler Layer** (`internal/handler/facility/field_device.go`):
  - `MultiCreateFieldDevices`: HTTP handler that processes requests

- **DTO Layer** (`internal/handler/dto/facility_field_device.go`):
  - Request and response DTOs for API communication

## Benefits

1. **Efficiency**: Create multiple devices in one request instead of multiple round-trips
2. **Resilience**: Partial failures don't prevent successful creations
3. **User-Friendly**: Clear error messages help users fix validation issues
4. **Clean Architecture**: Follows Go best practices and hexagonal architecture
5. **Type Safety**: Strong typing throughout the stack
