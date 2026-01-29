<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import {
		createAlarmDefinition,
		updateAlarmDefinition
	} from '$lib/infrastructure/api/facility.adapter.js';
	import { getErrorMessage, getFieldError, getFieldErrors } from '$lib/api/client.js';
	import type { AlarmDefinition } from '$lib/domain/facility/index.js';
	import { createEventDispatcher } from 'svelte';

	export let initialData: AlarmDefinition | undefined = undefined;

	let name = initialData?.name ?? '';
	let alarm_note = initialData?.alarm_note ?? '';
	let loading = false;
	let error = '';
	let fieldErrors: Record<string, string> = {};

	$: if (initialData) {
		name = initialData.name;
		alarm_note = initialData.alarm_note ?? '';
	}

	const dispatch = createEventDispatcher();

	const fieldError = (name: string) => getFieldError(fieldErrors, name, ['alarmdefinition']);

	async function handleSubmit() {
		loading = true;
		error = '';
		fieldErrors = {};

		try {
			if (initialData) {
				const res = await updateAlarmDefinition(initialData.id, {
					name,
					alarm_note: alarm_note || undefined
				});
				dispatch('success', res);
			} else {
				const res = await createAlarmDefinition({
					name,
					alarm_note: alarm_note || undefined
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
		<h3 class="text-lg font-medium">
			{initialData ? 'Edit Alarm Definition' : 'New Alarm Definition'}
		</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="alarm_name">Name</Label>
			<Input id="alarm_name" bind:value={name} required />
			{#if fieldError('name')}
				<p class="text-sm text-red-500">{fieldError('name')}</p>
			{/if}
		</div>
		<div class="space-y-2 md:col-span-2">
			<Label for="alarm_note">Alarm Note</Label>
			<Textarea id="alarm_note" bind:value={alarm_note} rows={3} />
			{#if fieldError('alarm_note')}
				<p class="text-sm text-red-500">{fieldError('alarm_note')}</p>
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
