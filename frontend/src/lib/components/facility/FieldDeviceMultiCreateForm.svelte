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
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import { Plus, Trash2, AlertCircle } from 'lucide-svelte';
	import {
		getFieldDeviceOptions,
		getAvailableApparatNumbers,
		multiCreateFieldDevices,
		listSPSControllerSystemTypes,
		getSPSControllerSystemType
	} from '$lib/infrastructure/api/facility.adapter.js';
	import { addToast } from '$lib/components/toast.svelte';
	import type {
		FieldDeviceOptions,
		SPSControllerSystemType,
		Apparat,
		SystemPart,
		ObjectData,
		CreateFieldDeviceRequest,
		AvailableApparatNumbersResponse
	} from '$lib/domain/facility/index.js';

	// Props
	interface Props {
		projectId?: string;
		onSuccess?: (createdDevices: FieldDevice[]) => void;
		onCancel?: () => void;
	}

	let { projectId, onSuccess, onCancel }: Props = $props();

	// State for selection flow
	let spsControllerSystemTypeId = $state('');
	let apparatId = $state('');
	let systemPartId = $state('');
	let objectDataId = $state('');

	// State for metadata
	let options = $state<FieldDeviceOptions | null>(null);
	let loadingOptions = $state(true);
	let selectedSpsSystemType = $state<SPSControllerSystemType | null>(null);

	// State for available apparat numbers
	let availableNumbers = $state<number[]>([]);
	let loadingAvailableNumbers = $state(false);

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
	const canAddRow = $derived(
		spsControllerSystemTypeId &&
			apparatId &&
			systemPartId &&
			objectDataId &&
			availableNumbers.length > rows.length
	);

	const filteredApparats = $derived.by((): Apparat[] => {
		if (!options) return [];
		if (!objectDataId) return options.apparats;

		const allowedApparatIds = options.object_data_to_apparat[objectDataId] || [];
		return options.apparats.filter((app) => allowedApparatIds.includes(app.id));
	});

	const filteredSystemParts = $derived.by((): SystemPart[] => {
		if (!options) return [];
		if (!apparatId) return options.system_parts;

		const allowedSystemPartIds = options.apparat_to_system_part[apparatId] || [];
		return options.system_parts.filter((sp) => allowedSystemPartIds.includes(sp.id));
	});

	const filteredObjectDatas = $derived.by((): ObjectData[] => {
		if (!options) return [];
		if (!apparatId) return options.object_datas;

		return options.object_datas.filter((od) => {
			const apparatIds = options.object_data_to_apparat[od.id] || [];
			return apparatIds.includes(apparatId);
		});
	});

	// Load metadata on mount
	onMount(async () => {
		try {
			options = await getFieldDeviceOptions();
			loadPersistedState();
		} catch (err: any) {
			addToast({
				type: 'error',
				message: 'Failed to load field device options',
				description: err.message
			});
		} finally {
			loadingOptions = false;
		}
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

	// Load SPS Controller System Type details when selected
	$effect(() => {
		if (spsControllerSystemTypeId) {
			loadSpsSystemType();
		}
	});

	async function loadSpsSystemType() {
		try {
			selectedSpsSystemType = await getSPSControllerSystemType(spsControllerSystemTypeId);
		} catch (err) {
			console.error('Failed to load SPS system type:', err);
		}
	}

	async function fetchAvailableNumbers() {
		loadingAvailableNumbers = true;
		try {
			const response: AvailableApparatNumbersResponse = await getAvailableApparatNumbers(
				spsControllerSystemTypeId,
				apparatId,
				systemPartId || undefined
			);
			availableNumbers = response.available;

			// Update rows with available numbers
			updateRowsWithAvailableNumbers();
		} catch (err: any) {
			addToast({
				type: 'error',
				message: 'Failed to fetch available apparat numbers',
				description: err.message
			});
			availableNumbers = [];
		} finally {
			loadingAvailableNumbers = false;
		}
	}

	function updateRowsWithAvailableNumbers() {
		// Auto-assign available numbers to rows that don't have a number yet
		rows.forEach((row, index) => {
			if (row.apparatNr === null && availableNumbers[index] !== undefined) {
				row.apparatNr = availableNumbers[index];
			}
		});
	}

	function addRow() {
		if (!canAddRow) return;

		const nextAvailableNr = availableNumbers.find(
			(nr) => !rows.some((row) => row.apparatNr === nr)
		);

		if (nextAvailableNr === undefined) {
			addToast({
				type: 'warning',
				message: 'No more available apparat numbers',
				description: 'All apparat numbers are already assigned.'
			});
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
		const num = parseInt(value);
		if (isNaN(num)) {
			rows[index].apparatNr = null;
			return;
		}

		// Validate range
		if (num < 1 || num > 99) {
			rows[index].error = 'Apparat number must be between 1 and 99';
			rows[index].errorField = 'apparat_nr';
			return;
		}

		// Check if available
		if (!availableNumbers.includes(num)) {
			rows[index].error = 'This apparat number is already assigned';
			rows[index].errorField = 'apparat_nr';
		} else {
			rows[index].error = '';
			rows[index].errorField = '';
		}

		rows[index].apparatNr = num;
	}

	async function handleSubmit() {
		if (rows.length === 0) {
			addToast({
				type: 'warning',
				message: 'No field devices to create',
				description: 'Please add at least one field device row.'
			});
			return;
		}

		// Validate all rows
		let hasErrors = false;
		rows.forEach((row, index) => {
			if (row.apparatNr === null) {
				row.error = 'Apparat number is required';
				row.errorField = 'apparat_nr';
				hasErrors = true;
			}
		});

		if (hasErrors) {
			addToast({
				type: 'error',
				message: 'Validation errors',
				description: 'Please fix the errors in the form.'
			});
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

			const response = await multiCreateFieldDevices({ field_devices: fieldDevices });

			// Update rows with errors from backend
			response.results.forEach((result) => {
				if (!result.success && result.index < rows.length) {
					rows[result.index].error = result.error;
					rows[result.index].errorField = result.error_field;
				}
			});

			if (response.failure_count > 0) {
				addToast({
					type: 'warning',
					message: `Created ${response.success_count} of ${response.total_requests} field devices`,
					description: `${response.failure_count} failed. See errors below.`
				});
			} else {
				addToast({
					type: 'success',
					message: 'Field devices created successfully',
					description: `Created ${response.success_count} field device(s).`
				});

				// Clear state and notify parent
				clearState();
				rows = [];
				if (onSuccess) {
					const createdDevices = response.results
						.filter((r) => r.success)
						.map((r) => r.field_device);
					onSuccess(createdDevices);
				}
			}
		} catch (err: any) {
			globalError = err.message || 'Failed to create field devices';
			addToast({
				type: 'error',
				message: 'Failed to create field devices',
				description: err.message
			});
		} finally {
			submitting = false;
		}
	}

	// State persistence
	const STORAGE_KEY = 'fieldDeviceMultiCreate';

	function saveState() {
		const state = {
			spsControllerSystemTypeId,
			apparatId,
			systemPartId,
			objectDataId,
			rows
		};
		sessionStorage.setItem(STORAGE_KEY, JSON.stringify(state));
	}

	function loadPersistedState() {
		const stored = sessionStorage.getItem(STORAGE_KEY);
		if (stored) {
			try {
				const state = JSON.parse(stored);
				spsControllerSystemTypeId = state.spsControllerSystemTypeId || '';
				apparatId = state.apparatId || '';
				systemPartId = state.systemPartId || '';
				objectDataId = state.objectDataId || '';
				rows = state.rows || [];
			} catch (err) {
				console.error('Failed to load persisted state:', err);
			}
		}
	}

	function clearState() {
		sessionStorage.removeItem(STORAGE_KEY);
	}

	// Handle selection changes - clear dependent fields
	function handleSpsSystemTypeChange(value: string) {
		if (value !== spsControllerSystemTypeId) {
			apparatId = '';
			systemPartId = '';
			objectDataId = '';
			rows = [];
			availableNumbers = [];
		}
		spsControllerSystemTypeId = value;
	}

	function handleObjectDataChange(value: string) {
		const oldValue = objectDataId;
		objectDataId = value;

		// Reset apparat if it's no longer valid
		if (oldValue !== value && apparatId) {
			const isValid = filteredApparats.some((app) => app.id === apparatId);
			if (!isValid) {
				apparatId = '';
				systemPartId = '';
				rows = [];
			}
		}
	}

	function handleApparatChange(value: string) {
		const oldValue = apparatId;
		apparatId = value;

		// Reset system part if it's no longer valid
		if (oldValue !== value && systemPartId) {
			const isValid = filteredSystemParts.some((sp) => sp.id === systemPartId);
			if (!isValid) {
				systemPartId = '';
				rows = [];
			}
		}
	}

	function handleSystemPartChange(value: string) {
		const oldValue = systemPartId;
		systemPartId = value;

		// Clear rows when system part changes
		if (oldValue !== value) {
			rows = [];
		}
	}

	// Format functions
	function formatSpsSystemType(item: SPSControllerSystemType): string {
		return `${item.designation} (${item.system_type_name})`;
	}

	function formatApparat(item: Apparat): string {
		return `${item.name} (${item.short_name})`;
	}

	function formatSystemPart(item: SystemPart): string {
		return item.name;
	}

	function formatObjectData(item: ObjectData): string {
		return item.name;
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
					fetchItems={async (query) => {
						const response = await listSPSControllerSystemTypes({ search: query, limit: 50 });
						return response.items;
					}}
					formatItem={formatSpsSystemType}
					value={spsControllerSystemTypeId}
					onValueChange={handleSpsSystemTypeChange}
					disabled={loadingOptions}
				/>
			</div>

			<!-- Object Data -->
			<div class="space-y-2">
				<Label for="object-data">Object Data *</Label>
				<AsyncCombobox
					id="object-data"
					placeholder="Select object data..."
					fetchItems={async (query) => {
						const filtered = filteredObjectDatas.filter(
							(od) =>
								od.name.toLowerCase().includes(query.toLowerCase()) ||
								od.description?.toLowerCase().includes(query.toLowerCase())
						);
						return Promise.resolve(filtered);
					}}
					formatItem={formatObjectData}
					value={objectDataId}
					onValueChange={handleObjectDataChange}
					disabled={loadingOptions || !spsControllerSystemTypeId}
				/>
			</div>

			<!-- Apparat -->
			<div class="space-y-2">
				<Label for="apparat">Apparat *</Label>
				<AsyncCombobox
					id="apparat"
					placeholder="Select apparat..."
					fetchItems={async (query) => {
						const filtered = filteredApparats.filter(
							(app) =>
								app.name.toLowerCase().includes(query.toLowerCase()) ||
								app.short_name.toLowerCase().includes(query.toLowerCase())
						);
						return Promise.resolve(filtered);
					}}
					formatItem={formatApparat}
					value={apparatId}
					onValueChange={handleApparatChange}
					disabled={loadingOptions || !objectDataId}
				/>
			</div>

			<!-- System Part -->
			<div class="space-y-2">
				<Label for="system-part">System Part *</Label>
				<AsyncCombobox
					id="system-part"
					placeholder="Select system part..."
					fetchItems={async (query) => {
						const filtered = filteredSystemParts.filter((sp) =>
							sp.name.toLowerCase().includes(query.toLowerCase())
						);
						return Promise.resolve(filtered);
					}}
					formatItem={formatSystemPart}
					value={systemPartId}
					onValueChange={handleSystemPartChange}
					disabled={loadingOptions || !apparatId}
				/>
			</div>
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
									<li class="text-destructive">• No available apparat numbers for this configuration</li>
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
							All apparat numbers are assigned. Cannot create more field devices for this configuration.
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
							{submitting ? 'Creating...' : `Create ${rows.length} Field Device${rows.length !== 1 ? 's' : ''}`}
						</Button>
					</div>
				</div>
			{/if}
		</Card.Root>
	{/if}
</div>
