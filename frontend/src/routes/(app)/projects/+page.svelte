<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus } from 'lucide-svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { projectsStore } from '$lib/stores/list/entityStores.js';
	import type { Project } from '$lib/domain/entities/project.js';

	function getStatusClass(status: string): string {
		switch (status) {
			case 'completed':
				return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
			case 'ongoing':
				return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200';
			default:
				return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200';
		}
	}

	onMount(() => {
		projectsStore.load();
	});
</script>

<svelte:head>
	<title>Projects | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Projects</h1>
			<p class="text-sm text-muted-foreground">
				Manage your infrastructure projects. You can only see projects you have permission to
				access.
			</p>
		</div>
		<Button href="/projects/new">
			<Plus class="mr-2 size-4" />
			New Project
		</Button>
	</div>

	<PaginatedList
		state={$projectsStore}
		columns={[
			{ key: 'name', label: 'Name' },
			{ key: 'status', label: 'Status' },
			{ key: 'start_date', label: 'Start Date' },
			{ key: 'created', label: 'Created' },
			{ key: 'actions', label: 'Actions', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search projects..."
		emptyMessage="No projects found. Create your first project to get started."
		onSearch={(text) => projectsStore.search(text)}
		onPageChange={(page) => projectsStore.goToPage(page)}
		onReload={() => projectsStore.reload()}
	>
		{#snippet rowSnippet(project: Project)}
			<Table.Cell class="font-medium">
				<a href="/projects/{project.id}" class="hover:underline">
					{project.name}
				</a>
			</Table.Cell>
			<Table.Cell>
				<span
					class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {getStatusClass(
						project.status
					)}"
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
		{/snippet}
	</PaginatedList>
</div>
