<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { ManageSystemTypeUseCase } from '$lib/application/useCases/facility/manageSystemTypeUseCase.js';
	import { systemTypeRepository } from '$lib/infrastructure/api/systemTypeRepository.js';
	const manageSystemType = new ManageSystemTypeUseCase(systemTypeRepository);
	import type { SystemType } from '$lib/domain/facility/index.js';
	import { useFormState } from '$lib/hooks/useFormState.svelte.js';

	interface SystemTypeFormProps {
		initialData?: SystemType;
		onSuccess?: (systemType: SystemType) => void;
		onCancel?: () => void;
	}

	let { initialData, onSuccess, onCancel }: SystemTypeFormProps = $props();

	const NUMBER_MIN_LIMIT = 1;
	const NUMBER_MAX_LIMIT = 9999;

	let name = $state('');
	let number_min = $state('');
	let number_max = $state('');

	let systemTypes = $state<SystemType[]>([]);
	let availableRanges = $state<{ start: number; end: number }[]>([]);
	let availableRangesLoading = $state(false);
	let availableRangesError = $state<string | null>(null);

	$effect(() => {
		if (initialData) {
			name = initialData.name;
			number_min = formatNumber(initialData.number_min);
			number_max = formatNumber(initialData.number_max);
		}
	});

	$effect(() => {
		if (systemTypes.length > 0) {
			availableRanges = calculateAvailableRanges(systemTypes, initialData?.id);
		}
	});

	const formState = useFormState({
		onSuccess: (result: SystemType) => {
			onSuccess?.(result);
		}
	});

	let localErrors = $state<Record<string, string>>({});

	function clearLocalErrors() {
		if (Object.keys(localErrors).length > 0) {
			localErrors = {};
		}
	}

	function validateLocal(): boolean {
		localErrors = {};
		const minValue = Number(number_min);
		const maxValue = Number(number_max);
		if (number_min.length !== 4) {
			localErrors = {
				...localErrors,
				number_min: 'Min number must be 4 digits.'
			};
		}
		if (number_max.length !== 4) {
			localErrors = {
				...localErrors,
				number_max: 'Max number must be 4 digits.'
			};
		}
		if (Number.isFinite(minValue) && (minValue < NUMBER_MIN_LIMIT || minValue > NUMBER_MAX_LIMIT)) {
			localErrors = {
				...localErrors,
				number_min: `Min number must be between ${formatNumber(NUMBER_MIN_LIMIT)} and ${formatNumber(NUMBER_MAX_LIMIT)}.`
			};
		}
		if (Number.isFinite(maxValue) && (maxValue < NUMBER_MIN_LIMIT || maxValue > NUMBER_MAX_LIMIT)) {
			localErrors = {
				...localErrors,
				number_max: `Max number must be between ${formatNumber(NUMBER_MIN_LIMIT)} and ${formatNumber(NUMBER_MAX_LIMIT)}.`
			};
		}
		if (Number.isFinite(minValue) && Number.isFinite(maxValue) && minValue >= maxValue) {
			localErrors = {
				...localErrors,
				number_max: 'Max number must be greater than min number.'
			};
		}
		return Object.keys(localErrors).length === 0;
	}

	function getError(field: string) {
		return localErrors[field] ?? formState.getFieldError(field, ['systemtype']);
	}

	function formatNumber(value: number): string {
		return String(value).padStart(4, '0');
	}

	function formatRange(range: { start: number; end: number }): string {
		return range.start === range.end
			? formatNumber(range.start)
			: `${formatNumber(range.start)}-${formatNumber(range.end)}`;
	}

	function normalizeNumber(value: string): string {
		const trimmed = value.trim();
		if (trimmed === '') return '';
		const parsed = Number(trimmed);
		if (!Number.isFinite(parsed)) return trimmed;
		return formatNumber(parsed);
	}

	function calculateAvailableRanges(items: SystemType[], excludeId?: string) {
		const occupied = items
			.filter((item) => item.id !== excludeId)
			.map((item) => ({
				start: Math.max(NUMBER_MIN_LIMIT, item.number_min),
				end: Math.min(NUMBER_MAX_LIMIT, item.number_max)
			}))
			.filter((range) => range.start <= range.end)
			.sort((a, b) => a.start - b.start);

		const merged: { start: number; end: number }[] = [];
		for (const range of occupied) {
			const last = merged[merged.length - 1];
			if (!last || range.start > last.end + 1) {
				merged.push({ ...range });
			} else if (range.end > last.end) {
				last.end = range.end;
			}
		}

		const available: { start: number; end: number }[] = [];
		let cursor = NUMBER_MIN_LIMIT;
		for (const range of merged) {
			if (range.start > cursor) {
				available.push({ start: cursor, end: range.start - 1 });
			}
			cursor = Math.max(cursor, range.end + 1);
		}
		if (cursor <= NUMBER_MAX_LIMIT) {
			available.push({ start: cursor, end: NUMBER_MAX_LIMIT });
		}
		return available;
	}

	async function loadAvailableRanges() {
		availableRangesLoading = true;
		availableRangesError = null;
		try {
			const collected: SystemType[] = [];
			let page = 1;
			let totalPages = 1;
			const limit = 200;
			while (page <= totalPages) {
				const result = await systemTypeRepository.list({
					pagination: { page, pageSize: limit },
					search: { text: '' }
				});
				collected.push(...result.items);
				totalPages = result.metadata.totalPages;
				page += 1;
			}
			systemTypes = collected;
			availableRanges = calculateAvailableRanges(collected, initialData?.id);
		} catch (error) {
			console.error('Failed to load system types:', error);
			availableRangesError = 'Failed to load available numbers.';
		} finally {
			availableRangesLoading = false;
		}
	}

	function handleNumberInput(event: Event, field: 'min' | 'max') {
		const target = event.currentTarget as HTMLInputElement;
		const sanitized = target.value.replace(/\D/g, '').slice(0, 4);
		if (field === 'min') {
			number_min = sanitized;
		} else {
			number_max = sanitized;
		}
		clearLocalErrors();
	}

	function handleNumberBlur(field: 'min' | 'max') {
		if (field === 'min') {
			number_min = normalizeNumber(number_min);
		} else {
			number_max = normalizeNumber(number_max);
		}
	}

	async function handleSubmit() {
		number_min = normalizeNumber(number_min);
		number_max = normalizeNumber(number_max);
		if (!validateLocal()) {
			return;
		}
		const minValue = Number(number_min);
		const maxValue = Number(number_max);
		await formState.handleSubmit(async () => {
			if (initialData) {
				return await manageSystemType.update(initialData.id, {
					name,
					number_min: minValue,
					number_max: maxValue
				});
			} else {
				return await manageSystemType.create({
					name,
					number_min: minValue,
					number_max: maxValue
				});
			}
		});
	}

	onMount(() => {
		loadAvailableRanges();
	});
</script>

<form
	onsubmit={(e) => {
		e.preventDefault();
		handleSubmit();
	}}
	class="space-y-4 rounded-md border bg-muted/20 p-4"
>
	<div class="mb-4 flex items-center justify-between">
		<h3 class="text-lg font-medium">{initialData ? 'Edit System Type' : 'New System Type'}</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-3">
		<div class="space-y-2 md:col-span-1">
			<Label for="system_type_name">Name</Label>
			<Input
				id="system_type_name"
				bind:value={name}
				required
				maxlength={150}
				oninput={clearLocalErrors}
			/>
			{#if getError('name')}
				<p class="text-sm text-red-500">{getError('name')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="system_type_min">Min Number</Label>
			<Input
				id="system_type_min"
				type="text"
				inputmode="numeric"
				pattern="[0-9]*"
				maxlength={4}
				placeholder="0001"
				value={number_min}
				min={NUMBER_MIN_LIMIT}
				max={NUMBER_MAX_LIMIT}
				required
				oninput={(event) => handleNumberInput(event, 'min')}
				onblur={() => handleNumberBlur('min')}
			/>
			{#if getError('number_min')}
				<p class="text-sm text-red-500">{getError('number_min')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="system_type_max">Max Number</Label>
			<Input
				id="system_type_max"
				type="text"
				inputmode="numeric"
				pattern="[0-9]*"
				maxlength={4}
				placeholder="9999"
				value={number_max}
				min={NUMBER_MIN_LIMIT}
				max={NUMBER_MAX_LIMIT}
				required
				oninput={(event) => handleNumberInput(event, 'max')}
				onblur={() => handleNumberBlur('max')}
			/>
			{#if getError('number_max')}
				<p class="text-sm text-red-500">{getError('number_max')}</p>
			{/if}
		</div>
	</div>

	{#if availableRangesLoading}
		<p class="text-xs text-muted-foreground">Loading available numbers...</p>
	{:else if availableRangesError}
		<p class="text-xs text-red-500">{availableRangesError}</p>
	{:else if availableRanges.length > 0}
		<p class="text-xs text-muted-foreground">
			Available numbers: {availableRanges.map(formatRange).join(', ')}
		</p>
	{:else}
		<p class="text-xs text-muted-foreground">
			No available numbers in {formatNumber(NUMBER_MIN_LIMIT)}-{formatNumber(NUMBER_MAX_LIMIT)}.
		</p>
	{/if}

	{#if formState.error}
		<p class="text-sm text-red-500">{formState.error}</p>
	{/if}

	<div class="flex justify-end gap-2 pt-2">
		<Button type="button" variant="ghost" onclick={onCancel}>Cancel</Button>
		<Button type="submit" disabled={formState.loading}>{initialData ? 'Update' : 'Create'}</Button>
	</div>
</form>
