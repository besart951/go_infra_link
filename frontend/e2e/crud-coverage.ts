/**
 * Each documented facility/project mutation must have a browser acceptance
 * scenario, except technical bulk reads and validations, which use handler
 * contract coverage. This turns a newly documented write route into an E2E
 * review item instead of silently leaving it untested.
 */
export const facilityProjectMutationCoverage: Record<string, string> = {
  'POST /api/v1/facility/alarm-definitions': 'facility-standard-crud.spec.ts',
  'PUT /api/v1/facility/alarm-definitions/{id}': 'facility-standard-crud.spec.ts',
  'DELETE /api/v1/facility/alarm-definitions/{id}': 'facility-standard-crud.spec.ts',
  'POST /api/v1/facility/alarm-fields': 'facility-alarm-catalog.spec.ts',
  'PUT /api/v1/facility/alarm-fields/{id}': 'facility-alarm-catalog.spec.ts',
  'DELETE /api/v1/facility/alarm-fields/{id}': 'facility-alarm-catalog.spec.ts',
  'PUT /api/v1/facility/alarm-type-fields/{id}': 'facility-alarm-catalog.spec.ts',
  'DELETE /api/v1/facility/alarm-type-fields/{id}': 'facility-alarm-catalog.spec.ts',
  'POST /api/v1/facility/alarm-types': 'facility-alarm-catalog.spec.ts',
  'PUT /api/v1/facility/alarm-types/{id}': 'facility-alarm-catalog.spec.ts',
  'DELETE /api/v1/facility/alarm-types/{id}': 'facility-alarm-catalog.spec.ts',
  'POST /api/v1/facility/alarm-types/{id}/fields': 'facility-alarm-catalog.spec.ts',
  'POST /api/v1/facility/alarm-units': 'facility-alarm-catalog.spec.ts',
  'PUT /api/v1/facility/alarm-units/{id}': 'facility-alarm-catalog.spec.ts',
  'DELETE /api/v1/facility/alarm-units/{id}': 'facility-alarm-catalog.spec.ts',
  'POST /api/v1/facility/apparats': 'facility-reference-crud.spec.ts',
  'PUT /api/v1/facility/apparats/{id}': 'facility-reference-crud.spec.ts',
  'DELETE /api/v1/facility/apparats/{id}': 'facility-reference-crud.spec.ts',
  'POST /api/v1/facility/bacnet-objects': 'facility-field-devices.spec.ts',
  'PUT /api/v1/facility/bacnet-objects/{id}': 'facility-field-devices.spec.ts',
  'PUT /api/v1/facility/bacnet-objects/{id}/alarm-values': 'facility-field-devices.spec.ts',
  'POST /api/v1/facility/buildings': 'facility-hierarchy-crud.spec.ts',
  'PUT /api/v1/facility/buildings/{id}': 'facility-hierarchy-crud.spec.ts',
  'DELETE /api/v1/facility/buildings/{id}': 'facility-hierarchy-crud.spec.ts',
  'POST /api/v1/facility/control-cabinets': 'facility-hierarchy-crud.spec.ts',
  'PUT /api/v1/facility/control-cabinets/{id}': 'facility-hierarchy-crud.spec.ts',
  'DELETE /api/v1/facility/control-cabinets/{id}': 'facility-hierarchy-crud.spec.ts',
  'POST /api/v1/facility/control-cabinets/{id}/copy': 'facility-hierarchy-crud.spec.ts',
  'DELETE /api/v1/facility/field-devices/bulk-delete': 'facility-field-devices.spec.ts',
  'PATCH /api/v1/facility/field-devices/bulk-update': 'facility-field-devices.spec.ts',
  'POST /api/v1/facility/field-devices/multi-create': 'facility-field-devices.spec.ts',
  'PUT /api/v1/facility/field-devices/{id}': 'facility-field-devices.spec.ts',
  'DELETE /api/v1/facility/field-devices/{id}': 'facility-field-devices.spec.ts',
  'POST /api/v1/facility/field-devices/{id}/specification': 'facility-field-devices.spec.ts',
  'PUT /api/v1/facility/field-devices/{id}/specification': 'facility-field-devices.spec.ts',
  'POST /api/v1/facility/notification-classes': 'facility-standard-crud.spec.ts',
  'PUT /api/v1/facility/notification-classes/{id}': 'facility-standard-crud.spec.ts',
  'DELETE /api/v1/facility/notification-classes/{id}': 'facility-standard-crud.spec.ts',
  'POST /api/v1/facility/object-data': 'facility-standard-crud.spec.ts',
  'PUT /api/v1/facility/object-data/{id}': 'facility-standard-crud.spec.ts',
  'DELETE /api/v1/facility/object-data/{id}': 'facility-standard-crud.spec.ts',
  'DELETE /api/v1/facility/sps-controller-system-types/{id}': 'facility-hierarchy-crud.spec.ts',
  'PUT /api/v1/facility/sps-controller-system-types/{id}': 'facility-field-devices.spec.ts',
  'POST /api/v1/facility/sps-controller-system-types/{id}/copy': 'facility-hierarchy-crud.spec.ts',
  'POST /api/v1/facility/sps-controllers': 'facility-hierarchy-crud.spec.ts',
  'PUT /api/v1/facility/sps-controllers/{id}': 'facility-hierarchy-crud.spec.ts',
  'DELETE /api/v1/facility/sps-controllers/{id}': 'facility-hierarchy-crud.spec.ts',
  'POST /api/v1/facility/sps-controllers/{id}/copy': 'facility-hierarchy-crud.spec.ts',
  'POST /api/v1/facility/state-texts': 'facility-standard-crud.spec.ts',
  'PUT /api/v1/facility/state-texts/{id}': 'facility-standard-crud.spec.ts',
  'DELETE /api/v1/facility/state-texts/{id}': 'facility-standard-crud.spec.ts',
  'POST /api/v1/facility/system-parts': 'facility-reference-crud.spec.ts',
  'PUT /api/v1/facility/system-parts/{id}': 'facility-reference-crud.spec.ts',
  'DELETE /api/v1/facility/system-parts/{id}': 'facility-reference-crud.spec.ts',
  'POST /api/v1/facility/system-types': 'facility-standard-crud.spec.ts',
  'PUT /api/v1/facility/system-types/{id}': 'facility-standard-crud.spec.ts',
  'DELETE /api/v1/facility/system-types/{id}': 'facility-standard-crud.spec.ts',
  'POST /api/v1/phases': 'project-crud.spec.ts',
  'PUT /api/v1/phases/{id}': 'project-crud.spec.ts',
  'PATCH /api/v1/phases/{id}': 'project-crud.spec.ts',
  'DELETE /api/v1/phases/{id}': 'project-crud.spec.ts',
  'POST /api/v1/projects': 'project-crud.spec.ts',
  'PUT /api/v1/projects/{id}': 'project-crud.spec.ts',
  'PATCH /api/v1/projects/{id}': 'project-crud.spec.ts',
  'DELETE /api/v1/projects/{id}': 'project-crud.spec.ts',
  'POST /api/v1/projects/{id}/control-cabinets': 'project-facility-links.spec.ts',
  'POST /api/v1/projects/{id}/control-cabinets/{controlCabinetId}/copy':
    'project-facility-links.spec.ts',
  'PUT /api/v1/projects/{id}/control-cabinets/{linkId}':
    'backend/internal/handler/project/controlcabinet/handler_contract_test.go',
  'DELETE /api/v1/projects/{id}/control-cabinets/{linkId}': 'project-facility-links.spec.ts',
  'POST /api/v1/projects/{id}/field-devices':
    'backend/internal/handler/project/fielddevice/handler_contract_test.go',
  'POST /api/v1/projects/{id}/field-devices/multi-create': 'project-facility-links.spec.ts',
  'PUT /api/v1/projects/{id}/field-devices/{linkId}':
    'backend/internal/handler/project/fielddevice/handler_contract_test.go',
  'DELETE /api/v1/projects/{id}/field-devices/{linkId}': 'project-facility-links.spec.ts',
  'POST /api/v1/projects/{id}/object-data': 'project-crud.spec.ts',
  'DELETE /api/v1/projects/{id}/object-data/{objectDataId}': 'project-crud.spec.ts',
  'POST /api/v1/projects/{id}/sps-controller-system-types/{systemTypeId}/copy':
    'project-facility-links.spec.ts',
  'POST /api/v1/projects/{id}/sps-controllers': 'project-facility-links.spec.ts',
  'PUT /api/v1/projects/{id}/sps-controllers/{linkId}':
    'backend/internal/handler/project/spscontroller/handler_contract_test.go',
  'DELETE /api/v1/projects/{id}/sps-controllers/{linkId}': 'project-facility-links.spec.ts',
  'POST /api/v1/projects/{id}/sps-controllers/{spsControllerId}/copy':
    'project-facility-links.spec.ts',
  'POST /api/v1/projects/{id}/users': 'project-crud.spec.ts',
  'DELETE /api/v1/projects/{id}/users/{userId}': 'project-crud.spec.ts',
  'POST /api/v1/facility/apparats/bulk':
    'backend/internal/handler/facility/realtime_changes_test.go',
  'POST /api/v1/facility/buildings/bulk':
    'backend/internal/handler/facility/realtime_changes_test.go',
  'POST /api/v1/facility/control-cabinets/bulk':
    'backend/internal/handler/facility/realtime_changes_test.go',
  'POST /api/v1/facility/sps-controllers/bulk':
    'backend/internal/handler/facility/realtime_changes_test.go',
  'POST /api/v1/facility/buildings/validate':
    'backend/internal/handler/facility/realtime_changes_test.go',
  'POST /api/v1/facility/control-cabinets/validate':
    'backend/internal/handler/facility/realtime_changes_test.go',
  'POST /api/v1/facility/sps-controllers/validate':
    'backend/internal/handler/facility/realtime_changes_test.go'
};
