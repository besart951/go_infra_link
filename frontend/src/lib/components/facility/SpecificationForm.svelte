<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import FieldDeviceSelect from '$lib/components/facility/FieldDeviceSelect.svelte';
	import {
		createSpecification,
		updateSpecification
	} from '$lib/infrastructure/api/facility.adapter.js';
	import { getErrorMessage } from '$lib/api/client.js';
	import type { Specification } from '$lib/domain/facility/index.js';
	import { createEventDispatcher } from 'svelte';

	export let initialData: Specification | undefined = undefined;

	let field_device_id = initialData?.field_device_id ?? '';
	let specification_supplier = initialData?.specification_supplier ?? '';
	let specification_brand = initialData?.specification_brand ?? '';
	let specification_type = initialData?.specification_type ?? '';
	let additional_info_motor_valve = initialData?.additional_info_motor_valve ?? '';
	let additional_info_size = initialData?.additional_info_size?.toString() ?? '';
	let additional_information_installation_location =
		initialData?.additional_information_installation_location ?? '';
	let electrical_connection_ph = initialData?.electrical_connection_ph?.toString() ?? '';
	let electrical_connection_acdc = initialData?.electrical_connection_acdc ?? '';
	let electrical_connection_amperage =
		initialData?.electrical_connection_amperage?.toString() ?? '';
	let electrical_connection_power = initialData?.electrical_connection_power?.toString() ?? '';
	let electrical_connection_rotation =
		initialData?.electrical_connection_rotation?.toString() ?? '';

	let loading = false;
	let error = '';

	$: if (initialData) {
		field_device_id = initialData.field_device_id ?? '';
		specification_supplier = initialData.specification_supplier ?? '';
		specification_brand = initialData.specification_brand ?? '';
		specification_type = initialData.specification_type ?? '';
		additional_info_motor_valve = initialData.additional_info_motor_valve ?? '';
		additional_info_size = initialData.additional_info_size?.toString() ?? '';
		additional_information_installation_location =
			initialData.additional_information_installation_location ?? '';
		electrical_connection_ph = initialData.electrical_connection_ph?.toString() ?? '';
		electrical_connection_acdc = initialData.electrical_connection_acdc ?? '';
		electrical_connection_amperage = initialData.electrical_connection_amperage?.toString() ?? '';
		electrical_connection_power = initialData.electrical_connection_power?.toString() ?? '';
		electrical_connection_rotation = initialData.electrical_connection_rotation?.toString() ?? '';
	}

	const dispatch = createEventDispatcher();

	function toNumber(value: string): number | undefined {
		const parsed = Number(value);
		return Number.isFinite(parsed) ? parsed : undefined;
	}

	async function handleSubmit() {
		loading = true;
		error = '';

		if (!field_device_id) {
			error = 'Please select a field device';
			loading = false;
			return;
		}

		try {
			if (initialData) {
				const res = await updateSpecification(initialData.id, {
					specification_supplier: specification_supplier || undefined,
					specification_brand: specification_brand || undefined,
					specification_type: specification_type || undefined,
					additional_info_motor_valve: additional_info_motor_valve || undefined,
					additional_info_size: toNumber(additional_info_size),
					additional_information_installation_location:
						additional_information_installation_location || undefined,
					electrical_connection_ph: toNumber(electrical_connection_ph),
					electrical_connection_acdc: electrical_connection_acdc || undefined,
					electrical_connection_amperage: toNumber(electrical_connection_amperage),
					electrical_connection_power: toNumber(electrical_connection_power),
					electrical_connection_rotation: toNumber(electrical_connection_rotation)
				});
				dispatch('success', res);
			} else {
				const res = await createSpecification({
					field_device_id,
					specification_supplier: specification_supplier || undefined,
					specification_brand: specification_brand || undefined,
					specification_type: specification_type || undefined,
					additional_info_motor_valve: additional_info_motor_valve || undefined,
					additional_info_size: toNumber(additional_info_size),
					additional_information_installation_location:
						additional_information_installation_location || undefined,
					electrical_connection_ph: toNumber(electrical_connection_ph),
					electrical_connection_acdc: electrical_connection_acdc || undefined,
					electrical_connection_amperage: toNumber(electrical_connection_amperage),
					electrical_connection_power: toNumber(electrical_connection_power),
					electrical_connection_rotation: toNumber(electrical_connection_rotation)
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
		<h3 class="text-lg font-medium">{initialData ? 'Edit Specification' : 'New Specification'}</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2 md:col-span-2">
			<Label>Field Device</Label>
			<div class="block">
				<FieldDeviceSelect bind:value={field_device_id} width="w-full" />
			</div>
		</div>
		<div class="space-y-2">
			<Label for="spec_supplier">Supplier</Label>
			<Input id="spec_supplier" bind:value={specification_supplier} maxlength={250} />
		</div>
		<div class="space-y-2">
			<Label for="spec_brand">Brand</Label>
			<Input id="spec_brand" bind:value={specification_brand} maxlength={250} />
		</div>
		<div class="space-y-2">
			<Label for="spec_type">Type</Label>
			<Input id="spec_type" bind:value={specification_type} maxlength={250} />
		</div>
		<div class="space-y-2">
			<Label for="spec_motor_valve">Motor Valve Info</Label>
			<Input id="spec_motor_valve" bind:value={additional_info_motor_valve} maxlength={250} />
		</div>
		<div class="space-y-2">
			<Label for="spec_size">Size</Label>
			<Input id="spec_size" type="number" bind:value={additional_info_size} />
		</div>
		<div class="space-y-2">
			<Label for="spec_ph">Electrical PH</Label>
			<Input id="spec_ph" type="number" bind:value={electrical_connection_ph} />
		</div>
		<div class="space-y-2">
			<Label for="spec_acdc">Electrical AC/DC</Label>
			<Input id="spec_acdc" bind:value={electrical_connection_acdc} maxlength={2} />
		</div>
		<div class="space-y-2">
			<Label for="spec_amp">Amperage</Label>
			<Input id="spec_amp" type="number" bind:value={electrical_connection_amperage} />
		</div>
		<div class="space-y-2">
			<Label for="spec_power">Power</Label>
			<Input id="spec_power" type="number" bind:value={electrical_connection_power} />
		</div>
		<div class="space-y-2">
			<Label for="spec_rotation">Rotation</Label>
			<Input id="spec_rotation" type="number" bind:value={electrical_connection_rotation} />
		</div>
		<div class="space-y-2 md:col-span-2">
			<Label for="spec_location">Installation Location</Label>
			<Textarea
				id="spec_location"
				bind:value={additional_information_installation_location}
				rows={2}
				maxlength={250}
			/>
		</div>
	</div>

	{#if error}
		<p class="text-sm text-red-500">{error}</p>
	{/if}

	<div class="flex justify-end gap-2 pt-2">
		<Button type="button" variant="ghost" onclick={() => dispatch('cancel')}>Cancel</Button>
		<Button type="submit" disabled={loading}>{initialData ? 'Update' : 'Create'}</Button>
	</div>
</form>
