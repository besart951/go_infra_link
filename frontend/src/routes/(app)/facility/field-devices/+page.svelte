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
	<title>Field Devices | Infra Link</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Field Devices</h1>
			<p class="text-sm text-muted-foreground">
				Manage field devices, BMK identifiers, and specifications.
			</p>
		</div>
		<Button href="/facility/field-devices/new">
			<PlusIcon class="mr-2 size-4" />
			New Field Device
		</Button>
	</div>

	<div class="flex items-center gap-4">
		<div class="relative max-w-sm flex-1">
			<SearchIcon class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
			<Input
				type="search"
				placeholder="Search field devices..."
				class="pl-10"
				bind:value={searchQuery}
			/>
		</div>
	</div>

	<div class="rounded-md border">
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head>BMK</Table.Head>
					<Table.Head>Description</Table.Head>
					<Table.Head>Apparat Nr</Table.Head>
					<Table.Head>Created</Table.Head>
					<Table.Head class="w-[100px]">Actions</Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#if data.fieldDevices && data.fieldDevices.length > 0}
					{#each data.fieldDevices as device (device.id)}
						<Table.Row>
							<Table.Cell class="font-medium">
								<a href="/facility/field-devices/{device.id}" class="hover:underline">
									{device.bmk}
								</a>
							</Table.Cell>
							<Table.Cell>{device.description}</Table.Cell>
							<Table.Cell>
								<code class="rounded bg-muted px-1.5 py-0.5 text-sm">
									{device.apparat_nr}
								</code>
							</Table.Cell>
							<Table.Cell>
								{new Date(device.created_at).toLocaleDateString()}
							</Table.Cell>
							<Table.Cell>
								<Button variant="ghost" size="sm" href="/facility/field-devices/{device.id}">
									View
								</Button>
							</Table.Cell>
						</Table.Row>
					{/each}
				{:else}
					<Table.Row>
						<Table.Cell colspan={5} class="h-24 text-center text-muted-foreground">
							No field devices found. Create your first field device to get started.
						</Table.Cell>
					</Table.Row>
				{/if}
			</Table.Body>
		</Table.Root>
	</div>
</div>
