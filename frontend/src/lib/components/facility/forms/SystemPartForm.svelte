<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import type { SystemPart } from '$lib/domain/facility/index.js';
	import { ManageEntityUseCase } from '$lib/application/useCases/manageEntityUseCase.js';
	import { systemPartRepository } from '$lib/infrastructure/api/systemPartRepository.js';
	const manageSystemPart = new ManageEntityUseCase(systemPartRepository);
	import { useFormState } from '$lib/hooks/useFormState.svelte.js';
	import { createTranslator } from '$lib/i18n/translator.js';

	interface SystemPartFormProps {
		initialData?: SystemPart;
		onSuccess?: (systemPart: SystemPart) => void;
		onCancel?: () => void;
	}

	let { initialData, onSuccess, onCancel }: SystemPartFormProps = $props();

	const t = createTranslator();

	let short_name = $state('');
	let name = $state('');
	let description = $state('');
	let shortNameError = $state('');

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
		const trimmedShortName = short_name.trim();
		if (trimmedShortName.length !== 3) {
			shortNameError = $t('facility.forms.system_part.short_name_length');
			return;
		}
		shortNameError = '';

		await formState.handleSubmit(async () => {
			if (initialData) {
				return await manageSystemPart.update(initialData.id, {
					short_name: trimmedShortName,
					name,
					description: description || undefined
				});
			} else {
				return await manageSystemPart.create({
					short_name: trimmedShortName,
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
		<h3 class="text-lg font-medium">
			{initialData
				? $t('facility.forms.system_part.title_edit')
				: $t('facility.forms.system_part.title_new')}
		</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="system_part_short">{$t('facility.forms.system_part.short_name_label')}</Label>
			<Input id="system_part_short" bind:value={short_name} required minlength={3} maxlength={3} />
			{#if shortNameError}
				<p class="text-sm text-red-500">{shortNameError}</p>
			{:else if formState.getFieldError('short_name', ['systempart'])}
				<p class="text-sm text-red-500">
					{formState.getFieldError('short_name', ['systempart'])}
				</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="system_part_name">{$t('common.name')}</Label>
			<Input id="system_part_name" bind:value={name} required maxlength={250} />
			{#if formState.getFieldError('name', ['systempart'])}
				<p class="text-sm text-red-500">{formState.getFieldError('name', ['systempart'])}</p>
			{/if}
		</div>
		<div class="space-y-2 md:col-span-2">
			<Label for="system_part_desc">{$t('common.description')}</Label>
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
		<Button type="button" variant="ghost" onclick={onCancel}>{$t('common.cancel')}</Button>
		<Button type="submit" disabled={formState.loading}>
			{initialData ? $t('common.update') : $t('common.create')}
		</Button>
	</div>
</form>
