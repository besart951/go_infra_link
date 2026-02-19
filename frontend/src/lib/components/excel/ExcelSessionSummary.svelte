<script lang="ts">
	import { CircleCheck } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import type { ExcelReadSession } from '$lib/domain/excel/index.js';
	import type { CreateObjectDataRequest } from '$lib/domain/facility/object-data.js';
	import type { CreateBacnetObjectRequest } from '$lib/domain/facility/bacnet-object.js';
	import {
		createAlarmDefinition,
		listApparats,
		listStateTexts,
		listNotificationClasses,
		updateBacnetObject
	} from '$lib/infrastructure/api/facility.adapter.js';
	import { objectDataRepository } from '$lib/infrastructure/api/objectDataRepository.js';
	import { alarmTypeRepository } from '$lib/infrastructure/api/alarmTypeRepository.js';

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
			.replace(/ß/g, 'ss')
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
		const normalizedNumber = normalizePart(softwareNumber);
		if (normalizedType.length === 0 || normalizedNumber.length === 0) return '';
		return `${normalizedType}${normalizedNumber}`;
	}

	function toSoftwareId(softwareType: string, softwareNumber: number | string): string {
		return `${String(softwareType || '').trim().toUpperCase()}${String(softwareNumber ?? '').trim()}`;
	}

	function inferAlarmTypeCodeFromLabel(label: string): string {
		const normalized = normalizeLookupKey(label);
		if (!normalized) return 'custom_value';

		if (normalized.includes('cov')) return 'cov_logging';
		if (normalized.includes('pid')) return 'pid_control';
		if (normalized.includes('position')) return 'position_control';
		if (normalized.includes('state') || normalized.includes('zustand')) return 'state_mapping';
		if (normalized.includes('priority') || normalized.includes('prioritat')) return 'priority_write';
		if (normalized.includes('io') || normalized.includes('ruckmeldung')) return 'io_monitoring';
		if (normalized.includes('limit') || normalized.includes('grenz')) return 'limit_high_low';
		if (
			normalized.includes('active') ||
			normalized.includes('inactive') ||
			normalized.includes('aktiv') ||
			normalized.includes('inaktiv') ||
			normalized.includes('alarm')
		)
			return 'active_inactive';

		return 'custom_value';
	}

	function parseHardwareLabel(label: string): { type: string; quantity: number } {
		const normalized = label.trim();
		if (normalized.length === 0) return { type: '', quantity: 0 };
		const match = normalized.match(/^([A-Za-z]+)(\d+)$/);
		if (!match) return { type: '', quantity: 0 };
		return { type: match[1].toLowerCase(), quantity: Number.parseInt(match[2], 10) };
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
				return hasIssueDetails(item) || item.plannedAlarmDefinitions.length > 0 || item.plannedSoftwareReferenceLinks.length > 0;
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

	function isFilterActive(filter: PrepareFilterKey): boolean {
		return activePrepareFilter === filter;
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
		fetchPage: (page: number, limit: number) => Promise<{ items: T[]; page: number; total_pages: number }>
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
				fetchAllPages((page, limit) => listApparats({ page, limit })),
				fetchAllPages((page, limit) => listStateTexts({ page, limit })),
				fetchAllPages((page, limit) => listNotificationClasses({ page, limit })),
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
			alarmTypes.forEach((alarmType) => {
				if (alarmType.code) {
					alarmTypeByCode.set(alarmType.code, alarmType.id);
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
							`${bacnetObject.text_fix || '(no text_fix)'} | hardware="${bacnetObject.hardware_label || ''}"`
						);
					}

					const softwareType = (bacnetObject.software_type || '').trim().toLowerCase();
					const softwareNumber = Number.parseInt(bacnetObject.software_number || '', 10);
					if (!Number.isFinite(softwareNumber)) {
						localMissingSoftwareNumbers += 1;
						localMissingSoftwareNumberEntries.add(
							`${bacnetObject.text_fix || '(no text_fix)'} | software_type="${softwareType}" software_number="${bacnetObject.software_number || ''}"`
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
					if (isMeaningfulLabel(alarmLabelRaw) && fromSoftwareId.length > 0) {
						const inferredAlarmTypeCode = inferAlarmTypeCodeFromLabel(alarmLabelRaw);
						plannedAlarmDefinitions.push({
							bacnetIndex,
							bacnetSoftwareId: fromSoftwareId,
							name: alarmLabelRaw,
							alarmTypeCode: inferredAlarmTypeCode,
							alarmTypeId: alarmTypeByCode.get(inferredAlarmTypeCode)
						});
					}

					const softwareReferenceLabel = normalizePart(
						bacnetObject.software_reference_label || ''
					);
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
						alarm_definition_id: undefined
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
				localMissingNotificationClasses.forEach((label) => missingNotificationClassLabels.add(label));
				localMissingSoftwareReferences.forEach((label) => missingSoftwareReferenceLabels.add(label));

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
			prepareError = error instanceof Error ? error.message : 'Failed to prepare create payloads.';
		} finally {
			preparing = false;
		}
	}

	async function createAllPreparedSequentially(): Promise<void> {
		if (creating) return;
		if (!preparedPayloads || preparedPayloads.length === 0) {
			createError = 'Please run "Prepare create payload" first.';
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
				const alarmIdByBacnetIndex = new Map<number, string>();
				for (const alarmPlan of item.plannedAlarmDefinitions) {
					const createdAlarm = await createAlarmDefinition({
						name: alarmPlan.name,
						alarm_type_id: alarmPlan.alarmTypeId
					});
					alarmIdByBacnetIndex.set(alarmPlan.bacnetIndex, createdAlarm.id);
				}

				const withAlarmIds = (item.request.bacnet_objects ?? []).map((bacnet, bacnetIndex) => {
					return {
						...bacnet,
						alarm_definition_id: alarmIdByBacnetIndex.get(bacnetIndex)
					};
				});

				const createdObjectData = await objectDataRepository.create({
					...item.request,
					bacnet_objects: withAlarmIds
				});

				const createdBacnetObjects = await objectDataRepository.getBacnetObjects(createdObjectData.id);
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
					reason: error instanceof Error ? error.message : 'Unknown error'
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
			createError = `${failed.length} object data entries failed during create.`;
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
	<div class="mb-3 flex items-center gap-2">
		<CircleCheck class="size-4 text-primary" />
		<h2 class="text-sm font-semibold">Object data loaded</h2>
	</div>
	<div class="mb-4 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground">
		<span>{session.fileName}</span>
		<span>{session.objectDataExcel.length} object data</span>
		<Button type="button" size="sm" variant="outline" onclick={runDuplicateSoftwareCheck}>
			Check duplicate software ids
		</Button>
		<Button type="button" size="sm" variant="outline" onclick={prepareCreatePayloads} disabled={preparing}>
			{preparing ? 'Preparing...' : 'Prepare create payload'}
		</Button>
		<Button
			type="button"
			size="sm"
			variant="outline"
			onclick={createAllPreparedSequentially}
			disabled={creating || preparing}
		>
			{creating ? 'Creating...' : 'Create all sequentially'}
		</Button>
		{#if duplicateCheckDone}
			<span>
				{duplicateSoftwareIds.size > 0
					? `${duplicateSoftwareIds.size} non-unique BACnet rows marked`
					: 'All BACnet software ids are unique'}
			</span>
		{/if}
	</div>
	{#if prepareError}
		<div class="mb-4 rounded-md border border-destructive/40 bg-destructive/10 p-3 text-xs text-destructive">
			{prepareError}
		</div>
	{/if}
	{#if createError}
		<div class="mb-4 rounded-md border border-destructive/40 bg-destructive/10 p-3 text-xs text-destructive">
			{createError}
		</div>
	{/if}
	{#if createReport}
		<div class="mb-4 rounded-md border bg-muted/20 p-3 text-xs text-muted-foreground">
			<span>Created: {createReport.success}/{createReport.total}</span>
			<span class="ml-3">Failed: {createReport.failed.length}</span>
			<span class="ml-3">Unresolved software links: {createReport.unresolvedSoftwareLinks.length}</span>
			{#if createReport.failed.length > 0}
				<div class="mt-2">
					<strong class="text-foreground">Failed object data:</strong>
					<p>
						{createReport.failed
							.map((item) => `${item.objectDataId} (${item.reason})`)
							.join(' | ')}
					</p>
				</div>
			{/if}
			{#if createReport.unresolvedSoftwareLinks.length > 0}
				<div class="mt-2">
					<strong class="text-foreground">Unresolved software links:</strong>
					<p>
						{createReport.unresolvedSoftwareLinks
							.map((item) => `${item.objectDataId}: ${item.from} -> ${item.to}`)
							.join(' | ')}
					</p>
				</div>
			{/if}
		</div>
	{/if}
	{#if preparedSummary}
		<div class="mb-4 rounded-md border bg-muted/20 p-3 text-xs text-muted-foreground">
			<button
				type="button"
				onclick={() => setPrepareFilter('all')}
				class={`cursor-pointer ${isFilterActive('all') ? 'font-semibold text-foreground underline' : ''}`}
			>
				{preparedSummary.objectDataCount} object data prepared
			</button>
			<button type="button" onclick={() => setPrepareFilter('all')} class="ml-3 cursor-pointer">
				{preparedSummary.bacnetCount} bacnet objects
			</button>
			<button
				type="button"
				onclick={() => setPrepareFilter('missingApparats')}
				class={`ml-3 cursor-pointer ${isFilterActive('missingApparats') ? 'font-semibold text-foreground underline' : ''}`}
			>
				Missing apparats: {preparedSummary.missingApparats}
			</button>
			<button
				type="button"
				onclick={() => setPrepareFilter('missingStateTexts')}
				class={`ml-3 cursor-pointer ${isFilterActive('missingStateTexts') ? 'font-semibold text-foreground underline' : ''}`}
			>
				Missing state texts: {preparedSummary.missingStateTexts}
			</button>
			<button
				type="button"
				onclick={() => setPrepareFilter('missingNotificationClasses')}
				class={`ml-3 cursor-pointer ${isFilterActive('missingNotificationClasses') ? 'font-semibold text-foreground underline' : ''}`}
			>
				Missing notif classes: {preparedSummary.missingNotificationClasses}
			</button>
			<button
				type="button"
				onclick={() => setPrepareFilter('missingSoftwareReferences')}
				class={`ml-3 cursor-pointer ${isFilterActive('missingSoftwareReferences') ? 'font-semibold text-foreground underline' : ''}`}
			>
				Missing software refs: {preparedSummary.missingSoftwareReferences}
			</button>
			<button
				type="button"
				onclick={() => setPrepareFilter('missingHardware')}
				class={`ml-3 cursor-pointer ${isFilterActive('missingHardware') ? 'font-semibold text-foreground underline' : ''}`}
			>
				Missing hardware: {preparedSummary.missingHardware}
			</button>
			<button
				type="button"
				onclick={() => setPrepareFilter('missingSoftwareNumbers')}
				class={`ml-3 cursor-pointer ${isFilterActive('missingSoftwareNumbers') ? 'font-semibold text-foreground underline' : ''}`}
			>
				Missing software numbers: {preparedSummary.missingSoftwareNumbers}
			</button>
			<button
				type="button"
				onclick={() => setPrepareFilter('plannedAlarmDefinitions')}
				class={`ml-3 cursor-pointer ${isFilterActive('plannedAlarmDefinitions') ? 'font-semibold text-foreground underline' : ''}`}
			>
				Planned alarm creates: {preparedSummary.plannedAlarmDefinitionCreates}
			</button>
			<button
				type="button"
				onclick={() => setPrepareFilter('plannedSoftwareLinks')}
				class={`ml-3 cursor-pointer ${isFilterActive('plannedSoftwareLinks') ? 'font-semibold text-foreground underline' : ''}`}
			>
				Planned software links: {preparedSummary.plannedSoftwareReferenceLinks}
			</button>
		</div>
	{/if}
	{#if preparedPayloads}
		<div class="mb-4 space-y-2">
			{#if filteredPreparedPayloads.length === 0}
				<div class="rounded-md border border-dashed p-3 text-xs text-muted-foreground">
					No entries found for the selected filter.
				</div>
			{/if}
			{#each filteredPreparedPayloads as preparedItem}
					<details class="rounded-md border bg-background p-3 text-xs">
						<summary class="cursor-pointer font-medium">
							{preparedItem.objectDataId} - missing details
						</summary>
						<div class="mt-2 space-y-2 text-muted-foreground">
							{#if preparedItem.issues.missingApparatLabels.length > 0}
								<div>
									<strong class="text-foreground">Missing apparat labels:</strong>
									<p>{preparedItem.issues.missingApparatLabels.join(', ')}</p>
								</div>
							{/if}
							{#if preparedItem.issues.missingStateTextLabels.length > 0}
								<div>
									<strong class="text-foreground">Missing state text labels:</strong>
									<p>{preparedItem.issues.missingStateTextLabels.join(', ')}</p>
								</div>
							{/if}
							{#if preparedItem.issues.missingNotificationClassLabels.length > 0}
								<div>
									<strong class="text-foreground">Missing notification class labels:</strong>
									<p>{preparedItem.issues.missingNotificationClassLabels.join(', ')}</p>
								</div>
							{/if}
							{#if preparedItem.plannedAlarmDefinitions.length > 0}
								<div>
									<strong class="text-foreground">Planned alarm definition creates:</strong>
									<p>
										{preparedItem.plannedAlarmDefinitions
											.map((entry) => `${entry.bacnetSoftwareId} -> ${entry.name}${entry.alarmTypeCode ? ` [${entry.alarmTypeCode}]` : ''}`)
											.join(' | ')}
									</p>
								</div>
							{/if}
							{#if preparedItem.plannedSoftwareReferenceLinks.length > 0}
								<div>
									<strong class="text-foreground">Planned software reference links:</strong>
									<p>
										{preparedItem.plannedSoftwareReferenceLinks
											.map((entry) => `${entry.fromSoftwareId} -> ${entry.toSoftwareId}`)
											.join(' | ')}
									</p>
								</div>
							{/if}
							{#if preparedItem.issues.missingSoftwareReferences.length > 0}
								<div>
									<strong class="text-foreground">Missing software references:</strong>
									<p>{preparedItem.issues.missingSoftwareReferences.join(', ')}</p>
								</div>
							{/if}
							{#if preparedItem.issues.missingHardwareEntries.length > 0}
								<div>
									<strong class="text-foreground">Invalid or missing hardware rows:</strong>
									<p>{preparedItem.issues.missingHardwareEntries.join(' | ')}</p>
								</div>
							{/if}
							{#if preparedItem.issues.missingSoftwareNumberEntries.length > 0}
								<div>
									<strong class="text-foreground">Invalid software number rows:</strong>
									<p>{preparedItem.issues.missingSoftwareNumberEntries.join(' | ')}</p>
								</div>
							{/if}
						</div>
					</details>
			{/each}
		</div>
	{/if}

	{#if session.objectDataExcel.length === 0}
		<div class="rounded-md border border-dashed p-4 text-sm text-muted-foreground">
			No ObjectData entries were found in the Excel sheet.
		</div>
	{:else}
		<div class="overflow-x-auto rounded-md border">
			<div class="min-w-[840px] text-xs">
				<div class="grid grid-cols-[32px_240px_1fr_96px_88px] border-b bg-muted/30 font-medium text-muted-foreground">
					<div class="px-2 py-1"></div>
					<div class="px-2 py-1">ObjectData ID</div>
					<div class="px-2 py-1">Description</div>
					<div class="px-2 py-1">BACnet</div>
					<div class="px-2 py-1">Optional</div>
				</div>

				{#each session.objectDataExcel as objectData}
					<details class="group border-b last:border-b-0">
						<summary class="list-none cursor-pointer">
							<div
								class="grid grid-cols-[32px_240px_1fr_96px_88px] items-center hover:bg-muted/20"
							>
								<div class="px-2 py-1.5 text-muted-foreground">
									<span class="group-open:hidden">▸</span><span class="hidden group-open:inline">▾</span>
								</div>
								<div class="truncate px-2 py-1.5 font-medium">{objectData.id}</div>
								<div class="truncate px-2 py-1.5 text-muted-foreground">{objectData.description || '-'}</div>
								<div class="px-2 py-1.5 text-muted-foreground">{objectData.bacnet_objects.length}</div>
								<div class="px-2 py-1.5 text-muted-foreground">{objectData.is_optional_anchor ? 'Yes' : 'No'}</div>
							</div>
						</summary>

						<div class="border-t bg-muted/10 px-2 py-2">
							{#if objectData.bacnet_objects.length === 0}
								<p class="px-2 py-1 text-muted-foreground">No BACnet objects for this entry.</p>
							{:else}
								<div class="overflow-x-auto rounded-sm border bg-background">
									<div class="min-w-245 text-[11px]">
										<div class="grid grid-cols-[180px_220px_70px_70px_180px_80px_90px_110px_140px_150px_150px_140px] border-b bg-muted/30 font-medium text-muted-foreground">
											<div class="px-1 py-1">Text fix</div>
											<div class="px-1 py-1">Description</div>
											<div class="px-1 py-1">Visible</div>
											<div class="px-1 py-1">Optional</div>
											<div class="px-1 py-1">Text individual</div>
											<div class="px-1 py-1">Type</div>
											<div class="px-1 py-1">Number</div>
											<div class="px-1 py-1">Hardware</div>
											<div class="px-1 py-1">Software ref</div>
											<div class="px-1 py-1">State text</div>
											<div class="px-1 py-1">Notification class</div>
											<div class="px-1 py-1">Alarm definition</div>
											<div class="px-1 py-1">Apparat</div>
										</div>

										{#each objectData.bacnet_objects as bacnetObject}
											<div
												class={`grid grid-cols-[180px_220px_70px_70px_180px_80px_90px_110px_140px_150px_150px_140px] border-b last:border-b-0 ${duplicateSoftwareIds.has(rowIdentifier(objectData.id, bacnetObject.id)) ? 'bg-destructive/10' : ''}`}
											>
												<div class="truncate px-1 py-1">{bacnetObject.text_fix || '-'}</div>
												<div class="truncate px-1 py-1 text-muted-foreground">{bacnetObject.description || '-'}</div>
												<div class="px-1 py-1">{bacnetObject.gms_visible ? 'Yes' : 'No'}</div>
												<div class="px-1 py-1">{bacnetObject.is_optional ? 'Yes' : 'No'}</div>
												<div class="truncate px-1 py-1">{bacnetObject.text_individual || '-'}</div>
												<div class="px-1 py-1">{bacnetObject.software_type || '-'}</div>
												<div class="px-1 py-1">{bacnetObject.software_number || '-'}</div>
												<div class="px-1 py-1">{bacnetObject.hardware_label || '-'}</div>
												<div class="truncate px-1 py-1">{bacnetObject.software_reference_label || '-'}</div>
												<div class="truncate px-1 py-1">{bacnetObject.state_text_label || '-'}</div>
												<div class="truncate px-1 py-1">{bacnetObject.notification_class_label || '-'}</div>
												<div class="truncate px-1 py-1">{bacnetObject.alarm_definition_label || '-'}</div>
												<div class="truncate px-1 py-1">{bacnetObject.apparat_label || '-'}</div>
											</div>
										{/each}
									</div>
								</div>
							{/if}
						</div>
					</details>
				{/each}
			</div>
		</div>
	{/if}
</div>
