<script lang="ts">
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import SearchIcon from '@lucide/svelte/icons/search';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import type { PageData } from './$types.js';

	let { data }: { data: PageData } = $props();
	let searchQuery = $state('');
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
		<Button href="/facility/buildings/new">
			<PlusIcon class="mr-2 size-4" />
			New Building
		</Button>
	</div>

	<div class="flex items-center gap-4">
		<div class="relative max-w-sm flex-1">
			<SearchIcon class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
			<Input
				type="search"
				placeholder="Search buildings..."
				class="pl-10"
				bind:value={searchQuery}
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
								<Button variant="ghost" size="sm" href="/facility/buildings/{building.id}">
									View
								</Button>
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
</div>
