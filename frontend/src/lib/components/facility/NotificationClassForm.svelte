<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import {
		createNotificationClass,
		updateNotificationClass
	} from '$lib/infrastructure/api/facility.adapter.js';
	import { getErrorMessage } from '$lib/api/client.js';
	import type { NotificationClass } from '$lib/domain/facility/index.js';
	import { createEventDispatcher } from 'svelte';

	export let initialData: NotificationClass | undefined = undefined;

	let event_category = initialData?.event_category ?? '';
	let nc = initialData?.nc ?? 0;
	let object_description = initialData?.object_description ?? '';
	let internal_description = initialData?.internal_description ?? '';
	let meaning = initialData?.meaning ?? '';
	let ack_required_not_normal = initialData?.ack_required_not_normal ?? false;
	let ack_required_error = initialData?.ack_required_error ?? false;
	let ack_required_normal = initialData?.ack_required_normal ?? false;
	let norm_not_normal = initialData?.norm_not_normal ?? 0;
	let norm_error = initialData?.norm_error ?? 0;
	let norm_normal = initialData?.norm_normal ?? 0;

	let loading = false;
	let error = '';

	$: if (initialData) {
		event_category = initialData.event_category;
		nc = initialData.nc;
		object_description = initialData.object_description;
		internal_description = initialData.internal_description;
		meaning = initialData.meaning;
		ack_required_not_normal = initialData.ack_required_not_normal;
		ack_required_error = initialData.ack_required_error;
		ack_required_normal = initialData.ack_required_normal;
		norm_not_normal = initialData.norm_not_normal;
		norm_error = initialData.norm_error;
		norm_normal = initialData.norm_normal;
	}

	const dispatch = createEventDispatcher();

	async function handleSubmit() {
		loading = true;
		error = '';

		try {
			if (initialData) {
				const res = await updateNotificationClass(initialData.id, {
					event_category,
					nc,
					object_description,
					internal_description,
					meaning,
					ack_required_not_normal,
					ack_required_error,
					ack_required_normal,
					norm_not_normal,
					norm_error,
					norm_normal
				});
				dispatch('success', res);
			} else {
				const res = await createNotificationClass({
					event_category,
					nc,
					object_description,
					internal_description,
					meaning,
					ack_required_not_normal,
					ack_required_error,
					ack_required_normal,
					norm_not_normal,
					norm_error,
					norm_normal
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
		<h3 class="text-lg font-medium">
			{initialData ? 'Edit Notification Class' : 'New Notification Class'}
		</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="nc_event">Event Category</Label>
			<Input id="nc_event" bind:value={event_category} required />
		</div>
		<div class="space-y-2">
			<Label for="nc_value">NC</Label>
			<Input id="nc_value" type="number" bind:value={nc} required />
		</div>
		<div class="space-y-2">
			<Label for="nc_object_desc">Object Description</Label>
			<Textarea id="nc_object_desc" bind:value={object_description} rows={2} required />
		</div>
		<div class="space-y-2">
			<Label for="nc_internal_desc">Internal Description</Label>
			<Textarea id="nc_internal_desc" bind:value={internal_description} rows={2} required />
		</div>
		<div class="space-y-2 md:col-span-2">
			<Label for="nc_meaning">Meaning</Label>
			<Textarea id="nc_meaning" bind:value={meaning} rows={2} required />
		</div>
		<div class="flex items-center gap-2">
			<input
				id="nc_ack_not_normal"
				type="checkbox"
				bind:checked={ack_required_not_normal}
				class="h-4 w-4"
			/>
			<Label for="nc_ack_not_normal">Ack Required (Not Normal)</Label>
		</div>
		<div class="flex items-center gap-2">
			<input id="nc_ack_error" type="checkbox" bind:checked={ack_required_error} class="h-4 w-4" />
			<Label for="nc_ack_error">Ack Required (Error)</Label>
		</div>
		<div class="flex items-center gap-2">
			<input
				id="nc_ack_normal"
				type="checkbox"
				bind:checked={ack_required_normal}
				class="h-4 w-4"
			/>
			<Label for="nc_ack_normal">Ack Required (Normal)</Label>
		</div>
		<div class="space-y-2">
			<Label for="nc_norm_not_normal">Norm Not Normal</Label>
			<Input id="nc_norm_not_normal" type="number" bind:value={norm_not_normal} />
		</div>
		<div class="space-y-2">
			<Label for="nc_norm_error">Norm Error</Label>
			<Input id="nc_norm_error" type="number" bind:value={norm_error} />
		</div>
		<div class="space-y-2">
			<Label for="nc_norm_normal">Norm Normal</Label>
			<Input id="nc_norm_normal" type="number" bind:value={norm_normal} />
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
