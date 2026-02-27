<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import FieldDeviceRow from '../FieldDeviceRow.svelte';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import { Plus, AlertCircle } from '@lucide/svelte';
	import { createTranslator } from '$lib/i18n/translator.js';

	import type {
		FieldDeviceRowData,
		FieldDeviceRowError
	} from '$lib/domain/facility/fieldDeviceMultiCreate.js';

	type Props = {
		rows: FieldDeviceRowData[];
		rowErrors: Map<number, FieldDeviceRowError>;
		availableNumbersCount: number;
		loadingAvailableNumbers: boolean;
		canAddRow: boolean;
		hasValidationErrors: boolean;
		submitting: boolean;
		onCancel?: () => void;
		onAddRow: () => void;
		onSubmit: () => void;
		onRowBmkChange: (index: number, value: string) => void;
		onRowDescriptionChange: (index: number, value: string) => void;
		onRowTextFixChange: (index: number, value: string) => void;
		onRowApparatNrChange: (index: number, value: string) => void;
		onRowRemove: (index: number) => void;
		getPlaceholderForRow: (index: number) => string;
	};

	let {
		rows,
		rowErrors,
		availableNumbersCount,
		loadingAvailableNumbers,
		canAddRow,
		hasValidationErrors,
		submitting,
		onCancel,
		onAddRow,
		onSubmit,
		onRowBmkChange,
		onRowDescriptionChange,
		onRowTextFixChange,
		onRowApparatNrChange,
		onRowRemove,
		getPlaceholderForRow
	}: Props = $props();

	const t = createTranslator();
</script>

<Card.Root class="p-6">
	<div class="mb-4 flex items-center justify-between">
		<div>
			<h3 class="text-lg font-semibold">{$t('field_device.multi_create.rows.title')}</h3>
			<p class="text-sm text-muted-foreground">
				{$t('field_device.multi_create.rows.description')}
			</p>
		</div>
		<Button onclick={onAddRow} disabled={!canAddRow} size="sm">
			<Plus class="mr-2 size-4" />
			{$t('field_device.multi_create.rows.add')}
		</Button>
	</div>

	{#if rows.length === 0}
		<Alert.Root>
			<AlertCircle class="size-4" />
			<Alert.Description>
				{#if availableNumbersCount === 0 && !loadingAvailableNumbers}
					{$t('field_device.multi_create.rows.none_available')}
				{:else if loadingAvailableNumbers}
					{$t('field_device.multi_create.rows.loading_numbers')}
				{:else}
					{$t('field_device.multi_create.rows.empty_prompt')}
				{/if}
			</Alert.Description>
		</Alert.Root>
	{/if}

	{#if rows.length > 0}
		<div class="space-y-4">
			{#each rows as row, index (row.id)}
				<FieldDeviceRow
					{index}
					{row}
					error={rowErrors.get(index) ?? null}
					placeholder={getPlaceholderForRow(index)}
					disabled={submitting}
					onBmkChange={(value) => onRowBmkChange(index, value)}
					onDescriptionChange={(value) => onRowDescriptionChange(index, value)}
					onTextFixChange={(value) => onRowTextFixChange(index, value)}
					onApparatNrChange={(value) => onRowApparatNrChange(index, value)}
					onRemove={() => onRowRemove(index)}
				/>
			{/each}
		</div>

		<Separator class="my-4" />

		<div class="flex items-center justify-between">
			<p class="text-sm text-muted-foreground">
				{$t('field_device.multi_create.rows.summary', { count: rows.length })}
				{#if hasValidationErrors}
					<span class="text-destructive">
						{$t('field_device.multi_create.rows.errors', { count: rowErrors.size })}
					</span>
				{/if}
			</p>
			<div class="flex gap-2">
				{#if onCancel}
					<Button variant="outline" onclick={onCancel} disabled={submitting}>
						{$t('common.cancel')}
					</Button>
				{/if}
				<Button
					onclick={onSubmit}
					disabled={submitting || rows.length === 0 || hasValidationErrors}
				>
					{submitting
						? $t('field_device.multi_create.actions.creating')
						: $t('field_device.multi_create.actions.create', { count: rows.length })}
				</Button>
			</div>
		</div>
	{/if}
</Card.Root>
