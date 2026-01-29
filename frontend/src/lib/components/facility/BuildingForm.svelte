<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { createBuilding, updateBuilding } from '$lib/infrastructure/api/facility.adapter.js';
	import { getErrorMessage, getFieldError, getFieldErrors } from '$lib/api/client.js';
	import type { Building } from '$lib/domain/facility/index.js';
	import { createEventDispatcher } from 'svelte';

	export let initialData: Building | undefined = undefined;

	let iws_code = initialData?.iws_code ?? '';
	let building_group = initialData?.building_group ?? 0;
	let loading = false;
	let error = '';
	let fieldErrors: Record<string, string> = {};

	$: if (initialData) {
		iws_code = initialData.iws_code;
		building_group = initialData.building_group;
	}

	const dispatch = createEventDispatcher();

	const fieldError = (name: string) => getFieldError(fieldErrors, name, ['building']);

	async function handleSubmit() {
		loading = true;
		error = '';
		fieldErrors = {};
		try {
			if (initialData) {
				const res = await updateBuilding(initialData.id, {
					iws_code,
					building_group: Number(building_group)
				});
				dispatch('success', res);
			} else {
				const res = await createBuilding({
					iws_code,
					building_group: Number(building_group)
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
		<h3 class="text-lg font-medium">{initialData ? 'Edit Building' : 'New Building'}</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="iws_code">IWS Code</Label>
			<Input
				id="iws_code"
				bind:value={iws_code}
				required
				placeholder="e.g. ABCD"
				minlength={4}
				maxlength={4}
			/>
			<p class="text-xs text-muted-foreground">Exactly 4 characters</p>
			{#if fieldError('iws_code')}
				<p class="text-sm text-red-500">{fieldError('iws_code')}</p>
			{/if}
		</div>

		<div class="space-y-2">
			<Label for="building_group">Building Group</Label>
			<Input
				id="building_group"
				type="number"
				bind:value={building_group}
				required
				placeholder="e.g. 1"
			/>
			{#if fieldError('building_group')}
				<p class="text-sm text-red-500">{fieldError('building_group')}</p>
			{/if}
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
