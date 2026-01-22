<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import Toasts, { addToast } from '$lib/components/toast.svelte';
	import ConfirmDialog, { confirm } from '$lib/components/confirm-dialog.svelte';
	import {
		listUsers,
		setUserRole,
		disableUser,
		enableUser,
		deleteUser,
		type User,
		type PaginatedUserResponse
	} from '$lib/api/users.js';
	import {
		Search,
		UserCircle,
		Shield,
		ShieldCheck,
		MoreVertical,
		UserMinus,
		UserCheck,
		Trash2,
		ChevronUp,
		ChevronDown
	} from 'lucide-svelte';

	let users = $state<User[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let searchQuery = $state('');
	let currentPage = $state(1);
	let totalPages = $state(1);
	let total = $state(0);
	let orderBy = $state('last_login_at');
	let order = $state<'asc' | 'desc'>('desc');

	async function loadUsers() {
		loading = true;
		error = null;
		try {
			const response: PaginatedUserResponse = await listUsers({
				page: currentPage,
				limit: 10,
				search: searchQuery,
				order_by: orderBy,
				order
			});
			users = response.items;
			totalPages = response.total_pages;
			total = response.total;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load users';
		} finally {
			loading = false;
		}
	}

	function handleSort(field: string) {
		if (orderBy === field) {
			order = order === 'asc' ? 'desc' : 'asc';
		} else {
			orderBy = field;
			order = 'desc';
		}
		loadUsers();
	}

	function handleSearch(event: Event) {
		const target = event.target as HTMLInputElement;
		searchQuery = target.value;
		currentPage = 1;
		loadUsers();
	}

	async function handleRoleChange(userId: string, newRole: 'user' | 'admin' | 'superadmin') {
		try {
			await setUserRole(userId, newRole);
			await loadUsers();
			addToast('Role updated successfully', 'success');
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to change role', 'error');
		}
	}

	async function handleToggleActive(userId: string, isActive: boolean) {
		try {
			if (isActive) {
				await disableUser(userId);
				addToast('User disabled successfully', 'success');
			} else {
				await enableUser(userId);
				addToast('User enabled successfully', 'success');
			}
			await loadUsers();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to toggle user status', 'error');
		}
	}

	async function handleDeleteUser(userId: string, userName: string) {
		const confirmed = await confirm({
			title: 'Delete User',
			message: `Are you sure you want to delete ${userName}? This action cannot be undone.`,
			confirmText: 'Delete',
			cancelText: 'Cancel',
			variant: 'destructive'
		});

		if (confirmed) {
			try {
				await deleteUser(userId);
				await loadUsers();
				addToast('User deleted successfully', 'success');
			} catch (err) {
				addToast(err instanceof Error ? err.message : 'Failed to delete user', 'error');
			}
		}
	}

	function formatDate(dateString: string | null | undefined): string {
		if (!dateString) return 'Never';
		const date = new Date(dateString);
		const now = new Date();
		const diffInMs = now.getTime() - date.getTime();
		const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24));

		if (diffInDays === 0) return 'Today';
		if (diffInDays === 1) return 'Yesterday';
		if (diffInDays < 7) return `${diffInDays} days ago`;
		if (diffInDays < 30) return `${Math.floor(diffInDays / 7)} weeks ago`;
		if (diffInDays < 365) return `${Math.floor(diffInDays / 30)} months ago`;
		return `${Math.floor(diffInDays / 365)} years ago`;
	}

	function getRoleBadgeVariant(role: string) {
		if (role === 'superadmin') return 'default';
		if (role === 'admin') return 'secondary';
		return 'outline';
	}

	function getRoleIcon(role: string) {
		if (role === 'superadmin' || role === 'admin') return ShieldCheck;
		return UserCircle;
	}

	onMount(() => {
		loadUsers();
	});

	let showActionsMenu = $state<string | null>(null);
</script>

<div class="flex flex-col gap-6">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-3xl font-bold tracking-tight">User Management</h1>
			<p class="text-muted-foreground mt-1">Manage all users and their permissions</p>
		</div>
	</div>

	<!-- Search and Filters -->
	<div class="flex items-center gap-4">
		<div class="relative flex-1 max-w-sm">
			<Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
			<Input
				type="search"
				placeholder="Search users by name or email..."
				class="pl-9"
				value={searchQuery}
				oninput={handleSearch}
			/>
		</div>
		<div class="text-sm text-muted-foreground">
			{total} {total === 1 ? 'user' : 'users'} total
		</div>
	</div>

	{#if error}
		<div
			class="bg-destructive/15 text-destructive border-destructive/50 rounded-md border px-4 py-3"
		>
			<p class="font-medium">Error loading users</p>
			<p class="text-sm">{error}</p>
		</div>
	{/if}

	<!-- Table -->
	<div class="border rounded-lg">
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head>
						<button
							class="flex items-center gap-1 font-medium hover:underline"
							onclick={() => handleSort('first_name')}
						>
							Name/Email
							{#if orderBy === 'first_name'}
								{#if order === 'asc'}
									<ChevronUp class="h-4 w-4" />
								{:else}
									<ChevronDown class="h-4 w-4" />
								{/if}
							{/if}
						</button>
					</Table.Head>
					<Table.Head>
						<button
							class="flex items-center gap-1 font-medium hover:underline"
							onclick={() => handleSort('role')}
						>
							Role
							{#if orderBy === 'role'}
								{#if order === 'asc'}
									<ChevronUp class="h-4 w-4" />
								{:else}
									<ChevronDown class="h-4 w-4" />
								{/if}
							{/if}
						</button>
					</Table.Head>
					<Table.Head>Status</Table.Head>
					<Table.Head>
						<button
							class="flex items-center gap-1 font-medium hover:underline"
							onclick={() => handleSort('last_login_at')}
						>
							Last Active
							{#if orderBy === 'last_login_at'}
								{#if order === 'asc'}
									<ChevronUp class="h-4 w-4" />
								{:else}
									<ChevronDown class="h-4 w-4" />
								{/if}
							{/if}
						</button>
					</Table.Head>
					<Table.Head class="text-right">Actions</Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#if loading}
					{#each Array(5) as _}
						<Table.Row>
							<Table.Cell>
								<Skeleton class="h-10 w-full" />
							</Table.Cell>
							<Table.Cell>
								<Skeleton class="h-6 w-20" />
							</Table.Cell>
							<Table.Cell>
								<Skeleton class="h-6 w-16" />
							</Table.Cell>
							<Table.Cell>
								<Skeleton class="h-5 w-24" />
							</Table.Cell>
							<Table.Cell>
								<Skeleton class="h-8 w-8" />
							</Table.Cell>
						</Table.Row>
					{/each}
				{:else if users.length === 0}
					<Table.Row>
						<Table.Cell colspan={5} class="h-24 text-center">
							<div class="flex flex-col items-center justify-center gap-2 text-muted-foreground">
								<UserCircle class="h-12 w-12" />
								<p class="font-medium">No users found</p>
								<p class="text-sm">
									{searchQuery ? 'Try adjusting your search query' : 'No users in the system yet'}
								</p>
							</div>
						</Table.Cell>
					</Table.Row>
				{:else}
					{#each users as user (user.id)}
						<Table.Row>
							<Table.Cell>
								<div class="flex flex-col">
									<div class="font-medium">
										{user.first_name}
										{user.last_name}
									</div>
									<div class="text-sm text-muted-foreground">{user.email}</div>
								</div>
							</Table.Cell>
							<Table.Cell>
								<Badge variant={getRoleBadgeVariant(user.role)}>
									{@const Icon = getRoleIcon(user.role)}
									<Icon class="mr-1 h-3 w-3" />
									{user.role}
								</Badge>
							</Table.Cell>
							<Table.Cell>
								{#if user.disabled_at}
									<Badge variant="destructive">Disabled</Badge>
								{:else if user.locked_until}
									<Badge variant="warning">Locked</Badge>
								{:else if user.is_active}
									<Badge variant="success">Active</Badge>
								{:else}
									<Badge variant="outline">Inactive</Badge>
								{/if}
							</Table.Cell>
							<Table.Cell>
								<span class="text-sm">{formatDate(user.last_login_at)}</span>
							</Table.Cell>
							<Table.Cell class="text-right">
								<div class="relative inline-block">
									<Button
										variant="ghost"
										size="sm"
										onclick={() =>
											(showActionsMenu = showActionsMenu === user.id ? null : user.id)}
									>
										<MoreVertical class="h-4 w-4" />
									</Button>
									{#if showActionsMenu === user.id}
										<div
											class="absolute right-0 z-10 mt-2 w-56 rounded-md border bg-popover p-1 shadow-md"
										>
											<div class="px-2 py-1.5 text-sm font-medium">Change Role</div>
											<button
												class="flex w-full items-center rounded-sm px-2 py-1.5 text-sm hover:bg-accent"
												onclick={() => {
													handleRoleChange(user.id, 'user');
													showActionsMenu = null;
												}}
											>
												<UserCircle class="mr-2 h-4 w-4" />
												User
											</button>
											<button
												class="flex w-full items-center rounded-sm px-2 py-1.5 text-sm hover:bg-accent"
												onclick={() => {
													handleRoleChange(user.id, 'admin');
													showActionsMenu = null;
												}}
											>
												<Shield class="mr-2 h-4 w-4" />
												Admin
											</button>
											<button
												class="flex w-full items-center rounded-sm px-2 py-1.5 text-sm hover:bg-accent"
												onclick={() => {
													handleRoleChange(user.id, 'superadmin');
													showActionsMenu = null;
												}}
											>
												<ShieldCheck class="mr-2 h-4 w-4" />
												Super Admin
											</button>
											<div class="my-1 h-px bg-border"></div>
											<button
												class="flex w-full items-center rounded-sm px-2 py-1.5 text-sm hover:bg-accent"
												onclick={() => {
													handleToggleActive(user.id, user.is_active);
													showActionsMenu = null;
												}}
											>
												{#if user.is_active}
													<UserMinus class="mr-2 h-4 w-4" />
													Disable User
												{:else}
													<UserCheck class="mr-2 h-4 w-4" />
													Enable User
												{/if}
											</button>
											<div class="my-1 h-px bg-border"></div>
											<button
												class="text-destructive flex w-full items-center rounded-sm px-2 py-1.5 text-sm hover:bg-destructive/10"
												onclick={() => {
													handleDeleteUser(user.id, `${user.first_name} ${user.last_name}`);
													showActionsMenu = null;
												}}
											>
												<Trash2 class="mr-2 h-4 w-4" />
												Delete User
											</button>
										</div>
									{/if}
								</div>
							</Table.Cell>
						</Table.Row>
					{/each}
				{/if}
			</Table.Body>
		</Table.Root>
	</div>

	<!-- Pagination -->
	{#if totalPages > 1}
		<div class="flex items-center justify-between">
			<div class="text-sm text-muted-foreground">
				Page {currentPage} of {totalPages}
			</div>
			<div class="flex items-center gap-2">
				<Button
					variant="outline"
					size="sm"
					disabled={currentPage <= 1}
					onclick={() => {
						currentPage--;
						loadUsers();
					}}
				>
					Previous
				</Button>
				<Button
					variant="outline"
					size="sm"
					disabled={currentPage >= totalPages}
					onclick={() => {
						currentPage++;
						loadUsers();
					}}
				>
					Next
				</Button>
			</div>
		</div>
	{/if}
</div>

<Toasts />
<ConfirmDialog />

