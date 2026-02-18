<script lang="ts">
	/**
	 * Multi-Create Field Device Form (Refactored)
	 *
	 * Architecture: Hexagonal / Clean Architecture
	 * - Domain logic in $lib/domain/facility/fieldDeviceMultiCreate.ts
	 * - UI components are thin wrappers around domain logic
	 * - Clear separation of concerns
	 *
	 * Features:
	 * - Selection flow: SPS Controller System Type → Object Data → Apparat → System Part
	 * - Dynamic row generation
	 * - Real-time validation with swap detection
	 * - State persistence
	 */
	import { onMount, onDestroy } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import FieldDevicePreselection from './FieldDevicePreselection.svelte';
	import FieldDeviceRow from './FieldDeviceRow.svelte';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import { Plus, AlertCircle } from '@lucide/svelte';
	import { addToast } from '$lib/components/toast.svelte';
	import { ApiException } from '$lib/api/client.js';
	import { createTranslator } from '$lib/i18n/translator.js';
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
	import { ListSPSControllersUseCase } from '$lib/application/useCases/facility/listSPSControllersUseCase.js';
	import { fieldDeviceRepository } from '$lib/infrastructure/api/fieldDeviceRepository.js';
	import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';

	import type {
		FieldDevice,
		SPSControllerSystemType,
		CreateFieldDeviceRequest
	} from '$lib/domain/facility/index.js';
	import type { FieldDevicePreselection as PreselectionType } from '$lib/domain/facility/preselectionFilter.js';

	// Use case instances
	const manageUseCase = new ManageFieldDeviceUseCase(fieldDeviceRepository);
	const spsUseCase = new ListSPSControllersUseCase(spsControllerRepository);

	const t = createTranslator();

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

	// Form state
	let submitting = $state(false);
	let globalError = $state('');

	// Internal tracking
	let selectionKey = $state('');
	let abortController: AbortController | null = null;

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
		abortController?.abort();
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

	// ─────────────────────────────────────────────────────────────────────────────
	// API Functions
	// ─────────────────────────────────────────────────────────────────────────────

	async function fetchAvailableNumbers() {
		if (!canFetchAvailableNumbers(selection)) {
			availableNumbers = [];
			return;
		}

		// Abort previous request
		abortController?.abort();

		loadingAvailableNumbers = true;
		const controller = new AbortController();
		abortController = controller;

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
			if (abortController === controller) {
				loadingAvailableNumbers = false;
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
			abortController?.abort();
			selection = {
				spsControllerSystemTypeId: value,
				objectDataId: '',
				apparatId: '',
				systemPartId: ''
			};
			preselectionValue = { objectDataId: '', apparatId: '', systemPartId: '' };
			rows = [];
			availableNumbers = [];
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
		const number = item.number ?? '';
		const documentName = item.document_name ?? '';
		return `${number} - ${documentName}`;
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

	<!-- Selection Flow Card -->
	<Card.Root class="p-6">
		<h3 class="mb-4 text-lg font-semibold">{$t('field_device.multi_create.selection.title')}</h3>
		<p class="mb-4 text-sm text-muted-foreground">
			{$t('field_device.multi_create.selection.description')}
		</p>

		<div class="grid gap-4 md:grid-cols-2">
			<!-- SPS Controller System Type -->
			<div class="space-y-2">
				<Label for="sps-system-type"
					>{$t('field_device.multi_create.selection.sps_system_type')}</Label
				>
				<AsyncCombobox
					id="sps-system-type"
					placeholder={$t('field_device.multi_create.selection.sps_system_type_placeholder')}
					searchPlaceholder={$t('field_device.multi_create.selection.sps_system_type_search')}
					emptyText={$t('field_device.multi_create.selection.sps_system_type_empty')}
					fetcher={fetchSpsControllerSystemTypes}
					fetchById={fetchSpsControllerSystemTypeById}
					labelKey="system_type_name"
					labelFormatter={formatSpsControllerSystemTypeLabel}
					width="w-full"
					value={selection.spsControllerSystemTypeId}
					onValueChange={handleSpsSystemTypeChange}
					clearable
					clearText={$t('field_device.multi_create.selection.sps_system_type_clear')}
					disabled={submitting}
				/>
			</div>
		</div>

		<div class="mt-4">
			<FieldDevicePreselection
				value={preselectionValue}
				onChange={handlePreselectionChange}
				{projectId}
				disabled={!selection.spsControllerSystemTypeId || submitting}
				className="grid grid-cols-1 gap-4 md:grid-cols-3"
			/>
		</div>

		<!-- Configuration Status -->
		{#if canFetchAvailableNumbers(selection)}
			<div class="mt-4">
				<Alert.Root>
					<Alert.Description>
						<div class="text-sm">
							<p class="font-medium">
								{$t('field_device.multi_create.status.title')}
							</p>
							<ul class="mt-2 space-y-1 text-muted-foreground">
								<li>
									• {$t('field_device.multi_create.status.available', {
										count: availableNumbers.length
									})}
									{#if loadingAvailableNumbers}
										{$t('field_device.multi_create.status.loading_suffix')}
									{/if}
								</li>
								{#if availableNumbers.length === 0 && !loadingAvailableNumbers}
									<li class="text-destructive">
										• {$t('field_device.multi_create.status.none_available')}
									</li>
								{/if}
							</ul>
						</div>
					</Alert.Description>
				</Alert.Root>
			</div>
		{/if}
	</Card.Root>

	<!-- Field Device Rows -->
	{#if showRowsSection}
		<Card.Root class="p-6">
			<div class="mb-4 flex items-center justify-between">
				<div>
					<h3 class="text-lg font-semibold">{$t('field_device.multi_create.rows.title')}</h3>
					<p class="text-sm text-muted-foreground">
						{$t('field_device.multi_create.rows.description')}
					</p>
				</div>
				<Button onclick={addRow} disabled={!canAddRow} size="sm">
					<Plus class="mr-2 size-4" />
					{$t('field_device.multi_create.rows.add')}
				</Button>
			</div>

			<!-- Empty State -->
			{#if rows.length === 0}
				<Alert.Root>
					<AlertCircle class="size-4" />
					<Alert.Description>
						{#if availableNumbers.length === 0 && !loadingAvailableNumbers}
							{$t('field_device.multi_create.rows.none_available')}
						{:else if loadingAvailableNumbers}
							{$t('field_device.multi_create.rows.loading_numbers')}
						{:else}
							{$t('field_device.multi_create.rows.empty_prompt')}
						{/if}
					</Alert.Description>
				</Alert.Root>
			{/if}

			<!-- Rows -->
			{#if rows.length > 0}
				<div class="space-y-4">
					{#each rows as row, index (row.id)}
						<FieldDeviceRow
							{index}
							{row}
							error={rowErrors.get(index) ?? null}
							placeholder={getPlaceholderForRow(index)}
							disabled={submitting}
							onBmkChange={(v) => handleRowBmkChange(index, v)}
							onDescriptionChange={(v) => handleRowDescriptionChange(index, v)}
							onApparatNrChange={(v) => handleRowApparatNrChange(index, v)}
							onRemove={() => removeRow(index)}
						/>
					{/each}
				</div>

				<Separator class="my-4" />

				<!-- Summary & Actions -->
				<div class="flex items-center justify-between">
					<p class="text-sm text-muted-foreground">
						{$t('field_device.multi_create.rows.summary', { count: rows.length })}
						{#if hasValidationErrors}
							<span class="text-destructive">
								{$t('field_device.multi_create.rows.errors', { count: rowErrors.size })}
							</span>
						{/if}
					</p>
					<div class="flex gap-2">
						{#if onCancel}
							<Button variant="outline" onclick={onCancel} disabled={submitting}>
								{$t('common.cancel')}
							</Button>
						{/if}
						<Button
							onclick={handleSubmit}
							disabled={submitting || rows.length === 0 || hasValidationErrors}
						>
							{submitting
								? $t('field_device.multi_create.actions.creating')
								: $t('field_device.multi_create.actions.create', { count: rows.length })}
						</Button>
					</div>
				</div>
			{/if}
		</Card.Root>
	{/if}
</div>
