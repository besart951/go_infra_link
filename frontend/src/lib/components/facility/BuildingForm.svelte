<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { createBuilding, updateBuilding } from '$lib/infrastructure/api/facility.adapter.js';
	import type { Building } from '$lib/domain/facility/index.js';
	import { useFormState } from '$lib/hooks/useFormState.svelte.js';

	interface BuildingFormProps {
		initialData?: Building;
		onSuccess?: (building: Building) => void;
		onCancel?: () => void;
	}

	let { initialData, onSuccess, onCancel }: BuildingFormProps = $props();

	let iws_code = $state('');
	let building_group = $state(0);

	$effect(() => {
		iws_code = initialData?.iws_code ?? '';
		building_group = initialData?.building_group ?? 0;
	});

	const formState = useFormState({
		onSuccess: (result: Building) => {
			onSuccess?.(result);
		}
	});

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		await formState.handleSubmit(async () => {
			if (initialData) {
				return await updateBuilding(initialData.id, {
					iws_code,
					building_group: Number(building_group)
				});
			} else {
				return await createBuilding({
					iws_code,
					building_group: Number(building_group)
				});
			}
		});
	}
</script>

<form onsubmit={handleSubmit} class="space-y-4 rounded-md border bg-muted/20 p-4">
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
			{#if formState.getFieldError('iws_code', ['building'])}
				<p class="text-sm text-red-500">{formState.getFieldError('iws_code', ['building'])}</p>
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
			{#if formState.getFieldError('building_group', ['building'])}
				<p class="text-sm text-red-500">
					{formState.getFieldError('building_group', ['building'])}
				</p>
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
