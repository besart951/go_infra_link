<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { createApparat, updateApparat } from '$lib/infrastructure/api/facility.adapter.js';
	import { getErrorMessage, getFieldError, getFieldErrors } from '$lib/api/client.js';
	import type { Apparat } from '$lib/domain/facility/index.js';
	import { createEventDispatcher } from 'svelte';
	import SystemPartMultiSelect from './SystemPartMultiSelect.svelte';

	export let initialData: Apparat | undefined = undefined;

	let short_name = initialData?.short_name ?? '';
	let name = initialData?.name ?? '';
	let description = initialData?.description ?? '';
	let system_part_ids = initialData?.system_parts?.map((sp) => sp.id) ?? [];
	let loading = false;
	let error = '';
	let fieldErrors: Record<string, string> = {};

	$: if (initialData) {
		short_name = initialData.short_name;
		name = initialData.name;
		description = initialData.description ?? '';
		system_part_ids = initialData.system_parts?.map((sp) => sp.id) ?? [];
	}

	const dispatch = createEventDispatcher();

	const fieldError = (name: string) => getFieldError(fieldErrors, name, ['apparat']);

	async function handleSubmit() {
		loading = true;
		error = '';
		fieldErrors = {};

		try {
			if (initialData) {
				const res = await updateApparat(initialData.id, {
					short_name,
					name,
					description: description || undefined,
					system_part_ids
				});
				dispatch('success', res);
			} else {
				const res = await createApparat({
					short_name,
					name,
					description: description || undefined,
					system_part_ids
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
		<h3 class="text-lg font-medium">{initialData ? 'Edit Apparat' : 'New Apparat'}</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="apparat_short">Short Name</Label>
			<Input id="apparat_short" bind:value={short_name} required maxlength={255} />
			{#if fieldError('short_name')}
				<p class="text-sm text-red-500">{fieldError('short_name')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="apparat_name">Name</Label>
			<Input id="apparat_name" bind:value={name} required maxlength={250} />
			{#if fieldError('name')}
				<p class="text-sm text-red-500">{fieldError('name')}</p>
			{/if}
		</div>
		<div class="space-y-2 md:col-span-2">
			<Label for="apparat_desc">Description</Label>
			<Textarea id="apparat_desc" bind:value={description} rows={3} maxlength={250} />
			{#if fieldError('description')}
				<p class="text-sm text-red-500">{fieldError('description')}</p>
			{/if}
		</div>
		<div class="space-y-2 md:col-span-2">
			<Label for="apparat_system_parts">System Parts</Label>
			<SystemPartMultiSelect id="apparat_system_parts" bind:value={system_part_ids} />
			{#if fieldError('system_part_ids')}
				<p class="text-sm text-red-500">{fieldError('system_part_ids')}</p>
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
