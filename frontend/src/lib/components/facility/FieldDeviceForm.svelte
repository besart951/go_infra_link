<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import {
		createFieldDevice,
		updateFieldDevice,
		getFieldDeviceOptions,
		getFieldDeviceOptionsForProject,
		listSPSControllerSystemTypes,
		getSPSControllerSystemType,
		getObjectDataBacnetObjects
	} from '$lib/infrastructure/api/facility.adapter.js';
	import { getErrorMessage, getFieldError, getFieldErrors } from '$lib/api/client.js';
	import type {
		FieldDevice,
		FieldDeviceOptions,
		SPSControllerSystemType,
		BacnetObject,
		Apparat,
		SystemPart,
		ObjectData
	} from '$lib/domain/facility/index.js';
	import { createEventDispatcher } from 'svelte';

	export let initialData: FieldDevice | undefined = undefined;
	export let projectId: string | undefined = undefined;
	export let onSuccess: ((device: FieldDevice) => void) | undefined = undefined;
	export let onCancel: (() => void) | undefined = undefined;

	let bmk = initialData?.bmk ?? '';
	let description = initialData?.description ?? '';
	let apparat_nr = initialData?.apparat_nr ?? '';
	let sps_controller_system_type_id = initialData?.sps_controller_system_type_id ?? '';
	let system_part_id = initialData?.system_part_id ?? '';
	let apparat_id = initialData?.apparat_id ?? '';
	let object_data_id = '';

	// Search queries for client-side filtering
	let objectDataSearch = '';
	let apparatSearch = '';
	let systemPartSearch = '';

	let bacnetObjects: BacnetObject[] = [];
	let loadingBacnet = false;

	let loading = false;
	let error = '';
	let fieldErrors: Record<string, string> = {};

	// Single-Fetch Metadata Strategy state
	let options: FieldDeviceOptions | null = null;
	let loadingOptions = true;

	$: if (initialData) {
		bmk = initialData.bmk ?? '';
		description = initialData.description ?? '';
		apparat_nr = initialData.apparat_nr ?? '';
		sps_controller_system_type_id = initialData.sps_controller_system_type_id ?? '';
		system_part_id = initialData.system_part_id ?? '';
		apparat_id = initialData.apparat_id ?? '';
	}

	// ============================================================================
	// REACTIVE FILTERING (Single-Fetch Strategy)
	// ============================================================================

	/**
	 * Filter apparats based on selected ObjectData
	 * If an ObjectData is selected, show only apparats that belong to it
	 */
	$: filteredApparats = filterApparats(
		options?.apparats || [],
		object_data_id,
		options?.object_data_to_apparat || {},
		apparatSearch
	);

	/**
	 * Filter system parts based on selected Apparat
	 * If an Apparat is selected, show only system parts that belong to it
	 */
	$: filteredSystemParts = filterSystemParts(
		options?.system_parts || [],
		apparat_id,
		options?.apparat_to_system_part || {},
		systemPartSearch
	);

	/**
	 * Filter object datas based on selected Apparat
	 * If an Apparat is selected, show only object datas that support it
	 */
	$: filteredObjectDatas = filterObjectDatas(
		options?.object_datas || [],
		apparat_id,
		options?.object_data_to_apparat || {},
		objectDataSearch
	);

	function filterApparats(
		apparats: Apparat[],
		objectDataId: string,
		objectDataToApparat: Record<string, string[]>,
		searchQuery: string
	): Apparat[] {
		let result = apparats;

		// Filter by ObjectData relationship
		if (objectDataId) {
			const allowedApparatIds = objectDataToApparat[objectDataId] || [];
			result = result.filter((app) => allowedApparatIds.includes(app.id));
		}

		// Filter by search query (client-side)
		if (searchQuery.trim()) {
			const query = searchQuery.toLowerCase();
			result = result.filter(
				(app) =>
					app.short_name.toLowerCase().includes(query) ||
					app.name.toLowerCase().includes(query) ||
					(app.description && app.description.toLowerCase().includes(query))
			);
		}

		return result;
	}

	function filterSystemParts(
		systemParts: SystemPart[],
		apparatId: string,
		apparatToSystemPart: Record<string, string[]>,
		searchQuery: string
	): SystemPart[] {
		let result = systemParts;

		// Filter by Apparat relationship
		if (apparatId) {
			const allowedSystemPartIds = apparatToSystemPart[apparatId] || [];
			result = result.filter((sp) => allowedSystemPartIds.includes(sp.id));
		}

		// Filter by search query (client-side)
		if (searchQuery.trim()) {
			const query = searchQuery.toLowerCase();
			result = result.filter(
				(sp) =>
					sp.short_name.toLowerCase().includes(query) ||
					sp.name.toLowerCase().includes(query) ||
					(sp.description && sp.description.toLowerCase().includes(query))
			);
		}

		return result;
	}

	function filterObjectDatas(
		objectDatas: ObjectData[],
		apparatId: string,
		objectDataToApparat: Record<string, string[]>,
		searchQuery: string
	): ObjectData[] {
		let result = objectDatas;

		// Filter by Apparat relationship
		if (apparatId) {
			result = result.filter((od) => {
				const apparatIds = objectDataToApparat[od.id] || [];
				return apparatIds.includes(apparatId);
			});
		}

		// Filter by search query (client-side)
		if (searchQuery.trim()) {
			const query = searchQuery.toLowerCase();
			result = result.filter(
				(od) =>
					od.description.toLowerCase().includes(query) || od.version.toLowerCase().includes(query)
			);
		}

		return result;
	}

	// Handler for ObjectData selection change
	function handleObjectDataChange(newObjectDataId: string) {
		object_data_id = newObjectDataId;

		// Reset Apparat if it's no longer valid
		if (apparat_id) {
			const filteredApps = filterApparats(
				options?.apparats || [],
				newObjectDataId,
				options?.object_data_to_apparat || {},
				''
			);
			const isApparatValid = filteredApps.some((app) => app.id === apparat_id);
			if (!isApparatValid) {
				apparat_id = '';
				system_part_id = '';
			}
		}
	}

	// Handler for Apparat selection change
	function handleApparatChange(newApparatId: string) {
		apparat_id = newApparatId;

		// Reset SystemPart if it's no longer valid
		if (system_part_id) {
			const filteredSps = filterSystemParts(
				options?.system_parts || [],
				newApparatId,
				options?.apparat_to_system_part || {},
				''
			);
			const isSystemPartValid = filteredSps.some((sp) => sp.id === system_part_id);
			if (!isSystemPartValid) {
				system_part_id = '';
			}
		}
	}

	// Handler for SystemPart selection change
	function handleSystemPartChange(newSystemPartId: string) {
		system_part_id = newSystemPartId;
	}

	// ============================================================================
	// LIFECYCLE & DATA LOADING
	// ============================================================================

	onMount(async () => {
		try {
			// Load options based on whether we're in a project context or not
			if (projectId) {
				options = await getFieldDeviceOptionsForProject(projectId);
			} else {
				options = await getFieldDeviceOptions();
			}
			loadingOptions = false;
		} catch (e) {
			console.error('Failed to load field device options:', e);
			error = 'Failed to load options. Please try again.';
			loadingOptions = false;
		}
	});

	async function fetchSPSControllerSystemTypes(search: string): Promise<SPSControllerSystemType[]> {
		try {
			const res = await listSPSControllerSystemTypes({ page: 1, limit: 50, search });
			return res.items;
		} catch (e) {
			console.error(e);
			return [];
		}
	}

	async function loadBacnetObjects(objectDataId: string) {
		loadingBacnet = true;
		try {
			bacnetObjects = await getObjectDataBacnetObjects(objectDataId);
		} catch (e) {
			console.error(e);
			bacnetObjects = [];
		} finally {
			loadingBacnet = false;
		}
	}

	$: if (object_data_id) {
		loadBacnetObjects(object_data_id);
	} else {
		bacnetObjects = [];
	}

	// ============================================================================
	// FORM SUBMISSION
	// ============================================================================

	const dispatch = createEventDispatcher();

	const fieldError = (name: string) => getFieldError(fieldErrors, name, ['fielddevice']);

	async function handleSubmit() {
		loading = true;
		error = '';
		fieldErrors = {};

		if (!sps_controller_system_type_id) {
			error = 'Please select an SPS controller system type';
			loading = false;
			return;
		}

		if (!system_part_id) {
			error = 'Please select a system part';
			loading = false;
			return;
		}

		if (!apparat_id) {
			error = 'Please select an apparat';
			loading = false;
			return;
		}

		const apparatNumber = Number(apparat_nr);
		if (!Number.isFinite(apparatNumber)) {
			error = 'Please provide a valid apparat number';
			loading = false;
			return;
		}

		try {
			if (initialData) {
				const res = await updateFieldDevice(initialData.id, {
					bmk: bmk || undefined,
					description: description || undefined,
					apparat_nr: apparatNumber,
					sps_controller_system_type_id,
					system_part_id,
					apparat_id,
					object_data_id: object_data_id || undefined
				});
				if (onSuccess) onSuccess(res);
			} else {
				const res = await createFieldDevice({
					bmk: bmk || undefined,
					description: description || undefined,
					apparat_nr: apparatNumber,
					sps_controller_system_type_id,
					system_part_id,
					apparat_id,
					object_data_id: object_data_id || undefined
				});
				if (onSuccess) onSuccess(res);
			}
		} catch (e) {
			console.error(e);
			fieldErrors = getFieldErrors(e);
			error = Object.keys(fieldErrors).length ? '' : getErrorMessage(e);
		} finally {
			loading = false;
		}
	}
</script>

<form
	onsubmit={(e) => {
		e.preventDefault();
		handleSubmit();
	}}
	class="space-y-4 rounded-md border bg-muted/20 p-4"
>
	<div class="mb-4 flex items-center justify-between">
		<h3 class="text-lg font-medium">
			{initialData ? 'Edit Field Device' : 'New Field Device'}
		</h3>
	</div>

	{#if loadingOptions}
		<div class="flex items-center justify-center py-8">
			<p class="text-muted-foreground">Loading options...</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
			<div class="space-y-2">
				<Label for="field_device_bmk">BMK</Label>
				<Input id="field_device_bmk" bind:value={bmk} maxlength={255} />
				{#if fieldError('bmk')}
					<p class="text-sm text-red-500">{fieldError('bmk')}</p>
				{/if}
			</div>
			<div class="space-y-2">
				<Label for="field_device_desc">Description</Label>
				<Input id="field_device_desc" bind:value={description} maxlength={255} />
				{#if fieldError('description')}
					<p class="text-sm text-red-500">{fieldError('description')}</p>
				{/if}
			</div>
			<div class="space-y-2">
				<Label for="field_device_apparat_nr">Apparat Nr *</Label>
				<Input id="field_device_apparat_nr" type="number" bind:value={apparat_nr} required />
				{#if fieldError('apparat_nr')}
					<p class="text-sm text-red-500">{fieldError('apparat_nr')}</p>
				{/if}
			</div>
			<div class="space-y-2">
				<Label for="field_device_sps_type">SPS Controller System Type *</Label>
				<AsyncCombobox
					id="field_device_sps_type"
					bind:value={sps_controller_system_type_id}
					fetcher={fetchSPSControllerSystemTypes}
					fetchById={getSPSControllerSystemType}
					labelKey="system_type_name"
					width="w-full"
					placeholder="Select SPS controller system type"
					searchPlaceholder="Search SPS types..."
				/>
				{#if fieldError('sps_controller_system_type_id')}
					<p class="text-sm text-red-500">{fieldError('sps_controller_system_type_id')}</p>
				{/if}
			</div>

			<!-- Object Data Selection (with reactive filtering) -->
			<div class="space-y-2">
				<Label for="field_device_object_data">Object Data (optional)</Label>
				<Input
					type="text"
					placeholder="Search object data..."
					bind:value={objectDataSearch}
					class="mb-1"
				/>
				<select
					id="field_device_object_data"
					value={object_data_id}
					onchange={(e) => handleObjectDataChange(e.currentTarget.value)}
					class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-ring focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
				>
					<option value="">None</option>
					{#each filteredObjectDatas as od}
						<option value={od.id}>
							{od.description} (v{od.version})
						</option>
					{/each}
				</select>
				<p class="text-xs text-muted-foreground">
					{filteredObjectDatas.length} option{filteredObjectDatas.length !== 1 ? 's' : ''} available
				</p>
				{#if fieldError('object_data_id')}
					<p class="text-sm text-red-500">{fieldError('object_data_id')}</p>
				{/if}
			</div>

			<!-- Apparat Selection (with reactive filtering) -->
			<div class="space-y-2">
				<Label for="field_device_apparat">Apparat *</Label>
				<Input
					type="text"
					placeholder="Search apparat..."
					bind:value={apparatSearch}
					class="mb-1"
				/>
				<select
					id="field_device_apparat"
					value={apparat_id}
					onchange={(e) => handleApparatChange(e.currentTarget.value)}
					required
					class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-ring focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
				>
					<option value="">Select apparat</option>
					{#each filteredApparats as app}
						<option value={app.id}>
							{app.short_name} - {app.name}
						</option>
					{/each}
				</select>
				<p class="text-xs text-muted-foreground">
					{filteredApparats.length} option{filteredApparats.length !== 1 ? 's' : ''} available
				</p>
				{#if fieldError('apparat_id')}
					<p class="text-sm text-red-500">{fieldError('apparat_id')}</p>
				{/if}
			</div>

			<!-- System Part Selection (with reactive filtering) -->
			<div class="space-y-2">
				<Label for="field_device_system_part">System Part *</Label>
				<Input
					type="text"
					placeholder="Search system part..."
					bind:value={systemPartSearch}
					class="mb-1"
				/>
				<select
					id="field_device_system_part"
					value={system_part_id}
					onchange={(e) => handleSystemPartChange(e.currentTarget.value)}
					required
					class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-ring focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
				>
					<option value="">Select system part</option>
					{#each filteredSystemParts as sp}
						<option value={sp.id}>
							{sp.short_name} - {sp.name}
						</option>
					{/each}
				</select>
				<p class="text-xs text-muted-foreground">
					{filteredSystemParts.length} option{filteredSystemParts.length !== 1 ? 's' : ''} available
				</p>
				{#if fieldError('system_part_id')}
					<p class="text-sm text-red-500">{fieldError('system_part_id')}</p>
				{/if}
			</div>
		</div>

		{#if object_data_id}
			<div class="mt-4 rounded-md border border-muted bg-background p-4">
				<h4 class="mb-2 text-sm font-medium">BacNet Objects (will be copied to field device)</h4>
				{#if loadingBacnet}
					<p class="text-sm text-muted-foreground">Loading...</p>
				{:else if bacnetObjects.length > 0}
					<div class="max-h-48 overflow-y-auto">
						<table class="w-full text-sm">
							<thead class="border-b">
								<tr class="text-left text-muted-foreground">
									<th class="p-2">Text Fix</th>
									<th class="p-2">Description</th>
									<th class="p-2">Type</th>
								</tr>
							</thead>
							<tbody>
								{#each bacnetObjects as obj}
									<tr class="border-b border-muted last:border-0">
										<td class="p-2">{obj.text_fix || '-'}</td>
										<td class="p-2">{obj.description || '-'}</td>
										<td class="p-2">{obj.software_type || '-'}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				{:else}
					<p class="text-sm text-muted-foreground">No BacNet objects found.</p>
				{/if}
			</div>
		{/if}

		{#if error}
			<p class="text-sm text-red-500">{error}</p>
		{/if}

		<div class="flex justify-end gap-2 pt-2">
			<Button type="button" variant="ghost" onclick={onCancel}>Cancel</Button>
			<Button type="submit" disabled={loading}>
				{initialData ? 'Update' : 'Create'}
			</Button>
		</div>
	{/if}
</form>
