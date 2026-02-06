<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { createStateText, updateStateText } from '$lib/infrastructure/api/facility.adapter.js';
	import type { StateText } from '$lib/domain/facility/index.js';
	import { useFormState } from '$lib/hooks/useFormState.svelte.js';

	interface StateTextFormProps {
		initialData?: StateText;
		onSuccess?: (stateText: StateText) => void;
		onCancel?: () => void;
	}

	let { initialData, onSuccess, onCancel }: StateTextFormProps = $props();

	let ref_number = $state(0);
	let state_text1 = $state('');
	let state_text2 = $state('');
	let state_text3 = $state('');
	let state_text4 = $state('');

	$effect(() => {
		if (initialData) {
			ref_number = initialData.ref_number;
			state_text1 = initialData.state_text1 ?? '';
			state_text2 = initialData.state_text2 ?? '';
			state_text3 = initialData.state_text3 ?? '';
			state_text4 = initialData.state_text4 ?? '';
		}
	});

	const formState = useFormState({
		onSuccess: (result: StateText) => {
			onSuccess?.(result);
		}
	});

	async function handleSubmit() {
		await formState.handleSubmit(async () => {
			if (initialData) {
				return await updateStateText(initialData.id, {
					ref_number,
					state_text1: state_text1 || undefined,
					state_text2: state_text2 || undefined,
					state_text3: state_text3 || undefined,
					state_text4: state_text4 || undefined
				});
			} else {
				return await createStateText({
					ref_number,
					state_text1: state_text1 || undefined,
					state_text2: state_text2 || undefined,
					state_text3: state_text3 || undefined,
					state_text4: state_text4 || undefined
				});
			}
		});
	}
</script>

<form
	onsubmit={(e) => {
		e.preventDefault();
		handleSubmit();
	}}
	class="space-y-4 rounded-md border bg-muted/20 p-4"
>
	<div class="mb-4 flex items-center justify-between">
		<h3 class="text-lg font-medium">{initialData ? 'Edit State Text' : 'New State Text'}</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="state_ref">Ref Number</Label>
			<Input id="state_ref" type="number" bind:value={ref_number} required />
			{#if formState.getFieldError('ref_number', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('ref_number', ['statetext'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text1">State Text 1</Label>
			<Input id="state_text1" bind:value={state_text1} />
			{#if formState.getFieldError('state_text1', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('state_text1', ['statetext'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text2">State Text 2</Label>
			<Input id="state_text2" bind:value={state_text2} />
			{#if formState.getFieldError('state_text2', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('state_text2', ['statetext'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text3">State Text 3</Label>
			<Input id="state_text3" bind:value={state_text3} />
			{#if formState.getFieldError('state_text3', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('state_text3', ['statetext'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="state_text4">State Text 4</Label>
			<Input id="state_text4" bind:value={state_text4} />
			{#if formState.getFieldError('state_text4', ['statetext'])}
				<p class="text-sm text-red-500">{formState.getFieldError('state_text4', ['statetext'])}</p>
			{/if}
		</div>
	</div>

	{#if formState.error}
		<p class="text-sm text-red-500">{formState.error}</p>
	{/if}

	<div class="flex justify-end gap-2 pt-2">
		<Button type="button" variant="ghost" onclick={onCancel}>Cancel</Button>
		<Button type="submit" disabled={formState.loading}>{initialData ? 'Update' : 'Create'}</Button>
	</div>
</form>
