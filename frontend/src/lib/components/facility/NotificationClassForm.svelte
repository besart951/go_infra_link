<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { ManageNotificationClassUseCase } from '$lib/application/useCases/facility/manageNotificationClassUseCase.js';
	import { notificationClassRepository } from '$lib/infrastructure/api/notificationClassRepository.js';
	const manageNotificationClass = new ManageNotificationClassUseCase(notificationClassRepository);
	import { getErrorMessage, getFieldError, getFieldErrors } from '$lib/api/client.js';
	import type { NotificationClass } from '$lib/domain/facility/index.js';

	interface Props {
		initialData?: NotificationClass;
		onSuccess?: (nc: NotificationClass) => void;
		onCancel?: () => void;
	}

	let { initialData, onSuccess, onCancel }: Props = $props();

	let event_category = $state('');
	let nc = $state(0);
	let object_description = $state('');
	let internal_description = $state('');
	let meaning = $state('');
	let ack_required_not_normal = $state(false);
	let ack_required_error = $state(false);
	let ack_required_normal = $state(false);
	let norm_not_normal = $state(0);
	let norm_error = $state(0);
	let norm_normal = $state(0);

	let loading = $state(false);
	let error = $state('');
	let fieldErrors = $state<Record<string, string>>({});

	$effect(() => {
		if (initialData) {
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
	});

	const fieldError = (name: string) => getFieldError(fieldErrors, name, ['notificationclass']);

	async function handleSubmit() {
		loading = true;
		error = '';
		fieldErrors = {};

		try {
			if (initialData) {
				const res = await manageNotificationClass.update(initialData.id, {
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
				onSuccess?.(res);
			} else {
				const res = await manageNotificationClass.create({
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
				onSuccess?.(res);
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

<form
	onsubmit={(e) => {
		e.preventDefault();
		handleSubmit();
	}}
	class="space-y-4 rounded-md border bg-muted/20 p-4"
>
	<div class="mb-4 flex items-center justify-between">
		<h3 class="text-lg font-medium">
			{initialData ? 'Edit Notification Class' : 'New Notification Class'}
		</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="nc_event">Event Category</Label>
			<Input id="nc_event" bind:value={event_category} required />
			{#if fieldError('event_category')}
				<p class="text-sm text-red-500">{fieldError('event_category')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="nc_value">NC</Label>
			<Input id="nc_value" type="number" bind:value={nc} required />
			{#if fieldError('nc')}
				<p class="text-sm text-red-500">{fieldError('nc')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="nc_object_desc">Object Description</Label>
			<Textarea id="nc_object_desc" bind:value={object_description} rows={2} required />
			{#if fieldError('object_description')}
				<p class="text-sm text-red-500">{fieldError('object_description')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="nc_internal_desc">Internal Description</Label>
			<Textarea id="nc_internal_desc" bind:value={internal_description} rows={2} required />
			{#if fieldError('internal_description')}
				<p class="text-sm text-red-500">{fieldError('internal_description')}</p>
			{/if}
		</div>
		<div class="space-y-2 md:col-span-2">
			<Label for="nc_meaning">Meaning</Label>
			<Textarea id="nc_meaning" bind:value={meaning} rows={2} required />
			{#if fieldError('meaning')}
				<p class="text-sm text-red-500">{fieldError('meaning')}</p>
			{/if}
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
		{#if fieldError('ack_required_not_normal')}
			<p class="text-sm text-red-500 md:col-span-2">{fieldError('ack_required_not_normal')}</p>
		{/if}
		<div class="flex items-center gap-2">
			<input id="nc_ack_error" type="checkbox" bind:checked={ack_required_error} class="h-4 w-4" />
			<Label for="nc_ack_error">Ack Required (Error)</Label>
		</div>
		{#if fieldError('ack_required_error')}
			<p class="text-sm text-red-500 md:col-span-2">{fieldError('ack_required_error')}</p>
		{/if}
		<div class="flex items-center gap-2">
			<input
				id="nc_ack_normal"
				type="checkbox"
				bind:checked={ack_required_normal}
				class="h-4 w-4"
			/>
			<Label for="nc_ack_normal">Ack Required (Normal)</Label>
		</div>
		{#if fieldError('ack_required_normal')}
			<p class="text-sm text-red-500 md:col-span-2">{fieldError('ack_required_normal')}</p>
		{/if}
		<div class="space-y-2">
			<Label for="nc_norm_not_normal">Norm Not Normal</Label>
			<Input id="nc_norm_not_normal" type="number" bind:value={norm_not_normal} />
			{#if fieldError('norm_not_normal')}
				<p class="text-sm text-red-500">{fieldError('norm_not_normal')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="nc_norm_error">Norm Error</Label>
			<Input id="nc_norm_error" type="number" bind:value={norm_error} />
			{#if fieldError('norm_error')}
				<p class="text-sm text-red-500">{fieldError('norm_error')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="nc_norm_normal">Norm Normal</Label>
			<Input id="nc_norm_normal" type="number" bind:value={norm_normal} />
			{#if fieldError('norm_normal')}
				<p class="text-sm text-red-500">{fieldError('norm_normal')}</p>
			{/if}
		</div>
	</div>

	{#if error}
		<p class="text-sm text-red-500">{error}</p>
	{/if}

	<div class="flex justify-end gap-2 pt-2">
		<Button type="button" variant="ghost" onclick={onCancel}>Cancel</Button>
		<Button type="submit" disabled={loading}>{initialData ? 'Update' : 'Create'}</Button>
	</div>
</form>
