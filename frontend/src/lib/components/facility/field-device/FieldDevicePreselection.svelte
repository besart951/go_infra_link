<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { RefreshCw } from '@lucide/svelte';
	import { createTranslator } from '$lib/i18n/translator.js';
	import { t as translate } from '$lib/i18n/index.js';

	import type {
		Apparat,
		FieldDeviceOptions,
		ObjectData,
		SystemPart
	} from '$lib/domain/facility/index.js';
	import {
		getFilteredFieldDevicePreselectionOptions,
		type FieldDevicePreselection
	} from '$lib/domain/facility/preselectionFilter.js';
	import { GetFieldDeviceOptionsUseCase } from '$lib/application/useCases/facility/getFieldDeviceOptionsUseCase.js';
	import { facilityFieldDeviceOptionsRepository } from '$lib/infrastructure/api/facilityFieldDeviceOptionsRepository.js';

	type Props = {
		value: FieldDevicePreselection;
		onChange: (next: FieldDevicePreselection) => void;
		projectId?: string;
		disabled?: boolean;
		className?: string;
	};

	let {
		value,
		onChange,
		projectId,
		disabled = false,
		className = 'grid grid-cols-1 gap-4 md:grid-cols-3'
	}: Props = $props();

	const t = createTranslator();

	const useCase = new GetFieldDeviceOptionsUseCase(facilityFieldDeviceOptionsRepository);

	let options = $state<FieldDeviceOptions | null>(null);
	let loading = $state(false);
	let error = $state<string | null>(null);

	// Search queries for client-side filtering
	let objectDataSearch = $state('');
	let apparatSearch = $state('');
	let systemPartSearch = $state('');

	const maxRetries = 3;
	let retryCycles = $state(0);
	let hasFetched = $state(false);
	let initialFetchTriggered = $state(false);
	let abortController: AbortController | null = null;

	const canRetry = $derived(retryCycles < maxRetries);

	function filterByProject(objectDatas: ObjectData[]): ObjectData[] {
		// When using project-specific endpoint, all data is already filtered
		// No need for additional client-side project filtering
		return objectDatas;
	}

	// Apply search filter to object datas
	const searchFilteredObjectDatas = $derived.by(() => {
		const base = filtered.objectDatas;
		if (!objectDataSearch.trim()) return base;
		const query = objectDataSearch.toLowerCase();
		return base.filter(
			(od) =>
				od.description.toLowerCase().includes(query) || od.version.toLowerCase().includes(query)
		);
	});

	// Apply search filter to apparats
	const searchFilteredApparats = $derived.by(() => {
		const base = filtered.apparats;
		if (!apparatSearch.trim()) return base;
		const query = apparatSearch.toLowerCase();
		return base.filter(
			(a) =>
				a.short_name.toLowerCase().includes(query) ||
				a.name.toLowerCase().includes(query) ||
				(a.description && a.description.toLowerCase().includes(query))
		);
	});

	// Apply search filter to system parts
	const searchFilteredSystemParts = $derived.by(() => {
		const base = filtered.systemParts;
		if (!systemPartSearch.trim()) return base;
		const query = systemPartSearch.toLowerCase();
		return base.filter(
			(sp) =>
				sp.short_name.toLowerCase().includes(query) ||
				sp.name.toLowerCase().includes(query) ||
				(sp.description && sp.description.toLowerCase().includes(query))
		);
	});

	const filtered = $derived.by(() => {
		if (!options) {
			return { objectDatas: [], apparats: [], systemParts: [] };
		}

		const filteredOptions = getFilteredFieldDevicePreselectionOptions(options, value);

		return {
			objectDatas: filterByProject(filteredOptions.objectDatas),
			apparats: filteredOptions.apparats,
			systemParts: filteredOptions.systemParts
		};
	});

	function getAllowedForSelection(selection: FieldDevicePreselection) {
		if (!options) {
			return { objectDatas: [], apparats: [], systemParts: [] };
		}

		const filteredOptions = getFilteredFieldDevicePreselectionOptions(options, selection);
		return {
			objectDatas: filterByProject(filteredOptions.objectDatas),
			apparats: filteredOptions.apparats,
			systemParts: filteredOptions.systemParts
		};
	}

	function normalizeSelection(next: FieldDevicePreselection): FieldDevicePreselection {
		if (!options) return next;

		let current = next;
		for (let i = 0; i < 3; i += 1) {
			const allowed = getAllowedForSelection(current);
			const normalized: FieldDevicePreselection = {
				objectDataId:
					current.objectDataId && allowed.objectDatas.some((od) => od.id === current.objectDataId)
						? current.objectDataId
						: '',
				apparatId:
					current.apparatId && allowed.apparats.some((a) => a.id === current.apparatId)
						? current.apparatId
						: '',
				systemPartId:
					current.systemPartId && allowed.systemParts.some((sp) => sp.id === current.systemPartId)
						? current.systemPartId
						: ''
			};

			if (
				normalized.objectDataId === current.objectDataId &&
				normalized.apparatId === current.apparatId &&
				normalized.systemPartId === current.systemPartId
			) {
				return normalized;
			}
			current = normalized;
		}
		return current;
	}

	function applyChange(partial: Partial<FieldDevicePreselection>) {
		const next = normalizeSelection({ ...value, ...partial });
		onChange(next);
	}

	// View models for nicer labels
	type ObjectDataItem = { id: string; label: string; raw: ObjectData };
	type ApparatItem = { id: string; label: string; raw: Apparat };
	type SystemPartItem = { id: string; label: string; raw: SystemPart };

	// Full items from options (for fetchById lookup)
	const allObjectDataItems = $derived.by((): ObjectDataItem[] =>
		(options?.object_datas ?? []).map((od) => ({
			id: od.id,
			label: `${od.description} (v${od.version})`,
			raw: od
		}))
	);

	const allApparatItems = $derived.by((): ApparatItem[] =>
		(options?.apparats ?? []).map((a) => ({
			id: a.id,
			label: `${a.short_name} - ${a.name}`,
			raw: a
		}))
	);

	const allSystemPartItems = $derived.by((): SystemPartItem[] =>
		(options?.system_parts ?? []).map((sp) => ({
			id: sp.id,
			label: `${sp.short_name} - ${sp.name}`,
			raw: sp
		}))
	);

	// Filtered items for dropdown display
	const objectDataItems = $derived.by((): ObjectDataItem[] =>
		searchFilteredObjectDatas.map((od) => ({
			id: od.id,
			label: `${od.description} (v${od.version})`,
			raw: od
		}))
	);

	const apparatItems = $derived.by((): ApparatItem[] =>
		searchFilteredApparats.map((a) => ({
			id: a.id,
			label: `${a.short_name} - ${a.name}`,
			raw: a
		}))
	);

	const systemPartItems = $derived.by((): SystemPartItem[] =>
		searchFilteredSystemParts.map((sp) => ({
			id: sp.id,
			label: `${sp.short_name} - ${sp.name}`,
			raw: sp
		}))
	);

	async function fetchOptions(isUserRetry: boolean) {
		if (loading) return;
		loading = true;
		error = null;

		abortController?.abort();
		abortController = new AbortController();

		try {
			// Load options based on whether we're in a project context or not
			let res: FieldDeviceOptions;
			if (projectId) {
				res = await useCase.executeForProject(projectId, abortController.signal);
			} else {
				res = await useCase.execute(abortController.signal);
			}
			options = res;
			hasFetched = true;
		} catch (e: any) {
			if (e instanceof DOMException && e.name === 'AbortError') return;
			error = e?.message ?? translate('field_device.preselection.errors.load');
			if (isUserRetry) {
				retryCycles = Math.min(maxRetries, retryCycles);
			}
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		if (initialFetchTriggered) return;
		if (loading) return;
		if (hasFetched) return;
		initialFetchTriggered = true;
		fetchOptions(false);
	});

	function handleRetry() {
		if (!canRetry || loading) return;
		retryCycles += 1;
		fetchOptions(true);
	}

	function handleObjectDataChange(objectDataId: string) {
		applyChange({ objectDataId });
	}

	function handleApparatChange(apparatId: string) {
		applyChange({ apparatId });
	}

	function handleSystemPartChange(systemPartId: string) {
		applyChange({ systemPartId });
	}
</script>

<div class="space-y-4" class:pointer-events-none={disabled} class:opacity-60={disabled}>
	{#if error && !loading}
		<div
			class="flex items-center gap-3 rounded-md border border-destructive/50 bg-destructive/10 p-3"
		>
			<span class="flex-1 text-sm text-destructive">
				{$t('field_device.preselection.errors.load_with_message', { message: error })}
			</span>
			{#if canRetry}
				<Button variant="outline" size="sm" onclick={handleRetry} disabled={loading} class="gap-2">
					<RefreshCw class={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
					{$t('field_device.preselection.actions.retry', { count: maxRetries - retryCycles })}
				</Button>
			{:else}
				<span class="text-xs text-muted-foreground"
					>{$t('field_device.preselection.errors.max_retries')}</span
				>
			{/if}
		</div>
	{/if}

	<div class={className}>
		<div class="space-y-2">
			<Label for="fd-object-data">{$t('field_device.preselection.object_data')}</Label>
			<AsyncCombobox
				id="fd-object-data"
				placeholder={$t('field_device.preselection.object_data_placeholder')}
				searchPlaceholder={$t('field_device.preselection.object_data_search')}
				emptyText={loading
					? $t('field_device.preselection.loading')
					: $t('field_device.preselection.object_data_empty')}
				fetcher={async (search: string) => {
					const q = search.toLowerCase();
					return objectDataItems.filter((i) => i.label.toLowerCase().includes(q));
				}}
				fetchById={async (id: string) => allObjectDataItems.find((i) => i.id === id) ?? null}
				labelKey="label"
				width="w-full"
				value={value.objectDataId}
				onValueChange={handleObjectDataChange}
				clearable
				clearText={$t('field_device.preselection.object_data_clear')}
				disabled={disabled || loading}
			/>
			<p class="text-xs text-muted-foreground">
				{$t('field_device.preselection.options', { count: objectDataItems.length })}
			</p>
		</div>

		<div class="space-y-2">
			<Label for="fd-apparat">{$t('field_device.preselection.apparat')}</Label>
			<AsyncCombobox
				id="fd-apparat"
				placeholder={$t('field_device.preselection.apparat_placeholder')}
				searchPlaceholder={$t('field_device.preselection.apparat_search')}
				emptyText={loading
					? $t('field_device.preselection.loading')
					: $t('field_device.preselection.apparat_empty')}
				fetcher={async (search: string) => {
					const q = search.toLowerCase();
					return apparatItems.filter((i) => i.label.toLowerCase().includes(q));
				}}
				fetchById={async (id: string) => allApparatItems.find((i) => i.id === id) ?? null}
				labelKey="label"
				width="w-full"
				value={value.apparatId}
				onValueChange={handleApparatChange}
				clearable
				clearText={$t('field_device.preselection.apparat_clear')}
				disabled={disabled || loading}
			/>
			<p class="text-xs text-muted-foreground">
				{$t('field_device.preselection.options', { count: apparatItems.length })}
			</p>
		</div>

		<div class="space-y-2">
			<Label for="fd-system-part">{$t('field_device.preselection.system_part')}</Label>
			<AsyncCombobox
				id="fd-system-part"
				placeholder={$t('field_device.preselection.system_part_placeholder')}
				searchPlaceholder={$t('field_device.preselection.system_part_search')}
				emptyText={loading
					? $t('field_device.preselection.loading')
					: $t('field_device.preselection.system_part_empty')}
				fetcher={async (search: string) => {
					const q = search.toLowerCase();
					return systemPartItems.filter((i) => i.label.toLowerCase().includes(q));
				}}
				fetchById={async (id: string) => allSystemPartItems.find((i) => i.id === id) ?? null}
				labelKey="label"
				width="w-full"
				value={value.systemPartId}
				onValueChange={handleSystemPartChange}
				clearable
				clearText={$t('field_device.preselection.system_part_clear')}
				disabled={disabled || loading}
			/>
			<p class="text-xs text-muted-foreground">
				{$t('field_device.preselection.options', { count: systemPartItems.length })}
			</p>
		</div>
	</div>
</div>
