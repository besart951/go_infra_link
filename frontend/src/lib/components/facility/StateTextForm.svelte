<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { createStateText, updateStateText } from '$lib/infrastructure/api/facility.adapter.js';
	import { getErrorMessage, getFieldError, getFieldErrors } from '$lib/api/client.js';
	import type { StateText } from '$lib/domain/facility/index.js';
	import { createEventDispatcher } from 'svelte';

	export let initialData: StateText | undefined = undefined;

	let ref_number = initialData?.ref_number ?? 0;
	let state_text1 = initialData?.state_text1 ?? '';
	let state_text2 = initialData?.state_text2 ?? '';
	let state_text3 = initialData?.state_text3 ?? '';
	let state_text4 = initialData?.state_text4 ?? '';
	let loading = false;
	let error = '';
	let fieldErrors: Record<string, string> = {};

	$: if (initialData) {
		ref_number = initialData.ref_number;
		state_text1 = initialData.state_text1 ?? '';
		state_text2 = initialData.state_text2 ?? '';
		state_text3 = initialData.state_text3 ?? '';
		state_text4 = initialData.state_text4 ?? '';
	}

	const dispatch = createEventDispatcher();

	const fieldError = (name: string) => getFieldError(fieldErrors, name, ['statetext']);

	async function handleSubmit() {
		loading = true;
		error = '';
		fieldErrors = {};

		try {
			if (initialData) {
				const res = await updateStateText(initialData.id, {
					ref_number,
					state_text1: state_text1 || undefined,
					state_text2: state_text2 || undefined,
					state_text3: state_text3 || undefined,
					state_text4: state_text4 || undefined
				});
				dispatch('success', res);
			} else {
				const res = await createStateText({
					ref_number,
					state_text1: state_text1 || undefined,
					state_text2: state_text2 || undefined,
					state_text3: state_text3 || undefined,
					state_text4: state_text4 || undefined
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
		<h3 class="text-lg font-medium">{initialData ? 'Edit State Text' : 'New State Text'}</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="state_ref">Ref Number</Label>
			<Input id="state_ref" type="number" bind:value={ref_number} required />
			{#if fieldError('ref_number')}
				<p class="text-sm text-red-500">{fieldError('ref_number')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text1">State Text 1</Label>
			<Input id="state_text1" bind:value={state_text1} />
			{#if fieldError('state_text1')}
				<p class="text-sm text-red-500">{fieldError('state_text1')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text2">State Text 2</Label>
			<Input id="state_text2" bind:value={state_text2} />
			{#if fieldError('state_text2')}
				<p class="text-sm text-red-500">{fieldError('state_text2')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text3">State Text 3</Label>
			<Input id="state_text3" bind:value={state_text3} />
			{#if fieldError('state_text3')}
				<p class="text-sm text-red-500">{fieldError('state_text3')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text4">State Text 4</Label>
			<Input id="state_text4" bind:value={state_text4} />
			{#if fieldError('state_text4')}
				<p class="text-sm text-red-500">{fieldError('state_text4')}</p>
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
