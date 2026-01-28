<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import Toasts, { addToast } from '$lib/components/toast.svelte';
	import ProjectPhaseSelect from '$lib/components/project/ProjectPhaseSelect.svelte';
	import { projectListStore } from '$lib/stores/projects/projectListStore.js';
	import { createProject } from '$lib/infrastructure/api/project.adapter.js';
	import type { CreateProjectRequest, Project, ProjectStatus } from '$lib/domain/project/index.js';

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

	const createStatusOptions: Array<{ value: ProjectStatus; label: string }> = [
		{ value: 'planned', label: 'Planned' },
		{ value: 'ongoing', label: 'Ongoing' },
		{ value: 'completed', label: 'Completed' }
	];

	type CreateProjectForm = {
		name: string;
		description: string;
		status: ProjectStatus;
		start_date: string;
		phase_id: string;
	};

	function todayInputValue(): string {
		const now = new Date();
		const yyyy = now.getFullYear();
		const mm = String(now.getMonth() + 1).padStart(2, '0');
		const dd = String(now.getDate()).padStart(2, '0');
		return `${yyyy}-${mm}-${dd}`;
	}

	let createOpen = $state(false);
	let createBusy = $state(false);
	let form = $state<CreateProjectForm>({
		name: '',
		description: '',
		status: 'planned',
		start_date: todayInputValue(),
		phase_id: ''
	});

	function canSubmitCreate(): boolean {
		return form.name.trim().length > 0 && form.phase_id.trim().length > 0 && !createBusy;
	}

	async function submitCreate() {
		if (!canSubmitCreate()) return;
		createBusy = true;
		try {
			const payload: CreateProjectRequest = {
				name: form.name.trim(),
				description: form.description.trim() || undefined,
				status: form.status,
				start_date: form.start_date
					? new Date(`${form.start_date}T00:00:00Z`).toISOString()
					: undefined,
				phase_id: form.phase_id
			};

			const project = await createProject(payload);
			addToast('Project created', 'success');
			form = {
				name: '',
				description: '',
				status: 'planned',
				start_date: todayInputValue(),
				phase_id: ''
			};
			createOpen = false;
			projectListStore.reload();
			goto(`/projects/${project.id}`);
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to create project', 'error');
		} finally {
			createBusy = false;
		}
	}

	function handleStatusChange(e: Event) {
		const value = (e.target as HTMLSelectElement).value;
		projectListStore.setStatus(value === 'all' ? 'all' : (value as ProjectStatus));
	}

	onMount(() => {
		projectListStore.load();
	});
</script>

<Toasts />

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
		<Button onclick={() => (createOpen = !createOpen)}>
			<Plus class="mr-2 size-4" />
			New Project
		</Button>
	</div>

	{#if createOpen}
		<div class="rounded-lg border bg-background p-4">
			<div class="grid gap-4 md:grid-cols-2">
				<div class="flex flex-col gap-2">
					<label class="text-sm font-medium" for="project_name_create">Name</label>
					<Input
						id="project_name_create"
						placeholder="Project name"
						bind:value={form.name}
						disabled={createBusy}
					/>
				</div>

				<div class="flex flex-col gap-2">
					<label class="text-sm font-medium" for="project_status_create">Status</label>
					<select
						id="project_status_create"
						class="h-9 rounded-md border border-input bg-background px-3 text-sm font-medium shadow-xs"
						bind:value={form.status}
						disabled={createBusy}
					>
						{#each createStatusOptions as opt}
							<option value={opt.value}>{opt.label}</option>
						{/each}
					</select>
				</div>

				<div class="flex flex-col gap-2">
					<label class="text-sm font-medium" for="project_start_create">Start date</label>
					<Input
						id="project_start_create"
						type="date"
						bind:value={form.start_date}
						disabled={createBusy}
					/>
				</div>

				<div class="flex flex-col gap-2">
					<label class="text-sm font-medium" for="project_phase_create">Phase</label>
					<ProjectPhaseSelect id="project_phase_create" bind:value={form.phase_id} width="w-full" />
				</div>

				<div class="flex flex-col gap-2 md:col-span-2">
					<label class="text-sm font-medium" for="project_desc_create">Description</label>
					<Textarea
						id="project_desc_create"
						placeholder="Describe the project goals"
						rows={3}
						bind:value={form.description}
						disabled={createBusy}
					/>
				</div>
			</div>

			<div class="mt-4 flex items-center justify-end gap-2">
				<Button variant="outline" onclick={() => (createOpen = false)} disabled={createBusy}
					>Cancel</Button
				>
				<Button onclick={submitCreate} disabled={!canSubmitCreate()}>Create</Button>
			</div>
		</div>
	{/if}

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
