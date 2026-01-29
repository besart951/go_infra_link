<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import {
		createFieldDevice,
		updateFieldDevice,
		listSPSControllerSystemTypes,
		listSystemParts,
		listApparats,
		listObjectData,
		getSPSControllerSystemType,
		getObjectData,
		getApparat,
		getSystemPart,
		getObjectDataBacnetObjects
	} from '$lib/infrastructure/api/facility.adapter.js';
	import { listProjectObjectData } from '$lib/infrastructure/api/project.adapter.js';
	import { getErrorMessage, getFieldError, getFieldErrors } from '$lib/api/client.js';
	import type {
		FieldDevice,
		SPSControllerSystemType,
		SystemPart,
		Apparat,
		ObjectData,
		BacnetObject
	} from '$lib/domain/facility/index.js';
	import { createEventDispatcher } from 'svelte';

	export let initialData: FieldDevice | undefined = undefined;
	export let projectId: string | undefined = undefined;

	let bmk = initialData?.bmk ?? '';
	let description = initialData?.description ?? '';
	let apparat_nr = initialData?.apparat_nr ?? '';
	let sps_controller_system_type_id = initialData?.sps_controller_system_type_id ?? '';
	let system_part_id = initialData?.system_part_id ?? '';
	let apparat_id = initialData?.apparat_id ?? '';
	let object_data_id = '';

	let bacnetObjects: BacnetObject[] = [];
	let loadingBacnet = false;

	let loading = false;
	let error = '';
	let fieldErrors: Record<string, string> = {};

	$: if (initialData) {
		bmk = initialData.bmk ?? '';
		description = initialData.description ?? '';
		apparat_nr = initialData.apparat_nr ?? '';
		sps_controller_system_type_id = initialData.sps_controller_system_type_id ?? '';
		system_part_id = initialData.system_part_id ?? '';
		apparat_id = initialData.apparat_id ?? '';
	}

	const dispatch = createEventDispatcher();

	const fieldError = (name: string) => getFieldError(fieldErrors, name, ['fielddevice']);

	async function fetchSPSControllerSystemTypes(
		search: string
	): Promise<SPSControllerSystemType[]> {
		try {
			const res = await listSPSControllerSystemTypes({ page: 1, limit: 50, search });
			return res.items;
		} catch (e) {
			console.error(e);
			return [];
		}
	}

	async function fetchObjectData(search: string): Promise<ObjectData[]> {
		try {
			const res = projectId
				? await listProjectObjectData(projectId, { page: 1, limit: 50, search })
				: await listObjectData({ page: 1, limit: 50, search });
			return projectId ? res.items.filter((obj) => obj.is_active) : res.items;
		} catch (e) {
			console.error(e);
			return [];
		}
	}

	async function fetchApparats(search: string): Promise<Apparat[]> {
		try {
			const params: any = { page: 1, limit: 50, search };
			if (object_data_id) {
				params.object_data_id = object_data_id;
			}
			const res = await listApparats(params);
			return res.items;
		} catch (e) {
			console.error(e);
			return [];
		}
	}

	async function fetchSystemParts(search: string): Promise<SystemPart[]> {
		try {
			const params: any = { page: 1, limit: 50, search };
			if (apparat_id) {
				params.apparat_id = apparat_id;
			}
			const res = await listSystemParts(params);
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
					id: initialData.id,
					bmk: bmk || undefined,
					description: description || undefined,
					apparat_nr: apparatNumber,
					sps_controller_system_type_id,
					system_part_id,
					apparat_id,
					object_data_id: object_data_id || undefined
				});
				dispatch('success', res);
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
				dispatch('success', res);
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

<form on:submit|preventDefault={handleSubmit} class="space-y-4 rounded-md border bg-muted/20 p-4">
	<div class="mb-4 flex items-center justify-between">
		<h3 class="text-lg font-medium">
			{initialData ? 'Edit Field Device' : 'New Field Device'}
		</h3>
	</div>

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
			<Label for="field_device_apparat_nr">Apparat Nr</Label>
			<Input id="field_device_apparat_nr" type="number" bind:value={apparat_nr} required />
			{#if fieldError('apparat_nr')}
				<p class="text-sm text-red-500">{fieldError('apparat_nr')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="field_device_sps_type">SPS Controller System Type</Label>
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
		<div class="space-y-2">
			<Label for="field_device_system_part">System Part</Label>
			<AsyncCombobox
				id="field_device_system_part"
				bind:value={system_part_id}
				fetcher={fetchSystemParts}
				fetchById={getSystemPart}
				labelKey="name"
				width="w-full"
				placeholder="Select system part"
				searchPlaceholder="Search system parts..."
			/>
			{#if fieldError('system_part_id')}
				<p class="text-sm text-red-500">{fieldError('system_part_id')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="field_device_apparat">Apparat</Label>
			<AsyncCombobox
				id="field_device_apparat"
				bind:value={apparat_id}
				fetcher={fetchApparats}
				fetchById={getApparat}
				labelKey="name"
				width="w-full"
				placeholder="Select apparat"
				searchPlaceholder="Search apparats..."
			/>
			{#if fieldError('apparat_id')}
				<p class="text-sm text-red-500">{fieldError('apparat_id')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="field_device_object_data">Object Data (optional)</Label>
			<AsyncCombobox
				id="field_device_object_data"
				bind:value={object_data_id}
				fetcher={fetchObjectData}
				fetchById={getObjectData}
				labelKey="description"
				width="w-full"
				placeholder="None"
				searchPlaceholder="Search object data..."
			/>
			{#if fieldError('object_data_id')}
				<p class="text-sm text-red-500">{fieldError('object_data_id')}</p>
			{/if}
		</div>
	</div>

	{#if object_data_id && bacnetObjects.length > 0}
		<div class="mt-4 rounded-md border border-muted bg-background p-4">
			<h4 class="mb-2 text-sm font-medium">BacNet Objects (will be copied to field device)</h4>
			{#if loadingBacnet}
				<p class="text-sm text-muted-foreground">Loading...</p>
			{:else}
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
			{/if}
		</div>
	{/if}

	{#if error}
		<p class="text-sm text-red-500">{error}</p>
	{/if}

	<div class="flex justify-end gap-2 pt-2">
		<Button type="button" variant="ghost" onclick={() => dispatch('cancel')}>Cancel</Button>
		<Button type="submit" disabled={loading}>
			{initialData ? 'Update' : 'Create'}
		</Button>
	</div>
</form>
