<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { createSystemPart, updateSystemPart } from '$lib/infrastructure/api/facility.adapter.js';
	import type { SystemPart } from '$lib/domain/facility/index.js';
	import { useFormState } from '$lib/hooks/useFormState.svelte.js';

	interface SystemPartFormProps {
		initialData?: SystemPart;
		onSuccess?: (systemPart: SystemPart) => void;
		onCancel?: () => void;
	}

	let { initialData, onSuccess, onCancel }: SystemPartFormProps = $props();

	let short_name = $state(initialData?.short_name ?? '');
	let name = $state(initialData?.name ?? '');
	let description = $state(initialData?.description ?? '');

	$effect(() => {
		if (initialData) {
			short_name = initialData.short_name;
			name = initialData.name;
			description = initialData.description ?? '';
		}
	});

	const formState = useFormState({
		onSuccess: (result: SystemPart) => {
			onSuccess?.(result);
		}
	});

	async function handleSubmit() {
		await formState.handleSubmit(async () => {
			if (initialData) {
				return await updateSystemPart(initialData.id, {
					short_name,
					name,
					description: description || undefined
				});
			} else {
				return await createSystemPart({
					short_name,
					name,
					description: description || undefined
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
		<h3 class="text-lg font-medium">{initialData ? 'Edit System Part' : 'New System Part'}</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="system_part_short">Short Name</Label>
			<Input id="system_part_short" bind:value={short_name} required maxlength={10} />
			{#if formState.getFieldError('short_name', ['systempart'])}
				<p class="text-sm text-red-500">{formState.getFieldError('short_name', ['systempart'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="system_part_name">Name</Label>
			<Input id="system_part_name" bind:value={name} required maxlength={250} />
			{#if formState.getFieldError('name', ['systempart'])}
				<p class="text-sm text-red-500">{formState.getFieldError('name', ['systempart'])}</p>
			{/if}
		</div>
		<div class="space-y-2 md:col-span-2">
			<Label for="system_part_desc">Description</Label>
			<Textarea id="system_part_desc" bind:value={description} rows={3} maxlength={250} />
			{#if formState.getFieldError('description', ['systempart'])}
				<p class="text-sm text-red-500">{formState.getFieldError('description', ['systempart'])}</p>
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
