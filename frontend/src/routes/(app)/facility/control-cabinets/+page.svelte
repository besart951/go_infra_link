<script lang="ts">
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import SearchIcon from '@lucide/svelte/icons/search';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import type { PageData } from './$types.js';
	import ControlCabinetForm from '$lib/components/facility/ControlCabinetForm.svelte';
	import type { ControlCabinet } from '$lib/domain/facility/index.js';
	import { invalidateAll } from '$app/navigation';

	let { data }: { data: PageData } = $props();
	let searchQuery = $state('');
	let showForm = $state(false);
	let editingItem: ControlCabinet | undefined = $state(undefined);

	function handleEdit(item: ControlCabinet) {
		editingItem = item;
		showForm = true;
	}

	function handleCreate() {
		editingItem = undefined;
		showForm = true;
	}

	function handleSuccess() {
		showForm = false;
		editingItem = undefined;
		invalidateAll();
	}

	function handleCancel() {
		showForm = false;
		editingItem = undefined;
	}
</script>

<svelte:head>
	<title>Control Cabinets | Infra Link</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Control Cabinets</h1>
			<p class="text-sm text-muted-foreground">
				Manage control cabinet configurations within buildings.
			</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<PlusIcon class="mr-2 size-4" />
				New Control Cabinet
			</Button>
		{/if}
	</div>

	{#if showForm}
		<ControlCabinetForm
			initialData={editingItem}
			on:success={handleSuccess}
			on:cancel={handleCancel}
		/>
	{/if}

	<div class="flex items-center gap-4">
		<div class="relative flex-1">
			<SearchIcon class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
			<Input
				type="search"
				placeholder="Search control cabinets..."
				class="pl-10"
				bind:value={searchQuery}
			/>
		</div>
	</div>

	<div class="rounded-md border">
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head>Cabinet Nr</Table.Head>
					<Table.Head>Building</Table.Head>
					<Table.Head>Created</Table.Head>
					<Table.Head class="w-[100px]">Actions</Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#if data.controlCabinets && data.controlCabinets.length > 0}
					{#each data.controlCabinets as cabinet (cabinet.id)}
						<Table.Row>
							<Table.Cell class="font-medium">
								<a href="/facility/control-cabinets/{cabinet.id}" class="hover:underline">
									{cabinet.control_cabinet_nr}
								</a>
							</Table.Cell>
							<Table.Cell>{cabinet.building_id}</Table.Cell>
							<Table.Cell>
								{new Date(cabinet.created_at).toLocaleDateString()}
							</Table.Cell>
							<Table.Cell>
								<div class="flex items-center gap-2">
									<Button variant="ghost" size="icon" onclick={() => handleEdit(cabinet)}>
										<PencilIcon class="size-4" />
									</Button>
									<Button variant="ghost" size="sm" href="/facility/control-cabinets/{cabinet.id}">
										View
									</Button>
								</div>
							</Table.Cell>
						</Table.Row>
					{/each}
				{:else}
					<Table.Row>
						<Table.Cell colspan={4} class="h-24 text-center text-muted-foreground">
							No control cabinets found. Create your first control cabinet to get started.
						</Table.Cell>
					</Table.Row>
				{/if}
			</Table.Body>
		</Table.Root>
	</div>
</div>
