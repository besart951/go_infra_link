<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { createSystemType, updateSystemType } from '$lib/infrastructure/api/facility.adapter.js';
	import type { SystemType } from '$lib/domain/facility/index.js';
	import { useFormState } from '$lib/hooks/useFormState.svelte.js';

	interface SystemTypeFormProps {
		initialData?: SystemType;
		onSuccess?: (systemType: SystemType) => void;
		onCancel?: () => void;
	}

	let { initialData, onSuccess, onCancel }: SystemTypeFormProps = $props();

	let name = $state('');
	let number_min = $state(0);
	let number_max = $state(0);

	$effect(() => {
		if (initialData) {
			name = initialData.name;
			number_min = initialData.number_min;
			number_max = initialData.number_max;
		}
	});

	const formState = useFormState({
		onSuccess: (result: SystemType) => {
			onSuccess?.(result);
		}
	});

	let localErrors = $state<Record<string, string>>({});

	function clearLocalErrors() {
		if (Object.keys(localErrors).length > 0) {
			localErrors = {};
		}
	}

	function validateLocal(): boolean {
		localErrors = {};
		const minValue = Number(number_min);
		const maxValue = Number(number_max);
		if (Number.isFinite(minValue) && Number.isFinite(maxValue) && minValue >= maxValue) {
			localErrors = {
				number_max: 'Max number must be greater than min number.'
			};
		}
		return Object.keys(localErrors).length === 0;
	}

	function getError(field: string) {
		return localErrors[field] ?? formState.getFieldError(field, ['systemtype']);
	}

	async function handleSubmit() {
		if (!validateLocal()) {
			return;
		}
		await formState.handleSubmit(async () => {
			if (initialData) {
				return await updateSystemType(initialData.id, {
					name,
					number_min,
					number_max
				});
			} else {
				return await createSystemType({
					name,
					number_min,
					number_max
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
		<h3 class="text-lg font-medium">{initialData ? 'Edit System Type' : 'New System Type'}</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-3">
		<div class="space-y-2 md:col-span-1">
			<Label for="system_type_name">Name</Label>
			<Input id="system_type_name" bind:value={name} required maxlength={150} oninput={clearLocalErrors} />
			{#if getError('name')}
				<p class="text-sm text-red-500">{getError('name')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="system_type_min">Min Number</Label>
			<Input
				id="system_type_min"
				type="number"
				bind:value={number_min}
				required
				oninput={clearLocalErrors}
			/>
			{#if getError('number_min')}
				<p class="text-sm text-red-500">{getError('number_min')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="system_type_max">Max Number</Label>
			<Input
				id="system_type_max"
				type="number"
				bind:value={number_max}
				required
				oninput={clearLocalErrors}
			/>
			{#if getError('number_max')}
				<p class="text-sm text-red-500">{getError('number_max')}</p>
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
