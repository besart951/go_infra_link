<script lang="ts">
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import SearchIcon from '@lucide/svelte/icons/search';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import PencilIcon from '@lucide/svelte/icons/pencil';
	import type { PageData } from './$types.js';
	import SPSControllerForm from '$lib/components/facility/SPSControllerForm.svelte';
	import type { SPSController } from '$lib/domain/facility/index.js';
	import { invalidateAll } from '$app/navigation';

	let { data }: { data: PageData } = $props();
	let searchQuery = $state('');
	let showForm = $state(false);
	let editingItem: SPSController | undefined = $state(undefined);

	function handleEdit(item: SPSController) {
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
	<title>SPS Controllers | Infra Link</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">SPS Controllers</h1>
			<p class="text-sm text-muted-foreground">
				Manage SPS controller devices and their configurations.
			</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<PlusIcon class="mr-2 size-4" />
				New SPS Controller
			</Button>
		{/if}
	</div>

	{#if showForm}
		<SPSControllerForm
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
				placeholder="Search SPS controllers..."
				class="pl-10"
				bind:value={searchQuery}
			/>
		</div>
	</div>

	<div class="rounded-md border">
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head>Device Name</Table.Head>
					<Table.Head>GA Device</Table.Head>
					<Table.Head>IP Address</Table.Head>
					<Table.Head>Cabinet</Table.Head>
					<Table.Head>Created</Table.Head>
					<Table.Head class="w-[100px]">Actions</Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#if data.spsControllers && data.spsControllers.length > 0}
					{#each data.spsControllers as controller (controller.id)}
						<Table.Row>
							<Table.Cell class="font-medium">
								<a href="/facility/sps-controllers/{controller.id}" class="hover:underline">
									{controller.device_name}
								</a>
							</Table.Cell>
							<Table.Cell>{controller.ga_device}</Table.Cell>
							<Table.Cell>
								<code class="rounded bg-muted px-1.5 py-0.5 text-sm">
									{controller.ip_address}
								</code>
							</Table.Cell>
							<Table.Cell>{controller.control_cabinet_id}</Table.Cell>
							<Table.Cell>
								{new Date(controller.created_at).toLocaleDateString()}
							</Table.Cell>
							<Table.Cell>
								<div class="flex items-center gap-2">
									<Button variant="ghost" size="icon" onclick={() => handleEdit(controller)}>
										<PencilIcon class="size-4" />
									</Button>
									<Button
										variant="ghost"
										size="sm"
										href="/facility/sps-controllers/{controller.id}"
									>
										View
									</Button>
								</div>
							</Table.Cell>
						</Table.Row>
					{/each}
				{:else}
					<Table.Row>
						<Table.Cell colspan={6} class="h-24 text-center text-muted-foreground">
							No SPS controllers found. Create your first SPS controller to get started.
						</Table.Cell>
					</Table.Row>
				{/if}
			</Table.Body>
		</Table.Root>
	</div>
</div>
