<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import type { Phase } from '$lib/domain/phase/index.js';
	import { createPhase, updatePhase } from '$lib/infrastructure/api/phase.adapter.js';
	import { useFormState } from '$lib/hooks/useFormState.svelte.js';

	interface PhaseFormProps {
		initialData?: Phase;
		onSuccess?: (phase: Phase) => void;
		onCancel?: () => void;
	}

	let { initialData, onSuccess, onCancel }: PhaseFormProps = $props();

	let name = $state(initialData?.name ?? '');

	// Watch for changes to initialData
	$effect(() => {
		if (initialData) {
			name = initialData.name ?? '';
		}
	});

	const formState = useFormState({
		onSuccess: (result: Phase) => {
			onSuccess?.(result);
		}
	});

	async function handleSubmit() {
		if (!name.trim()) {
			formState.resetErrors();
			// We can't set error directly on formState, so we'll handle validation inline
			return;
		}

		await formState.handleSubmit(async () => {
			if (initialData) {
				return await updatePhase(initialData.id, { name: name.trim() });
			} else {
				return await createPhase({ name: name.trim() });
			}
		});
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
			{#if formState.getFieldError('name')}
				<p class="text-sm text-red-500">{formState.getFieldError('name')}</p>
			{/if}
		</div>
	</div>

	{#if formState.error}
		<p class="text-sm text-red-500">{formState.error}</p>
	{/if}

	<div class="flex justify-end gap-2 pt-2">
		<Button type="button" variant="ghost" onclick={onCancel}>Cancel</Button>
		<Button type="submit" disabled={formState.loading}>
			{initialData ? 'Update' : 'Create'}
		</Button>
	</div>
</form>
