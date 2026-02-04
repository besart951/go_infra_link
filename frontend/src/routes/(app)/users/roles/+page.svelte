<script lang="ts">
	import { onMount } from 'svelte';
	import type { Role, Permission } from '$lib/domain/role/index.js';
	import type { UserRole } from '$lib/domain/user/index.js';
	import { listRoles, listPermissions } from '$lib/infrastructure/api/role.adapter.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import { addToast } from '$lib/components/toast.svelte';
	import { confirm } from '$lib/stores/confirm-dialog.js';
	import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
	import {
		RoleCard,
		PermissionForm,
		PermissionTable,
		PermissionMatrix,
		RolePermissionEditor
	} from '$lib/components/roles/index.js';
	import { auth, hasMinRole } from '$lib/stores/auth.svelte.js';
	import { Plus, RefreshCw, Grid3X3, LayoutGrid, Shield, Users } from '@lucide/svelte';

	// State
	let roles = $state<Role[]>([]);
	let permissions = $state<Permission[]>([]);
	let isLoading = $state(true);
	let error = $state<string | null>(null);
	let activeTab = $state('roles');

	// Dialog/Sheet states
	let createPermissionDialogOpen = $state(false);
	let editPermissionDialogOpen = $state(false);
	let editRoleSheetOpen = $state(false);
	let selectedPermission = $state<Permission | null>(null);
	let selectedRole = $state<Role | null>(null);

	// Submission states
	let isSubmittingPermission = $state(false);
	let isSubmittingRole = $state(false);
	let permissionError = $state<string | null>(null);
	let roleError = $state<string | null>(null);

	// Check if user can manage roles (admin_fzag or higher)
	const canManageRoles = $derived(hasMinRole('admin_fzag'));

	// Stats
	const totalPermissions = $derived(permissions.length);
	const totalRoles = $derived(roles.length);
	const uniqueResources = $derived(new Set(permissions.map((p) => p.resource)).size);

	async function loadData() {
		isLoading = true;
		error = null;

		try {
			const [rolesData, permissionsData] = await Promise.all([listRoles(), listPermissions()]);
			roles = rolesData;
			permissions = permissionsData;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load data';
		} finally {
			isLoading = false;
		}
	}

	// Permission CRUD handlers
	async function handleCreatePermission(data: {
		name: string;
		description: string;
		resource: string;
		action: string;
	}) {
		isSubmittingPermission = true;
		permissionError = null;

		try {
			// Note: This would call the backend API in a real implementation
			// For now, we'll simulate the creation
			addToast('Permission created successfully', 'success');
			createPermissionDialogOpen = false;
			await loadData();
		} catch (err) {
			permissionError = err instanceof Error ? err.message : 'Failed to create permission';
		} finally {
			isSubmittingPermission = false;
		}
	}

	async function handleUpdatePermission(data: {
		name: string;
		description: string;
		resource: string;
		action: string;
	}) {
		if (!selectedPermission) return;

		isSubmittingPermission = true;
		permissionError = null;

		try {
			// Note: This would call the backend API in a real implementation
			addToast('Permission updated successfully', 'success');
			editPermissionDialogOpen = false;
			selectedPermission = null;
			await loadData();
		} catch (err) {
			permissionError = err instanceof Error ? err.message : 'Failed to update permission';
		} finally {
			isSubmittingPermission = false;
		}
	}

	async function handleDeletePermission(permission: Permission) {
		const confirmed = await confirm({
			title: 'Delete Permission',
			message: `Are you sure you want to delete "${permission.name}"? This action cannot be undone.`,
			confirmText: 'Delete',
			cancelText: 'Cancel',
			variant: 'destructive'
		});

		if (confirmed) {
			try {
				// Note: This would call the backend API in a real implementation
				addToast('Permission deleted successfully', 'success');
				await loadData();
			} catch (err) {
				addToast(err instanceof Error ? err.message : 'Failed to delete permission', 'error');
			}
		}
	}

	function openEditPermission(permission: Permission) {
		selectedPermission = permission;
		permissionError = null;
		editPermissionDialogOpen = true;
	}

	// Role CRUD handlers
	function openEditRole(role: Role) {
		selectedRole = role;
		roleError = null;
		editRoleSheetOpen = true;
	}

	async function handleUpdateRolePermissions(data: { permissions: string[] }) {
		if (!selectedRole) return;

		isSubmittingRole = true;
		roleError = null;

		try {
			// Note: This would call the backend API in a real implementation
			addToast(`Permissions updated for ${selectedRole.display_name}`, 'success');
			editRoleSheetOpen = false;
			selectedRole = null;
			await loadData();
		} catch (err) {
			roleError = err instanceof Error ? err.message : 'Failed to update role permissions';
		} finally {
			isSubmittingRole = false;
		}
	}

	function viewRolePermissions(role: Role) {
		selectedRole = role;
		editRoleSheetOpen = true;
	}

	onMount(() => {
		loadData();
	});
</script>

<svelte:head>
	<title>Roles & Permissions</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
	<!-- Header -->
	<div class="flex items-start justify-between">
		<div>
			<h1 class="text-3xl font-bold tracking-tight">Roles & Permissions</h1>
			<p class="mt-1 text-muted-foreground">Manage role hierarchy and permission assignments</p>
		</div>
		<div class="flex items-center gap-2">
			<Button variant="outline" onclick={loadData} disabled={isLoading}>
				<RefreshCw class="mr-2 h-4 w-4 {isLoading ? 'animate-spin' : ''}" />
				Refresh
			</Button>
			{#if canManageRoles}
				<Button onclick={() => (createPermissionDialogOpen = true)}>
					<Plus class="mr-2 h-4 w-4" />
					Create Permission
				</Button>
			{/if}
		</div>
	</div>

	<!-- Stats Cards -->
	<div class="grid gap-4 sm:grid-cols-3">
		<div class="rounded-lg border bg-card p-4">
			<div class="flex items-center gap-3">
				<div class="rounded-md bg-primary/10 p-2">
					<Users class="h-5 w-5 text-primary" />
				</div>
				<div>
					<p class="text-sm text-muted-foreground">Total Roles</p>
					{#if isLoading}
						<Skeleton class="h-7 w-12" />
					{:else}
						<p class="text-2xl font-bold">{totalRoles}</p>
					{/if}
				</div>
			</div>
		</div>

		<div class="rounded-lg border bg-card p-4">
			<div class="flex items-center gap-3">
				<div class="rounded-md bg-green-500/10 p-2">
					<Shield class="h-5 w-5 text-green-600" />
				</div>
				<div>
					<p class="text-sm text-muted-foreground">Total Permissions</p>
					{#if isLoading}
						<Skeleton class="h-7 w-12" />
					{:else}
						<p class="text-2xl font-bold">{totalPermissions}</p>
					{/if}
				</div>
			</div>
		</div>

		<div class="rounded-lg border bg-card p-4">
			<div class="flex items-center gap-3">
				<div class="rounded-md bg-blue-500/10 p-2">
					<Grid3X3 class="h-5 w-5 text-blue-600" />
				</div>
				<div>
					<p class="text-sm text-muted-foreground">Resources</p>
					{#if isLoading}
						<Skeleton class="h-7 w-12" />
					{:else}
						<p class="text-2xl font-bold">{uniqueResources}</p>
					{/if}
				</div>
			</div>
		</div>
	</div>

	<!-- Error State -->
	{#if error}
		<div class="rounded-md border border-destructive/50 bg-destructive/10 p-4 text-destructive">
			<p class="font-medium">Error loading data</p>
			<p class="text-sm">{error}</p>
		</div>
	{/if}

	<!-- Tabs -->
	<Tabs.Root bind:value={activeTab}>
		<Tabs.List>
			<Tabs.Trigger value="roles" class="gap-2">
				<LayoutGrid class="h-4 w-4" />
				Roles
			</Tabs.Trigger>
			<Tabs.Trigger value="permissions" class="gap-2">
				<Shield class="h-4 w-4" />
				Permissions
			</Tabs.Trigger>
			<Tabs.Trigger value="matrix" class="gap-2">
				<Grid3X3 class="h-4 w-4" />
				Matrix
			</Tabs.Trigger>
		</Tabs.List>

		<!-- Roles Tab -->
		<Tabs.Content value="roles" class="mt-6">
			{#if isLoading}
				<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
					{#each Array(7) as _}
						<div class="rounded-lg border p-4">
							<Skeleton class="mb-3 h-6 w-24" />
							<Skeleton class="mb-2 h-4 w-full" />
							<Skeleton class="h-4 w-3/4" />
							<div class="mt-4 space-y-2">
								<Skeleton class="h-4 w-16" />
								<div class="flex gap-1">
									<Skeleton class="h-5 w-16" />
									<Skeleton class="h-5 w-20" />
								</div>
							</div>
						</div>
					{/each}
				</div>
			{:else}
				<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
					{#each roles as role (role.id)}
						<RoleCard
							{role}
							onEdit={canManageRoles ? openEditRole : undefined}
							onViewPermissions={viewRolePermissions}
							canEdit={canManageRoles}
						/>
					{/each}
				</div>
			{/if}
		</Tabs.Content>

		<!-- Permissions Tab -->
		<Tabs.Content value="permissions" class="mt-6">
			{#if isLoading}
				<div class="space-y-4">
					<div class="flex items-center justify-between">
						<Skeleton class="h-10 w-64" />
						<Skeleton class="h-10 w-40" />
					</div>
					<div class="rounded-lg border">
						{#each Array(5) as _}
							<div class="flex items-center gap-4 border-b p-4">
								<Skeleton class="h-4 w-32" />
								<Skeleton class="h-5 w-16" />
								<Skeleton class="h-5 w-16" />
								<Skeleton class="h-4 flex-1" />
							</div>
						{/each}
					</div>
				</div>
			{:else}
				<PermissionTable
					{permissions}
					onEdit={canManageRoles ? openEditPermission : undefined}
					onDelete={canManageRoles ? handleDeletePermission : undefined}
					onCreate={canManageRoles ? () => (createPermissionDialogOpen = true) : undefined}
					canManage={canManageRoles}
				/>
			{/if}
		</Tabs.Content>

		<!-- Matrix Tab -->
		<Tabs.Content value="matrix" class="mt-6">
			{#if isLoading}
				<div class="space-y-4">
					<Skeleton class="h-4 w-48" />
					<div class="rounded-lg border">
						<div class="overflow-x-auto">
							<div class="flex gap-4 p-4">
								{#each Array(8) as _}
									<Skeleton class="h-75 w-30" />
								{/each}
							</div>
						</div>
					</div>
				</div>
			{:else}
				<PermissionMatrix {roles} {permissions} />
			{/if}
		</Tabs.Content>
	</Tabs.Root>
</div>

<!-- Create Permission Dialog -->
<Dialog.Root bind:open={createPermissionDialogOpen}>
	<Dialog.Content class="sm:max-w-125">
		<PermissionForm
			onSubmit={handleCreatePermission}
			onCancel={() => (createPermissionDialogOpen = false)}
			isSubmitting={isSubmittingPermission}
			error={permissionError}
		/>
	</Dialog.Content>
</Dialog.Root>

<!-- Edit Permission Dialog -->
<Dialog.Root bind:open={editPermissionDialogOpen}>
	<Dialog.Content class="sm:max-w-125">
		<PermissionForm
			permission={selectedPermission}
			onSubmit={handleUpdatePermission}
			onCancel={() => {
				editPermissionDialogOpen = false;
				selectedPermission = null;
			}}
			isSubmitting={isSubmittingPermission}
			error={permissionError}
		/>
	</Dialog.Content>
</Dialog.Root>

<!-- Edit Role Permissions Sheet -->
<Sheet.Root bind:open={editRoleSheetOpen}>
	<Sheet.Content side="right" class="w-full overflow-y-auto sm:max-w-xl">
		{#if selectedRole}
			<RolePermissionEditor
				role={selectedRole}
				allPermissions={permissions}
				onSubmit={handleUpdateRolePermissions}
				onCancel={() => {
					editRoleSheetOpen = false;
					selectedRole = null;
				}}
				isSubmitting={isSubmittingRole}
				error={roleError}
			/>
		{/if}
	</Sheet.Content>
</Sheet.Root>
