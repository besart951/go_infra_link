<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { createSystemType, updateSystemType } from '$lib/infrastructure/api/facility.adapter.js';
	import { getErrorMessage, getFieldError, getFieldErrors } from '$lib/api/client.js';
	import type { SystemType } from '$lib/domain/facility/index.js';
	import { createEventDispatcher } from 'svelte';

	export let initialData: SystemType | undefined = undefined;

	let name = initialData?.name ?? '';
	let number_min = initialData?.number_min ?? 0;
	let number_max = initialData?.number_max ?? 0;
	let loading = false;
	let error = '';
	let fieldErrors: Record<string, string> = {};

	$: if (initialData) {
		name = initialData.name;
		number_min = initialData.number_min;
		number_max = initialData.number_max;
	}

	const dispatch = createEventDispatcher();

	const fieldError = (name: string) => getFieldError(fieldErrors, name, ['systemtype']);

	async function handleSubmit() {
		loading = true;
		error = '';
		fieldErrors = {};

		try {
			if (initialData) {
				const res = await updateSystemType(initialData.id, {
					name,
					number_min,
					number_max
				});
				dispatch('success', res);
			} else {
				const res = await createSystemType({
					name,
					number_min,
					number_max
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
		<h3 class="text-lg font-medium">{initialData ? 'Edit System Type' : 'New System Type'}</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-3">
		<div class="space-y-2 md:col-span-1">
			<Label for="system_type_name">Name</Label>
			<Input id="system_type_name" bind:value={name} required maxlength={150} />
			{#if fieldError('name')}
				<p class="text-sm text-red-500">{fieldError('name')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="system_type_min">Min Number</Label>
			<Input id="system_type_min" type="number" bind:value={number_min} required />
			{#if fieldError('number_min')}
				<p class="text-sm text-red-500">{fieldError('number_min')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="system_type_max">Max Number</Label>
			<Input id="system_type_max" type="number" bind:value={number_max} required />
			{#if fieldError('number_max')}
				<p class="text-sm text-red-500">{fieldError('number_max')}</p>
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
