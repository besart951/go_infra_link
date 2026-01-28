<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import Toasts, { addToast } from '$lib/components/toast.svelte';
	import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
	import { confirm } from '$lib/stores/confirm-dialog.js';
	import ProjectPhaseSelect from '$lib/components/project/ProjectPhaseSelect.svelte';
	import {
		getProject,
		updateProject,
		listProjectUsers,
		addProjectUser,
		removeProjectUser,
		listProjectObjectData,
		addProjectObjectData,
		removeProjectObjectData
	} from '$lib/infrastructure/api/project.adapter.js';
	import { listUsers } from '$lib/infrastructure/api/user.adapter.js';
	import type { Project, ProjectStatus, UpdateProjectRequest } from '$lib/domain/project/index.js';
	import type { User } from '$lib/domain/user/index.js';
	import type { ObjectData } from '$lib/domain/facility/index.js';
	import { ArrowLeft } from '@lucide/svelte';

	const projectId = $derived($page.params.id ?? '');

	let project = $state<Project | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let saving = $state(false);
	let activeTab = $state<'settings' | 'users' | 'object-data'>('settings');

	let form = $state<{
		name: string;
		description: string;
		status: ProjectStatus;
		start_date: string;
		phase_id: string;
	}>({
		name: '',
		description: '',
		status: 'planned',
		start_date: '',
		phase_id: ''
	});

	const statusOptions: Array<{ value: ProjectStatus; label: string }> = [
		{ value: 'planned', label: 'Planned' },
		{ value: 'ongoing', label: 'Ongoing' },
		{ value: 'completed', label: 'Completed' }
	];

	let projectUsers = $state<User[]>([]);
	let availableUsers = $state<User[]>([]);
	let usersLoading = $state(false);
	let selectedUserId = $state('');
	let usersLoaded = $state(false);

	const availableUsersToInvite = $derived(
		availableUsers.filter((user) => !projectUsers.some((member) => member.id === user.id))
	);

	let projectObjectData = $state<ObjectData[]>([]);
	let availableObjectData = $state<ObjectData[]>([]);
	let objectDataLoading = $state(false);
	let selectedObjectDataId = $state('');
	let objectDataLoaded = $state(false);

	const availableObjectDataToActivate = $derived(
		availableObjectData.filter((obj) => !obj.is_active)
	);

	function formatDate(value?: string): string {
		if (!value) return '-';
		try {
			return new Date(value).toLocaleDateString();
		} catch {
			return value;
		}
	}

	function hydrateForm(p: Project) {
		form = {
			name: p.name ?? '',
			description: p.description ?? '',
			status: p.status,
			start_date: p.start_date ? p.start_date.slice(0, 10) : '',
			phase_id: p.phase_id ?? ''
		};
	}

	async function saveSettings() {
		if (!projectId || !project) return;
		if (!form.phase_id.trim()) {
			addToast('Phase is required', 'error');
			return;
		}
		saving = true;
		try {
			const payload: UpdateProjectRequest = {
				id: projectId,
				name: form.name.trim(),
				description: form.description.trim() || undefined,
				status: form.status,
				start_date: form.start_date || undefined,
				phase_id: form.phase_id
			};
			project = await updateProject(projectId, payload);
			hydrateForm(project);
			addToast('Project updated', 'success');
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to update project', 'error');
		} finally {
			saving = false;
		}
	}

	async function load() {
		if (!projectId) {
			error = 'Missing project id';
			loading = false;
			return;
		}
		loading = true;
		error = null;
		usersLoaded = false;
		objectDataLoaded = false;
		try {
			project = await getProject(projectId);
			hydrateForm(project);
		} catch (err) {
			const message = err instanceof Error ? err.message : 'Failed to load project';
			error = message;
			addToast(message, 'error');
		} finally {
			loading = false;
		}
	}

	async function loadUsers() {
		if (!projectId) return;
		usersLoading = true;
		try {
			const [projectUsersRes, usersRes] = await Promise.all([
				listProjectUsers(projectId),
				listUsers({ page: 1, limit: 100 })
			]);
			projectUsers = projectUsersRes.items;
			availableUsers = usersRes.items;
			usersLoaded = true;
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to load users', 'error');
		} finally {
			usersLoading = false;
		}
	}

	async function inviteUser() {
		if (!projectId || !selectedUserId) return;
		try {
			await addProjectUser(projectId, selectedUserId);
			addToast('User added', 'success');
			selectedUserId = '';
			await loadUsers();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to add user', 'error');
		}
	}

	async function removeUser(userId: string) {
		if (!projectId) return;
		const ok = await confirm({
			title: 'Remove user',
			message: 'Remove this user from the project?',
			confirmText: 'Remove',
			cancelText: 'Cancel',
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await removeProjectUser(projectId, userId);
			addToast('User removed', 'success');
			await loadUsers();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to remove user', 'error');
		}
	}

	async function loadObjectData() {
		if (!projectId) return;
		objectDataLoading = true;
		try {
			const projectRes = await listProjectObjectData(projectId, { page: 1, limit: 100 });
			projectObjectData = projectRes.items;
			availableObjectData = projectRes.items;
			objectDataLoaded = true;
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to load object data', 'error');
		} finally {
			objectDataLoading = false;
		}
	}

	async function attachObjectData() {
		if (!projectId || !selectedObjectDataId) return;
		try {
			await addProjectObjectData(projectId, selectedObjectDataId);
			addToast('Object data activated', 'success');
			selectedObjectDataId = '';
			await loadObjectData();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to add object data', 'error');
		}
	}

	async function detachObjectData(objectDataId: string) {
		if (!projectId) return;
		const ok = await confirm({
			title: 'Deactivate object data',
			message: 'Deactivate this object data for the project?',
			confirmText: 'Deactivate',
			cancelText: 'Cancel',
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await removeProjectObjectData(projectId, objectDataId);
			addToast('Object data deactivated', 'success');
			await loadObjectData();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to remove object data', 'error');
		}
	}

	$effect(() => {
		if (activeTab === 'users' && !usersLoaded && !usersLoading) {
			loadUsers();
		}
	});

	$effect(() => {
		if (activeTab === 'object-data' && !objectDataLoaded && !objectDataLoading) {
			loadObjectData();
		}
	});

	onMount(() => {
		load();
	});
</script>

<Toasts />
<ConfirmDialog />

<div class="flex flex-col gap-6">
	<div class="flex items-start gap-3">
		<Button variant="outline" onclick={() => goto(`/projects/${projectId}`)}>
			<ArrowLeft class="mr-2 h-4 w-4" />
			Back
		</Button>
		<div>
			<h1 class="text-3xl font-bold tracking-tight">Project Settings</h1>
			<p class="mt-1 text-muted-foreground">Update project metadata and phase.</p>
		</div>
	</div>

	{#if error}
		<div class="rounded-md border bg-muted px-4 py-3 text-muted-foreground">
			<p class="font-medium">Could not load project</p>
			<p class="text-sm">{error}</p>
		</div>
	{/if}

	<div class="rounded-lg border bg-background">
		<div class="flex flex-wrap gap-2 border-b px-6 py-3">
			<Button
				variant={activeTab === 'settings' ? 'default' : 'ghost'}
				onclick={() => (activeTab = 'settings')}
			>
				Settings
			</Button>
			<Button
				variant={activeTab === 'users' ? 'default' : 'ghost'}
				onclick={() => (activeTab = 'users')}
			>
				Users
			</Button>
			<Button
				variant={activeTab === 'object-data' ? 'default' : 'ghost'}
				onclick={() => (activeTab = 'object-data')}
			>
				Object data
			</Button>
		</div>

		{#if loading}
			<div class="p-6">
				<div class="grid gap-4 md:grid-cols-2">
					{#each Array(6) as _}
						<Skeleton class="h-6 w-full" />
					{/each}
				</div>
			</div>
		{:else if !project}
			<div class="p-6 text-sm text-muted-foreground">Project not found.</div>
		{:else if activeTab === 'settings'}
			<div class="p-6">
				<div class="grid gap-4 md:grid-cols-2">
					<div class="flex flex-col gap-2">
						<label class="text-sm font-medium" for="project_name">Name</label>
						<Input id="project_name" bind:value={form.name} disabled={saving} />
					</div>

					<div class="flex flex-col gap-2">
						<label class="text-sm font-medium" for="project_status">Status</label>
						<select
							id="project_status"
							class="h-9 rounded-md border border-input bg-background px-3 text-sm font-medium shadow-xs"
							bind:value={form.status}
							disabled={saving}
						>
							{#each statusOptions as opt}
								<option value={opt.value}>{opt.label}</option>
							{/each}
						</select>
					</div>

					<div class="flex flex-col gap-2">
						<label class="text-sm font-medium" for="project_start">Start date</label>
						<Input id="project_start" type="date" bind:value={form.start_date} disabled={saving} />
					</div>

					<div class="flex flex-col gap-2">
						<label class="text-sm font-medium" for="project_phase_edit">Phase</label>
						<ProjectPhaseSelect id="project_phase_edit" bind:value={form.phase_id} width="w-full" />
					</div>

					<div class="flex flex-col gap-2 md:col-span-2">
						<label class="text-sm font-medium" for="project_desc">Description</label>
						<Textarea id="project_desc" rows={4} bind:value={form.description} disabled={saving} />
					</div>
				</div>

				<div class="mt-6 flex items-center justify-end gap-2">
					<Button
						variant="outline"
						onclick={() => project && hydrateForm(project)}
						disabled={saving}
					>
						Reset
					</Button>
					<Button onclick={saveSettings} disabled={saving}>Save changes</Button>
				</div>

				<div class="mt-8 grid gap-6 md:grid-cols-2">
					<div class="space-y-2">
						<div class="text-xs text-muted-foreground uppercase">Created</div>
						<div class="text-sm font-medium">{formatDate(project.created_at)}</div>
					</div>
					<div class="space-y-2">
						<div class="text-xs text-muted-foreground uppercase">Updated</div>
						<div class="text-sm font-medium">{formatDate(project.updated_at)}</div>
					</div>
				</div>
			</div>
		{:else if activeTab === 'users'}
			<div class="p-6">
				<div class="flex flex-wrap items-end gap-3">
					<div class="flex flex-col gap-2">
						<label class="text-sm font-medium" for="project_user">Add user</label>
						<select
							id="project_user"
							class="h-9 rounded-md border border-input bg-background px-3 text-sm font-medium shadow-xs"
							bind:value={selectedUserId}
						>
							<option value="">Select user</option>
							{#each availableUsersToInvite as user}
								<option value={user.id}>
									{user.first_name}
									{user.last_name} ({user.email})
								</option>
							{/each}
						</select>
					</div>
					<Button
						onclick={inviteUser}
						disabled={!selectedUserId || usersLoading || availableUsersToInvite.length === 0}
					>
						Add
					</Button>
					<Button variant="outline" onclick={loadUsers} disabled={usersLoading}>Refresh</Button>
				</div>

				<div class="mt-6 rounded-lg border bg-background">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Name</Table.Head>
								<Table.Head>Email</Table.Head>
								<Table.Head class="w-32"></Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#if usersLoading}
								{#each Array(5) as _}
									<Table.Row>
										<Table.Cell><Skeleton class="h-4 w-40" /></Table.Cell>
										<Table.Cell><Skeleton class="h-4 w-60" /></Table.Cell>
										<Table.Cell><Skeleton class="h-8 w-20" /></Table.Cell>
									</Table.Row>
								{/each}
							{:else if projectUsers.length === 0}
								<Table.Row>
									<Table.Cell colspan={3} class="h-20 text-center text-sm text-muted-foreground">
										No users in this project yet.
									</Table.Cell>
								</Table.Row>
							{:else}
								{#each projectUsers as user (user.id)}
									<Table.Row>
										<Table.Cell class="font-medium">
											{user.first_name}
											{user.last_name}
										</Table.Cell>
										<Table.Cell class="text-muted-foreground">{user.email}</Table.Cell>
										<Table.Cell class="text-right">
											<Button variant="outline" onclick={() => removeUser(user.id)}>Remove</Button>
										</Table.Cell>
									</Table.Row>
								{/each}
							{/if}
						</Table.Body>
					</Table.Root>
				</div>
			</div>
		{:else}
			<div class="p-6">
				<div class="flex flex-wrap items-end gap-3">
					<div class="flex flex-col gap-2">
						<label class="text-sm font-medium" for="project_object_data">Add object data</label>
						<select
							id="project_object_data"
							class="h-9 rounded-md border border-input bg-background px-3 text-sm font-medium shadow-xs"
							bind:value={selectedObjectDataId}
						>
							<option value="">Select object data</option>
							{#each availableObjectDataToActivate as obj}
								<option value={obj.id}>{obj.description}</option>
							{/each}
						</select>
					</div>
					<Button
						onclick={attachObjectData}
						disabled={!selectedObjectDataId ||
							objectDataLoading ||
							availableObjectDataToActivate.length === 0}>Add</Button
					>
					<Button variant="outline" onclick={loadObjectData} disabled={objectDataLoading}
						>Refresh</Button
					>
				</div>

				<div class="mt-6 rounded-lg border bg-background">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Description</Table.Head>
								<Table.Head>Version</Table.Head>
								<Table.Head>Active</Table.Head>
								<Table.Head class="w-32"></Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#if objectDataLoading}
								{#each Array(5) as _}
									<Table.Row>
										<Table.Cell><Skeleton class="h-4 w-60" /></Table.Cell>
										<Table.Cell><Skeleton class="h-4 w-24" /></Table.Cell>
										<Table.Cell><Skeleton class="h-4 w-16" /></Table.Cell>
										<Table.Cell><Skeleton class="h-8 w-20" /></Table.Cell>
									</Table.Row>
								{/each}
							{:else if projectObjectData.length === 0}
								<Table.Row>
									<Table.Cell colspan={4} class="h-20 text-center text-sm text-muted-foreground">
										No object data assigned yet.
									</Table.Cell>
								</Table.Row>
							{:else}
								{#each projectObjectData as obj (obj.id)}
									<Table.Row>
										<Table.Cell class="font-medium">{obj.description}</Table.Cell>
										<Table.Cell class="text-muted-foreground">{obj.version}</Table.Cell>
										<Table.Cell>
											{obj.is_active ? 'Yes' : 'No'}
										</Table.Cell>
										<Table.Cell class="text-right">
											<Button variant="outline" onclick={() => detachObjectData(obj.id)}>
												Deactivate
											</Button>
										</Table.Cell>
									</Table.Row>
								{/each}
							{/if}
						</Table.Body>
					</Table.Root>
				</div>
			</div>
		{/if}
	</div>
</div>
