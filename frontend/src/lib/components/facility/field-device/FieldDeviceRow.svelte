<script lang="ts">
	/**
	 * Field Device Row Component
	 * Displays a single row in the multi-create form
	 */
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import { Trash2, AlertCircle } from '@lucide/svelte';
	import type { FieldDeviceRowData, FieldDeviceRowError } from '$lib/domain/facility/fieldDeviceMultiCreate.js';

	interface Props {
		index: number;
		row: FieldDeviceRowData;
		error: FieldDeviceRowError | null;
		placeholder: string;
		disabled?: boolean;
		onBmkChange: (value: string) => void;
		onDescriptionChange: (value: string) => void;
		onApparatNrChange: (value: string) => void;
		onRemove: () => void;
	}

	let {
		index,
		row,
		error,
		placeholder,
		disabled = false,
		onBmkChange,
		onDescriptionChange,
		onApparatNrChange,
		onRemove
	}: Props = $props();

	const hasApparatNrError = $derived(error?.field === 'apparat_nr');
</script>

<div class="rounded-lg border p-4">
	<div class="mb-3 flex items-center justify-between">
		<h4 class="font-medium">Field Device #{index + 1}</h4>
		<Button variant="ghost" size="sm" onclick={onRemove} {disabled}>
			<Trash2 class="size-4 text-destructive" />
		</Button>
	</div>

	<div class="grid gap-4 md:grid-cols-3">
		<!-- BMK -->
		<div class="space-y-2">
			<Label for={`bmk-${index}`}>BMK</Label>
			<Input
				id={`bmk-${index}`}
				value={row.bmk}
				oninput={(e) => onBmkChange((e.target as HTMLInputElement).value)}
				placeholder="BMK identifier (optional)"
				maxlength={10}
				{disabled}
			/>
		</div>

		<!-- Description -->
		<div class="space-y-2">
			<Label for={`description-${index}`}>Description</Label>
			<Input
				id={`description-${index}`}
				value={row.description}
				oninput={(e) => onDescriptionChange((e.target as HTMLInputElement).value)}
				placeholder="Description (optional)"
				maxlength={250}
				{disabled}
			/>
		</div>

		<!-- Apparat Nr -->
		<div class="space-y-2">
			<Label for={`apparat-nr-${index}`}>Apparat Nr *</Label>
			<Input
				id={`apparat-nr-${index}`}
				type="number"
				value={row.apparatNr?.toString() ?? ''}
				oninput={(e) => onApparatNrChange((e.target as HTMLInputElement).value)}
				{placeholder}
				min={1}
				max={99}
				{disabled}
				class={hasApparatNrError ? 'border-destructive' : ''}
			/>
			{#if hasApparatNrError && error}
				<p class="text-sm text-destructive">{error.message}</p>
			{/if}
		</div>
	</div>

	<!-- Row Error (non-apparat_nr errors) -->
	{#if error && error.field !== 'apparat_nr'}
		<Alert.Root variant="destructive" class="mt-3">
			<AlertCircle class="size-4" />
			<Alert.Description>{error.message}</Alert.Description>
		</Alert.Root>
	{/if}
</div>
