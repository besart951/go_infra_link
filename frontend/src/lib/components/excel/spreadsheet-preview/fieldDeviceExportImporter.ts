import type { SpreadsheetRow, WorksheetPreview } from '$lib/domain/excel/index.js';
import type {
  AlarmType,
  Apparat,
  BacnetObjectInput,
  Building,
  ControlCabinet,
  CreateControlCabinetRequest,
  CreateFieldDeviceRequest,
  CreateSPSControllerRequest,
  FieldDevice,
  NotificationClass,
  SPSController,
  SPSControllerSystemType,
  StateText,
  SystemPart,
  SystemType
} from '$lib/domain/facility/index.js';
import { toSpreadsheetColumnLabel } from '$lib/domain/excel/index.js';
import { t as translate } from '$lib/i18n/index.js';

export type ImportDiagnosticSeverity = 'error' | 'warning';
export type ImportEntityKind =
  | 'worksheet'
  | 'building'
  | 'control_cabinet'
  | 'sps_controller'
  | 'sps_controller_system_type'
  | 'field_device'
  | 'bacnet_object';

export interface ImportSourceCell {
  rowNumber: number;
  columnIndex: number;
  columnLabel: string;
  address: string;
}

export interface ImportDiagnostic {
  id: string;
  severity: ImportDiagnosticSeverity;
  message: string;
  entity: ImportEntityKind;
  cell?: ImportSourceCell;
  entityKey?: string;
}

export interface ImportCellMarker {
  severity: ImportDiagnosticSeverity;
  messages: string[];
}

export interface FieldDeviceImportLookups {
  buildings: Building[];
  controlCabinets: ControlCabinet[];
  spsControllers: SPSController[];
  spsControllerSystemTypes: SPSControllerSystemType[];
  fieldDevices: FieldDevice[];
  systemTypes: SystemType[];
  systemParts: SystemPart[];
  apparats: Apparat[];
  stateTexts: StateText[];
  notificationClasses: NotificationClass[];
  alarmTypes: AlarmType[];
}

export interface FieldDeviceImportLookupCriteria {
  iwsCode: string;
  buildingGroup?: number;
  controlCabinetNr: string;
  gaDevice: string;
  deviceName: string;
  systemTypeNumbers: number[];
  systemPartLabels: string[];
  apparatLabels: string[];
  notificationNumbers: number[];
  stateTextLabels: string[];
  alarmTypeLabels: string[];
}

export interface ParsedObjectName {
  iwsCode: string;
  buildingGroup?: number;
  spsSystemTypeNumber?: number;
  gaDevice: string;
  devicePart: string;
  address: string;
}

export interface PendingFieldDeviceRequest extends CreateFieldDeviceRequest {
  spsControllerSystemTypeKey: string;
}

export interface FieldDeviceImportBacnetObjectPlan {
  key: string;
  sourceRowNumber: number;
  sourceCell: ImportSourceCell;
  sourceCells: {
    objectName: ImportSourceCell;
    address: ImportSourceCell;
    textFix: ImportSourceCell;
  };
  address: string;
  textFix: string;
  request: BacnetObjectInput;
}

export interface FieldDeviceImportDevicePlan {
  key: string;
  sourceRowNumber: number;
  sourceCell: ImportSourceCell;
  objectName: string;
  displayName: string;
  spsControllerSystemTypeKey: string;
  spsSystemTypeNumber?: number;
  existingFieldDeviceId?: string;
  systemPartLabel: string;
  apparatLabel: string;
  apparatNr?: number;
  sourceCells: {
    objectName: ImportSourceCell;
    bmk: ImportSourceCell;
    description: ImportSourceCell;
    apparatNr: ImportSourceCell;
    systemPart: ImportSourceCell;
    apparat: ImportSourceCell;
    spsControllerSystemType: ImportSourceCell;
  };
  request: PendingFieldDeviceRequest;
  bacnetObjects: FieldDeviceImportBacnetObjectPlan[];
}

export interface FieldDeviceImportSystemTypePlan {
  key: string;
  number: number;
  systemTypeId: string;
  systemTypeName: string;
  sourceCell: ImportSourceCell;
  fieldDeviceCount: number;
  existingSpsControllerSystemTypeId?: string;
}

export interface FieldDeviceImportControllerPlan {
  iwsCode: string;
  buildingGroup?: number;
  buildingId?: string;
  buildingSourceCell?: ImportSourceCell;
  controlCabinetNr: string;
  controlCabinetSourceCell: ImportSourceCell;
  spsControllerSourceCell: ImportSourceCell;
  existingControlCabinet?: ControlCabinet;
  existingSpsController?: SPSController;
  controlCabinetRequest?: CreateControlCabinetRequest;
  spsControllerRequest?: Omit<CreateSPSControllerRequest, 'control_cabinet_id'>;
  systemTypes: FieldDeviceImportSystemTypePlan[];
  fieldDevices: FieldDeviceImportDevicePlan[];
}

export interface FieldDeviceImportPlan {
  worksheetName: string;
  headerRowNumber?: number;
  controller: FieldDeviceImportControllerPlan;
  diagnostics: ImportDiagnostic[];
  errorCount: number;
  warningCount: number;
  fieldDeviceCount: number;
  bacnetObjectCount: number;
  canImport: boolean;
}

const COLUMNS = {
  objectName: 0,
  description: 1,
  stateTextAggregate: 2,
  notification: 3,
  bmk: 4,
  gmsVisible: 5,
  systemPartName: 6,
  apparatName: 7,
  systemPartShortName: 8,
  apparatShortName: 9,
  textFix: 10,
  address: 26,
  alarmDefinition: 27,
  stateText: 28,
  hardwareStart: 29,
  remark: 33
} as const;

const SOFTWARE_COLUMNS = [
  { type: 'ai', column: 11 },
  { type: 'ao', column: 12 },
  { type: 'av', column: 13 },
  { type: 'bi', column: 14 },
  { type: 'bo', column: 15 },
  { type: 'bv', column: 16 },
  { type: 'mi', column: 17 },
  { type: 'mo', column: 18 },
  { type: 'mv', column: 19 },
  { type: 'ca', column: 20 },
  { type: 'ee', column: 21 },
  { type: 'lp', column: 22 },
  { type: 'nc', column: 23 },
  { type: 'sc', column: 24 },
  { type: 'tl', column: 25 }
] as const;

const HARDWARE_COLUMNS = [
  { type: 'do', column: 29 },
  { type: 'ao', column: 30 },
  { type: 'di', column: 31 },
  { type: 'ai', column: 32 }
] as const;

const REQUIRED_HEADING_KEYS = new Set(['bacnetobjectname', 'description', 'bmk']);
const SOFTWARE_TYPES: Set<string> = new Set(SOFTWARE_COLUMNS.map((item) => item.type));

export function normalizeLookupKey(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .normalize('NFKD')
    .replace(/[\u0300-\u036f]/g, '')
    .replace(/ae/g, 'a')
    .replace(/oe/g, 'o')
    .replace(/ue/g, 'u')
    .replace(/[^a-z0-9]+/g, '');
}

export function buildImportCellMarkers(
  diagnostics: readonly ImportDiagnostic[]
): Record<string, ImportCellMarker> {
  const markers: Record<string, ImportCellMarker> = {};

  for (const diagnostic of diagnostics) {
    if (!diagnostic.cell) continue;

    const key = sourceCellKey(diagnostic.cell.rowNumber, diagnostic.cell.columnIndex);
    const existing = markers[key];
    if (!existing) {
      markers[key] = {
        severity: diagnostic.severity,
        messages: [diagnostic.message]
      };
      continue;
    }

    if (diagnostic.severity === 'error') {
      existing.severity = 'error';
    }
    existing.messages.push(diagnostic.message);
  }

  return markers;
}

export function sourceCellKey(rowNumber: number, columnIndex: number): string {
  return `${rowNumber}:${columnIndex}`;
}

export function parseExportObjectName(value: string): ParsedObjectName | null {
  const parts = value
    .split('_')
    .map((part) => part.trim())
    .filter(Boolean);

  if (parts.length < 5) return null;

  const maybeAddress = parts[parts.length - 1] ?? '';
  const hasAddress = /^[a-z]{2}\d+$/i.test(maybeAddress);
  const requiredTailLength = hasAddress ? 5 : 4;
  if (parts.length <= requiredTailLength) return null;

  const tailStart = parts.length - requiredTailLength;
  const iwsCode = parts.slice(0, tailStart).join('_');
  const buildingGroup = parseInteger(parts[tailStart]);
  const spsSystemTypeNumber = parseInteger(parts[tailStart + 1]);
  const gaDevice = parts[tailStart + 2] ?? '';
  const devicePart = parts[tailStart + 3] ?? '';
  const address = hasAddress ? maybeAddress : '';

  return {
    iwsCode,
    buildingGroup,
    spsSystemTypeNumber,
    gaDevice,
    devicePart,
    address
  };
}

export function collectFieldDeviceImportLookupCriteria(
  worksheet: WorksheetPreview
): FieldDeviceImportLookupCriteria {
  const rows = worksheet.rows;
  const headerRowIndex = findHeaderRowIndex(rows);
  const firstDataRowIndex = headerRowIndex === -1 ? 0 : headerRowIndex + 1;
  const firstObjectNameCell = findFirstObjectNameCell(rows, firstDataRowIndex);
  const firstParsedName = firstObjectNameCell
    ? parseExportObjectName(cellString(rows[firstObjectNameCell.rowIndex], COLUMNS.objectName))
    : null;
  const headerValues = readControllerHeader(rows, headerRowIndex);

  const systemTypeNumbers = new Set<number>();
  const systemPartLabels = new Set<string>();
  const apparatLabels = new Set<string>();
  const notificationNumbers = new Set<number>();
  const stateTextLabels = new Set<string>();
  const alarmTypeLabels = new Set<string>();

  for (let rowIndex = firstDataRowIndex; rowIndex < rows.length; rowIndex += 1) {
    const row = rows[rowIndex];
    if (isBlankRow(row)) continue;

    if (isFieldDeviceRow(row)) {
      const parsedName = parseExportObjectName(cellString(row, COLUMNS.objectName));
      if (parsedName?.spsSystemTypeNumber) {
        systemTypeNumbers.add(parsedName.spsSystemTypeNumber);
      }

      addMeaningfulLabels(
        systemPartLabels,
        cellString(row, COLUMNS.systemPartShortName),
        cellString(row, COLUMNS.systemPartName)
      );
      addMeaningfulLabels(
        apparatLabels,
        cellString(row, COLUMNS.apparatShortName),
        cellString(row, COLUMNS.apparatName)
      );
      continue;
    }

    if (isBacnetObjectRow(row)) {
      const parsedName = parseExportObjectName(cellString(row, COLUMNS.objectName));
      if (parsedName?.spsSystemTypeNumber) {
        systemTypeNumbers.add(parsedName.spsSystemTypeNumber);
      }

      const notificationNumber = parseInteger(cellString(row, COLUMNS.notification));
      if (notificationNumber !== undefined) {
        notificationNumbers.add(notificationNumber);
      }

      addMeaningfulLabels(
        stateTextLabels,
        cellString(row, COLUMNS.stateText),
        cellString(row, COLUMNS.stateTextAggregate)
      );
      addMeaningfulLabels(alarmTypeLabels, cellString(row, COLUMNS.alarmDefinition));
    }
  }

  return {
    iwsCode: firstParsedName?.iwsCode ?? parseIwsFromDeviceName(headerValues.deviceName.value),
    buildingGroup:
      firstParsedName?.buildingGroup ??
      parseBuildingGroupFromDeviceName(headerValues.deviceName.value),
    controlCabinetNr: headerValues.controlCabinetNr.value.trim(),
    gaDevice: normalizeGADevice(headerValues.gaDevice.value || firstParsedName?.gaDevice || ''),
    deviceName: headerValues.deviceName.value.trim(),
    systemTypeNumbers: sortedNumbers(systemTypeNumbers),
    systemPartLabels: sortedLabels(systemPartLabels),
    apparatLabels: sortedLabels(apparatLabels),
    notificationNumbers: sortedNumbers(notificationNumbers),
    stateTextLabels: sortedLabels(stateTextLabels),
    alarmTypeLabels: sortedLabels(alarmTypeLabels)
  };
}

export function transformWorksheetToFieldDeviceImport(
  worksheet: WorksheetPreview,
  lookups: FieldDeviceImportLookups
): FieldDeviceImportPlan {
  const diagnostics: ImportDiagnostic[] = [];
  const addDiagnostic = createDiagnosticCollector(diagnostics);
  const rows = worksheet.rows;
  const headerRowIndex = findHeaderRowIndex(rows);

  if (headerRowIndex === -1) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.header_missing'),
      'worksheet',
      sourceCell(0, 0)
    );
  }

  const firstDataRowIndex = headerRowIndex === -1 ? 0 : headerRowIndex + 1;
  const firstObjectNameCell = findFirstObjectNameCell(rows, firstDataRowIndex);
  const firstParsedName = firstObjectNameCell
    ? parseExportObjectName(cellString(rows[firstObjectNameCell.rowIndex], COLUMNS.objectName))
    : null;

  if (!firstObjectNameCell) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.no_field_device_rows'),
      'worksheet',
      sourceCell(Math.max(0, firstDataRowIndex), COLUMNS.objectName)
    );
  } else if (!firstParsedName) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.object_name_format'),
      'worksheet',
      sourceCell(firstObjectNameCell.rowIndex, COLUMNS.objectName)
    );
  }

  const headerValues = readControllerHeader(rows, headerRowIndex);
  const iwsCode = firstParsedName?.iwsCode ?? parseIwsFromDeviceName(headerValues.deviceName.value);
  const buildingGroup =
    firstParsedName?.buildingGroup ??
    parseBuildingGroupFromDeviceName(headerValues.deviceName.value);
  const gaDevice = normalizeGADevice(
    headerValues.gaDevice.value || firstParsedName?.gaDevice || ''
  );
  const controlCabinetNr = headerValues.controlCabinetNr.value.trim();
  const deviceName = headerValues.deviceName.value.trim();

  const lookupIndex = createLookupIndex(lookups);
  const building = resolveBuilding(lookupIndex, iwsCode, buildingGroup);

  if (!iwsCode) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.iws_missing'),
      'building',
      firstObjectNameCell
        ? sourceCell(firstObjectNameCell.rowIndex, COLUMNS.objectName)
        : headerValues.deviceName.cell
    );
  }

  if (!buildingGroup) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.building_group_missing'),
      'building',
      firstObjectNameCell
        ? sourceCell(firstObjectNameCell.rowIndex, COLUMNS.objectName)
        : headerValues.deviceName.cell
    );
  }

  if (iwsCode && buildingGroup && !building) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.building_not_found', {
        iws: iwsCode,
        group: buildingGroup
      }),
      'building',
      firstObjectNameCell
        ? sourceCell(firstObjectNameCell.rowIndex, COLUMNS.objectName)
        : headerValues.deviceName.cell
    );
  }

  if (!controlCabinetNr) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.control_cabinet_required'),
      'control_cabinet',
      headerValues.controlCabinetNr.cell,
      'control-cabinet'
    );
  }

  let existingControlCabinet: ControlCabinet | undefined;
  if (building && controlCabinetNr) {
    const cabinetKey = scopedKey(building.id, controlCabinetNr);
    existingControlCabinet = lookupIndex.controlCabinetByBuildingAndNr.get(cabinetKey);
    if (existingControlCabinet) {
      addDiagnostic(
        'warning',
        translate('field_device.importer.validation.control_cabinet_exists', {
          value: controlCabinetNr
        }),
        'control_cabinet',
        headerValues.controlCabinetNr.cell,
        'control-cabinet'
      );
    }
  }

  if (!gaDevice) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.ga_device_required'),
      'sps_controller',
      headerValues.gaDevice.cell,
      'sps-controller'
    );
  } else if (!/^[A-Z]{3}$/.test(gaDevice)) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.ga_device_format'),
      'sps_controller',
      headerValues.gaDevice.cell,
      'sps-controller'
    );
  }

  if (!deviceName) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.device_name_required'),
      'sps_controller',
      headerValues.deviceName.cell,
      'sps-controller'
    );
  } else if (deviceName.length > 100) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.device_name_max'),
      'sps_controller',
      headerValues.deviceName.cell,
      'sps-controller'
    );
  }

  const existingSpsController = existingControlCabinet
    ? resolveExistingSpsController(lookupIndex, existingControlCabinet.id, deviceName, gaDevice)
    : undefined;
  if (existingSpsController) {
    addDiagnostic(
      'warning',
      translate('field_device.importer.validation.sps_controller_exists', {
        value: existingSpsController.device_name || existingSpsController.ga_device || deviceName
      }),
      'sps_controller',
      headerValues.deviceName.cell,
      'sps-controller'
    );
  }

  validateOptionalNetworkField(
    headerValues.ipAddress.value,
    translate('field_device.importer.validation.ip_invalid'),
    'ip',
    headerValues.ipAddress.cell,
    addDiagnostic
  );
  validateOptionalNetworkField(
    headerValues.gateway.value,
    translate('field_device.importer.validation.gateway_invalid'),
    'ip',
    headerValues.gateway.cell,
    addDiagnostic
  );
  validateOptionalNetworkField(
    headerValues.subnet.value,
    translate('field_device.importer.validation.subnet_invalid'),
    'subnet',
    headerValues.subnet.cell,
    addDiagnostic
  );
  validateVlan(headerValues.vlan.value, headerValues.vlan.cell, addDiagnostic);

  const fieldDevices = parseFieldDeviceRows(
    rows,
    firstDataRowIndex,
    {
      iwsCode,
      buildingGroup,
      gaDevice
    },
    lookupIndex,
    addDiagnostic
  );

  const systemTypes = buildSystemTypePlans(fieldDevices, lookupIndex, addDiagnostic);
  markExistingSystemTypePlans(systemTypes, existingSpsController?.id, lookupIndex, addDiagnostic);
  validateExistingFieldDevices(fieldDevices, deviceName, lookupIndex, addDiagnostic);
  validateDuplicateFieldDevices(fieldDevices, addDiagnostic);

  const controlCabinetRequest =
    building && controlCabinetNr
      ? {
          building_id: building.id,
          control_cabinet_nr: controlCabinetNr
        }
      : undefined;

  const spsControllerRequest =
    deviceName && gaDevice
      ? {
          ga_device: gaDevice,
          device_name: deviceName,
          device_description: emptyToUndefined(headerValues.deviceDescription.value),
          device_location: emptyToUndefined(headerValues.deviceLocation.value),
          ip_address: emptyToUndefined(headerValues.ipAddress.value),
          subnet: emptyToUndefined(headerValues.subnet.value),
          gateway: emptyToUndefined(headerValues.gateway.value),
          vlan: emptyToUndefined(headerValues.vlan.value),
          system_types: systemTypes.map((item) => ({
            system_type_id: item.systemTypeId,
            number: item.number
          }))
        }
      : undefined;

  const errorCount = diagnostics.filter((item) => item.severity === 'error').length;
  const warningCount = diagnostics.filter((item) => item.severity === 'warning').length;
  const bacnetObjectCount = fieldDevices.reduce((sum, item) => sum + item.bacnetObjects.length, 0);

  return {
    worksheetName: worksheet.name,
    headerRowNumber: headerRowIndex === -1 ? undefined : headerRowIndex + 1,
    controller: {
      iwsCode,
      buildingGroup,
      buildingId: building?.id,
      buildingSourceCell: firstObjectNameCell
        ? sourceCell(firstObjectNameCell.rowIndex, COLUMNS.objectName)
        : headerValues.deviceName.cell,
      controlCabinetNr,
      controlCabinetSourceCell: headerValues.controlCabinetNr.cell,
      spsControllerSourceCell: headerValues.deviceName.cell,
      existingControlCabinet,
      existingSpsController,
      controlCabinetRequest,
      spsControllerRequest,
      systemTypes,
      fieldDevices
    },
    diagnostics,
    errorCount,
    warningCount,
    fieldDeviceCount: fieldDevices.length,
    bacnetObjectCount,
    canImport:
      errorCount === 0 &&
      Boolean(controlCabinetRequest) &&
      Boolean(spsControllerRequest) &&
      fieldDevices.length > 0
  };
}

function parseFieldDeviceRows(
  rows: SpreadsheetRow[],
  startRowIndex: number,
  expectedHeader: { iwsCode: string; buildingGroup?: number; gaDevice: string },
  lookupIndex: LookupIndex,
  addDiagnostic: AddDiagnostic
): FieldDeviceImportDevicePlan[] {
  const fieldDevices: FieldDeviceImportDevicePlan[] = [];
  let current: FieldDeviceImportDevicePlan | null = null;

  for (let rowIndex = startRowIndex; rowIndex < rows.length; rowIndex += 1) {
    const row = rows[rowIndex];
    if (isBlankRow(row)) continue;

    if (isFieldDeviceRow(row)) {
      current = parseFieldDeviceRow(row, rowIndex, expectedHeader, lookupIndex, addDiagnostic);
      fieldDevices.push(current);
      continue;
    }

    if (isBacnetObjectRow(row)) {
      if (!current) {
        addDiagnostic(
          'error',
          translate('field_device.importer.validation.bacnet_without_field_device'),
          'bacnet_object',
          sourceCell(rowIndex, COLUMNS.objectName)
        );
        continue;
      }

      const bacnetObject = parseBacnetObjectRow(row, rowIndex, lookupIndex, addDiagnostic);
      if (bacnetObject) {
        current.bacnetObjects.push(bacnetObject);
        current.request.bacnet_objects = current.bacnetObjects.map((item) => item.request);
      }
      continue;
    }

    addDiagnostic(
      'warning',
      translate('field_device.importer.validation.unclassified_row'),
      'worksheet',
      sourceCell(rowIndex, COLUMNS.objectName)
    );
  }

  return fieldDevices;
}

function parseFieldDeviceRow(
  row: SpreadsheetRow,
  rowIndex: number,
  expectedHeader: { iwsCode: string; buildingGroup?: number; gaDevice: string },
  lookupIndex: LookupIndex,
  addDiagnostic: AddDiagnostic
): FieldDeviceImportDevicePlan {
  const objectName = cellString(row, COLUMNS.objectName);
  const parsedName = parseExportObjectName(objectName);
  const source = sourceCell(rowIndex, COLUMNS.objectName);
  const key = `field-device:${rowIndex + 1}`;

  if (!parsedName) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.field_device_object_name_format'),
      'field_device',
      source,
      key
    );
  } else {
    if (
      expectedHeader.iwsCode &&
      parsedName.iwsCode &&
      parsedName.iwsCode !== expectedHeader.iwsCode
    ) {
      addDiagnostic(
        'error',
        translate('field_device.importer.validation.iws_mismatch', {
          value: parsedName.iwsCode,
          expected: expectedHeader.iwsCode
        }),
        'field_device',
        source,
        key
      );
    }
    if (
      expectedHeader.buildingGroup &&
      parsedName.buildingGroup &&
      parsedName.buildingGroup !== expectedHeader.buildingGroup
    ) {
      addDiagnostic(
        'error',
        translate('field_device.importer.validation.building_group_mismatch', {
          value: parsedName.buildingGroup,
          expected: expectedHeader.buildingGroup
        }),
        'field_device',
        source,
        key
      );
    }
    if (
      expectedHeader.gaDevice &&
      parsedName.gaDevice &&
      parsedName.gaDevice !== expectedHeader.gaDevice
    ) {
      addDiagnostic(
        'error',
        translate('field_device.importer.validation.ga_device_mismatch', {
          value: parsedName.gaDevice,
          expected: expectedHeader.gaDevice
        }),
        'field_device',
        source,
        key
      );
    }
  }

  const spsSystemTypeNumber = parsedName?.spsSystemTypeNumber;
  if (!spsSystemTypeNumber) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.sps_system_type_number_missing'),
      'sps_controller_system_type',
      source,
      key
    );
  }

  const devicePartNr = parseDevicePartNumber(parsedName?.devicePart ?? '');
  if (!devicePartNr || devicePartNr < 1 || devicePartNr > 99) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.apparat_nr_missing_or_range'),
      'field_device',
      source,
      key
    );
  }

  const systemPartShort = cellString(row, COLUMNS.systemPartShortName);
  const systemPartName = cellString(row, COLUMNS.systemPartName);
  const systemPart = resolveByShortOrName(
    lookupIndex.systemPartByLookup,
    systemPartShort,
    systemPartName
  );
  if (!systemPart) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.system_part_not_found', {
        value:
          systemPartShort ||
          systemPartName ||
          translate('field_device.importer.validation.empty_value')
      }),
      'field_device',
      sourceCell(rowIndex, systemPartShort ? COLUMNS.systemPartShortName : COLUMNS.systemPartName),
      key
    );
  }

  const apparatShort = cellString(row, COLUMNS.apparatShortName);
  const apparatName = cellString(row, COLUMNS.apparatName);
  const apparat = resolveByShortOrName(lookupIndex.apparatByLookup, apparatShort, apparatName);
  if (!apparat) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.apparat_not_found', {
        value:
          apparatShort || apparatName || translate('field_device.importer.validation.empty_value')
      }),
      'field_device',
      sourceCell(rowIndex, apparatShort ? COLUMNS.apparatShortName : COLUMNS.apparatName),
      key
    );
  }

  if (systemPart && apparat && !apparat.system_parts?.some((item) => item.id === systemPart.id)) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.apparat_system_part_unlinked', {
        apparat: apparat.short_name,
        systemPart: systemPart.short_name
      }),
      'field_device',
      sourceCell(rowIndex, COLUMNS.apparatShortName),
      key
    );
  }

  const bmk = cellString(row, COLUMNS.bmk);
  if (bmk.length > 10) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.bmk_max'),
      'field_device',
      sourceCell(rowIndex, COLUMNS.bmk),
      key
    );
  }

  const description = cellString(row, COLUMNS.remark);
  if (description.length > 250) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.remark_max'),
      'field_device',
      sourceCell(rowIndex, COLUMNS.remark),
      key
    );
  }

  const spsControllerSystemTypeKey = systemTypeKey(spsSystemTypeNumber ?? 0);
  const request: PendingFieldDeviceRequest = {
    bmk: emptyToUndefined(bmk),
    description: emptyToUndefined(description),
    apparat_nr: devicePartNr ?? 0,
    sps_controller_system_type_id: '',
    spsControllerSystemTypeKey,
    system_part_id: systemPart?.id ?? '',
    apparat_id: apparat?.id ?? '',
    bacnet_objects: []
  };

  return {
    key,
    sourceRowNumber: rowIndex + 1,
    sourceCell: source,
    objectName,
    displayName:
      objectName ||
      translate('field_device.importer.validation.row_label', {
        row: rowIndex + 1
      }),
    spsControllerSystemTypeKey,
    spsSystemTypeNumber,
    systemPartLabel: systemPart?.short_name ?? systemPartShort ?? systemPartName,
    apparatLabel: apparat?.short_name ?? apparatShort ?? apparatName,
    apparatNr: devicePartNr,
    sourceCells: {
      objectName: source,
      bmk: sourceCell(rowIndex, COLUMNS.bmk),
      description: sourceCell(rowIndex, COLUMNS.remark),
      apparatNr: source,
      systemPart: sourceCell(
        rowIndex,
        systemPartShort ? COLUMNS.systemPartShortName : COLUMNS.systemPartName
      ),
      apparat: sourceCell(rowIndex, apparatShort ? COLUMNS.apparatShortName : COLUMNS.apparatName),
      spsControllerSystemType: source
    },
    request,
    bacnetObjects: []
  };
}

function parseBacnetObjectRow(
  row: SpreadsheetRow,
  rowIndex: number,
  lookupIndex: LookupIndex,
  addDiagnostic: AddDiagnostic
): FieldDeviceImportBacnetObjectPlan | null {
  const objectName = cellString(row, COLUMNS.objectName);
  const parsedName = parseExportObjectName(objectName);
  const source = sourceCell(rowIndex, COLUMNS.objectName);
  const key = `bacnet-object:${rowIndex + 1}`;

  if (!parsedName) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.bacnet_object_name_format'),
      'bacnet_object',
      source,
      key
    );
  }

  const addressCell = sourceCell(rowIndex, COLUMNS.address);
  const address = cellString(row, COLUMNS.address) || parsedName?.address || '';
  const software = parseSoftware(address);
  if (!software) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.bacnet_address_format'),
      'bacnet_object',
      addressCell,
      key
    );
  }

  if (software && !SOFTWARE_TYPES.has(software.type)) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.software_type_unsupported', {
        type: software.type.toUpperCase()
      }),
      'bacnet_object',
      addressCell,
      key
    );
  }

  const textFix = cellString(row, COLUMNS.textFix);
  if (!textFix) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.text_fix_required'),
      'bacnet_object',
      sourceCell(rowIndex, COLUMNS.textFix),
      key
    );
  } else if (textFix.length > 250) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.text_fix_max'),
      'bacnet_object',
      sourceCell(rowIndex, COLUMNS.textFix),
      key
    );
  }

  const hardware = parseHardware(row, rowIndex, addDiagnostic, key);
  const notificationClassId = resolveNotificationClass(
    row,
    rowIndex,
    lookupIndex,
    addDiagnostic,
    key
  );
  const stateTextId = resolveStateText(row, rowIndex, lookupIndex, addDiagnostic, key);
  const alarmTypeId = resolveAlarmType(row, rowIndex, lookupIndex, addDiagnostic, key);

  const request: BacnetObjectInput = {
    text_fix: textFix,
    description: emptyToUndefined(cellString(row, COLUMNS.description)),
    gms_visible: parseBoolean(cellValue(row, COLUMNS.gmsVisible), true),
    optional: false,
    software_type: software?.type ?? '',
    software_number: software?.number ?? 0,
    hardware_type: hardware.type,
    hardware_quantity: hardware.quantity,
    notification_class_id: notificationClassId,
    state_text_id: stateTextId,
    alarm_type_id: alarmTypeId
  };

  return {
    key,
    sourceRowNumber: rowIndex + 1,
    sourceCell: source,
    sourceCells: {
      objectName: source,
      address: addressCell,
      textFix: sourceCell(rowIndex, COLUMNS.textFix)
    },
    address,
    textFix,
    request
  };
}

function buildSystemTypePlans(
  fieldDevices: FieldDeviceImportDevicePlan[],
  lookupIndex: LookupIndex,
  addDiagnostic: AddDiagnostic
): FieldDeviceImportSystemTypePlan[] {
  const byKey = new Map<string, FieldDeviceImportSystemTypePlan>();

  for (const fieldDevice of fieldDevices) {
    const number = fieldDevice.spsSystemTypeNumber;
    if (!number) continue;

    const key = systemTypeKey(number);
    const existing = byKey.get(key);
    if (existing) {
      existing.fieldDeviceCount += 1;
      continue;
    }

    const systemType = lookupIndex.systemTypes.find(
      (item) => number >= item.number_min && number <= item.number_max
    );

    if (!systemType) {
      addDiagnostic(
        'error',
        translate('field_device.importer.validation.system_type_definition_missing', {
          number
        }),
        'sps_controller_system_type',
        fieldDevice.sourceCell,
        fieldDevice.key
      );
      continue;
    }

    byKey.set(key, {
      key,
      number,
      systemTypeId: systemType.id,
      systemTypeName: systemType.name,
      sourceCell: fieldDevice.sourceCell,
      fieldDeviceCount: 1
    });
  }

  return Array.from(byKey.values()).sort((a, b) => a.number - b.number);
}

function markExistingSystemTypePlans(
  systemTypes: FieldDeviceImportSystemTypePlan[],
  spsControllerId: string | undefined,
  lookupIndex: LookupIndex,
  addDiagnostic: AddDiagnostic
): void {
  if (!spsControllerId) return;

  for (const systemType of systemTypes) {
    const existing = lookupIndex.spsControllerSystemTypeByControllerTypeAndNumber.get(
      spsControllerSystemTypeKey(spsControllerId, systemType.systemTypeId, systemType.number)
    );
    if (!existing) continue;

    systemType.existingSpsControllerSystemTypeId = existing.id;
    addDiagnostic(
      'warning',
      translate('field_device.importer.validation.sps_controller_system_type_exists', {
        number: systemType.number,
        name: systemType.systemTypeName
      }),
      'sps_controller_system_type',
      systemType.sourceCell,
      systemType.key
    );
  }
}

function validateExistingFieldDevices(
  fieldDevices: FieldDeviceImportDevicePlan[],
  controllerDeviceName: string,
  lookupIndex: LookupIndex,
  addDiagnostic: AddDiagnostic
): void {
  for (const fieldDevice of fieldDevices) {
    const key = fieldDeviceBusinessKey(
      controllerDeviceName,
      fieldDevice.spsSystemTypeNumber,
      fieldDevice.request.system_part_id,
      fieldDevice.request.apparat_id,
      fieldDevice.apparatNr
    );
    if (!key) continue;

    const existing = lookupIndex.existingFieldDeviceByBusinessKey.get(key);
    if (!existing) continue;

    fieldDevice.existingFieldDeviceId = existing.id;
    const systemTypeNumber = fieldDevice.spsSystemTypeNumber
      ? translate('field_device.importer.validation.sps_system_type_with_number', {
          number: fieldDevice.spsSystemTypeNumber
        })
      : translate('field_device.importer.validation.sps_system_type');
    addDiagnostic(
      'warning',
      translate('field_device.importer.validation.field_device_exists', {
        systemType: systemTypeNumber,
        systemPart: fieldDevice.systemPartLabel,
        apparat: fieldDevice.apparatLabel,
        apparatNr: fieldDevice.apparatNr
      }),
      'field_device',
      fieldDevice.sourceCells.apparatNr,
      fieldDevice.key
    );
  }
}

function validateDuplicateFieldDevices(
  fieldDevices: FieldDeviceImportDevicePlan[],
  addDiagnostic: AddDiagnostic
): void {
  const seen = new Map<string, FieldDeviceImportDevicePlan>();

  for (const fieldDevice of fieldDevices) {
    const systemPartId = fieldDevice.request.system_part_id;
    const apparatId = fieldDevice.request.apparat_id;
    const apparatNr = fieldDevice.apparatNr;
    if (!systemPartId || !apparatId || !apparatNr || !fieldDevice.spsSystemTypeNumber) continue;

    const key = [fieldDevice.spsSystemTypeNumber, systemPartId, apparatId, apparatNr].join('|');
    const existing = seen.get(key);
    if (!existing) {
      seen.set(key, fieldDevice);
      continue;
    }

    addDiagnostic(
      'error',
      translate('field_device.importer.validation.duplicate_apparat_nr', {
        row: existing.sourceRowNumber
      }),
      'field_device',
      fieldDevice.sourceCell,
      fieldDevice.key
    );
  }
}

function resolveNotificationClass(
  row: SpreadsheetRow,
  rowIndex: number,
  lookupIndex: LookupIndex,
  addDiagnostic: AddDiagnostic,
  key: string
): string | undefined {
  const value = cellString(row, COLUMNS.notification);
  if (!isMeaningfulLabel(value)) return undefined;

  const notificationNumber = parseInteger(value);
  if (notificationNumber === undefined) {
    addDiagnostic(
      'warning',
      translate('field_device.importer.validation.notification_not_numeric', {
        value
      }),
      'bacnet_object',
      sourceCell(rowIndex, COLUMNS.notification),
      key
    );
    return undefined;
  }

  const notification = lookupIndex.notificationClassByNc.get(notificationNumber);
  if (!notification) {
    addDiagnostic(
      'warning',
      translate('field_device.importer.validation.notification_class_not_found', {
        value: notificationNumber
      }),
      'bacnet_object',
      sourceCell(rowIndex, COLUMNS.notification),
      key
    );
    return undefined;
  }

  return notification.id;
}

function resolveStateText(
  row: SpreadsheetRow,
  rowIndex: number,
  lookupIndex: LookupIndex,
  addDiagnostic: AddDiagnostic,
  key: string
): string | undefined {
  const stateText = cellString(row, COLUMNS.stateText);
  const aggregate = cellString(row, COLUMNS.stateTextAggregate);
  const label = stateText || aggregate;
  if (!isMeaningfulLabel(label)) return undefined;

  const byNumber = parseInteger(label);
  if (byNumber !== undefined) {
    const match = lookupIndex.stateTextByRefNumber.get(byNumber);
    if (match) return match.id;
  }

  const normalized = normalizeLookupKey(label);
  const match = lookupIndex.stateTextByLookup.get(normalized);
  if (!match) {
    addDiagnostic(
      'warning',
      translate('field_device.importer.validation.state_text_not_found', {
        value: label
      }),
      'bacnet_object',
      sourceCell(rowIndex, stateText ? COLUMNS.stateText : COLUMNS.stateTextAggregate),
      key
    );
    return undefined;
  }

  return match.id;
}

function resolveAlarmType(
  row: SpreadsheetRow,
  rowIndex: number,
  lookupIndex: LookupIndex,
  addDiagnostic: AddDiagnostic,
  key: string
): string | undefined {
  const label = cellString(row, COLUMNS.alarmDefinition);
  if (!isMeaningfulLabel(label)) return undefined;

  const match = lookupIndex.alarmTypeByLookup.get(normalizeLookupKey(label));
  if (!match) {
    addDiagnostic(
      'warning',
      translate('field_device.importer.validation.alarm_type_not_found', {
        value: label
      }),
      'bacnet_object',
      sourceCell(rowIndex, COLUMNS.alarmDefinition),
      key
    );
    return undefined;
  }

  return match.id;
}

function parseHardware(
  row: SpreadsheetRow,
  rowIndex: number,
  addDiagnostic: AddDiagnostic,
  key: string
): { type: string; quantity: number } {
  let hardware = { type: '', quantity: 0 };

  for (const item of HARDWARE_COLUMNS) {
    const raw = cellString(row, item.column);
    if (!raw) continue;

    const numeric = Number(raw);
    if (!Number.isFinite(numeric) || numeric < 0 || numeric > 255) {
      addDiagnostic(
        'error',
        translate('field_device.importer.validation.hardware_quantity_range'),
        'bacnet_object',
        sourceCell(rowIndex, item.column),
        key
      );
      continue;
    }

    if (numeric === 0) continue;

    if (!hardware.type) {
      hardware = { type: item.type, quantity: Math.trunc(numeric) };
      continue;
    }

    addDiagnostic(
      'warning',
      translate('field_device.importer.validation.multiple_hardware_columns'),
      'bacnet_object',
      sourceCell(rowIndex, item.column),
      key
    );
  }

  return hardware;
}

function parseSoftware(address: string): { type: string; number: number } | null {
  const match = address.trim().match(/^([a-z]{2})(\d+)$/i);
  if (!match) return null;

  const number = Number.parseInt(match[2], 10);
  if (!Number.isFinite(number) || number < 0 || number > 65535) return null;

  return {
    type: match[1].toLowerCase(),
    number
  };
}

function isFieldDeviceRow(row: SpreadsheetRow): boolean {
  return (
    cellString(row, COLUMNS.objectName).length > 0 &&
    cellString(row, COLUMNS.textFix).length === 0 &&
    cellString(row, COLUMNS.address).length === 0 &&
    (cellString(row, COLUMNS.systemPartName).length > 0 ||
      cellString(row, COLUMNS.apparatName).length > 0 ||
      cellString(row, COLUMNS.systemPartShortName).length > 0 ||
      cellString(row, COLUMNS.apparatShortName).length > 0)
  );
}

function isBacnetObjectRow(row: SpreadsheetRow): boolean {
  return (
    cellString(row, COLUMNS.objectName).length > 0 &&
    (cellString(row, COLUMNS.textFix).length > 0 ||
      cellString(row, COLUMNS.address).length > 0 ||
      SOFTWARE_COLUMNS.some((item) => numericCellValue(row, item.column) > 0))
  );
}

function isBlankRow(row: SpreadsheetRow): boolean {
  return row.every((cell) => String(cell ?? '').trim().length === 0);
}

function findHeaderRowIndex(rows: SpreadsheetRow[]): number {
  return rows.findIndex((row) => {
    const keys = [0, 1, 4].map((column) => normalizeLookupKey(cellString(row, column)));
    return keys.every((key) => REQUIRED_HEADING_KEYS.has(key));
  });
}

function findFirstObjectNameCell(
  rows: SpreadsheetRow[],
  startRowIndex: number
): { rowIndex: number; columnIndex: number } | null {
  for (let rowIndex = startRowIndex; rowIndex < rows.length; rowIndex += 1) {
    if (cellString(rows[rowIndex], COLUMNS.objectName).length > 0) {
      return { rowIndex, columnIndex: COLUMNS.objectName };
    }
  }

  return null;
}

function readControllerHeader(rows: SpreadsheetRow[], headerRowIndex: number) {
  const maxRowIndex = headerRowIndex === -1 ? Math.min(rows.length, 20) : headerRowIndex;

  return {
    gaDevice: findHeaderValue(rows, maxRowIndex, ['gagerat', 'projektcontroller'], 1),
    controlCabinetNr: findHeaderValue(rows, maxRowIndex, ['schaltschranknr'], 2),
    deviceName: findHeaderValue(rows, maxRowIndex, ['devicename'], 3),
    deviceDescription: findHeaderValue(rows, maxRowIndex, ['devicedescription'], 5),
    deviceLocation: findHeaderValue(rows, maxRowIndex, ['devicelocation'], 6),
    ipAddress: findHeaderValue(rows, maxRowIndex, ['ipadresse'], 7),
    subnet: findHeaderValue(rows, maxRowIndex, ['subnetz'], 8),
    gateway: findHeaderValue(rows, maxRowIndex, ['gateway'], 9),
    vlan: findHeaderValue(rows, maxRowIndex, ['vlan'], 10)
  };
}

function findHeaderValue(
  rows: SpreadsheetRow[],
  maxRowIndex: number,
  labelKeys: string[],
  fallbackRowIndex: number
): { value: string; cell: ImportSourceCell } {
  for (let rowIndex = 0; rowIndex < maxRowIndex; rowIndex += 1) {
    const key = normalizeLookupKey(cellString(rows[rowIndex], 0));
    if (labelKeys.includes(key)) {
      return {
        value: cellString(rows[rowIndex], 1),
        cell: sourceCell(rowIndex, 1)
      };
    }
  }

  return {
    value: cellString(rows[fallbackRowIndex], 1),
    cell: sourceCell(fallbackRowIndex, 1)
  };
}

interface LookupIndex {
  buildingsByScope: Map<string, Building>;
  controlCabinetByBuildingAndNr: Map<string, ControlCabinet>;
  spsControllerByCabinetAndDeviceName: Map<string, SPSController>;
  spsControllerByCabinetAndGADevice: Map<string, SPSController>;
  spsControllerSystemTypeByControllerTypeAndNumber: Map<string, SPSControllerSystemType>;
  existingFieldDeviceByBusinessKey: Map<string, FieldDevice>;
  systemTypes: SystemType[];
  systemPartByLookup: Map<string, SystemPart>;
  apparatByLookup: Map<string, Apparat>;
  notificationClassByNc: Map<number, NotificationClass>;
  stateTextByRefNumber: Map<number, StateText>;
  stateTextByLookup: Map<string, StateText>;
  alarmTypeByLookup: Map<string, AlarmType>;
}

function createLookupIndex(lookups: FieldDeviceImportLookups): LookupIndex {
  const buildingsByScope = new Map<string, Building>();
  for (const building of lookups.buildings) {
    buildingsByScope.set(buildingKey(building.iws_code, building.building_group), building);
  }

  const controlCabinetByBuildingAndNr = new Map<string, ControlCabinet>();
  for (const cabinet of lookups.controlCabinets) {
    controlCabinetByBuildingAndNr.set(
      scopedKey(cabinet.building_id, cabinet.control_cabinet_nr),
      cabinet
    );
  }

  const spsControllerByCabinetAndDeviceName = new Map<string, SPSController>();
  const spsControllerByCabinetAndGADevice = new Map<string, SPSController>();
  for (const controller of lookups.spsControllers) {
    spsControllerByCabinetAndDeviceName.set(
      scopedKey(controller.control_cabinet_id, controller.device_name),
      controller
    );
    if (controller.ga_device) {
      spsControllerByCabinetAndGADevice.set(
        scopedKey(controller.control_cabinet_id, controller.ga_device),
        controller
      );
    }
  }

  const spsControllerSystemTypeByControllerTypeAndNumber = new Map<
    string,
    SPSControllerSystemType
  >();
  for (const item of lookups.spsControllerSystemTypes) {
    if (item.number === undefined || item.number === null) continue;
    spsControllerSystemTypeByControllerTypeAndNumber.set(
      spsControllerSystemTypeKey(item.sps_controller_id, item.system_type_id, item.number),
      item
    );
  }

  const existingFieldDeviceByBusinessKey = new Map<string, FieldDevice>();
  for (const fieldDevice of lookups.fieldDevices) {
    const key = fieldDeviceBusinessKey(
      fieldDevice.sps_controller_system_type?.sps_controller_name,
      fieldDevice.sps_controller_system_type?.number,
      fieldDevice.system_part_id,
      fieldDevice.apparat_id,
      parseInteger(fieldDevice.apparat_nr)
    );
    if (key && !existingFieldDeviceByBusinessKey.has(key)) {
      existingFieldDeviceByBusinessKey.set(key, fieldDevice);
    }
  }

  const systemPartByLookup = new Map<string, SystemPart>();
  for (const systemPart of lookups.systemParts) {
    setLookup(systemPartByLookup, systemPart.short_name, systemPart);
    setLookup(systemPartByLookup, systemPart.name, systemPart);
  }

  const apparatByLookup = new Map<string, Apparat>();
  for (const apparat of lookups.apparats) {
    setLookup(apparatByLookup, apparat.short_name, apparat);
    setLookup(apparatByLookup, apparat.name, apparat);
  }

  const notificationClassByNc = new Map<number, NotificationClass>();
  for (const item of lookups.notificationClasses) {
    notificationClassByNc.set(item.nc, item);
  }

  const stateTextByRefNumber = new Map<number, StateText>();
  const stateTextByLookup = new Map<string, StateText>();
  for (const item of lookups.stateTexts) {
    stateTextByRefNumber.set(item.ref_number, item);
    for (const value of stateTextValues(item)) {
      setLookup(stateTextByLookup, value, item);
    }
  }

  const alarmTypeByLookup = new Map<string, AlarmType>();
  for (const item of lookups.alarmTypes) {
    setLookup(alarmTypeByLookup, item.code, item);
    setLookup(alarmTypeByLookup, item.name, item);
  }

  return {
    buildingsByScope,
    controlCabinetByBuildingAndNr,
    spsControllerByCabinetAndDeviceName,
    spsControllerByCabinetAndGADevice,
    spsControllerSystemTypeByControllerTypeAndNumber,
    existingFieldDeviceByBusinessKey,
    systemTypes: [...lookups.systemTypes].sort((a, b) => a.number_min - b.number_min),
    systemPartByLookup,
    apparatByLookup,
    notificationClassByNc,
    stateTextByRefNumber,
    stateTextByLookup,
    alarmTypeByLookup
  };
}

function stateTextValues(item: StateText): string[] {
  return Array.from({ length: 16 }, (_, index) => {
    const key = `state_text${index + 1}` as keyof StateText;
    return typeof item[key] === 'string' ? (item[key] as string) : '';
  }).filter(Boolean);
}

function resolveBuilding(
  lookupIndex: LookupIndex,
  iwsCode: string,
  buildingGroup?: number
): Building | undefined {
  if (!iwsCode || !buildingGroup) return undefined;
  return lookupIndex.buildingsByScope.get(buildingKey(iwsCode, buildingGroup));
}

function resolveByShortOrName<T>(
  map: Map<string, T>,
  shortName: string,
  name: string
): T | undefined {
  if (shortName) {
    const byShort = map.get(normalizeLookupKey(shortName));
    if (byShort) return byShort;
  }

  if (name) {
    return map.get(normalizeLookupKey(name));
  }

  return undefined;
}

function setLookup<T>(map: Map<string, T>, value: string | undefined, item: T): void {
  if (!value) return;
  const key = normalizeLookupKey(value);
  if (!key || map.has(key)) return;
  map.set(key, item);
}

function buildingKey(iwsCode: string, buildingGroup: number): string {
  return `${normalizeLookupKey(iwsCode)}|${buildingGroup}`;
}

function scopedKey(scopeId: string, value: string): string {
  return `${scopeId}|${normalizeLookupKey(value)}`;
}

function spsControllerSystemTypeKey(
  spsControllerId: string,
  systemTypeId: string,
  number: number
): string {
  return `${spsControllerId}|${systemTypeId}|${number}`;
}

function resolveExistingSpsController(
  lookupIndex: LookupIndex,
  controlCabinetId: string,
  deviceName: string,
  gaDevice: string
): SPSController | undefined {
  const byDeviceName = deviceName
    ? lookupIndex.spsControllerByCabinetAndDeviceName.get(scopedKey(controlCabinetId, deviceName))
    : undefined;
  if (byDeviceName) return byDeviceName;

  return gaDevice
    ? lookupIndex.spsControllerByCabinetAndGADevice.get(scopedKey(controlCabinetId, gaDevice))
    : undefined;
}

function fieldDeviceBusinessKey(
  controllerDeviceName: string | undefined,
  spsSystemTypeNumber: number | undefined,
  systemPartId: string | undefined,
  apparatId: string | undefined,
  apparatNr: number | undefined
): string | undefined {
  if (!controllerDeviceName || !spsSystemTypeNumber || !systemPartId || !apparatId || !apparatNr) {
    return undefined;
  }

  return [
    normalizeLookupKey(controllerDeviceName),
    spsSystemTypeNumber,
    systemPartId,
    apparatId,
    apparatNr
  ].join('|');
}

function systemTypeKey(number: number): string {
  return `system-type:${number}`;
}

function parseDevicePartNumber(devicePart: string): number | undefined {
  const match = devicePart.trim().match(/(\d{1,2})$/);
  if (!match) return undefined;
  return parseInteger(match[1]);
}

function parseIwsFromDeviceName(deviceName: string): string {
  const parts = deviceName
    .split('_')
    .map((part) => part.trim())
    .filter(Boolean);
  return parts[0] ?? '';
}

function parseBuildingGroupFromDeviceName(deviceName: string): number | undefined {
  const parts = deviceName
    .split('_')
    .map((part) => part.trim())
    .filter(Boolean);
  return parseInteger(parts[1]);
}

function normalizeGADevice(value: string): string {
  return value.trim().toUpperCase();
}

function parseInteger(value: unknown): number | undefined {
  const raw = String(value ?? '').trim();
  if (!raw) return undefined;
  const parsed = Number.parseInt(raw, 10);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function numericCellValue(row: SpreadsheetRow, columnIndex: number): number {
  const value = Number(cellString(row, columnIndex));
  return Number.isFinite(value) ? value : 0;
}

function cellValue(row: SpreadsheetRow, columnIndex: number): unknown {
  return row[columnIndex];
}

function cellString(row: SpreadsheetRow | undefined, columnIndex: number): string {
  const value = row?.[columnIndex];
  if (value === null || value === undefined) return '';
  if (value instanceof Date) return value.toISOString();
  return String(value).trim();
}

function sourceCell(rowIndex: number, columnIndex: number): ImportSourceCell {
  const rowNumber = Math.max(0, rowIndex) + 1;
  const columnLabel = toSpreadsheetColumnLabel(columnIndex);
  return {
    rowNumber,
    columnIndex,
    columnLabel,
    address: `${columnLabel}${rowNumber}`
  };
}

type AddDiagnostic = (
  severity: ImportDiagnosticSeverity,
  message: string,
  entity: ImportEntityKind,
  cell?: ImportSourceCell,
  entityKey?: string
) => void;

function createDiagnosticCollector(diagnostics: ImportDiagnostic[]): AddDiagnostic {
  return (severity, message, entity, cell, entityKey) => {
    diagnostics.push({
      id: `${diagnostics.length + 1}`,
      severity,
      message,
      entity,
      cell,
      entityKey
    });
  };
}

function isMeaningfulLabel(value: string): boolean {
  const trimmed = value.trim();
  return trimmed.length > 0 && trimmed !== '-';
}

function addMeaningfulLabels(target: Set<string>, ...values: string[]): void {
  for (const value of values) {
    const trimmed = value.trim();
    if (!isMeaningfulLabel(trimmed)) continue;
    target.add(trimmed);
  }
}

function sortedLabels(values: Set<string>): string[] {
  return Array.from(values).sort((a, b) => a.localeCompare(b));
}

function sortedNumbers(values: Set<number>): number[] {
  return Array.from(values).sort((a, b) => a - b);
}

function emptyToUndefined(value: string): string | undefined {
  const trimmed = value.trim();
  return trimmed.length > 0 ? trimmed : undefined;
}

function parseBoolean(value: unknown, defaultValue: boolean): boolean {
  if (typeof value === 'boolean') return value;
  if (typeof value === 'number') return value === 1;
  const normalized = String(value ?? '')
    .trim()
    .toLowerCase();
  if (!normalized) return defaultValue;
  if (['1', 'true', 'wahr', 'yes', 'ja'].includes(normalized)) return true;
  if (['0', 'false', 'falsch', 'no', 'nein'].includes(normalized)) return false;
  return defaultValue;
}

function validateOptionalNetworkField(
  value: string,
  message: string,
  kind: 'ip' | 'subnet',
  cell: ImportSourceCell,
  addDiagnostic: AddDiagnostic
): void {
  const trimmed = value.trim();
  if (!trimmed) return;

  if (kind === 'ip' && !isValidIPv4(trimmed)) {
    addDiagnostic('error', message, 'sps_controller', cell);
  }

  if (kind === 'subnet' && !isValidSubnetMask(trimmed)) {
    addDiagnostic('error', message, 'sps_controller', cell);
  }
}

function validateVlan(value: string, cell: ImportSourceCell, addDiagnostic: AddDiagnostic): void {
  const trimmed = value.trim();
  if (!trimmed) return;
  const vlan = Number.parseInt(trimmed, 10);
  if (!Number.isFinite(vlan) || vlan < 1 || vlan > 4094) {
    addDiagnostic(
      'error',
      translate('field_device.importer.validation.vlan_range'),
      'sps_controller',
      cell
    );
  }
}

function isValidIPv4(value: string): boolean {
  const parts = value.split('.');
  if (parts.length !== 4) return false;
  return parts.every((part) => {
    if (!/^\d+$/.test(part)) return false;
    const number = Number(part);
    return number >= 0 && number <= 255;
  });
}

function isValidSubnetMask(value: string): boolean {
  if (!isValidIPv4(value)) return false;

  const bits = value
    .split('.')
    .map((part) => Number(part).toString(2).padStart(8, '0'))
    .join('');
  return /^1+0*$/.test(bits) && bits.includes('1');
}
