<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import ControlCabinetSelect from './ControlCabinetSelect.svelte';
	import {
		createSPSController,
		updateSPSController
	} from '$lib/infrastructure/api/facility.adapter.js';
	import { getErrorMessage } from '$lib/api/client.js';
	import type { SPSController } from '$lib/domain/facility/index.js';
	import { createEventDispatcher } from 'svelte';

	export let initialData: SPSController | undefined = undefined;

	let ga_device = initialData?.ga_device ?? '';
	let device_name = initialData?.device_name ?? '';
	let ip_address = initialData?.ip_address ?? '';
	let control_cabinet_id = initialData?.control_cabinet_id ?? '';

	let loading = false;
	let error = '';

	// React to initialData changes
	$: if (initialData) {
		ga_device = initialData.ga_device;
		device_name = initialData.device_name;
		ip_address = initialData.ip_address;
		control_cabinet_id = initialData.control_cabinet_id;
	}

	const dispatch = createEventDispatcher();

	async function handleSubmit() {
		loading = true;
		error = '';

		if (!control_cabinet_id) {
			error = 'Please select a control cabinet';
			loading = false;
			return;
		}

		try {
			if (initialData) {
				const res = await updateSPSController(initialData.id, {
					id: initialData.id,
					ga_device,
					device_name,
					ip_address,
					control_cabinet_id
				});
				dispatch('success', res);
			} else {
				const res = await createSPSController({
					ga_device,
					device_name,
					ip_address,
					control_cabinet_id
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
</script>

<form on:submit|preventDefault={handleSubmit} class="space-y-4 rounded-md border bg-muted/20 p-4">
	<div class="mb-4 flex items-center justify-between">
		<h3 class="text-lg font-medium">
			{initialData ? 'Edit SPS Controller' : 'New SPS Controller'}
		</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="ga_device">GA Device</Label>
			<Input id="ga_device" bind:value={ga_device} required maxlength={10} />
		</div>
		<div class="space-y-2">
			<Label for="device_name">Device Name</Label>
			<Input id="device_name" bind:value={device_name} required maxlength={100} />
		</div>
		<div class="space-y-2">
			<Label for="ip_address">IP Address</Label>
			<Input id="ip_address" bind:value={ip_address} required maxlength={50} />
		</div>

		<div class="space-y-2">
			<Label>Control Cabinet</Label>
			<div class="block">
				<ControlCabinetSelect bind:value={control_cabinet_id} width="w-full" />
			</div>
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
