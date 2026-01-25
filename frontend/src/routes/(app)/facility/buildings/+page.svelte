<script lang="ts">
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import * as Pagination from '$lib/components/ui/pagination/index.js';
	import SearchIcon from '@lucide/svelte/icons/search';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import type { PageData } from './$types.js';
	import BuildingForm from '$lib/components/facility/BuildingForm.svelte';
	import type { Building } from '$lib/domain/facility/index.js';
	import { invalidateAll, goto } from '$app/navigation';
	import { page } from '$app/state';
	import { debounce } from '$lib/utils.js';

	let { data }: { data: PageData } = $props();
	let searchQuery = $state(page.url.searchParams.get('search') || '');
	let showForm = $state(false);
	let editingBuilding: Building | undefined = $state(undefined);

	function handleEdit(building: Building) {
		editingBuilding = building;
		showForm = true;
	}

	function handleCreate() {
		editingBuilding = undefined;
		showForm = true;
	}

	function handleSuccess() {
		showForm = false;
		editingBuilding = undefined;
		invalidateAll();
	}

	function handleCancel() {
		showForm = false;
		editingBuilding = undefined;
	}

	function handlePageChange(newPage: number) {
		const url = new URL(page.url);
		url.searchParams.set('page', newPage.toString());
		goto(url.toString(), { keepFocus: true });
	}

	function handleSearch() {
		const url = new URL(page.url);
		url.searchParams.set('search', searchQuery);
		url.searchParams.set('page', '1');
		goto(url.toString(), { keepFocus: true });
	}

	const debouncedSearch = debounce(handleSearch, 300);
</script>

<svelte:head>
	<title>Buildings | Infra Link</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Buildings</h1>
			<p class="text-sm text-muted-foreground">Manage building infrastructure and IWS codes.</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<PlusIcon class="mr-2 size-4" />
				New Building
			</Button>
		{/if}
	</div>

	{#if showForm}
		<BuildingForm
			initialData={editingBuilding}
			on:success={handleSuccess}
			on:cancel={handleCancel}
		/>
	{/if}

	<div class="flex items-center gap-4">
		<div class="relative flex-1">
			<SearchIcon class="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
			<Input
				type="search"
				placeholder="Search buildings..."
				class="pl-10"
				bind:value={searchQuery}
				oninput={debouncedSearch}
			/>
		</div>
	</div>

	<div class="rounded-md border">
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head>IWS Code</Table.Head>
					<Table.Head>Building Group</Table.Head>
					<Table.Head>Created</Table.Head>
					<Table.Head class="w-[100px]">Actions</Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#if data.buildings && data.buildings.length > 0}
					{#each data.buildings as building (building.id)}
						<Table.Row>
							<Table.Cell class="font-medium">
								<a href="/facility/buildings/{building.id}" class="hover:underline">
									{building.iws_code}
								</a>
							</Table.Cell>
							<Table.Cell>{building.building_group}</Table.Cell>
							<Table.Cell>
								{new Date(building.created_at).toLocaleDateString()}
							</Table.Cell>
							<Table.Cell>
								<div class="flex items-center gap-2">
									<Button variant="ghost" size="icon" onclick={() => handleEdit(building)}>
										<PencilIcon class="size-4" />
									</Button>
									<Button variant="ghost" size="sm" href="/facility/buildings/{building.id}">
										View
									</Button>
								</div>
							</Table.Cell>
						</Table.Row>
					{/each}
				{:else}
					<Table.Row>
						<Table.Cell colspan={4} class="h-24 text-center text-muted-foreground">
							No buildings found. Create your first building to get started.
						</Table.Cell>
					</Table.Row>
				{/if}
			</Table.Body>
		</Table.Root>
	</div>

	{#if (data.total_pages ?? 0) > 1}
		<Pagination.Root
			count={data.total ?? 0}
			perPage={data.limit ?? 10}
			page={data.page ?? 1}
			onPageChange={handlePageChange}
		>
			{#snippet children({ pages, currentPage })}
				<Pagination.Content>
					<Pagination.Item>
						<Pagination.Previous />
					</Pagination.Item>
					{#each pages as page (page.key)}
						{#if page.type === 'ellipsis'}
							<Pagination.Item>
								<Pagination.Ellipsis />
							</Pagination.Item>
						{:else}
							<Pagination.Item>
								<Pagination.Link {page} isActive={currentPage === page.value}>
									{page.value}
								</Pagination.Link>
							</Pagination.Item>
						{/if}
					{/each}
					<Pagination.Item>
						<Pagination.Next />
					</Pagination.Item>
				</Pagination.Content>
			{/snippet}
		</Pagination.Root>
	{/if}
</div>