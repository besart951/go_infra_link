<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { createObjectData, updateObjectData } from '$lib/infrastructure/api/facility.adapter.js';
	import { getErrorMessage } from '$lib/api/client.js';
	import type { ObjectData } from '$lib/domain/facility/index.js';
	import { createEventDispatcher } from 'svelte';

	export let initialData: ObjectData | undefined = undefined;

	let description = initialData?.description ?? '';
	let version = initialData?.version ?? '';
	let is_active = initialData?.is_active ?? true;
	let loading = false;
	let error = '';

	$: if (initialData) {
		description = initialData.description ?? '';
		version = initialData.version ?? '';
		is_active = initialData.is_active ?? true;
	}

	const dispatch = createEventDispatcher();

	async function handleSubmit() {
		loading = true;
		error = '';

		try {
			if (initialData) {
				const res = await updateObjectData(initialData.id, {
					description,
					version,
					is_active
				});
				dispatch('success', res);
			} else {
				const res = await createObjectData({
					description,
					version,
					is_active
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
		<h3 class="text-lg font-medium">{initialData ? 'Edit Object Data' : 'New Object Data'}</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-3">
		<div class="space-y-2 md:col-span-2">
			<Label for="object_data_description">Description</Label>
			<Input id="object_data_description" bind:value={description} required maxlength={250} />
		</div>
		<div class="space-y-2">
			<Label for="object_data_version">Version</Label>
			<Input id="object_data_version" bind:value={version} required maxlength={100} />
		</div>
		<div class="flex items-center gap-2 md:col-span-3">
			<input id="object_data_active" type="checkbox" bind:checked={is_active} class="h-4 w-4" />
			<Label for="object_data_active">Active</Label>
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
