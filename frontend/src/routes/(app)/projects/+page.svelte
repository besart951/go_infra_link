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
	<title>Projects | Infra Link</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Projects</h1>
			<p class="text-sm text-muted-foreground">
				Manage your infrastructure projects. You can only see projects you have permission to
				access.
			</p>
		</div>
		<Button href="/projects/new">
			<PlusIcon class="mr-2 size-4" />
			New Project
		</Button>
	</div>

	<div class="flex items-center gap-4">
		<div class="relative max-w-sm flex-1">
			<SearchIcon class="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
			<Input
				type="search"
				placeholder="Search projects..."
				class="pl-10"
				bind:value={searchQuery}
			/>
		</div>
	</div>

	<div class="rounded-md border">
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head>Name</Table.Head>
					<Table.Head>Status</Table.Head>
					<Table.Head>Start Date</Table.Head>
					<Table.Head>Created</Table.Head>
					<Table.Head class="w-[100px]">Actions</Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#if data.projects && data.projects.length > 0}
					{#each data.projects as project (project.id)}
						<Table.Row>
							<Table.Cell class="font-medium">
								<a href="/projects/{project.id}" class="hover:underline">
									{project.name}
								</a>
							</Table.Cell>
							<Table.Cell>
								<span
									class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium
									{project.status === 'completed'
										? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
										: project.status === 'ongoing'
											? 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
											: 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200'}"
								>
									{project.status}
								</span>
							</Table.Cell>
							<Table.Cell>
								{project.start_date ? new Date(project.start_date).toLocaleDateString() : '-'}
							</Table.Cell>
							<Table.Cell>
								{new Date(project.created_at).toLocaleDateString()}
							</Table.Cell>
							<Table.Cell>
								<Button variant="ghost" size="sm" href="/projects/{project.id}">View</Button>
							</Table.Cell>
						</Table.Row>
					{/each}
				{:else}
					<Table.Row>
						<Table.Cell colspan={5} class="h-24 text-center text-muted-foreground">
							No projects found. Create your first project to get started.
						</Table.Cell>
					</Table.Row>
				{/if}
			</Table.Body>
		</Table.Root>
	</div>
</div>
