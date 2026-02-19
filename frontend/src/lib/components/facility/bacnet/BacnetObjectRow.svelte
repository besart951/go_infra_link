<script lang="ts">
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Trash2 } from '@lucide/svelte';
	import { BACNET_SOFTWARE_TYPES, BACNET_HARDWARE_TYPES } from '$lib/domain/facility/index.js';
	import { createTranslator } from '$lib/i18n/translator.js';
	import AlarmDefinitionSelect from '$lib/components/facility/selects/AlarmDefinitionSelect.svelte';

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
		alarmDefinitionId?: string;
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
		alarmDefinitionId = $bindable(''),
		errors = {},
		onRemove,
		onUpdate
	}: Props = $props();

	const t = createTranslator();

	let textIndividualEnabled = $state(!!textIndividual);
	let prevGmsVisible = $state<boolean | null>(null);
	let prevOptional = $state<boolean | null>(null);
	let prevAlarmDefinitionId = $state<string | null>(null);

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
		if (prevAlarmDefinitionId === null) {
			prevAlarmDefinitionId = alarmDefinitionId ?? '';
			return;
		}
		const current = alarmDefinitionId ?? '';
		if (current !== prevAlarmDefinitionId) {
			prevAlarmDefinitionId = current;
			onUpdate('alarm_definition_id', current || null);
		}
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

	<!-- Alarm Definition Section -->
	<div class="col-span-12 space-y-1 border-t pt-2 md:col-span-12">
		<Label class="text-xs">{$t('field_device.bacnet.row.alarm_definition')}</Label>
		<div class="flex items-center gap-2">
			<AlarmDefinitionSelect bind:value={alarmDefinitionId} width="w-full" />
			{#if alarmDefinitionId}
				<Button
					variant="ghost"
					size="sm"
					onclick={() => {
						alarmDefinitionId = '';
					}}
					class="h-8 w-8 shrink-0 p-0"
					title={$t('field_device.bacnet.row.alarm_definition_remove')}
				>
					✕
				</Button>
			{/if}
		</div>
	</div>
</div>
