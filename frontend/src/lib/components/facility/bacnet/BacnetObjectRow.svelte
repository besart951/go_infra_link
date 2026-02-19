<script lang="ts">
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Trash2 } from '@lucide/svelte';
	import { BACNET_SOFTWARE_TYPES, BACNET_HARDWARE_TYPES } from '$lib/domain/facility/index.js';
	import { createTranslator } from '$lib/i18n/translator.js';
	import AlarmTypeSelect from '$lib/components/facility/selects/AlarmTypeSelect.svelte';
	import { alarmTypeRepository } from '$lib/infrastructure/api/alarmTypeRepository.js';
	import type { AlarmTypeField } from '$lib/domain/facility/index.js';

	type BacnetRowErrors = Partial<
		Record<
			| 'text_fix'
			| 'description'
			| 'software_type'
			| 'software_number'
			| 'hardware_type'
			| 'hardware_quantity',
			string
		>
	>;

	interface Props {
		index: number;
		textFix: string;
		description?: string;
		gmsVisible?: boolean;
		optional?: boolean;
		textIndividual?: string;
		softwareType: string;
		softwareNumber: number;
		hardwareType: string;
		hardwareQuantity: number;
		alarmTypeId?: string;
		errors?: BacnetRowErrors;
		onRemove: () => void;
		onUpdate: (field: string, value: any) => void;
	}

	let {
		index,
		textFix = $bindable(),
		description = $bindable(),
		gmsVisible = $bindable(false),
		optional = $bindable(false),
		textIndividual = $bindable(''),
		softwareType = $bindable('ai'),
		softwareNumber = $bindable(1),
		hardwareType = $bindable('ai'),
		hardwareQuantity = $bindable(1),
		alarmTypeId = $bindable(),
		errors = {},
		onRemove,
		onUpdate
	}: Props = $props();

	const t = createTranslator();

	let textIndividualEnabled = $state(!!textIndividual);
	let prevGmsVisible = $state<boolean | null>(null);
	let prevOptional = $state<boolean | null>(null);
	let prevAlarmTypeId = $state<string | null>(null);
	let alarmTypeFields = $state<AlarmTypeField[]>([]);
	let alarmTypeFieldsLoading = $state(false);
	let alarmTypeFieldsError = $state('');
	const requiredAlarmTypeFields = $derived(alarmTypeFields.filter((field) => field.is_required));

	$effect(() => {
		if (prevGmsVisible === null) {
			prevGmsVisible = gmsVisible;
			return;
		}
		if (gmsVisible !== prevGmsVisible) {
			prevGmsVisible = gmsVisible;
			onUpdate('gms_visible', gmsVisible);
		}
	});

	$effect(() => {
		if (prevOptional === null) {
			prevOptional = optional;
			return;
		}
		if (optional !== prevOptional) {
			prevOptional = optional;
			onUpdate('optional', optional);
		}
	});

	$effect(() => {
		if (prevAlarmTypeId === null) {
			prevAlarmTypeId = alarmTypeId ?? '';
			return;
		}
		const current = alarmTypeId ?? '';
		if (current !== prevAlarmTypeId) {
			prevAlarmTypeId = current;
			onUpdate('alarm_type_id', current || null);
		}
	});

	$effect(() => {
		const selectedAlarmTypeId = alarmTypeId?.trim() ?? '';
		if (!selectedAlarmTypeId) {
			alarmTypeFields = [];
			alarmTypeFieldsError = '';
			alarmTypeFieldsLoading = false;
			return;
		}

		alarmTypeFieldsLoading = true;
		alarmTypeFieldsError = '';

		void alarmTypeRepository
			.getWithFields(selectedAlarmTypeId)
			.then((alarmType) => {
				alarmTypeFields = [...(alarmType.fields ?? [])].sort(
					(a, b) => (a.display_order ?? 0) - (b.display_order ?? 0)
				);
			})
			.catch(() => {
				alarmTypeFields = [];
				alarmTypeFieldsError = 'Felder konnten nicht geladen werden';
			})
			.finally(() => {
				alarmTypeFieldsLoading = false;
			});
	});

	$effect(() => {
		const value = textIndividualEnabled ? $t('field_device.bacnet.row.text_individual_value') : '';
		if (textIndividual !== value) {
			textIndividual = value;
			onUpdate('text_individual', textIndividual);
		}
	});

	$effect(() => {
		if (textIndividual && !textIndividualEnabled) {
			textIndividualEnabled = true;
		}
		if (!textIndividual && textIndividualEnabled) {
			textIndividualEnabled = false;
		}
	});
</script>

<div class="grid grid-cols-12 gap-2 rounded-md border p-3">
	<!-- Row number and remove button -->
	<div class="col-span-12 mb-2 flex items-center justify-between">
		<h4 class="text-sm font-semibold text-muted-foreground">
			{$t('field_device.bacnet.row.title', { index: index + 1 })}
		</h4>
		<Button variant="ghost" size="sm" onclick={onRemove} class="h-7 w-7 p-0">
			<Trash2 class="size-4 text-destructive" />
		</Button>
	</div>

	<!-- Text Fix -->
	<div class="col-span-12 space-y-1 md:col-span-6">
		<Label for="text_fix_{index}" class="text-xs">{$t('field_device.bacnet.row.text_fix')}</Label>
		<Input
			id="text_fix_{index}"
			bind:value={textFix}
			onchange={() => onUpdate('text_fix', textFix)}
			required
			maxlength={250}
			placeholder={$t('field_device.bacnet.row.text_fix_placeholder')}
			class="h-8 text-sm"
		/>
		{#if errors.text_fix}
			<p class="text-xs text-red-500">{errors.text_fix}</p>
		{/if}
	</div>

	<!-- Description -->
	<div class="col-span-12 space-y-1 md:col-span-6">
		<Label for="description_{index}" class="text-xs">
			{$t('field_device.bacnet.row.description')}
		</Label>
		<Input
			id="description_{index}"
			bind:value={description}
			onchange={() => onUpdate('description', description)}
			maxlength={250}
			placeholder={$t('field_device.bacnet.row.description_placeholder')}
			class="h-8 text-sm"
		/>
	</div>

	<!-- Software Group: Type + Number -->
	<div class="col-span-12 space-y-1 md:col-span-6">
		<Label class="text-xs">{$t('field_device.bacnet.row.software')}</Label>
		<div class="grid grid-cols-2 gap-2">
			<div class="space-y-1">
				<Label for="software_type_{index}" class="text-xs text-muted-foreground">
					{$t('field_device.bacnet.row.type')}
				</Label>
				<select
					id="software_type_{index}"
					bind:value={softwareType}
					onchange={() => onUpdate('software_type', softwareType)}
					required
					class="flex h-8 w-full rounded-md border border-input bg-background px-2 py-1 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
				>
					<option value="">{$t('field_device.bacnet.row.select')}</option>
					{#each BACNET_SOFTWARE_TYPES as type}
						<option value={type.value}>{type.label}</option>
					{/each}
				</select>
				{#if errors.software_type}
					<p class="text-xs text-red-500">{errors.software_type}</p>
				{/if}
			</div>
			<div class="space-y-1">
				<Label for="software_number_{index}" class="text-xs text-muted-foreground">
					{$t('field_device.bacnet.row.number')}
				</Label>
				<Input
					id="software_number_{index}"
					type="number"
					bind:value={softwareNumber}
					onchange={() => onUpdate('software_number', softwareNumber)}
					required
					min={0}
					max={65535}
					placeholder={$t('field_device.bacnet.row.software_number_placeholder')}
					class="h-8 text-sm"
				/>
				{#if errors.software_number}
					<p class="text-xs text-red-500">{errors.software_number}</p>
				{/if}
			</div>
		</div>
	</div>

	<!-- Hardware Group: Type + Quantity -->
	<div class="col-span-12 space-y-1 md:col-span-6">
		<Label class="text-xs">{$t('field_device.bacnet.row.hardware')}</Label>
		<div class="grid grid-cols-2 gap-2">
			<div class="space-y-1">
				<Label for="hardware_type_{index}" class="text-xs text-muted-foreground">
					{$t('field_device.bacnet.row.type')}
				</Label>
				<select
					id="hardware_type_{index}"
					bind:value={hardwareType}
					onchange={() => onUpdate('hardware_type', hardwareType)}
					class="flex h-8 w-full rounded-md border border-input bg-background px-2 py-1 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
				>
					<option value="">{$t('field_device.bacnet.row.select')}</option>
					{#each BACNET_HARDWARE_TYPES as type}
						<option value={type.value}>{type.label}</option>
					{/each}
				</select>
				{#if errors.hardware_type}
					<p class="text-xs text-red-500">{errors.hardware_type}</p>
				{/if}
			</div>
			<div class="space-y-1">
				<Label for="hardware_quantity_{index}" class="text-xs text-muted-foreground">
					{$t('field_device.bacnet.row.quantity')}
				</Label>
				<Input
					id="hardware_quantity_{index}"
					type="number"
					bind:value={hardwareQuantity}
					onchange={() => onUpdate('hardware_quantity', hardwareQuantity)}
					min={0}
					max={255}
					placeholder={$t('field_device.bacnet.row.hardware_quantity_placeholder')}
					class="h-8 text-sm"
				/>
				{#if errors.hardware_quantity}
					<p class="text-xs text-red-500">{errors.hardware_quantity}</p>
				{/if}
			</div>
		</div>
	</div>

	<!-- Checkboxes -->
	<div class="col-span-12 flex flex-wrap items-center gap-4 md:col-span-6">
		<div class="flex items-center gap-2">
			<Checkbox id="gms_visible_{index}" bind:checked={gmsVisible} />
			<Label for="gms_visible_{index}" class="cursor-pointer text-xs">
				{$t('field_device.bacnet.row.gms_visible')}
			</Label>
		</div>
		<div class="flex items-center gap-2">
			<Checkbox id="optional_{index}" bind:checked={optional} />
			<Label for="optional_{index}" class="cursor-pointer text-xs">
				{$t('field_device.bacnet.row.optional')}
			</Label>
		</div>
		<div class="flex items-center gap-2">
			<Checkbox id="text_individual_{index}" bind:checked={textIndividualEnabled} />
			<Label for="text_individual_{index}" class="cursor-pointer text-xs">
				{$t('field_device.bacnet.row.text_individual')}
			</Label>
		</div>
	</div>

	<!-- Alarm Type Section -->
	<div class="col-span-12 space-y-1 border-t pt-2 md:col-span-12">
		<Label class="text-xs">Alarmtyp</Label>
		<div class="space-y-2">
			<AlarmTypeSelect bind:value={alarmTypeId} width="w-full" />
			{#if alarmTypeId}
				<div class="flex justify-end">
					<Button
						variant="ghost"
						size="sm"
						onclick={() => {
							alarmTypeId = '';
						}}
						class="h-7 px-2 text-xs"
						title="Alarmtyp entfernen"
					>
						Alarmtyp entfernen
					</Button>
				</div>
			{/if}
		</div>

		{#if alarmTypeFieldsLoading}
			<p class="text-xs text-muted-foreground">Alarmfelder werden geladen…</p>
		{:else if alarmTypeFieldsError}
			<p class="text-xs text-red-500">{alarmTypeFieldsError}</p>
		{:else if requiredAlarmTypeFields.length > 0}
			<div class="rounded-md border bg-muted/30 p-2">
				<p class="mb-1 text-xs font-medium text-muted-foreground">Pflichtfelder</p>
				<div class="space-y-1">
					{#each requiredAlarmTypeFields as field (field.id)}
						<div class="flex items-center justify-between gap-2 text-xs">
							<span class="truncate">
								{field.alarm_field?.label ?? field.alarm_field_id}
								({field.alarm_field?.data_type ?? 'unknown'})
							</span>
							<span class="shrink-0 text-muted-foreground">Pflicht</span>
						</div>
					{/each}
				</div>
			</div>
		{:else if alarmTypeId}
			<p class="text-xs text-muted-foreground">Keine Pflichtfelder vorhanden</p>
		{/if}
	</div>
</div>
