<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import type { AlarmDefinition, AlarmType, AlarmTypeField } from '$lib/domain/facility/index.js';
	import { ManageEntityUseCase } from '$lib/application/useCases/manageEntityUseCase.js';
	import { alarmDefinitionRepository } from '$lib/infrastructure/api/alarmDefinitionRepository.js';
	import { alarmTypeRepository } from '$lib/infrastructure/api/alarmTypeRepository.js';
	const manageAlarmDefinition = new ManageEntityUseCase(alarmDefinitionRepository);
	import { useFormState } from '$lib/hooks/useFormState.svelte.js';
	import { createTranslator } from '$lib/i18n/translator.js';

	interface AlarmDefinitionFormProps {
		initialData?: AlarmDefinition;
		onSuccess?: (alarmDefinition: AlarmDefinition) => void;
		onCancel?: () => void;
	}

	let { initialData, onSuccess, onCancel }: AlarmDefinitionFormProps = $props();

	const t = createTranslator();

	let name = $state('');
	let alarm_note = $state('');
	let alarm_type_id = $state('');
	let typeFields = $state<AlarmTypeField[]>([]);

	$effect(() => {
		if (initialData) {
			name = initialData.name;
			alarm_note = initialData.alarm_note ?? '';
			alarm_type_id = initialData.alarm_type_id ?? '';
		}
	});

	$effect(() => {
		if (alarm_type_id) {
			alarmTypeRepository.getWithFields(alarm_type_id).then((t) => {
				typeFields = t.fields ?? [];
			}).catch(() => {
				typeFields = [];
			});
		} else {
			typeFields = [];
		}
	});

	async function alarmTypeFetcher(search: string) {
		const res = await alarmTypeRepository.list({ search, page: 1, pageSize: 20 });
		return res.items;
	}

	async function alarmTypeFetchById(id: string) {
		return alarmTypeRepository.get(id);
	}

	const formState = useFormState({
		onSuccess: (result: AlarmDefinition) => {
			onSuccess?.(result);
		}
	});

	async function handleSubmit() {
		await formState.handleSubmit(async () => {
			if (initialData) {
				return await manageAlarmDefinition.update(initialData.id, {
					name,
					alarm_note: alarm_note || undefined,
					alarm_type_id: alarm_type_id || undefined
				});
			} else {
				return await manageAlarmDefinition.create({
					name,
					alarm_note: alarm_note || undefined,
					alarm_type_id: alarm_type_id || undefined
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
				? $t('facility.forms.alarm_definition.title_edit')
				: $t('facility.forms.alarm_definition.title_new')}
		</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="alarm_name">{$t('common.name')}</Label>
			<Input id="alarm_name" bind:value={name} required />
			{#if formState.getFieldError('name', ['alarmdefinition'])}
				<p class="text-sm text-red-500">{formState.getFieldError('name', ['alarmdefinition'])}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label>{$t('facility.forms.alarm_definition.alarm_type_label')}</Label>
			<AsyncCombobox
				bind:value={alarm_type_id}
				fetcher={alarmTypeFetcher}
				fetchById={alarmTypeFetchById}
				labelKey="name"
				placeholder={$t('facility.forms.alarm_definition.alarm_type_placeholder')}
				width="w-full"
			/>
		</div>
		<div class="space-y-2 md:col-span-2">
			<Label for="alarm_note">{$t('facility.forms.alarm_definition.note_label')}</Label>
			<Textarea id="alarm_note" bind:value={alarm_note} rows={3} />
			{#if formState.getFieldError('alarm_note', ['alarmdefinition'])}
				<p class="text-sm text-red-500">
					{formState.getFieldError('alarm_note', ['alarmdefinition'])}
				</p>
			{/if}
		</div>
	</div>

	{#if typeFields.length > 0}
		<div class="space-y-2">
			<p class="text-sm font-medium">{$t('facility.forms.alarm_definition.fields_preview_title')}</p>
			<div class="rounded-md border p-3 text-sm">
				{#each typeFields as field}
					<div class="flex items-center gap-2 py-1">
						<span class="font-mono text-xs text-muted-foreground">{field.alarm_field?.key}</span>
						<span>{field.alarm_field?.label}</span>
						<span class="ml-auto text-xs text-muted-foreground">{field.alarm_field?.data_type}</span>
						{#if field.is_required}
							<span class="rounded bg-destructive/10 px-1 text-xs text-destructive">{$t('common.required')}</span>
						{/if}
					</div>
				{/each}
			</div>
		</div>
	{/if}

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
