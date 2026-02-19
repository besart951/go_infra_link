<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import type { AlarmTypeField, AlarmValueDraft } from '$lib/domain/facility/index.js';
	import { bacnetAlarmRepository } from '$lib/infrastructure/api/bacnetAlarmRepository.js';
	import { createTranslator } from '$lib/i18n/translator.js';

	interface Props {
		bacnetObjectId: string;
	}

	let { bacnetObjectId }: Props = $props();

	const t = createTranslator();

	let schema = $state<AlarmTypeField[]>([]);
	let values = $state<Record<string, AlarmValueDraft>>({});
	let saving = $state(false);
	let saveError = $state('');
	let loading = $state(false);

	$effect(() => {
		if (bacnetObjectId) {
			loadData();
		}
	});

	async function loadData() {
		loading = true;
		try {
			const [schemaResult, valuesResult] = await Promise.all([
				bacnetAlarmRepository.getSchema(bacnetObjectId),
				bacnetAlarmRepository.getValues(bacnetObjectId)
			]);
			schema = schemaResult?.fields ?? [];
			const newValues: Record<string, AlarmValueDraft> = {};
			for (const v of valuesResult) {
				newValues[v.alarm_type_field_id] = {
					alarm_type_field_id: v.alarm_type_field_id,
					value_number: v.value_number,
					value_integer: v.value_integer,
					value_boolean: v.value_boolean,
					value_string: v.value_string,
					value_json: v.value_json,
					unit_id: v.unit_id,
					source: v.source
				};
			}
			values = newValues;
		} catch (e) {
			console.error('Failed to load alarm schema/values:', e);
			schema = [];
			saveError = $t('common.error');
		} finally {
			loading = false;
		}
	}

	async function handleSave() {
		saving = true;
		saveError = '';
		try {
			const valueList = schema.map((field) => {
				const existing = values[field.id] ?? {};
				return {
					alarm_type_field_id: field.id,
					value_number: existing.value_number,
					value_integer: existing.value_integer,
					value_boolean: existing.value_boolean,
					value_string: existing.value_string,
					value_json: existing.value_json,
					unit_id: existing.unit_id,
					source: 'user'
				};
			});
			await bacnetAlarmRepository.putValues(bacnetObjectId, valueList);
		} catch (e: any) {
			saveError = e?.message ?? $t('common.error');
		} finally {
			saving = false;
		}
	}

	// Group fields by ui_group
	function groupFields(fields: AlarmTypeField[]): Record<string, AlarmTypeField[]> {
		const grouped: Record<string, AlarmTypeField[]> = {};
		for (const field of fields) {
			const group = field.ui_group ?? 'general';
			if (!grouped[group]) grouped[group] = [];
			grouped[group].push(field);
		}
		return grouped;
	}

	let groupedFields = $derived(groupFields(schema));

	let requiredTotal = $derived(schema.filter((f) => f.is_required).length);
	let requiredFilled = $derived(
		schema.filter((f) => {
			if (!f.is_required) return false;
			const v = values[f.id];
			if (!v) return false;
			const dataType = f.alarm_field?.data_type;
			if (dataType === 'boolean') return v.value_boolean !== undefined;
			if (dataType === 'number') return v.value_number !== undefined;
			if (dataType === 'integer') return v.value_integer !== undefined;
			return !!(v.value_string || v.value_json);
		}).length
	);
</script>

{#if loading}
	<div class="py-4 text-center text-sm text-muted-foreground">{$t('common.loading')}</div>
{:else if schema.length === 0}
	<div class="py-4 text-center text-sm text-muted-foreground">
		{$t('field_device.bacnet.alarm_editor.no_alarm')}
	</div>
{:else}
	<div class="space-y-4 p-3">
		<!-- Completeness badge -->
		<div class="flex items-center gap-2 text-sm">
			<span class="font-medium">{$t('field_device.bacnet.alarm_editor.title')}</span>
			<span
				class={`ml-auto rounded-full px-2 py-0.5 text-xs font-medium ${requiredFilled === requiredTotal ? 'bg-green-100 text-green-700' : 'bg-orange-100 text-orange-700'}`}
			>
				{$t('field_device.bacnet.alarm_editor.completeness', {
					filled: requiredFilled,
					total: requiredTotal
				})}
				{requiredFilled === requiredTotal
					? $t('field_device.bacnet.alarm_editor.complete')
					: $t('field_device.bacnet.alarm_editor.incomplete')}
			</span>
		</div>

		<!-- Fields grouped by ui_group -->
		{#each Object.entries(groupedFields) as [group, fields]}
			<fieldset class="rounded-md border p-3">
				<legend class="px-1 text-xs font-semibold text-muted-foreground">{group}</legend>
				<div class="space-y-3">
					{#each fields as field}
						{@const dataType = field.alarm_field?.data_type ?? 'string'}
						{@const label = field.alarm_field?.label ?? field.alarm_field?.key ?? field.id}
						<div class="space-y-1">
							<Label class="text-xs">
								{label}{#if field.is_required}<span class="text-destructive"> *</span>{/if}
							</Label>
							{#if dataType === 'number'}
								<Input
									type="number"
									step="any"
									value={values[field.id]?.value_number ?? ''}
									oninput={(e) => {
										const v = parseFloat((e.target as HTMLInputElement).value);
										values = {
											...values,
											[field.id]: { ...values[field.id], alarm_type_field_id: field.id, value_number: isNaN(v) ? undefined : v }
										};
									}}
									class="h-8 text-sm"
								/>
							{:else if dataType === 'integer'}
								<Input
									type="number"
									step="1"
									value={values[field.id]?.value_integer ?? ''}
									oninput={(e) => {
										const v = parseInt((e.target as HTMLInputElement).value, 10);
										values = {
											...values,
											[field.id]: { ...values[field.id], alarm_type_field_id: field.id, value_integer: isNaN(v) ? undefined : v }
										};
									}}
									class="h-8 text-sm"
								/>
							{:else if dataType === 'boolean'}
								<Checkbox
									checked={values[field.id]?.value_boolean ?? false}
									onCheckedChange={(checked) => {
										values = {
											...values,
											[field.id]: { ...values[field.id], alarm_type_field_id: field.id, value_boolean: !!checked }
										};
									}}
								/>
							{:else if dataType === 'json' || dataType === 'state_map'}
								<Textarea
									value={values[field.id]?.value_json ?? ''}
									oninput={(e) => {
										values = {
											...values,
											[field.id]: { ...values[field.id], alarm_type_field_id: field.id, value_json: (e.target as HTMLTextAreaElement).value }
										};
									}}
									rows={3}
									class="text-sm font-mono"
								/>
							{:else}
								<Input
									type="text"
									value={values[field.id]?.value_string ?? ''}
									oninput={(e) => {
										values = {
											...values,
											[field.id]: { ...values[field.id], alarm_type_field_id: field.id, value_string: (e.target as HTMLInputElement).value }
										};
									}}
									class="h-8 text-sm"
								/>
							{/if}
						</div>
					{/each}
				</div>
			</fieldset>
		{/each}

		{#if saveError}
			<p class="text-sm text-red-500">{saveError}</p>
		{/if}

		<div class="flex justify-end">
			<Button onclick={handleSave} disabled={saving}>
				{$t('field_device.bacnet.alarm_editor.save')}
			</Button>
		</div>
	</div>
{/if}
