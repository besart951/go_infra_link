<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { createSystemPart, updateSystemPart } from '$lib/infrastructure/api/facility.adapter.js';
	import { getErrorMessage } from '$lib/api/client.js';
	import type { SystemPart } from '$lib/domain/facility/index.js';
	import { createEventDispatcher } from 'svelte';

	export let initialData: SystemPart | undefined = undefined;

	let short_name = initialData?.short_name ?? '';
	let name = initialData?.name ?? '';
	let description = initialData?.description ?? '';
	let loading = false;
	let error = '';

	$: if (initialData) {
		short_name = initialData.short_name;
		name = initialData.name;
		description = initialData.description ?? '';
	}

	const dispatch = createEventDispatcher();

	async function handleSubmit() {
		loading = true;
		error = '';

		try {
			if (initialData) {
				const res = await updateSystemPart(initialData.id, {
					short_name,
					name,
					description: description || undefined
				});
				dispatch('success', res);
			} else {
				const res = await createSystemPart({
					short_name,
					name,
					description: description || undefined
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
		<h3 class="text-lg font-medium">{initialData ? 'Edit System Part' : 'New System Part'}</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="system_part_short">Short Name</Label>
			<Input id="system_part_short" bind:value={short_name} required maxlength={10} />
		</div>
		<div class="space-y-2">
			<Label for="system_part_name">Name</Label>
			<Input id="system_part_name" bind:value={name} required maxlength={250} />
		</div>
		<div class="space-y-2 md:col-span-2">
			<Label for="system_part_desc">Description</Label>
			<Textarea id="system_part_desc" bind:value={description} rows={3} maxlength={250} />
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
