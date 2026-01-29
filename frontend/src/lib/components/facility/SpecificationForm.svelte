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
	import { getErrorMessage, getFieldError, getFieldErrors } from '$lib/api/client.js';
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
	let fieldErrors: Record<string, string> = {};

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

	const fieldError = (name: string) => getFieldError(fieldErrors, name, ['specification']);

	function toNumber(value: string): number | undefined {
		const parsed = Number(value);
		return Number.isFinite(parsed) ? parsed : undefined;
	}

	async function handleSubmit() {
		loading = true;
		error = '';
		fieldErrors = {};

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
			fieldErrors = getFieldErrors(e);
			error = Object.keys(fieldErrors).length ? '' : getErrorMessage(e);
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
			{#if fieldError('field_device_id')}
				<p class="text-sm text-red-500">{fieldError('field_device_id')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="spec_supplier">Supplier</Label>
			<Input id="spec_supplier" bind:value={specification_supplier} maxlength={250} />
			{#if fieldError('specification_supplier')}
				<p class="text-sm text-red-500">{fieldError('specification_supplier')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="spec_brand">Brand</Label>
			<Input id="spec_brand" bind:value={specification_brand} maxlength={250} />
			{#if fieldError('specification_brand')}
				<p class="text-sm text-red-500">{fieldError('specification_brand')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="spec_type">Type</Label>
			<Input id="spec_type" bind:value={specification_type} maxlength={250} />
			{#if fieldError('specification_type')}
				<p class="text-sm text-red-500">{fieldError('specification_type')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="spec_motor_valve">Motor Valve Info</Label>
			<Input id="spec_motor_valve" bind:value={additional_info_motor_valve} maxlength={250} />
			{#if fieldError('additional_info_motor_valve')}
				<p class="text-sm text-red-500">{fieldError('additional_info_motor_valve')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="spec_size">Size</Label>
			<Input id="spec_size" type="number" bind:value={additional_info_size} />
			{#if fieldError('additional_info_size')}
				<p class="text-sm text-red-500">{fieldError('additional_info_size')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="spec_ph">Electrical PH</Label>
			<Input id="spec_ph" type="number" bind:value={electrical_connection_ph} />
			{#if fieldError('electrical_connection_ph')}
				<p class="text-sm text-red-500">{fieldError('electrical_connection_ph')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="spec_acdc">Electrical AC/DC</Label>
			<Input id="spec_acdc" bind:value={electrical_connection_acdc} maxlength={2} />
			{#if fieldError('electrical_connection_acdc')}
				<p class="text-sm text-red-500">{fieldError('electrical_connection_acdc')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="spec_amp">Amperage</Label>
			<Input id="spec_amp" type="number" bind:value={electrical_connection_amperage} />
			{#if fieldError('electrical_connection_amperage')}
				<p class="text-sm text-red-500">{fieldError('electrical_connection_amperage')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="spec_power">Power</Label>
			<Input id="spec_power" type="number" bind:value={electrical_connection_power} />
			{#if fieldError('electrical_connection_power')}
				<p class="text-sm text-red-500">{fieldError('electrical_connection_power')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="spec_rotation">Rotation</Label>
			<Input id="spec_rotation" type="number" bind:value={electrical_connection_rotation} />
			{#if fieldError('electrical_connection_rotation')}
				<p class="text-sm text-red-500">{fieldError('electrical_connection_rotation')}</p>
			{/if}
		</div>
		<div class="space-y-2 md:col-span-2">
			<Label for="spec_location">Installation Location</Label>
			<Textarea
				id="spec_location"
				bind:value={additional_information_installation_location}
				rows={2}
				maxlength={250}
			/>
			{#if fieldError('additional_information_installation_location')}
				<p class="text-sm text-red-500">
					{fieldError('additional_information_installation_location')}
				</p>
			{/if}
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
