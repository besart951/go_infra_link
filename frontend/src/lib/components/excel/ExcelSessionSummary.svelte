<script lang="ts">
  import type { ExcelReadSession } from '$lib/domain/excel/index.js';
  import type { CreateObjectDataRequest } from '$lib/domain/facility/object-data.js';
  import type { CreateBacnetObjectRequest } from '$lib/domain/facility/bacnet-object.js';
  import ExcelSessionActionSection from './ExcelSessionActionSection.svelte';
  import ExcelSessionWarningSection from './ExcelSessionWarningSection.svelte';
  import ExcelSessionPreparedSummary from './ExcelSessionPreparedSummary.svelte';
  import ExcelSessionPreparedDetails from './ExcelSessionPreparedDetails.svelte';
  import ExcelSessionWorkbookSection from './ExcelSessionWorkbookSection.svelte';
  import { apparatRepository } from '$lib/infrastructure/api/apparatRepository.js';
  import { notificationClassRepository } from '$lib/infrastructure/api/notificationClassRepository.js';
  import { stateTextRepository } from '$lib/infrastructure/api/stateTextRepository.js';
  import { updateBacnetObject } from '$lib/infrastructure/api/bacnetObjectEndpoint.js';
  import { objectDataRepository } from '$lib/infrastructure/api/objectDataRepository.js';
  import { alarmTypeRepository } from '$lib/infrastructure/api/alarmTypeRepository.js';
  import { ApiException } from '$lib/api/client.js';

  interface Props {
    session: ExcelReadSession;
  }

  interface PreparedObjectData {
    objectDataId: string;
    request: CreateObjectDataRequest;
    plannedAlarmDefinitions: Array<{
      bacnetIndex: number;
      bacnetSoftwareId: string;
      name: string;
      alarmTypeId?: string;
      alarmTypeCode?: string;
    }>;
    plannedSoftwareReferenceLinks: Array<{ fromSoftwareId: string; toSoftwareId: string }>;
    issues: {
      missingApparatLabels: string[];
      missingStateTextLabels: string[];
      missingNotificationClassLabels: string[];
      missingSoftwareReferences: string[];
      missingHardwareEntries: string[];
      missingSoftwareNumberEntries: string[];
      missingHardwareCount: number;
      missingSoftwareNumberCount: number;
    };
  }

  interface PreparedSummary {
    objectDataCount: number;
    bacnetCount: number;
    missingApparats: number;
    missingStateTexts: number;
    missingNotificationClasses: number;
    missingSoftwareReferences: number;
    missingHardware: number;
    missingSoftwareNumbers: number;
    plannedAlarmDefinitionCreates: number;
    plannedSoftwareReferenceLinks: number;
  }

  interface CreateExecutionReport {
    total: number;
    success: number;
    failed: Array<{ objectDataId: string; reason: string }>;
    unresolvedSoftwareLinks: Array<{ objectDataId: string; from: string; to: string }>;
  }

  type PrepareFilterKey =
    | 'all'
    | 'missingApparats'
    | 'missingStateTexts'
    | 'missingNotificationClasses'
    | 'missingSoftwareReferences'
    | 'missingHardware'
    | 'missingSoftwareNumbers'
    | 'plannedAlarmDefinitions'
    | 'plannedSoftwareLinks';

  let { session }: Props = $props();

  let duplicateSoftwareIds = $state<Set<string>>(new Set());
  let duplicateCheckDone = $state(false);
  let preparing = $state(false);
  let creating = $state(false);
  let prepareError = $state<string | null>(null);
  let preparedPayloads = $state<PreparedObjectData[] | null>(null);
  let preparedSummary = $state<PreparedSummary | null>(null);
  let createError = $state<string | null>(null);
  let createReport = $state<CreateExecutionReport | null>(null);
  let activePrepareFilter = $state<PrepareFilterKey>('all');

  function normalizePart(value: string): string {
    return value.trim().toLowerCase();
  }

  function normalizeLookupKey(value: string): string {
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

  function isMeaningfulLabel(value: string): boolean {
    const normalized = value.trim();
    return normalized.length > 0 && normalized !== '-';
  }

  function rowIdentifier(objectDataId: string, bacnetObjectId: string): string {
    return `${objectDataId}::${bacnetObjectId}`;
  }

  function buildSoftwareKey(softwareType: string, softwareNumber: string): string {
    const normalizedType = normalizePart(softwareType);
    const numberRaw = normalizePart(softwareNumber);
    const parsedNumber = Number.parseInt(numberRaw, 10);
    const normalizedNumber = Number.isFinite(parsedNumber) ? String(parsedNumber) : numberRaw;
    if (normalizedType.length === 0 || normalizedNumber.length === 0) return '';
    return `${normalizedType}${normalizedNumber}`;
  }

  function toSoftwareId(softwareType: string, softwareNumber: number | string): string {
    return `${String(softwareType || '')
      .trim()
      .toUpperCase()}${String(softwareNumber ?? '').trim()}`;
  }

  function inferAlarmTypeCodeFromLabel(label: string): string | undefined {
    const normalized = normalizeLookupKey(label);
    if (!normalized) return undefined;

    // 1. Exact or highly specific matches first
    // active_inactive
    if (
      normalized.includes('alarmactive') ||
      normalized.includes('alarminactive') ||
      normalized.includes('alarmaktiv') ||
      normalized.includes('alarminaktiv') ||
      normalized.includes('elapsedactivetimealarm') ||
      (normalized.includes('alarm') && normalized.includes('state')) ||
      (normalized.includes('state') && (normalized.includes('auf') || normalized.includes('zu')))
    ) {
      return 'active_inactive';
    }

    // priority_write
    if (normalized.includes('priorityforwrit') || normalized.includes('priorityforwriting')) {
      return 'priority_write';
    }

    // io_monitoring_limit and io_monitoring
    const hasIoMonitoringHint =
      normalized.includes('io') ||
      normalized.includes('uberwach') ||
      normalized.includes('ueberwach') ||
      normalized.includes('ruckmeldung');
    const hasLimitHint =
      normalized.includes('limit') ||
      normalized.includes('grenz') ||
      normalized.includes('high') ||
      normalized.includes('low') ||
      normalized.includes('voralarm');

    if (hasIoMonitoringHint && hasLimitHint) return 'io_monitoring_limit';
    if (hasIoMonitoringHint) return 'io_monitoring';

    // cov_logging
    if (normalized.includes('loggingtypecov') || normalized.includes('cov')) {
      return 'cov_logging';
    }

    // elapsed_active_time
    if (normalized.includes('elapsedactivetime') || normalized.includes('elapsedaktiv')) {
      return 'elapsed_active_time';
    }

    // pid_control
    if (
      normalized.includes('controlledvariable') ||
      normalized.includes('presentvalue') ||
      normalized.includes('setpoint') ||
      normalized.includes('proportionalconstant') ||
      normalized.includes('integralconstant') ||
      normalized.includes('maximumoutput') ||
      normalized.includes('minimumoutput') ||
      normalized.includes('errorlimit') ||
      normalized.includes('timedelay') ||
      normalized.includes('pid')
    ) {
      return 'pid_control';
    }

    // limit_high_low
    if (
      hasLimitHint ||
      normalized.includes('differenz') ||
      normalized.includes('diff') ||
      normalized.includes('alarmbei')
    ) {
      return 'limit_high_low';
    }

    // position_control
    if (
      normalized.includes('position') ||
      normalized.includes('positionskontrolle') ||
      normalized.includes('flusskontrolle') ||
      normalized.includes('leistungsregelung') ||
      normalized.includes('motstop') ||
      normalized.includes('vnom') ||
      normalized.includes('vmax') ||
      normalized.includes('pnom') ||
      normalized.includes('pmax')
    ) {
      return 'position_control';
    }

    // state_mapping
    if (
      normalized.includes('state1') ||
      normalized.includes('state') ||
      normalized.includes('zustandszuordnung') ||
      normalized.includes('storungsundserviceinformationen') ||
      /^\d+=/.test(label.replace(/\s+/g, '')) || // matches "0=DG0"
      normalized.includes('0dg0') // mapped normalized version
    ) {
      // Don't override clear active_inactive states from above
      return 'state_mapping';
    }

    // custom_value
    if (normalized.includes('individuell') || normalized.includes('custom')) {
      return 'custom_value';
    }

    // active_inactive (fallback for simple "Alarm")
    if (normalized.includes('alarmstate') || normalized.includes('alarm')) {
      return 'active_inactive';
    }

    // Check if it's purely a setting, unit, or range that should map to undefined
    const purelyUnitOrRangePattern =
      /^(\d+[kcpamhjsluxv]*|\d+-\d+%?|\d+°c|\d+\.\.\.\d+lux|%?|°c.*?|min\.?|%ref)$/i;
    const unitsOnly = [
      'c',
      'k',
      'pa',
      'hpa',
      'kw',
      'kwh',
      'm',
      'ms',
      's',
      'h',
      'j',
      'lux',
      'umin',
      'm3h',
      'gkg',
      'ref'
    ];

    if (
      unitsOnly.includes(normalized) ||
      purelyUnitOrRangePattern.test(label.trim().toLowerCase())
    ) {
      return undefined;
    }

    // Unmapped things or strings only containing numbers are not alarm definitions
    const hasOnlyDigitsOrUnits = /^[\d\s\WcKkPAMhjsLuxv]+$/.test(label.toLowerCase());
    if (hasOnlyDigitsOrUnits) {
      return undefined;
    }

    return undefined;
  }

  function resolveAlarmTypeFromLabel(
    label: string,
    alarmTypeByCode: Map<string, string>,
    alarmTypeByLookupKey: Map<string, { id: string; code: string }>
  ): { id?: string; code?: string } {
    const normalizedLabel = normalizeLookupKey(label);
    if (!normalizedLabel) return {};

    const directMatch = alarmTypeByLookupKey.get(normalizedLabel);
    if (directMatch) {
      return {
        id: directMatch.id,
        code: directMatch.code
      };
    }

    const inferredCode = inferAlarmTypeCodeFromLabel(label);
    if (!inferredCode) {
      return {};
    }
    const inferredId = alarmTypeByCode.get(inferredCode);
    if (!inferredId) {
      return {};
    }

    return {
      id: inferredId,
      code: inferredCode
    };
  }

  function parseHardwareLabel(label: string): { type: string; quantity: number } {
    const normalized = label.trim();
    if (normalized.length === 0) return { type: '', quantity: 0 };
    const match = normalized.match(/^([A-Za-z]+)(\d+)$/);
    if (!match) return { type: '', quantity: 0 };
    return { type: match[1].toLowerCase(), quantity: Number.parseInt(match[2], 10) };
  }

  function formatCreateError(error: unknown): string {
    if (error instanceof ApiException) {
      if (error.details && typeof error.details === 'object') {
        const fieldEntries = Object.entries(error.details as Record<string, unknown>).filter(
          ([, value]) => typeof value === 'string'
        ) as Array<[string, string]>;
        if (fieldEntries.length > 0) {
          return fieldEntries.map(([field, msg]) => `${field}: ${msg}`).join(', ');
        }
      }
      return error.message || `HTTP ${error.status}`;
    }

    if (error instanceof Error) return error.message;
    return 'Unbekannter Fehler';
  }

  function hasIssueDetails(item: PreparedObjectData): boolean {
    return (
      item.issues.missingApparatLabels.length > 0 ||
      item.issues.missingStateTextLabels.length > 0 ||
      item.issues.missingNotificationClassLabels.length > 0 ||
      item.issues.missingSoftwareReferences.length > 0 ||
      item.issues.missingHardwareEntries.length > 0 ||
      item.issues.missingSoftwareNumberEntries.length > 0
    );
  }

  function matchesPrepareFilter(item: PreparedObjectData, filter: PrepareFilterKey): boolean {
    switch (filter) {
      case 'all':
        return (
          hasIssueDetails(item) ||
          item.plannedAlarmDefinitions.length > 0 ||
          item.plannedSoftwareReferenceLinks.length > 0
        );
      case 'missingApparats':
        return item.issues.missingApparatLabels.length > 0;
      case 'missingStateTexts':
        return item.issues.missingStateTextLabels.length > 0;
      case 'missingNotificationClasses':
        return item.issues.missingNotificationClassLabels.length > 0;
      case 'missingSoftwareReferences':
        return item.issues.missingSoftwareReferences.length > 0;
      case 'missingHardware':
        return item.issues.missingHardwareEntries.length > 0;
      case 'missingSoftwareNumbers':
        return item.issues.missingSoftwareNumberEntries.length > 0;
      case 'plannedAlarmDefinitions':
        return item.plannedAlarmDefinitions.length > 0;
      case 'plannedSoftwareLinks':
        return item.plannedSoftwareReferenceLinks.length > 0;
    }
  }

  function setPrepareFilter(filter: PrepareFilterKey): void {
    activePrepareFilter = filter;
  }

  const filteredPreparedPayloads = $derived(
    preparedPayloads
      ? preparedPayloads.filter((item) => matchesPrepareFilter(item, activePrepareFilter))
      : []
  );

  async function fetchAllPages<T>(
    fetchPage: (
      page: number,
      limit: number
    ) => Promise<{ items: T[]; page: number; total_pages: number }>
  ): Promise<T[]> {
    const limit = 500;
    let page = 1;
    const items: T[] = [];
    while (true) {
      const response = await fetchPage(page, limit);
      items.push(...response.items);
      if (response.page >= response.total_pages) break;
      page += 1;
    }
    return items;
  }

  function repositoryPage<T>(response: {
    items: T[];
    metadata: { page: number; totalPages: number };
  }): { items: T[]; page: number; total_pages: number } {
    return {
      items: response.items,
      page: response.metadata.page,
      total_pages: response.metadata.totalPages
    };
  }

  function runDuplicateSoftwareCheck(): void {
    const duplicates = new Set<string>();

    for (const objectData of session.objectDataExcel) {
      const softwareCounts = new Map<string, number>();
      for (const bacnetObject of objectData.bacnet_objects) {
        const key = buildSoftwareKey(
          bacnetObject.software_type || '',
          bacnetObject.software_number || ''
        );
        if (key.length === 0) continue;

        softwareCounts.set(key, (softwareCounts.get(key) ?? 0) + 1);
      }

      for (const bacnetObject of objectData.bacnet_objects) {
        const key = buildSoftwareKey(
          bacnetObject.software_type || '',
          bacnetObject.software_number || ''
        );
        if (key.length === 0) continue;
        if ((softwareCounts.get(key) ?? 0) > 1) {
          duplicates.add(rowIdentifier(objectData.id, bacnetObject.id));
        }
      }
    }

    duplicateSoftwareIds = duplicates;
    duplicateCheckDone = true;
  }

  async function prepareCreatePayloads(): Promise<void> {
    if (preparing) return;
    preparing = true;
    prepareError = null;
    preparedPayloads = null;
    preparedSummary = null;
    activePrepareFilter = 'all';

    try {
      const [apparats, stateTexts, notificationClasses, alarmTypes] = await Promise.all([
        fetchAllPages(async (page, limit) =>
          repositoryPage(
            await apparatRepository.list({
              pagination: { page, pageSize: limit },
              search: { text: '' }
            })
          )
        ),
        fetchAllPages(async (page, limit) =>
          repositoryPage(
            await stateTextRepository.list({
              pagination: { page, pageSize: limit },
              search: { text: '' }
            })
          )
        ),
        fetchAllPages(async (page, limit) =>
          repositoryPage(
            await notificationClassRepository.list({
              pagination: { page, pageSize: limit },
              search: { text: '' }
            })
          )
        ),
        fetchAllPages(async (page, limit) => {
          const res = await alarmTypeRepository.list({ page, pageSize: limit });
          return {
            items: res.items,
            page: res.page,
            total_pages: res.totalPages
          };
        })
      ]);

      const apparatMap = new Map<string, string>();
      apparats.forEach((apparat) => {
        if (apparat.short_name) {
          apparatMap.set(normalizeLookupKey(apparat.short_name), apparat.id);
        }
        if (apparat.name) {
          apparatMap.set(normalizeLookupKey(apparat.name), apparat.id);
        }
      });

      const stateTextMap = new Map<number, string>();
      stateTexts.forEach((stateText) => {
        stateTextMap.set(stateText.ref_number, stateText.id);
      });

      const notificationClassMap = new Map<number, string>();
      notificationClasses.forEach((notificationClass) => {
        notificationClassMap.set(notificationClass.nc, notificationClass.id);
      });

      const alarmTypeByCode = new Map<string, string>();
      const alarmTypeByLookupKey = new Map<string, { id: string; code: string }>();
      alarmTypes.forEach((alarmType) => {
        if (alarmType.code) {
          alarmTypeByCode.set(alarmType.code, alarmType.id);
          alarmTypeByLookupKey.set(normalizeLookupKey(alarmType.code), {
            id: alarmType.id,
            code: alarmType.code
          });
        }
        if (alarmType.name) {
          alarmTypeByLookupKey.set(normalizeLookupKey(alarmType.name), {
            id: alarmType.id,
            code: alarmType.code
          });
        }
      });

      const preparedItems: PreparedObjectData[] = [];
      let totalBacnetObjects = 0;
      let missingHardware = 0;
      let missingSoftwareNumbers = 0;
      let plannedAlarmDefinitionCreateCount = 0;
      let plannedSoftwareReferenceLinkCount = 0;
      const missingApparatLabels = new Set<string>();
      const missingStateTextLabels = new Set<string>();
      const missingNotificationClassLabels = new Set<string>();
      const missingSoftwareReferenceLabels = new Set<string>();

      for (const objectData of session.objectDataExcel) {
        const apparatIds = new Set<string>();
        const softwareIdMap = new Map<string, string>();
        const plannedAlarmDefinitions: PreparedObjectData['plannedAlarmDefinitions'] = [];
        const plannedSoftwareLinks: Array<{ fromSoftwareId: string; toSoftwareId: string }> = [];
        objectData.bacnet_objects.forEach((bacnetObject) => {
          const softwareKey = buildSoftwareKey(
            bacnetObject.software_type || '',
            bacnetObject.software_number || ''
          );
          if (softwareKey.length > 0) {
            softwareIdMap.set(softwareKey, softwareKey.toUpperCase());
          }
        });

        const bacnetRequests: CreateBacnetObjectRequest[] = [];
        const localMissingApparats = new Set<string>();
        const localMissingStateTexts = new Set<string>();
        const localMissingNotificationClasses = new Set<string>();
        const localMissingSoftwareReferences = new Set<string>();
        const localMissingHardwareEntries = new Set<string>();
        const localMissingSoftwareNumberEntries = new Set<string>();
        let localMissingHardware = 0;
        let localMissingSoftwareNumbers = 0;

        for (const [bacnetIndex, bacnetObject] of objectData.bacnet_objects.entries()) {
          const apparatLabelRaw = (bacnetObject.apparat_label || '').trim();
          const apparatLabel = normalizeLookupKey(apparatLabelRaw);
          if (isMeaningfulLabel(apparatLabelRaw)) {
            const apparatId = apparatMap.get(apparatLabel);
            if (apparatId) {
              apparatIds.add(apparatId);
            } else {
              localMissingApparats.add(apparatLabelRaw);
            }
          }

          const hardware = parseHardwareLabel(bacnetObject.hardware_label || '');
          const rawHardwareLabel = (bacnetObject.hardware_label || '').trim();
          if (rawHardwareLabel.length > 0 && (!hardware.type || !hardware.quantity)) {
            localMissingHardware += 1;
            localMissingHardwareEntries.add(
              `${bacnetObject.text_fix || '(kein text_fix)'} | hardware="${bacnetObject.hardware_label || ''}"`
            );
          }

          const softwareType = (bacnetObject.software_type || '').trim().toLowerCase();
          const softwareNumber = Number.parseInt(bacnetObject.software_number || '', 10);
          if (!Number.isFinite(softwareNumber)) {
            localMissingSoftwareNumbers += 1;
            localMissingSoftwareNumberEntries.add(
              `${bacnetObject.text_fix || '(kein text_fix)'} | software_type="${softwareType}" software_number="${bacnetObject.software_number || ''}"`
            );
          }

          const stateTextLabel = (bacnetObject.state_text_label || '').trim();
          const stateTextNumber = Number.parseInt(stateTextLabel, 10);
          const stateTextId = Number.isFinite(stateTextNumber)
            ? stateTextMap.get(stateTextNumber)
            : undefined;
          if (isMeaningfulLabel(stateTextLabel) && !stateTextId) {
            localMissingStateTexts.add(stateTextLabel);
          }

          const notificationLabel = (bacnetObject.notification_class_label || '').trim();
          const notificationNumber = Number.parseInt(notificationLabel, 10);
          const notificationClassId = Number.isFinite(notificationNumber)
            ? notificationClassMap.get(notificationNumber)
            : undefined;
          if (isMeaningfulLabel(notificationLabel) && !notificationClassId) {
            localMissingNotificationClasses.add(notificationLabel);
          }

          const softwareKey = buildSoftwareKey(
            bacnetObject.software_type || '',
            bacnetObject.software_number || ''
          );
          const fromSoftwareId = softwareKey.toUpperCase();

          const alarmLabelRaw = (bacnetObject.alarm_definition_label || '').trim();
          const resolvedAlarmType = isMeaningfulLabel(alarmLabelRaw)
            ? resolveAlarmTypeFromLabel(alarmLabelRaw, alarmTypeByCode, alarmTypeByLookupKey)
            : {};
          if (isMeaningfulLabel(alarmLabelRaw) && fromSoftwareId.length > 0) {
            plannedAlarmDefinitions.push({
              bacnetIndex,
              bacnetSoftwareId: fromSoftwareId,
              name: alarmLabelRaw,
              alarmTypeCode: resolvedAlarmType.code,
              alarmTypeId: resolvedAlarmType.id
            });
          }

          const softwareReferenceLabel = normalizePart(bacnetObject.software_reference_label || '');
          if (softwareReferenceLabel.length > 0) {
            const targetSoftwareId = softwareIdMap.get(softwareReferenceLabel);
            if (targetSoftwareId && fromSoftwareId.length > 0) {
              plannedSoftwareLinks.push({
                fromSoftwareId,
                toSoftwareId: targetSoftwareId
              });
            } else {
              localMissingSoftwareReferences.add(bacnetObject.software_reference_label);
            }
          }

          const bacnetRequest: CreateBacnetObjectRequest = {
            text_fix: bacnetObject.text_fix,
            description: bacnetObject.description || undefined,
            gms_visible: bacnetObject.gms_visible,
            optional: bacnetObject.is_optional,
            text_individual: bacnetObject.text_individual || undefined,
            software_type: softwareType,
            software_number: Number.isFinite(softwareNumber) ? softwareNumber : 0,
            hardware_type: hardware.type,
            hardware_quantity: hardware.quantity,
            software_reference_id: undefined,
            state_text_id: stateTextId,
            notification_class_id: notificationClassId,
            alarm_type_id: resolvedAlarmType.id
          };

          bacnetRequests.push(bacnetRequest);
        }

        totalBacnetObjects += bacnetRequests.length;
        missingHardware += localMissingHardware;
        missingSoftwareNumbers += localMissingSoftwareNumbers;
        plannedAlarmDefinitionCreateCount += plannedAlarmDefinitions.length;
        plannedSoftwareReferenceLinkCount += plannedSoftwareLinks.length;
        localMissingApparats.forEach((label) => missingApparatLabels.add(label));
        localMissingStateTexts.forEach((label) => missingStateTextLabels.add(label));
        localMissingNotificationClasses.forEach((label) =>
          missingNotificationClassLabels.add(label)
        );
        localMissingSoftwareReferences.forEach((label) =>
          missingSoftwareReferenceLabels.add(label)
        );

        const request: CreateObjectDataRequest = {
          description: objectData.description,
          version: '1.0',
          is_active: true,
          apparat_ids: Array.from(apparatIds),
          bacnet_objects: bacnetRequests
        };

        preparedItems.push({
          objectDataId: objectData.id,
          request,
          plannedAlarmDefinitions,
          plannedSoftwareReferenceLinks: plannedSoftwareLinks,
          issues: {
            missingApparatLabels: Array.from(localMissingApparats),
            missingStateTextLabels: Array.from(localMissingStateTexts),
            missingNotificationClassLabels: Array.from(localMissingNotificationClasses),
            missingSoftwareReferences: Array.from(localMissingSoftwareReferences),
            missingHardwareEntries: Array.from(localMissingHardwareEntries),
            missingSoftwareNumberEntries: Array.from(localMissingSoftwareNumberEntries),
            missingHardwareCount: localMissingHardware,
            missingSoftwareNumberCount: localMissingSoftwareNumbers
          }
        });
      }

      preparedPayloads = preparedItems;
      preparedSummary = {
        objectDataCount: preparedItems.length,
        bacnetCount: totalBacnetObjects,
        missingApparats: missingApparatLabels.size,
        missingStateTexts: missingStateTextLabels.size,
        missingNotificationClasses: missingNotificationClassLabels.size,
        missingSoftwareReferences: missingSoftwareReferenceLabels.size,
        missingHardware,
        missingSoftwareNumbers,
        plannedAlarmDefinitionCreates: plannedAlarmDefinitionCreateCount,
        plannedSoftwareReferenceLinks: plannedSoftwareReferenceLinkCount
      };
    } catch (error) {
      prepareError =
        error instanceof Error
          ? error.message
          : 'Erstellungspayloads konnten nicht vorbereitet werden.';
    } finally {
      preparing = false;
    }
  }

  async function createAllPreparedSequentially(): Promise<void> {
    if (creating) return;
    if (!preparedPayloads || preparedPayloads.length === 0) {
      createError = 'Bitte zuerst "Erstellung vorbereiten" ausführen.';
      return;
    }

    creating = true;
    createError = null;
    createReport = null;

    const failed: Array<{ objectDataId: string; reason: string }> = [];
    const unresolvedSoftwareLinks: Array<{ objectDataId: string; from: string; to: string }> = [];
    let success = 0;

    for (const item of preparedPayloads) {
      try {
        const createdObjectData = await objectDataRepository.create({
          ...item.request
        });

        const createdBacnetObjects = await objectDataRepository.getBacnetObjects(
          createdObjectData.id
        );
        const createdSoftwareIdMap = new Map<string, string>();
        createdBacnetObjects.forEach((bacnet) => {
          createdSoftwareIdMap.set(
            toSoftwareId(bacnet.software_type, bacnet.software_number),
            bacnet.id
          );
        });

        for (const link of item.plannedSoftwareReferenceLinks) {
          const fromId = createdSoftwareIdMap.get(link.fromSoftwareId);
          const toId = createdSoftwareIdMap.get(link.toSoftwareId);
          if (!fromId || !toId) {
            unresolvedSoftwareLinks.push({
              objectDataId: item.objectDataId,
              from: link.fromSoftwareId,
              to: link.toSoftwareId
            });
            continue;
          }

          await updateBacnetObject(fromId, { software_reference_id: toId });
        }

        success += 1;
      } catch (error) {
        failed.push({
          objectDataId: item.objectDataId,
          reason: formatCreateError(error)
        });
      }
    }

    createReport = {
      total: preparedPayloads.length,
      success,
      failed,
      unresolvedSoftwareLinks
    };

    if (failed.length > 0) {
      createError = `${failed.length} Objektdaten-Einträge konnten nicht erstellt werden.`;
    }

    creating = false;
  }

  $effect(() => {
    session;
    duplicateSoftwareIds = new Set();
    duplicateCheckDone = false;
    prepareError = null;
    preparedPayloads = null;
    preparedSummary = null;
    createError = null;
    createReport = null;
    activePrepareFilter = 'all';
  });
</script>

<div class="rounded-lg border bg-background p-4">
  <ExcelSessionActionSection
    {session}
    {duplicateSoftwareIds}
    {duplicateCheckDone}
    {preparing}
    {creating}
    onRunDuplicateSoftwareCheck={runDuplicateSoftwareCheck}
    onPrepareCreatePayloads={prepareCreatePayloads}
    onCreateAllPreparedSequentially={createAllPreparedSequentially}
  />
  <ExcelSessionWarningSection {prepareError} {createError} {createReport} />
  <ExcelSessionPreparedSummary
    {preparedSummary}
    {activePrepareFilter}
    onSetPrepareFilter={setPrepareFilter}
  />
  <ExcelSessionPreparedDetails {preparedPayloads} {filteredPreparedPayloads} />
  <ExcelSessionWorkbookSection {session} {duplicateSoftwareIds} {rowIdentifier} />
</div>
