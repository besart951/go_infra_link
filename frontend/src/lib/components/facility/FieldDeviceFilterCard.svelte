<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { X } from '@lucide/svelte';
	import BuildingSelect from '$lib/components/facility/BuildingSelect.svelte';
	import ControlCabinetSelect from '$lib/components/facility/ControlCabinetSelect.svelte';
	import SPSControllerSelect from '$lib/components/facility/SPSControllerSelect.svelte';
	import SPSControllerSystemTypeSelect from '$lib/components/facility/SPSControllerSystemTypeSelect.svelte';
	import ProjectSelect from '$lib/components/project/ProjectSelect.svelte';
	import type { FieldDeviceFilters } from '$lib/stores/facility/fieldDeviceStore.js';

	interface Props {
		onApplyFilters: (filters: FieldDeviceFilters) => void;
		onClearFilters: () => void;
		showProjectFilter?: boolean;
	}

	let { onApplyFilters, onClearFilters, showProjectFilter = false }: Props = $props();

	let buildingId = $state('');
	let controlCabinetId = $state('');
	let spsControllerId = $state('');
	let spsControllerSystemTypeId = $state('');
	let projectId = $state('');

	const hasActiveFilters = $derived(
		buildingId ||
			controlCabinetId ||
			spsControllerId ||
			spsControllerSystemTypeId ||
			(showProjectFilter && projectId)
	);

	const gridClass = $derived(
		showProjectFilter
			? 'grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-5'
			: 'grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4'
	);

	function applyFilters() {
		onApplyFilters({
			buildingId: buildingId || undefined,
			controlCabinetId: controlCabinetId || undefined,
			spsControllerId: spsControllerId || undefined,
			spsControllerSystemTypeId: spsControllerSystemTypeId || undefined,
			projectId: showProjectFilter && projectId ? projectId : undefined
		});
	}

	function clearFilters() {
		buildingId = '';
		controlCabinetId = '';
		spsControllerId = '';
		spsControllerSystemTypeId = '';
		projectId = '';
		onClearFilters();
	}
</script>

<Card.Root>
	<Card.Header>
		<Card.Title>Filters</Card.Title>
		<Card.Description>
			Filter field devices by building, control cabinet, SPS controller, system type
			{#if showProjectFilter}
				, or project.
			{:else}
				.
			{/if}
		</Card.Description>
	</Card.Header>
	<Card.Content>
		<div class={gridClass}>
			<div class="flex flex-col gap-2">
				<label for="building-filter" class="text-sm font-medium">Building</label>
				<BuildingSelect bind:value={buildingId} width="w-full" />
			</div>
			<div class="flex flex-col gap-2">
				<label for="control-cabinet-filter" class="text-sm font-medium">Control Cabinet</label>
				<ControlCabinetSelect bind:value={controlCabinetId} width="w-full" />
			</div>
			<div class="flex flex-col gap-2">
				<label for="sps-controller-filter" class="text-sm font-medium">SPS Controller</label>
				<SPSControllerSelect bind:value={spsControllerId} width="w-full" />
			</div>
			<div class="flex flex-col gap-2">
				<label for="sps-controller-system-type-filter" class="text-sm font-medium">
					SPS Controller System Type
				</label>
				<SPSControllerSystemTypeSelect bind:value={spsControllerSystemTypeId} width="w-full" />
			</div>
			{#if showProjectFilter}
				<div class="flex flex-col gap-2">
					<label for="project-filter" class="text-sm font-medium">Project</label>
					<ProjectSelect bind:value={projectId} width="w-full" />
				</div>
			{/if}
		</div>
		<div class="mt-4 flex gap-2">
			<Button onclick={applyFilters}>Apply Filters</Button>
			{#if hasActiveFilters}
				<Button variant="outline" onclick={clearFilters}>
					<X class="mr-2 size-4" />
					Clear Filters
				</Button>
			{/if}
		</div>
	</Card.Content>
</Card.Root>
