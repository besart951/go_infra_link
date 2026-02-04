<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { RefreshCw } from '@lucide/svelte';

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
				od.description.toLowerCase().includes(query) ||
				od.version.toLowerCase().includes(query)
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
			error = e?.message ?? 'Failed to load selectable object data';
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
			<span class="flex-1 text-sm text-destructive">Failed to load object data: {error}</span>
			{#if canRetry}
				<Button variant="outline" size="sm" onclick={handleRetry} disabled={loading} class="gap-2">
					<RefreshCw class={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
					Retry ({maxRetries - retryCycles} left)
				</Button>
			{:else}
				<span class="text-xs text-muted-foreground"
					>Max retries reached. Please reload the page.</span
				>
			{/if}
		</div>
	{/if}

	<div class={className}>
		<div class="space-y-2">
			<Label for="fd-object-data">Object Data *</Label>
			<AsyncCombobox
				id="fd-object-data"
				placeholder="Select object data..."
				searchPlaceholder="Search object data..."
				emptyText={loading ? 'Loading...' : 'No object data found.'}
				fetcher={async (search: string) => {
					const q = search.toLowerCase();
					return objectDataItems.filter((i) => i.label.toLowerCase().includes(q));
				}}
				fetchById={async (id: string) => objectDataItems.find((i) => i.id === id) ?? null}
				labelKey="label"
				width="w-full"
				value={value.objectDataId}
				onValueChange={handleObjectDataChange}
				clearable
				clearText="Clear object data"
				disabled={disabled || loading}
			/>
			<p class="text-xs text-muted-foreground">{objectDataItems.length} option(s)</p>
		</div>

		<div class="space-y-2">
			<Label for="fd-apparat">Apparat *</Label>
			<AsyncCombobox
				id="fd-apparat"
				placeholder="Select apparat..."
				searchPlaceholder="Search apparats..."
				emptyText={loading ? 'Loading...' : 'No apparats found.'}
				fetcher={async (search: string) => {
					const q = search.toLowerCase();
					return apparatItems.filter((i) => i.label.toLowerCase().includes(q));
				}}
				fetchById={async (id: string) => apparatItems.find((i) => i.id === id) ?? null}
				labelKey="label"
				width="w-full"
				value={value.apparatId}
				onValueChange={handleApparatChange}
				clearable
				clearText="Clear apparat"
				disabled={disabled || loading}
			/>
			<p class="text-xs text-muted-foreground">{apparatItems.length} option(s)</p>
		</div>

		<div class="space-y-2">
			<Label for="fd-system-part">System Part *</Label>
			<AsyncCombobox
				id="fd-system-part"
				placeholder="Select system part..."
				searchPlaceholder="Search system parts..."
				emptyText={loading ? 'Loading...' : 'No system parts found.'}
				fetcher={async (search: string) => {
					const q = search.toLowerCase();
					return systemPartItems.filter((i) => i.label.toLowerCase().includes(q));
				}}
				fetchById={async (id: string) => systemPartItems.find((i) => i.id === id) ?? null}
				labelKey="label"
				width="w-full"
				value={value.systemPartId}
				onValueChange={handleSystemPartChange}
				clearable
				clearText="Clear system part"
				disabled={disabled || loading}
			/>
			<p class="text-xs text-muted-foreground">{systemPartItems.length} option(s)</p>
		</div>
	</div>
</div>
