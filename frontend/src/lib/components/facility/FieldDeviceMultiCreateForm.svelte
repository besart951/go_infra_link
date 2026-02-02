<script lang="ts">
	/**
	 * Multi-Create Field Device Form
	 * Allows creating multiple field devices with:
	 * - Selection flow: SPS Controller System Type → Object Data → Apparat → System Part
	 * - Dynamic row generation with + button
	 * - Auto-calculated apparat_nr with placeholder
	 * - Field-level validation display
	 * - State persistence on navigation
	 */
	import { onMount, onDestroy } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import FieldDevicePreselection from '$lib/components/facility/FieldDevicePreselection.svelte';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import { Plus, Trash2, AlertCircle } from '@lucide/svelte';
	import { FieldDeviceMultiCreateUseCase } from '$lib/application/useCases/facility/fieldDeviceMultiCreateUseCase.js';
	import { facilityFieldDeviceMultiCreateRepository } from '$lib/infrastructure/api/facilityFieldDeviceMultiCreateRepository.js';
	import { addToast } from '$lib/components/toast.svelte';
	import { ApiException } from '$lib/api/client.js';
	import type {
		FieldDevice,
		SPSControllerSystemType,
		CreateFieldDeviceRequest,
		AvailableApparatNumbersResponse
	} from '$lib/domain/facility/index.js';
	import type { FieldDevicePreselection as Preselection } from '$lib/domain/facility/preselectionFilter.js';

	const useCase = new FieldDeviceMultiCreateUseCase(facilityFieldDeviceMultiCreateRepository);

	// Props
	interface Props {
		projectId?: string;
		onSuccess?: (createdDevices: FieldDevice[]) => void;
		onCancel?: () => void;
	}

	let { projectId, onSuccess, onCancel }: Props = $props();

	// State for selection flow
	let spsControllerSystemTypeId = $state('');
	let preselection = $state<Preselection>({ objectDataId: '', apparatId: '', systemPartId: '' });
	let apparatId = $state('');
	let systemPartId = $state('');
	let objectDataId = $state('');
	let selectionKey = $state('');
	// State for available apparat numbers
	let availableNumbers = $state<number[]>([]);
	let loadingAvailableNumbers = $state(false);
	let availableNumbersAbortController: AbortController | null = null;

	// State for field device rows
	interface FieldDeviceRow {
		id: string; // temporary ID for UI
		bmk: string;
		description: string;
		apparatNr: number | null;
		error: string;
		errorField: string;
	}

	let rows = $state<FieldDeviceRow[]>([]);

	// Form state
	let submitting = $state(false);
	let globalError = $state('');

	// Derived states
	const hasRequiredSelection = $derived(
		Boolean(spsControllerSystemTypeId && objectDataId && apparatId && systemPartId)
	);

	const canAddRow = $derived(hasRequiredSelection && availableNumbers.length > rows.length);

	onMount(() => {
		loadPersistedState();
		preselection = {
			objectDataId,
			apparatId,
			systemPartId
		};
	});

	// Save state on unmount
	onDestroy(() => {
		if (rows.length > 0 || spsControllerSystemTypeId) {
			saveState();
		}
	});

	// Watch for changes in selection and fetch available numbers
	$effect(() => {
		if (spsControllerSystemTypeId && apparatId && systemPartId) {
			fetchAvailableNumbers();
		}
	});

	$effect(() => {
		const nextKey = `${spsControllerSystemTypeId}|${objectDataId}|${apparatId}|${systemPartId}`;
		if (!selectionKey) {
			selectionKey = nextKey;
			return;
		}
		if (nextKey !== selectionKey) {
			selectionKey = nextKey;
			availableNumbersAbortController?.abort();
			availableNumbers = [];
			rows = [];
		}
	});

	async function fetchAvailableNumbers() {
		if (!spsControllerSystemTypeId || !apparatId || !systemPartId) {
			availableNumbers = [];
			loadingAvailableNumbers = false;
			return;
		}

		loadingAvailableNumbers = true;
		availableNumbersAbortController?.abort();
		availableNumbersAbortController = new AbortController();
		const signal = availableNumbersAbortController.signal;
		try {
			const response: AvailableApparatNumbersResponse = await useCase.getAvailableApparatNumbers(
				spsControllerSystemTypeId,
				apparatId,
				systemPartId || undefined,
				signal
			);
			availableNumbers = response.available;

			// Update rows with available numbers
			updateRowsWithAvailableNumbers();
		} catch (err: any) {
			if (err instanceof DOMException && err.name === 'AbortError') {
				return;
			}
			const msg =
				err instanceof ApiException
					? `Failed to fetch available apparat numbers (${err.status} ${err.error}): ${err.message}`
					: `Failed to fetch available apparat numbers: ${err?.message ?? String(err)}`;
			addToast(msg, 'error');
			availableNumbers = [];
		} finally {
			loadingAvailableNumbers = false;
		}
	}

	function updateRowsWithAvailableNumbers() {
		// Auto-assign available numbers to rows that don't have a number yet
		const used = new Set<number>();
		rows.forEach((row) => {
			if (typeof row.apparatNr === 'number') used.add(row.apparatNr);
		});

		rows.forEach((row, index) => {
			if (row.apparatNr !== null) return;
			const nextAvailable = availableNumbers.find((nr) => !used.has(nr));
			if (nextAvailable === undefined) return;
			row.apparatNr = nextAvailable;
			used.add(nextAvailable);
			validateRow(index, { requireApparatNr: false });
		});

		// Re-validate rows when the available set changes
		rows.forEach((_, index) => validateRow(index, { requireApparatNr: false }));
	}

	function validateRow(index: number, opts: { requireApparatNr: boolean }) {
		const row = rows[index];
		if (!row) return false;

		row.error = '';
		row.errorField = '';

		if (row.apparatNr === null) {
			if (opts.requireApparatNr) {
				row.error = 'Apparat number is required';
				row.errorField = 'apparat_nr';
				return false;
			}
			return true;
		}

		if (row.apparatNr < 1 || row.apparatNr > 99) {
			row.error = 'Apparat number must be between 1 and 99';
			row.errorField = 'apparat_nr';
			return false;
		}

		if (!availableNumbers.includes(row.apparatNr)) {
			row.error = 'This apparat number is not available';
			row.errorField = 'apparat_nr';
			return false;
		}

		const duplicateIndex = rows.findIndex(
			(r, i) => i !== index && r.apparatNr !== null && r.apparatNr === row.apparatNr
		);
		if (duplicateIndex !== -1) {
			row.error = `Duplicate apparat number (also used in row #${duplicateIndex + 1})`;
			row.errorField = 'apparat_nr';
			return false;
		}

		return true;
	}

	function addRow() {
		if (!canAddRow) return;

		const nextAvailableNr = availableNumbers.find(
			(nr) => !rows.some((row) => row.apparatNr === nr)
		);

		if (nextAvailableNr === undefined) {
			addToast('No more available apparat numbers (all are assigned).', 'warning');
			return;
		}

		rows.push({
			id: crypto.randomUUID(),
			bmk: '',
			description: '',
			apparatNr: nextAvailableNr,
			error: '',
			errorField: ''
		});
	}

	function removeRow(index: number) {
		rows.splice(index, 1);
	}

	function getPlaceholderForRow(index: number): string {
		const usedNumbers = rows
			.filter((r, i) => i !== index && r.apparatNr !== null)
			.map((r) => r.apparatNr);
		const nextAvailable = availableNumbers.find((nr) => !usedNumbers.includes(nr));
		return nextAvailable !== undefined ? `Next available: ${nextAvailable}` : '';
	}

	function handleApparatNrChange(index: number, value: string) {
		const trimmed = value.trim();
		if (!trimmed) {
			rows[index].apparatNr = null;
			rows[index].error = '';
			rows[index].errorField = '';
			return;
		}

		const num = Number.parseInt(trimmed, 10);
		if (Number.isNaN(num)) {
			rows[index].apparatNr = null;
			rows[index].error = 'Please enter a valid number';
			rows[index].errorField = 'apparat_nr';
			return;
		}

		rows[index].apparatNr = num;
		validateRow(index, { requireApparatNr: false });
	}

	async function handleSubmit() {
		if (rows.length === 0) {
			addToast('Please add at least one field device row.', 'warning');
			return;
		}

		// Validate all rows
		let hasErrors = false;
		rows.forEach((_, index) => {
			const ok = validateRow(index, { requireApparatNr: true });
			if (!ok) hasErrors = true;
		});

		if (hasErrors) {
			addToast('Please fix the validation errors in the form.', 'error');
			return;
		}

		submitting = true;
		globalError = '';

		try {
			const fieldDevices: CreateFieldDeviceRequest[] = rows.map((row) => ({
				bmk: row.bmk || undefined,
				description: row.description || undefined,
				apparat_nr: row.apparatNr!,
				sps_controller_system_type_id: spsControllerSystemTypeId,
				system_part_id: systemPartId,
				apparat_id: apparatId,
				object_data_id: objectDataId || undefined
			}));

			const response = await useCase.multiCreate(fieldDevices);

			// Update rows with errors from backend
			response.results.forEach((result) => {
				if (!result.success && result.index < rows.length) {
					rows[result.index].error = result.error;
					rows[result.index].errorField = result.error_field;
				}
			});

			if (response.failure_count > 0) {
				addToast(
					`Created ${response.success_count} of ${response.total_requests} field devices (${response.failure_count} failed).`,
					'warning'
				);
			} else {
				addToast(`Created ${response.success_count} field device(s).`, 'success');

				// Clear state and notify parent
				clearState();
				rows = [];
				if (onSuccess) {
					const createdDevices = response.results
						.filter((r) => r.success)
						.map((r) => r.field_device)
						.filter((d): d is FieldDevice => Boolean(d));
					onSuccess(createdDevices);
				}
			}
		} catch (err: any) {
			globalError = err.message || 'Failed to create field devices';
			addToast(`Failed to create field devices: ${err?.message ?? String(err)}`, 'error');
		} finally {
			submitting = false;
		}
	}

	// State persistence
	const STORAGE_KEY = 'fieldDeviceMultiCreate';

	function saveState() {
		if (typeof sessionStorage === 'undefined') return;
		const state = {
			spsControllerSystemTypeId,
			apparatId,
			systemPartId,
			objectDataId,
			rows
		};
		try {
			sessionStorage.setItem(STORAGE_KEY, JSON.stringify(state));
		} catch (err) {
			console.error('Failed to persist state:', err);
		}
	}

	function loadPersistedState() {
		if (typeof sessionStorage === 'undefined') return;
		const stored = sessionStorage.getItem(STORAGE_KEY);
		if (stored) {
			try {
				const state = JSON.parse(stored);
				spsControllerSystemTypeId = state.spsControllerSystemTypeId || '';
				apparatId = state.apparatId || '';
				systemPartId = state.systemPartId || '';
				objectDataId = state.objectDataId || '';
				rows = Array.isArray(state.rows)
					? state.rows.map((r: any) => ({
							id: typeof r?.id === 'string' ? r.id : crypto.randomUUID(),
							bmk: typeof r?.bmk === 'string' ? r.bmk : '',
							description: typeof r?.description === 'string' ? r.description : '',
							apparatNr: typeof r?.apparatNr === 'number' ? r.apparatNr : null,
							error: typeof r?.error === 'string' ? r.error : '',
							errorField: typeof r?.errorField === 'string' ? r.errorField : ''
						}))
					: [];
			} catch (err) {
				console.error('Failed to load persisted state:', err);
			}
		}
	}

	function clearState() {
		if (typeof sessionStorage === 'undefined') return;
		sessionStorage.removeItem(STORAGE_KEY);
	}

	// Handle selection changes - clear dependent fields
	function handleSpsSystemTypeChange(value: string) {
		if (value !== spsControllerSystemTypeId) {
			availableNumbersAbortController?.abort();
			preselection = { objectDataId: '', apparatId: '', systemPartId: '' };
			apparatId = '';
			systemPartId = '';
			objectDataId = '';
			rows = [];
			availableNumbers = [];
		}
		spsControllerSystemTypeId = value;
	}

	function handlePreselectionChange(next: Preselection) {
		preselection = next;
		objectDataId = next.objectDataId;
		apparatId = next.apparatId;
		systemPartId = next.systemPartId;
	}

	async function fetchSpsControllerSystemTypes(search: string): Promise<SPSControllerSystemType[]> {
		return useCase.searchSpsControllerSystemTypes({ search, limit: 50 });
	}

	async function fetchSpsControllerSystemTypeById(
		id: string
	): Promise<SPSControllerSystemType | null> {
		try {
			return await useCase.getSpsControllerSystemType(id);
		} catch {
			return null;
		}
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
		<h3 class="mb-4 text-lg font-semibold">Selection Flow</h3>
		<p class="mb-4 text-sm text-muted-foreground">
			Select the configuration parameters for the field devices you want to create.
		</p>

		<div class="grid gap-4 md:grid-cols-2">
			<!-- SPS Controller System Type -->
			<div class="space-y-2">
				<Label for="sps-system-type">SPS Controller System Type *</Label>
				<AsyncCombobox
					id="sps-system-type"
					placeholder="Select SPS controller system type..."
					searchPlaceholder="Search SPS system types..."
					emptyText="No SPS system types found."
					fetcher={fetchSpsControllerSystemTypes}
					fetchById={fetchSpsControllerSystemTypeById}
					labelKey="system_type_name"
					width="w-full"
					value={spsControllerSystemTypeId}
					onValueChange={handleSpsSystemTypeChange}
					clearable
					clearText="Clear SPS controller system type"
					disabled={submitting}
				/>
			</div>
		</div>

		<div class="mt-4">
			<FieldDevicePreselection
				value={preselection}
				onChange={handlePreselectionChange}
				{projectId}
				disabled={!spsControllerSystemTypeId || submitting}
				className="grid grid-cols-1 gap-4 md:grid-cols-3"
			/>
		</div>

		{#if spsControllerSystemTypeId && apparatId && systemPartId && objectDataId}
			<div class="mt-4">
				<Alert.Root>
					<Alert.Description>
						<div class="text-sm">
							<p class="font-medium">Configuration selected:</p>
							<ul class="mt-2 space-y-1 text-muted-foreground">
								<li>
									• Available apparat numbers: {availableNumbers.length}
									{loadingAvailableNumbers ? '(loading...)' : ''}
								</li>
								{#if availableNumbers.length === 0 && !loadingAvailableNumbers}
									<li class="text-destructive">
										• No available apparat numbers for this configuration
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
	{#if spsControllerSystemTypeId && apparatId && systemPartId && objectDataId}
		<Card.Root class="p-6">
			<div class="mb-4 flex items-center justify-between">
				<div>
					<h3 class="text-lg font-semibold">Field Devices</h3>
					<p class="text-sm text-muted-foreground">
						Add field devices to create. Each will use the configuration selected above.
					</p>
				</div>
				<Button onclick={addRow} disabled={!canAddRow} size="sm">
					<Plus class="mr-2 size-4" />
					Add Field Device
				</Button>
			</div>

			{#if !canAddRow && rows.length === 0}
				<Alert.Root>
					<AlertCircle class="size-4" />
					<Alert.Description>
						{#if availableNumbers.length === 0 && !loadingAvailableNumbers}
							All apparat numbers are assigned. Cannot create more field devices for this
							configuration.
						{:else if loadingAvailableNumbers}
							Loading available apparat numbers...
						{:else}
							Click "Add Field Device" to create field device rows.
						{/if}
					</Alert.Description>
				</Alert.Root>
			{/if}

			{#if rows.length > 0}
				<div class="space-y-4">
					{#each rows as row, index (row.id)}
						<div class="rounded-lg border p-4">
							<div class="mb-3 flex items-center justify-between">
								<h4 class="font-medium">Field Device #{index + 1}</h4>
								<Button
									variant="ghost"
									size="sm"
									onclick={() => removeRow(index)}
									disabled={submitting}
								>
									<Trash2 class="size-4 text-destructive" />
								</Button>
							</div>

							<div class="grid gap-4 md:grid-cols-3">
								<!-- BMK -->
								<div class="space-y-2">
									<Label for={`bmk-${index}`}>BMK</Label>
									<Input
										id={`bmk-${index}`}
										bind:value={row.bmk}
										placeholder="BMK identifier (optional)"
										maxlength={10}
										disabled={submitting}
									/>
								</div>

								<!-- Description -->
								<div class="space-y-2">
									<Label for={`description-${index}`}>Description</Label>
									<Input
										id={`description-${index}`}
										bind:value={row.description}
										placeholder="Description (optional)"
										maxlength={250}
										disabled={submitting}
									/>
								</div>

								<!-- Apparat Nr -->
								<div class="space-y-2">
									<Label for={`apparat-nr-${index}`}>Apparat Nr *</Label>
									<Input
										id={`apparat-nr-${index}`}
										type="number"
										value={row.apparatNr?.toString() ?? ''}
										oninput={(e) =>
											handleApparatNrChange(index, (e.target as HTMLInputElement).value)}
										placeholder={getPlaceholderForRow(index)}
										min={1}
										max={99}
										disabled={submitting}
										class={row.errorField === 'apparat_nr' ? 'border-destructive' : ''}
									/>
									{#if row.errorField === 'apparat_nr' && row.error}
										<p class="text-sm text-destructive">{row.error}</p>
									{/if}
								</div>
							</div>

							<!-- Row Error -->
							{#if row.error && row.errorField !== 'apparat_nr'}
								<Alert.Root variant="destructive" class="mt-3">
									<AlertCircle class="size-4" />
									<Alert.Description>{row.error}</Alert.Description>
								</Alert.Root>
							{/if}
						</div>
					{/each}
				</div>

				<Separator class="my-4" />

				<!-- Summary -->
				<div class="flex items-center justify-between">
					<p class="text-sm text-muted-foreground">
						{rows.length} field device(s) ready to create
					</p>
					<div class="flex gap-2">
						{#if onCancel}
							<Button variant="outline" onclick={onCancel} disabled={submitting}>Cancel</Button>
						{/if}
						<Button onclick={handleSubmit} disabled={submitting || rows.length === 0}>
							{submitting
								? 'Creating...'
								: `Create ${rows.length} Field Device${rows.length !== 1 ? 's' : ''}`}
						</Button>
					</div>
				</div>
			{/if}
		</Card.Root>
	{/if}
</div>
