<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import {
		createFieldDevice,
		updateFieldDevice,
		listSPSControllerSystemTypes,
		listSystemParts,
		listApparats,
		listObjectData
	} from '$lib/infrastructure/api/facility.adapter.js';
	import { listProjectObjectData } from '$lib/infrastructure/api/project.adapter.js';
	import { getErrorMessage } from '$lib/api/client.js';
	import type {
		FieldDevice,
		SPSControllerSystemType,
		SystemPart,
		Apparat,
		ObjectData
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

	let spsControllerSystemTypes: SPSControllerSystemType[] = [];
	let systemParts: SystemPart[] = [];
	let apparats: Apparat[] = [];
	let objectData: ObjectData[] = [];

	let loading = false;
	let error = '';

	$: if (initialData) {
		bmk = initialData.bmk ?? '';
		description = initialData.description ?? '';
		apparat_nr = initialData.apparat_nr ?? '';
		sps_controller_system_type_id = initialData.sps_controller_system_type_id ?? '';
		system_part_id = initialData.system_part_id ?? '';
		apparat_id = initialData.apparat_id ?? '';
	}

	const dispatch = createEventDispatcher();

	async function loadLookups() {
		try {
			const [spsRes, partsRes, apparatsRes, objectRes] = await Promise.all([
				listSPSControllerSystemTypes({ page: 1, limit: 100 }),
				listSystemParts({ page: 1, limit: 100 }),
				listApparats({ page: 1, limit: 100 }),
				projectId
					? listProjectObjectData(projectId, { page: 1, limit: 100 })
					: listObjectData({ page: 1, limit: 100 })
			]);
			spsControllerSystemTypes = spsRes.items;
			systemParts = partsRes.items;
			apparats = apparatsRes.items;
			objectData = projectId
				? objectRes.items.filter((obj) => obj.is_active)
				: objectRes.items;
		} catch (e) {
			console.error(e);
			error = getErrorMessage(e);
		}
	}

	async function handleSubmit() {
		loading = true;
		error = '';

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
			error = getErrorMessage(e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadLookups();
	});
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
		</div>
		<div class="space-y-2">
			<Label for="field_device_desc">Description</Label>
			<Input id="field_device_desc" bind:value={description} maxlength={255} />
		</div>
		<div class="space-y-2">
			<Label for="field_device_apparat_nr">Apparat Nr</Label>
			<Input id="field_device_apparat_nr" type="number" bind:value={apparat_nr} required />
		</div>
		<div class="space-y-2">
			<Label for="field_device_sps_type">SPS Controller System Type</Label>
			<select
				id="field_device_sps_type"
				class="h-9 rounded-md border border-input bg-background px-3 text-sm font-medium shadow-xs"
				bind:value={sps_controller_system_type_id}
				required
			>
				<option value="">Select SPS controller system type</option>
				{#each spsControllerSystemTypes as item}
					<option value={item.id}>
						{item.sps_controller_name || item.sps_controller_id} -
						{item.system_type_name || item.system_type_id}
					</option>
				{/each}
			</select>
		</div>
		<div class="space-y-2">
			<Label for="field_device_system_part">System Part</Label>
			<select
				id="field_device_system_part"
				class="h-9 rounded-md border border-input bg-background px-3 text-sm font-medium shadow-xs"
				bind:value={system_part_id}
				required
			>
				<option value="">Select system part</option>
				{#each systemParts as part}
					<option value={part.id}>
						{part.short_name} - {part.name}
					</option>
				{/each}
			</select>
		</div>
		<div class="space-y-2">
			<Label for="field_device_apparat">Apparat</Label>
			<select
				id="field_device_apparat"
				class="h-9 rounded-md border border-input bg-background px-3 text-sm font-medium shadow-xs"
				bind:value={apparat_id}
				required
			>
				<option value="">Select apparat</option>
				{#each apparats as apparat}
					<option value={apparat.id}>
						{apparat.short_name} - {apparat.name}
					</option>
				{/each}
			</select>
		</div>
		<div class="space-y-2">
			<Label for="field_device_object_data">Object Data (optional)</Label>
			<select
				id="field_device_object_data"
				class="h-9 rounded-md border border-input bg-background px-3 text-sm font-medium shadow-xs"
				bind:value={object_data_id}
			>
				<option value="">None</option>
				{#each objectData as obj}
					<option value={obj.id}>{obj.description}</option>
				{/each}
			</select>
		</div>
	</div>

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
