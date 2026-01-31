<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import {
		createAlarmDefinition,
		updateAlarmDefinition
	} from '$lib/infrastructure/api/facility.adapter.js';
	import type { AlarmDefinition } from '$lib/domain/facility/index.js';
	import { useFormState } from '$lib/hooks/useFormState.svelte.js';

	interface AlarmDefinitionFormProps {
		initialData?: AlarmDefinition;
		onSuccess?: (alarmDefinition: AlarmDefinition) => void;
		onCancel?: () => void;
	}

	let { initialData, onSuccess, onCancel }: AlarmDefinitionFormProps = $props();

	let name = $state(initialData?.name ?? '');
	let alarm_note = $state(initialData?.alarm_note ?? '');

	$effect(() => {
		if (initialData) {
			name = initialData.name;
			alarm_note = initialData.alarm_note ?? '';
		}
	});

	const formState = useFormState({
		onSuccess: (result: AlarmDefinition) => {
			onSuccess?.(result);
		}
	});

	async function handleSubmit() {
		await formState.handleSubmit(async () => {
			if (initialData) {
				return await updateAlarmDefinition(initialData.id, {
					name,
					alarm_note: alarm_note || undefined
				});
			} else {
				return await createAlarmDefinition({
					name,
					alarm_note: alarm_note || undefined
				});
			}
		});
	}
</script>

<form on:submit|preventDefault={handleSubmit} class="space-y-4 rounded-md border bg-muted/20 p-4">
	<div class="mb-4 flex items-center justify-between">
		<h3 class="text-lg font-medium">
			{initialData ? 'Edit Alarm Definition' : 'New Alarm Definition'}
		</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="alarm_name">Name</Label>
			<Input id="alarm_name" bind:value={name} required />
			{#if formState.getFieldError('name', ['alarmdefinition'])}
				<p class="text-sm text-red-500">{formState.getFieldError('name', ['alarmdefinition'])}</p>
			{/if}
		</div>
		<div class="space-y-2 md:col-span-2">
			<Label for="alarm_note">Alarm Note</Label>
			<Textarea id="alarm_note" bind:value={alarm_note} rows={3} />
			{#if formState.getFieldError('alarm_note', ['alarmdefinition'])}
				<p class="text-sm text-red-500">
					{formState.getFieldError('alarm_note', ['alarmdefinition'])}
				</p>
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
