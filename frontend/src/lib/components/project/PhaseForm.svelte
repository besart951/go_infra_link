<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { createEventDispatcher } from 'svelte';
	import { getErrorMessage } from '$lib/api/client.js';
	import type { Phase } from '$lib/domain/phase/index.js';
	import { createPhase, updatePhase } from '$lib/infrastructure/api/phase.adapter.js';

	export let initialData: Phase | undefined = undefined;

	let name = initialData?.name ?? '';
	let loading = false;
	let error = '';

	$: if (initialData) {
		name = initialData.name ?? '';
	}

	const dispatch = createEventDispatcher();

	async function handleSubmit() {
		error = '';

		if (!name.trim()) {
			error = 'Phase name is required';
			return;
		}

		loading = true;
		try {
			if (initialData) {
				const res = await updatePhase(initialData.id, { name: name.trim() });
				dispatch('success', res);
			} else {
				const res = await createPhase({
					name: name.trim()
				});
				dispatch('success', res);
			}
		} catch (e) {
			error = getErrorMessage(e);
		} finally {
			loading = false;
		}
	}
</script>

<form on:submit|preventDefault={handleSubmit} class="space-y-4 rounded-md border bg-muted/20 p-4">
	<div class="mb-4 flex items-center justify-between">
		<h3 class="text-lg font-medium">{initialData ? 'Edit Phase' : 'New Phase'}</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="phase_name">Phase Name</Label>
			<Input id="phase_name" bind:value={name} required placeholder="e.g. SIA:51" />
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
