<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { projectListStore } from '$lib/stores/projects/projectListStore.js';
	import type { Project, ProjectStatus } from '$lib/domain/project/index.js';

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

	const statusOptions: Array<{ value: ProjectStatus | 'all'; label: string }> = [
		{ value: 'all', label: 'All statuses' },
		{ value: 'planned', label: 'Planned' },
		{ value: 'ongoing', label: 'Ongoing' },
		{ value: 'completed', label: 'Completed' }
	];

	function handleStatusChange(e: Event) {
		const value = (e.target as HTMLSelectElement).value;
		projectListStore.setStatus(value === 'all' ? 'all' : (value as ProjectStatus));
	}

	onMount(() => {
		projectListStore.load();
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

	<div class="flex flex-wrap items-center gap-3">
		<label class="text-sm font-medium" for="project_status_filter">Status</label>
		<select
			id="project_status_filter"
			class="h-9 rounded-md border border-input bg-background px-3 text-sm font-medium shadow-xs"
			value={$projectListStore.status}
			onchange={handleStatusChange}
		>
			{#each statusOptions as opt}
				<option value={opt.value}>{opt.label}</option>
			{/each}
		</select>
	</div>

	<PaginatedList
		state={$projectListStore}
		columns={[
			{ key: 'name', label: 'Name' },
			{ key: 'status', label: 'Status' },
			{ key: 'start_date', label: 'Start Date' },
			{ key: 'created', label: 'Created' },
			{ key: 'actions', label: 'Actions', width: 'w-[100px]' }
		]}
		searchPlaceholder="Search projects..."
		emptyMessage="No projects found. Create your first project to get started."
		onSearch={(text) => projectListStore.search(text)}
		onPageChange={(page) => projectListStore.goToPage(page)}
		onReload={() => projectListStore.reload()}
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
