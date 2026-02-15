<script lang="ts">
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import { Trash2 } from '@lucide/svelte';
	import { BACNET_SOFTWARE_TYPES, BACNET_HARDWARE_TYPES } from '$lib/domain/facility/index.js';

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
		textFixError?: string;
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
		textFixError,
		onRemove,
		onUpdate
	}: Props = $props();

	let textIndividualEnabled = $state(!!textIndividual);
	let prevGmsVisible = $state<boolean | null>(null);
	let prevOptional = $state<boolean | null>(null);

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
		const value = textIndividualEnabled ? 'Text Individuell' : '';
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
		<h4 class="text-sm font-semibold text-muted-foreground">BACnet Object #{index + 1}</h4>
		<Button variant="ghost" size="sm" onclick={onRemove} class="h-7 w-7 p-0">
			<Trash2 class="size-4 text-destructive" />
		</Button>
	</div>

	<!-- Text Fix -->
	<div class="col-span-12 space-y-1 md:col-span-6">
		<Label for="text_fix_{index}" class="text-xs">Text Fix *</Label>
		<Input
			id="text_fix_{index}"
			bind:value={textFix}
			onchange={() => onUpdate('text_fix', textFix)}
			required
			maxlength={250}
			placeholder="e.g., AI_001"
			class="h-8 text-sm"
		/>
		{#if textFixError}
			<p class="text-xs text-red-500">{textFixError}</p>
		{/if}
	</div>

	<!-- Description -->
	<div class="col-span-12 space-y-1 md:col-span-6">
		<Label for="description_{index}" class="text-xs">Description</Label>
		<Input
			id="description_{index}"
			bind:value={description}
			onchange={() => onUpdate('description', description)}
			maxlength={250}
			placeholder="Optional description"
			class="h-8 text-sm"
		/>
	</div>

	<!-- Software Group: Type + Number -->
	<div class="col-span-12 space-y-1 md:col-span-6">
		<Label class="text-xs">Software</Label>
		<div class="grid grid-cols-2 gap-2">
			<div class="space-y-1">
				<Label for="software_type_{index}" class="text-xs text-muted-foreground">Type *</Label>
				<select
					id="software_type_{index}"
					bind:value={softwareType}
					onchange={() => onUpdate('software_type', softwareType)}
					required
					class="flex h-8 w-full rounded-md border border-input bg-background px-2 py-1 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
				>
					<option value="">Select...</option>
					{#each BACNET_SOFTWARE_TYPES as type}
						<option value={type.value}>{type.label}</option>
					{/each}
				</select>
			</div>
			<div class="space-y-1">
				<Label for="software_number_{index}" class="text-xs text-muted-foreground">Number *</Label>
				<Input
					id="software_number_{index}"
					type="number"
					bind:value={softwareNumber}
					onchange={() => onUpdate('software_number', softwareNumber)}
					required
					min={0}
					max={65535}
					placeholder="0-65535"
					class="h-8 text-sm"
				/>
			</div>
		</div>
	</div>

	<!-- Hardware Group: Type + Quantity -->
	<div class="col-span-12 space-y-1 md:col-span-6">
		<Label class="text-xs">Hardware</Label>
		<div class="grid grid-cols-2 gap-2">
			<div class="space-y-1">
				<Label for="hardware_type_{index}" class="text-xs text-muted-foreground">Type *</Label>
				<select
					id="hardware_type_{index}"
					bind:value={hardwareType}
					onchange={() => onUpdate('hardware_type', hardwareType)}
					required
					class="flex h-8 w-full rounded-md border border-input bg-background px-2 py-1 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50"
				>
					<option value="">Select...</option>
					{#each BACNET_HARDWARE_TYPES as type}
						<option value={type.value}>{type.label}</option>
					{/each}
				</select>
			</div>
			<div class="space-y-1">
				<Label for="hardware_quantity_{index}" class="text-xs text-muted-foreground"
					>Quantity *</Label
				>
				<Input
					id="hardware_quantity_{index}"
					type="number"
					bind:value={hardwareQuantity}
					onchange={() => onUpdate('hardware_quantity', hardwareQuantity)}
					required
					min={1}
					max={255}
					placeholder="1-255"
					class="h-8 text-sm"
				/>
			</div>
		</div>
	</div>

	<!-- Checkboxes -->
	<div class="col-span-12 flex flex-wrap items-center gap-4 md:col-span-6">
		<div class="flex items-center gap-2">
			<Checkbox id="gms_visible_{index}" bind:checked={gmsVisible} />
			<Label for="gms_visible_{index}" class="cursor-pointer text-xs">GMS Visible</Label>
		</div>
		<div class="flex items-center gap-2">
			<Checkbox id="optional_{index}" bind:checked={optional} />
			<Label for="optional_{index}" class="cursor-pointer text-xs">Optional</Label>
		</div>
		<div class="flex items-center gap-2">
			<Checkbox id="text_individual_{index}" bind:checked={textIndividualEnabled} />
			<Label for="text_individual_{index}" class="cursor-pointer text-xs">Text Individual</Label>
		</div>
	</div>
</div>
