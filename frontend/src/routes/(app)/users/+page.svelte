<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import * as Tooltip from '$lib/components/ui/tooltip/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { addToast } from '$lib/components/toast.svelte';
	import { confirm } from '$lib/stores/confirm-dialog.js';
	import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
	import RoleBadge from '$lib/components/role-badge.svelte';
	import UserAvatar from '$lib/components/user-avatar.svelte';
	import UserManagementForm from '$lib/components/user-management-form.svelte';
	import { setUserRole, disableUser, enableUser, deleteUser } from '$lib/api/users.js';
	import type { UserRole } from '$lib/api/users.js';
	import { listTeams, listTeamMembers } from '$lib/api/teams.js';
	import type { Team } from '$lib/domain/entities/team.js';
	import type { User } from '$lib/domain/entities/user.js';
	import { getAllowedRolesForCreation } from '$lib/stores/auth.svelte.js';
	import { getRoleLabel } from '$lib/utils/permissions.js';
	import {
		MoreVertical,
		UserMinus,
		UserCheck,
		Trash2,
		BadgeCheck,
		BadgeX,
		KeyRound,
		UserPlus
	} from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { usersStore } from '$lib/stores/list/entityStores.js';
	import { createTranslator } from '$lib/i18n/translator';

	const t = createTranslator();

	let teams = $state<Team[]>([]);
	let teamByUserId = $state<Map<string, string[]>>(new Map());
	let selectedTeamId = $state<string>('all');
	let teamsLoading = $state(true);
	let teamsError = $state<string | null>(null);
	let createDialogOpen = $state(false);

	function getUserTeams(userId: string): string[] {
		return teamByUserId.get(userId) ?? [];
	}

	function userMatchesTeam(userId: string): boolean {
		if (selectedTeamId === 'all') return true;
		const names = getUserTeams(userId);
		const t = teams.find((x) => x.id === selectedTeamId);
		if (!t) return true;
		return names.includes(t.name);
	}

	function visibleUsers(): User[] {
		if (selectedTeamId === 'all') return $usersStore.items;
		return $usersStore.items.filter((u) => userMatchesTeam(u.id));
	}

	async function loadTeamsAndMembers() {
		teamsLoading = true;
		teamsError = null;
		try {
			const res = await listTeams({ page: 1, limit: 100, search: '' });
			teams = res.items;

			const memberLists = await Promise.all(
				teams.map(async (t) => ({
					team: t,
					members: await listTeamMembers(t.id, { page: 1, limit: 1000 })
				}))
			);

			const map = new Map<string, string[]>();
			for (const { team, members } of memberLists) {
				for (const m of members.items) {
					const arr = map.get(m.user_id) ?? [];
					arr.push(team.name);
					map.set(m.user_id, arr);
				}
			}
			teamByUserId = map;
		} catch (err) {
			teamsError = err instanceof Error ? err.message : 'Failed to load teams';
		} finally {
			teamsLoading = false;
		}
	}

	async function handleRoleChange(userId: string, newRole: UserRole) {
		try {
			await setUserRole(userId, newRole);
			usersStore.reload();
			addToast($t('messages.role_updated_success'), 'success');
		} catch (err) {
			addToast(err instanceof Error ? err.message : $t('errors.change_role_failed'), 'error');
		}
	}

	async function handleToggleActive(userId: string, isActive: boolean) {
		try {
			if (isActive) {
				await disableUser(userId);
				addToast($t('messages.user_disabled_success'), 'success');
			} else {
				await enableUser(userId);
				addToast($t('messages.user_enabled_success'), 'success');
			}
			usersStore.reload();
		} catch (err) {
			addToast(err instanceof Error ? err.message : $t('errors.toggle_user_status_failed'), 'error');
		}
	}

	async function handleDeleteUser(userId: string, userName: string) {
		const confirmed = await confirm({
			title: $t('common.delete_user'),
			message: $t('messages.delete_user_confirm', { name: userName }),
			confirmText: $t('common.delete'),
			cancelText: $t('common.cancel'),
			variant: 'destructive'
		});

		if (confirmed) {
			try {
				await deleteUser(userId);
				usersStore.reload();
				addToast($t('messages.user_deleted_success'), 'success');
			} catch (err) {
				addToast(err instanceof Error ? err.message : $t('errors.delete_user_failed'), 'error');
			}
		}
	}

	function formatDate(dateString: string | null | undefined): string {
		if (!dateString) return $t('messages.never');
		const date = new Date(dateString);
		const now = new Date();
		const diffInMs = now.getTime() - date.getTime();
		const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24));

		if (diffInDays === 0) return $t('messages.today');
		if (diffInDays === 1) return $t('messages.yesterday');
		if (diffInDays < 7) return $t('messages.days_ago').replace('{count}', String(diffInDays));
		if (diffInDays < 30) return $t('messages.weeks_ago').replace('{count}', String(Math.floor(diffInDays / 7)));
		if (diffInDays < 365) return $t('messages.months_ago').replace('{count}', String(Math.floor(diffInDays / 30)));
		return $t('messages.years_ago').replace('{count}', String(Math.floor(diffInDays / 365)));
	}

	function authVerified(user: User): boolean {
		return Boolean(user.is_active && !user.disabled_at);
	}

	function twoFactorEnabled(_user: User): boolean {
		return false;
	}

	onMount(() => {
		loadTeamsAndMembers();
		usersStore.load();
	});
</script>

<svelte:head>
	<title>{$t('navigation.users')} | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-3xl font-bold tracking-tight">{$t('pages.user_management')}</h1>
			<p class="mt-1 text-muted-foreground">{$t('pages.user_management_desc')}</p>
		</div>
		<Button onclick={() => (createDialogOpen = true)}>
			<UserPlus class="mr-2 h-4 w-4" />
			{$t('common.create_user')}
		</Button>
	</div>

	<!-- Team Filter -->
	<div class="flex items-center justify-end gap-3">
		<div class="text-sm text-muted-foreground">
			{#if selectedTeamId === 'all'}
				{$usersStore.total} {$usersStore.total === 1 ? $t('common.user') : $t('common.users')} {$t('common.total')}
			{:else}
				{visibleUsers().length} {$t('common.shown')} • {$usersStore.total} {$t('common.total')}
			{/if}
		</div>
		<div class="flex items-center gap-2">
			<span class="text-sm text-muted-foreground">{$t('common.team')}</span>
			<select
				class="h-9 rounded-md border bg-background px-3 text-sm"
				bind:value={selectedTeamId}
				disabled={teamsLoading || teams.length === 0}
			>
				<option value="all">{$t('common.all_teams')}</option>
				{#each teams as t (t.id)}
					<option value={t.id}>{t.name}</option>
				{/each}
			</select>
		</div>
	</div>

	{#if teamsError}
		<div class="rounded-md border bg-muted px-4 py-3 text-muted-foreground">
			<p class="font-medium">{$t('messages.teams_unavailable')}</p>
			<p class="text-sm">{teamsError}</p>
		</div>
	{/if}

	<PaginatedList
		state={$usersStore}
		columns={[
			{ key: 'name', label: $t('common.name_email') },
			{ key: 'team', label: $t('common.team') },
			{ key: 'role', label: $t('common.role') },
			{ key: 'auth', label: $t('common.auth') },
			{ key: 'status', label: $t('common.status') },
			{ key: 'last_active', label: $t('common.last_active') },
			{ key: 'actions', label: $t('common.actions'), width: 'text-right' }
		]}
		searchPlaceholder={$t('messages.search_users')}
		emptyMessage={$t('messages.no_users_found')}
		onSearch={(text) => usersStore.search(text)}
		onPageChange={(page) => usersStore.goToPage(page)}
		onReload={() => usersStore.reload()}
	>
		{#snippet rowSnippet(user: User)}
			{@const isVisible = userMatchesTeam(user.id)}
			{#if isVisible || selectedTeamId === 'all'}
				<Table.Cell>
					<div class="flex items-center gap-3">
						<UserAvatar firstName={user.first_name} lastName={user.last_name} />
						<div class="flex flex-col">
							<div class="font-medium">
								{user.first_name}
								{user.last_name}
							</div>
							<div class="text-sm text-muted-foreground">{user.email}</div>
						</div>
					</div>
				</Table.Cell>
				<Table.Cell>
					{@const tnames = getUserTeams(user.id)}
					{#if tnames.length === 0}
						<span class="text-sm text-muted-foreground">&mdash;</span>
					{:else}
						<div class="flex items-center gap-2">
							<span class="text-sm font-medium">{tnames[0]}</span>
							{#if tnames.length > 1}
								<Tooltip.Root>
									<Tooltip.Trigger class="inline-flex">
										<Badge variant="outline">+{tnames.length - 1}</Badge>
									</Tooltip.Trigger>
									<Tooltip.Content class="max-w-xs">
										<div class="text-sm">{tnames.join(', ')}</div>
									</Tooltip.Content>
								</Tooltip.Root>
							{/if}
						</div>
					{/if}
				</Table.Cell>
				<Table.Cell>
					<RoleBadge role={user.role} />
				</Table.Cell>
				<Table.Cell>
					<div class="flex items-center gap-2">
						<Tooltip.Root>
							<Tooltip.Trigger class="inline-flex">
								{#if authVerified(user)}
									<Badge variant="success">
										<BadgeCheck class="mr-1 h-3 w-3" />
										{$t('common.verified')}
									</Badge>
								{:else}
									<Badge variant="outline">
										<BadgeX class="mr-1 h-3 w-3" />
										{$t('common.unverified')}
									</Badge>
								{/if}
							</Tooltip.Trigger>
							<Tooltip.Content>
								<div class="text-sm">{$t('messages.email_verification_info')}</div>
							</Tooltip.Content>
						</Tooltip.Root>

						<Tooltip.Root>
							<Tooltip.Trigger class="inline-flex">
								{#if twoFactorEnabled(user)}
									<Badge variant="secondary">
										<KeyRound class="mr-1 h-3 w-3" />
										{$t('common.2fa')}
									</Badge>
								{:else}
									<Badge variant="outline">
										<KeyRound class="mr-1 h-3 w-3" />
										{$t('common.2fa_off')}
									</Badge>
								{/if}
							</Tooltip.Trigger>
							<Tooltip.Content>
								<div class="text-sm">{$t('messages.2fa_not_implemented')}</div>
							</Tooltip.Content>
						</Tooltip.Root>
					</div>
				</Table.Cell>
				<Table.Cell>
					{#if user.disabled_at}
						<Badge variant="destructive">{$t('common.disabled')}</Badge>
					{:else if user.locked_until}
						<Badge variant="warning">{$t('common.locked')}</Badge>
					{:else if user.is_active}
						<Badge variant="success">{$t('common.active')}</Badge>
					{:else}
						<Badge variant="outline">{$t('common.inactive')}</Badge>
					{/if}
				</Table.Cell>
				<Table.Cell>
					<span class="text-sm">{formatDate(user.last_login_at)}</span>
				</Table.Cell>
				<Table.Cell class="text-right">
					<DropdownMenu.Root>
						<DropdownMenu.Trigger>
							{#snippet child({ props })}
								<Button variant="ghost" size="sm" {...props}>
									<MoreVertical class="h-4 w-4" />
								</Button>
							{/snippet}
						</DropdownMenu.Trigger>
						<DropdownMenu.Content align="end" class="w-56">
							<DropdownMenu.Label>{$t('common.change_role')}</DropdownMenu.Label>
							<DropdownMenu.Separator />
							{#each getAllowedRolesForCreation() as role (role)}
								<DropdownMenu.Item
									disabled={user.role === role}
									onclick={() => handleRoleChange(user.id, role)}
								>
									{getRoleLabel(role)}
									{#if user.role === role}
										<DropdownMenu.Shortcut>{$t('common.current')}</DropdownMenu.Shortcut>
									{/if}
								</DropdownMenu.Item>
							{/each}
							<DropdownMenu.Separator />
							<DropdownMenu.Item
								onclick={() => handleToggleActive(user.id, user.is_active)}
							>
								{#if user.is_active}
									<UserMinus class="mr-2 h-4 w-4" />
									{$t('actions.disable_user')}
								{:else}
									<UserCheck class="mr-2 h-4 w-4" />
									{$t('actions.enable_user')}
								{/if}
							</DropdownMenu.Item>
							<DropdownMenu.Separator />
							<DropdownMenu.Item
								class="text-destructive"
								onclick={() =>
									handleDeleteUser(user.id, `${user.first_name} ${user.last_name}`)}
							>
								<Trash2 class="mr-2 h-4 w-4" />
								{$t('actions.delete_user')}
							</DropdownMenu.Item>
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				</Table.Cell>
			{/if}
		{/snippet}
	</PaginatedList>
</div>

<Dialog.Root bind:open={createDialogOpen}>
	<Dialog.Content class="sm:max-w-lg">
		<Dialog.Header>
			<Dialog.Title>{$t('common.create_user')}</Dialog.Title>
			<Dialog.Description>{$t('messages.add_new_user')}</Dialog.Description>
		</Dialog.Header>
		<UserManagementForm
			onSuccess={() => {
				createDialogOpen = false;
				usersStore.reload();
				addToast($t('messages.user_created_success'), 'success');
			}}
			onCancel={() => (createDialogOpen = false)}
		/>
	</Dialog.Content>
</Dialog.Root>

<ConfirmDialog />
