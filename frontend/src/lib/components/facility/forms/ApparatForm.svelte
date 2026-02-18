<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import type { Apparat } from '$lib/domain/facility/index.js';
	import { ManageEntityUseCase } from '$lib/application/useCases/manageEntityUseCase.js';
	import { apparatRepository } from '$lib/infrastructure/api/apparatRepository.js';
	const manageApparat = new ManageEntityUseCase(apparatRepository);
	import { useFormState } from '$lib/hooks/useFormState.svelte.js';
	import SystemPartMultiSelect from '../selects/SystemPartMultiSelect.svelte';
	import { createTranslator } from '$lib/i18n/translator.js';

	interface ApparatFormProps {
		initialData?: Apparat;
		onSuccess?: (apparat: Apparat) => void;
		onCancel?: () => void;
	}

	let { initialData, onSuccess, onCancel }: ApparatFormProps = $props();

	const t = createTranslator();

	let short_name = $state('');
	let name = $state('');
	let description = $state('');
	let system_part_ids = $state<string[]>([]);
	let shortNameError = $state('');

	$effect(() => {
		if (initialData) {
			short_name = initialData.short_name;
			name = initialData.name;
			description = initialData.description ?? '';
			system_part_ids = initialData.system_parts?.map((sp) => sp.id) ?? [];
		}
	});

	const formState = useFormState({
		onSuccess: (result: Apparat) => {
			onSuccess?.(result);
		}
	});

	async function handleSubmit() {
		const trimmedShortName = short_name.trim();
		if (trimmedShortName.length !== 3) {
			shortNameError = $t('facility.forms.apparat.short_name_length');
			return;
		}
		shortNameError = '';

		await formState.handleSubmit(async () => {
			if (initialData) {
				return await manageApparat.update(initialData.id, {
					short_name: trimmedShortName,
					name,
					description: description || undefined,
					system_part_ids
				});
			} else {
				return await manageApparat.create({
					short_name: trimmedShortName,
					name,
					description: description || undefined,
					system_part_ids
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
				? $t('facility.forms.apparat.title_edit')
				: $t('facility.forms.apparat.title_new')}
		</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="apparat_short">{$t('facility.forms.apparat.short_name_label')}</Label>
			<Input id="apparat_short" bind:value={short_name} required minlength={3} maxlength={3} />
			{#if shortNameError}
				<p class="text-sm text-red-500">{shortNameError}</p>
			{:else if formState.getFieldError('short_name', ['apparat'])}
				<p class="text-sm text-red-500">
					{formState.getFieldError('short_name', ['apparat'])}
				</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="apparat_name">{$t('common.name')}</Label>
			<Input id="apparat_name" bind:value={name} required maxlength={250} />
			{#if formState.getFieldError('name', ['apparat'])}
				<p class="text-sm text-red-500">{formState.getFieldError('name', ['apparat'])}</p>
			{/if}
		</div>
		<div class="space-y-2 md:col-span-2">
			<Label for="apparat_desc">{$t('common.description')}</Label>
			<Textarea id="apparat_desc" bind:value={description} rows={3} maxlength={250} />
			{#if formState.getFieldError('description', ['apparat'])}
				<p class="text-sm text-red-500">{formState.getFieldError('description', ['apparat'])}</p>
			{/if}
		</div>
		<div class="space-y-2 md:col-span-2">
			<Label for="apparat_system_parts">{$t('facility.forms.apparat.system_parts_label')}</Label>
			<SystemPartMultiSelect id="apparat_system_parts" bind:value={system_part_ids} />
			{#if formState.getFieldError('system_part_ids', ['apparat'])}
				<p class="text-sm text-red-500">
					{formState.getFieldError('system_part_ids', ['apparat'])}
				</p>
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
