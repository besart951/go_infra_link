<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import { AlertCircle } from '@lucide/svelte';
	import MultiCreateSelectionSection from './multi-create/MultiCreateSelectionSection.svelte';
	import MultiCreateRowsSection from './multi-create/MultiCreateRowsSection.svelte';
	import { addToast } from '$lib/components/toast.svelte';
	import { ApiException } from '$lib/api/client.js';
	import { t as translate } from '$lib/i18n/index.js';

	// Domain imports
	import {
		type FieldDeviceRowData,
		type FieldDeviceRowError,
		type MultiCreateSelection,
		createSelectionKey,
		hasRequiredSelections,
		canFetchAvailableNumbers,
		createNewRow,
		getUsedApparatNumbers,
		validateAllRows,
		loadPersistedState,
		savePersistedState,
		clearPersistedState
	} from '$lib/domain/facility/fieldDeviceMultiCreate.js';

	// Application layer imports
	import { ManageFieldDeviceUseCase } from '$lib/application/useCases/facility/manageFieldDeviceUseCase.js';
	import { ManageObjectDataUseCase } from '$lib/application/useCases/facility/manageObjectDataUseCase.js';
	import { ListSPSControllersUseCase } from '$lib/application/useCases/facility/listSPSControllersUseCase.js';
	import { fieldDeviceRepository } from '$lib/infrastructure/api/fieldDeviceRepository.js';
	import { objectDataRepository } from '$lib/infrastructure/api/objectDataRepository.js';
	import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';

	import type {
		FieldDevice,
		ObjectData,
		SPSControllerSystemType,
		CreateFieldDeviceRequest
	} from '$lib/domain/facility/index.js';
	import type { FieldDevicePreselection as PreselectionType } from '$lib/domain/facility/preselectionFilter.js';

	// Use case instances
	const manageUseCase = new ManageFieldDeviceUseCase(fieldDeviceRepository);
	const manageObjectDataUseCase = new ManageObjectDataUseCase(objectDataRepository);
	const spsUseCase = new ListSPSControllersUseCase(spsControllerRepository);

	// ─────────────────────────────────────────────────────────────────────────────
	// Props
	// ─────────────────────────────────────────────────────────────────────────────
	interface Props {
		projectId?: string;
		onSuccess?: (createdDevices: FieldDevice[]) => void;
		onCancel?: () => void;
	}

	let { projectId, onSuccess, onCancel }: Props = $props();

	// ─────────────────────────────────────────────────────────────────────────────
	// State
	// ─────────────────────────────────────────────────────────────────────────────

	// Selection state
	let selection = $state<MultiCreateSelection>({
		spsControllerSystemTypeId: '',
		objectDataId: '',
		apparatId: '',
		systemPartId: ''
	});

	// For FieldDevicePreselection component
	let preselectionValue = $state<PreselectionType>({
		objectDataId: '',
		apparatId: '',
		systemPartId: ''
	});

	// Row data
	let rows = $state<FieldDeviceRowData[]>([]);

	// Validation errors (reactive map)
	let rowErrors = $state<Map<number, FieldDeviceRowError>>(new Map());

	// Available apparat numbers from backend
	let availableNumbers = $state<number[]>([]);
	let loadingAvailableNumbers = $state(false);
	let selectedObjectData = $state<ObjectData | null>(null);
	let loadingObjectDataPreview = $state(false);
	let objectDataPreviewError = $state('');

	// Form state
	let submitting = $state(false);
	let globalError = $state('');

	// Internal tracking
	let selectionKey = $state('');
	let availableNumbersAbortController: AbortController | null = null;
	let objectDataPreviewAbortController: AbortController | null = null;

	// ─────────────────────────────────────────────────────────────────────────────
	// Derived State
	// ─────────────────────────────────────────────────────────────────────────────

	const showConfiguration = $derived(hasRequiredSelections(selection));
	const showRowsSection = $derived(showConfiguration);
	const canAddRow = $derived(
		showConfiguration && availableNumbers.length > rows.length && !loadingAvailableNumbers
	);
	const hasValidationErrors = $derived(rowErrors.size > 0);

	// ─────────────────────────────────────────────────────────────────────────────
	// Lifecycle
	// ─────────────────────────────────────────────────────────────────────────────

	onMount(() => {
		const persisted = loadPersistedState();
		if (persisted) {
			selection = persisted.selection;
			rows = persisted.rows;
			preselectionValue = {
				objectDataId: persisted.selection.objectDataId,
				apparatId: persisted.selection.apparatId,
				systemPartId: persisted.selection.systemPartId
			};
		}
	});

	onDestroy(() => {
		availableNumbersAbortController?.abort();
		objectDataPreviewAbortController?.abort();
		if (rows.length > 0 || selection.spsControllerSystemTypeId) {
			savePersistedState(selection, rows);
		}
	});

	// ─────────────────────────────────────────────────────────────────────────────
	// Effects
	// ─────────────────────────────────────────────────────────────────────────────

	// Fetch available numbers when selection changes
	$effect(() => {
		const newKey = createSelectionKey(selection);
		const keyChanged = selectionKey !== '' && newKey !== selectionKey;

		// Reset on key change
		if (keyChanged) {
			availableNumbers = [];
			rows = [];
			rowErrors = new Map();
		}

		selectionKey = newKey;

		// Fetch if we have the required selections
		if (canFetchAvailableNumbers(selection)) {
			fetchAvailableNumbers();
		}
	});

	// Re-validate rows when availableNumbers or rows change
	$effect(() => {
		// This effect depends on availableNumbers and rows
		if (availableNumbers.length > 0 && rows.length > 0) {
			revalidateAllRows();
		}
	});

	$effect(() => {
		const objectDataId = selection.objectDataId?.trim() ?? '';

		if (!objectDataId) {
			objectDataPreviewAbortController?.abort();
			selectedObjectData = null;
			loadingObjectDataPreview = false;
			objectDataPreviewError = '';
			return;
		}

		fetchSelectedObjectDataPreview(objectDataId);
	});

	// ─────────────────────────────────────────────────────────────────────────────
	// API Functions
	// ─────────────────────────────────────────────────────────────────────────────

	async function fetchAvailableNumbers() {
		if (!canFetchAvailableNumbers(selection)) {
			availableNumbers = [];
			return;
		}

		// Abort previous request
		availableNumbersAbortController?.abort();

		loadingAvailableNumbers = true;
		const controller = new AbortController();
		availableNumbersAbortController = controller;

		try {
			const response = await manageUseCase.getAvailableApparatNumbers(
				selection.spsControllerSystemTypeId,
				selection.apparatId,
				selection.systemPartId,
				controller.signal
			);

			if (!controller.signal.aborted) {
				availableNumbers = response.available;
				autoAssignApparatNumbers();
			}
		} catch (err) {
			if (err instanceof DOMException && err.name === 'AbortError') return;

			const msg =
				err instanceof ApiException
					? translate('field_device.multi_create.errors.fetch_numbers_status', {
							status: err.status,
							message: err.message
						})
					: translate('field_device.multi_create.errors.fetch_numbers_failed', {
							message: (err as Error)?.message ?? String(err)
						});
			addToast(msg, 'error');
			availableNumbers = [];
		} finally {
			if (availableNumbersAbortController === controller) {
				loadingAvailableNumbers = false;
			}
		}
	}

	async function fetchSelectedObjectDataPreview(objectDataId: string) {
		objectDataPreviewAbortController?.abort();
		const controller = new AbortController();
		objectDataPreviewAbortController = controller;

		loadingObjectDataPreview = true;
		objectDataPreviewError = '';

		try {
			const objectData = await manageObjectDataUseCase.get(objectDataId, controller.signal);
			if (!controller.signal.aborted) {
				selectedObjectData = objectData;
			}
		} catch (err) {
			if (err instanceof DOMException && err.name === 'AbortError') return;
			if (!controller.signal.aborted) {
				selectedObjectData = null;
				objectDataPreviewError = translate('field_device.multi_create.object_data_preview.load_failed');
			}
		} finally {
			if (objectDataPreviewAbortController === controller) {
				loadingObjectDataPreview = false;
			}
		}
	}

	async function handleSubmit() {
		if (rows.length === 0) {
			addToast(translate('field_device.multi_create.errors.no_rows'), 'warning');
			return;
		}

		// Validate all rows (requiring values)
		const errors = validateAllRows(rows, availableNumbers, true);
		rowErrors = errors;

		if (errors.size > 0) {
			addToast(translate('field_device.multi_create.errors.validation'), 'error');
			return;
		}

		submitting = true;
		globalError = '';

		try {
			const fieldDevices: CreateFieldDeviceRequest[] = rows.map((row) => ({
				bmk: row.bmk || undefined,
				description: row.description || undefined,
				text_fix: row.textFix || undefined,
				apparat_nr: row.apparatNr!,
				sps_controller_system_type_id: selection.spsControllerSystemTypeId,
				system_part_id: selection.systemPartId,
				apparat_id: selection.apparatId,
				object_data_id: selection.objectDataId || undefined
			}));

			const response = await manageUseCase.multiCreate({ field_devices: fieldDevices });

			// Map backend errors to rows
			const newErrors = new Map<number, FieldDeviceRowError>();
			response.results.forEach((result) => {
				if (!result.success && result.index < rows.length) {
					newErrors.set(result.index, {
						message: result.error,
						field: (result.error_field as FieldDeviceRowError['field']) || ''
					});
				}
			});
			rowErrors = newErrors;

			if (response.failure_count > 0) {
				addToast(
					translate('field_device.multi_create.toasts.partial_created', {
						success: response.success_count,
						total: response.total_requests,
						failed: response.failure_count
					}),
					'warning'
				);
			} else {
				addToast(
					translate('field_device.multi_create.toasts.created', {
						count: response.success_count
					}),
					'success'
				);
				clearPersistedState();
				rows = [];
				rowErrors = new Map();

				if (onSuccess) {
					const createdDevices = response.results
						.filter((r) => r.success)
						.map((r) => r.field_device)
						.filter((d): d is FieldDevice => Boolean(d));
					onSuccess(createdDevices);
				}
			}
		} catch (err) {
			globalError = (err as Error)?.message || translate('field_device.multi_create.errors.create');
			addToast(
				translate('field_device.multi_create.errors.create_with_message', { message: globalError }),
				'error'
			);
		} finally {
			submitting = false;
		}
	}

	// ─────────────────────────────────────────────────────────────────────────────
	// Row Management
	// ─────────────────────────────────────────────────────────────────────────────

	function addRow() {
		if (!canAddRow) return;

		const usedNumbers = getUsedApparatNumbers(rows);
		const newRow = createNewRow(availableNumbers, usedNumbers);

		if (!newRow) {
			addToast(translate('field_device.multi_create.errors.no_more_numbers'), 'warning');
			return;
		}

		rows = [...rows, newRow];
	}

	function removeRow(index: number) {
		rows = rows.filter((_, i) => i !== index);
		revalidateAllRows();
	}

	function autoAssignApparatNumbers() {
		const used = getUsedApparatNumbers(rows);
		let changed = false;

		rows.forEach((row) => {
			if (row.apparatNr === null) {
				const nextAvailable = availableNumbers.find((nr) => !used.has(nr));
				if (nextAvailable !== undefined) {
					row.apparatNr = nextAvailable;
					used.add(nextAvailable);
					changed = true;
				}
			}
		});

		if (changed) {
			rows = [...rows]; // Trigger reactivity
			revalidateAllRows();
		}
	}

	function revalidateAllRows() {
		rowErrors = validateAllRows(rows, availableNumbers, false);
	}

	// ─────────────────────────────────────────────────────────────────────────────
	// Row Event Handlers
	// ─────────────────────────────────────────────────────────────────────────────

	function handleRowBmkChange(index: number, value: string) {
		rows[index].bmk = value;
		rows = [...rows];
	}

	function handleRowDescriptionChange(index: number, value: string) {
		rows[index].description = value;
		rows = [...rows];
	}

	function handleRowTextFixChange(index: number, value: string) {
		rows[index].textFix = value;
		rows = [...rows];
	}

	function handleRowApparatNrChange(index: number, value: string) {
		const trimmed = value.trim();

		if (!trimmed) {
			rows[index].apparatNr = null;
		} else {
			const num = parseInt(trimmed, 10);
			rows[index].apparatNr = isNaN(num) ? null : num;
		}

		rows = [...rows];
		revalidateAllRows();
	}

	function getPlaceholderForRow(index: number): string {
		const usedNumbers = rows
			.filter((_, i) => i !== index)
			.map((r) => r.apparatNr)
			.filter((n): n is number => n !== null);
		const nextAvailable = availableNumbers.find((nr) => !usedNumbers.includes(nr));
		return nextAvailable !== undefined
			? translate('field_device.multi_create.next_available', { value: nextAvailable })
			: '';
	}

	// ─────────────────────────────────────────────────────────────────────────────
	// Selection Handlers
	// ─────────────────────────────────────────────────────────────────────────────

	function handleSpsSystemTypeChange(value: string) {
		if (value !== selection.spsControllerSystemTypeId) {
			availableNumbersAbortController?.abort();
			objectDataPreviewAbortController?.abort();
			selection = {
				spsControllerSystemTypeId: value,
				objectDataId: '',
				apparatId: '',
				systemPartId: ''
			};
			preselectionValue = { objectDataId: '', apparatId: '', systemPartId: '' };
			rows = [];
			availableNumbers = [];
			selectedObjectData = null;
			loadingObjectDataPreview = false;
			objectDataPreviewError = '';
			rowErrors = new Map();
		}
	}

	function handlePreselectionChange(next: PreselectionType) {
		preselectionValue = next;
		selection = {
			...selection,
			objectDataId: next.objectDataId,
			apparatId: next.apparatId,
			systemPartId: next.systemPartId
		};
	}

	// ─────────────────────────────────────────────────────────────────────────────
	// SPS Controller System Type Fetchers
	// ─────────────────────────────────────────────────────────────────────────────

	async function fetchSpsControllerSystemTypes(search: string): Promise<SPSControllerSystemType[]> {
		const res = await spsUseCase.listSystemTypes({ search, limit: 50 });
		return res.items || [];
	}

	async function fetchSpsControllerSystemTypeById(
		id: string
	): Promise<SPSControllerSystemType | null> {
		try {
			return await spsUseCase.getSystemType(id);
		} catch {
			return null;
		}
	}

	function formatSpsControllerSystemTypeLabel(item: SPSControllerSystemType): string {
		const deviceName = item.sps_controller_name ?? '';
		const number = item.number ?? '';
		const documentName = item.document_name ?? '';
		const sysTypePart = number || documentName ? `${number}_${documentName}` : '';
		return deviceName && sysTypePart
			? `${deviceName}_${sysTypePart}`
			: deviceName || sysTypePart || '';
	}
</script>

<div class="space-y-6">
	<!-- Global Error -->
	{#if globalError}
		<Alert.Root variant="destructive">
			<AlertCircle class="size-4" />
			<Alert.Description>{globalError}</Alert.Description>
		</Alert.Root>
	{/if}

	<MultiCreateSelectionSection
		{projectId}
		{selection}
		{preselectionValue}
		{submitting}
		availableNumbersCount={availableNumbers.length}
		{loadingAvailableNumbers}
		showStatus={canFetchAvailableNumbers(selection)}
		{selectedObjectData}
		{loadingObjectDataPreview}
		{objectDataPreviewError}
		onSpsSystemTypeChange={handleSpsSystemTypeChange}
		onPreselectionChange={handlePreselectionChange}
		{fetchSpsControllerSystemTypes}
		{fetchSpsControllerSystemTypeById}
		{formatSpsControllerSystemTypeLabel}
	/>

	<!-- Field Device Rows -->
	{#if showRowsSection}
		<MultiCreateRowsSection
			{rows}
			{rowErrors}
			availableNumbersCount={availableNumbers.length}
			{loadingAvailableNumbers}
			{canAddRow}
			{hasValidationErrors}
			{submitting}
			{onCancel}
			onAddRow={addRow}
			onSubmit={handleSubmit}
			onRowBmkChange={handleRowBmkChange}
			onRowDescriptionChange={handleRowDescriptionChange}
			onRowTextFixChange={handleRowTextFixChange}
			onRowApparatNrChange={handleRowApparatNrChange}
			onRowRemove={removeRow}
			{getPlaceholderForRow}
		/>
	{/if}
</div>
