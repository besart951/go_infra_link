<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import * as Command from '$lib/components/ui/command/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { addToast } from '$lib/components/toast.svelte';
	import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
	import { confirm } from '$lib/stores/confirm-dialog.js';
	import UserAvatar from '$lib/components/user-avatar.svelte';
	import { ArrowLeft, UserMinus, UserPlus } from '@lucide/svelte';

	import {
		addTeamMember,
		getTeam,
		listTeamMembers,
		removeTeamMember,
		type Team,
		type TeamMember
	} from '$lib/api/teams.js';
	import { listUsers, type User } from '$lib/api/users.js';

	const teamId = $derived($page.params.id ?? '');

	let team = $state<Team | null>(null);
	let members = $state<TeamMember[]>([]);
	let users = $state<User[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let busy = $state(false);

	// Add Member popover state
	let addMemberOpen = $state(false);
	let addMemberSearch = $state('');
	let addMemberResults = $state<User[]>([]);
	let addMemberLoading = $state(false);
	let debounceTimer: ReturnType<typeof setTimeout> | undefined;

	function userById(id: string): User | undefined {
		return users.find((u) => u.id === id);
	}

	function memberUserIds(): Set<string> {
		return new Set(members.map((m) => m.user_id));
	}

	async function load() {
		if (!teamId) {
			error = 'Missing team id';
			loading = false;
			return;
		}
		loading = true;
		error = null;
		try {
			const [t, m, u] = await Promise.all([
				getTeam(teamId),
				listTeamMembers(teamId, { page: 1, limit: 100 }),
				listUsers({ page: 1, limit: 100, search: '' })
			]);
			team = t;
			members = m.items;
			users = u.items;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load team';
		} finally {
			loading = false;
		}
	}

	async function searchUsers(query: string) {
		addMemberLoading = true;
		try {
			const res = await listUsers({ page: 1, limit: 20, search: query });
			const existingIds = memberUserIds();
			addMemberResults = res.items.filter((u) => !existingIds.has(u.id));
		} catch {
			addMemberResults = [];
		} finally {
			addMemberLoading = false;
		}
	}

	$effect(() => {
		const query = addMemberSearch;
		clearTimeout(debounceTimer);
		debounceTimer = setTimeout(() => {
			searchUsers(query);
		}, 300);
	});

	async function handleAddMember(userId: string) {
		if (!teamId) return;
		busy = true;
		try {
			await addTeamMember(teamId, { user_id: userId, role: 'member' });
			addToast('Member added', 'success');
			addMemberOpen = false;
			addMemberSearch = '';
			addMemberResults = [];
			await load();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to add member', 'error');
		} finally {
			busy = false;
		}
	}

	async function changeRole(userId: string, role: 'member' | 'manager' | 'owner') {
		if (!teamId) return;
		busy = true;
		try {
			await addTeamMember(teamId, { user_id: userId, role });
			addToast('Role updated', 'success');
			await load();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to update role', 'error');
		} finally {
			busy = false;
		}
	}

	async function remove(userId: string) {
		if (!teamId) return;
		const u = userById(userId);
		const ok = await confirm({
			title: 'Remove member',
			message: `Remove ${u ? `${u.first_name} ${u.last_name}` : 'this user'} from the team?`,
			confirmText: 'Remove',
			cancelText: 'Cancel',
			variant: 'destructive'
		});
		if (!ok) return;

		busy = true;
		try {
			await removeTeamMember(teamId, userId);
			addToast('Member removed', 'success');
			await load();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to remove member', 'error');
		} finally {
			busy = false;
		}
	}

	$effect(() => {
		if (addMemberOpen) {
			searchUsers('');
		}
	});

	onMount(() => {
		load();
	});
</script>

<ConfirmDialog />

<div class="flex flex-col gap-6">
	<div class="flex items-start justify-between">
		<div class="flex items-start gap-3">
			<Button variant="outline" onclick={() => goto('/teams')}>
				<ArrowLeft class="mr-2 h-4 w-4" />
				Back
			</Button>
			<div>
				<h1 class="text-3xl font-bold tracking-tight">{team?.name ?? 'Team'}</h1>
				<p class="mt-1 text-muted-foreground">Manage members and permissions.</p>
			</div>
		</div>
		<Popover.Root bind:open={addMemberOpen}>
			<Popover.Trigger>
				{#snippet child({ props })}
					<Button {...props}>
						<UserPlus class="mr-2 h-4 w-4" />
						Add Member
					</Button>
				{/snippet}
			</Popover.Trigger>
			<Popover.Content class="w-72 p-0" align="end">
				<Command.Root shouldFilter={false}>
					<Command.Input
						placeholder="Search users..."
						bind:value={addMemberSearch}
					/>
					<Command.List>
						<Command.Empty>
							{addMemberLoading ? 'Searching...' : 'No users found'}
						</Command.Empty>
						<Command.Group>
							{#each addMemberResults as user (user.id)}
								<Command.Item
									value={user.id}
									onSelect={() => handleAddMember(user.id)}
								>
									<div class="flex items-center gap-2">
										<UserAvatar
											firstName={user.first_name}
											lastName={user.last_name}
											class="h-6 w-6"
										/>
										<div class="flex flex-col">
											<span class="text-sm">{user.first_name} {user.last_name}</span>
											<span class="text-xs text-muted-foreground">{user.email}</span>
										</div>
									</div>
								</Command.Item>
							{/each}
						</Command.Group>
					</Command.List>
				</Command.Root>
			</Popover.Content>
		</Popover.Root>
	</div>

	{#if team?.description}
		<div class="text-sm text-muted-foreground">{team.description}</div>
	{/if}

	{#if error}
		<div class="rounded-md border bg-muted px-4 py-3 text-muted-foreground">
			<p class="font-medium">Could not load team</p>
			<p class="text-sm">{error}</p>
		</div>
	{/if}

	<div class="rounded-lg border bg-background">
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head>User</Table.Head>
					<Table.Head>Role</Table.Head>
					<Table.Head class="w-30"></Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#if loading}
					{#each Array(6) as _}
						<Table.Row>
							<Table.Cell><Skeleton class="h-4 w-70" /></Table.Cell>
							<Table.Cell><Skeleton class="h-4 w-30" /></Table.Cell>
							<Table.Cell><Skeleton class="h-8 w-24" /></Table.Cell>
						</Table.Row>
					{/each}
				{:else if members.length === 0}
					<Table.Row>
						<Table.Cell colspan={3}>
							<div class="flex flex-col items-center justify-center gap-2 py-10 text-center">
								<div class="text-sm font-medium">No members yet</div>
								<p class="text-sm text-muted-foreground">Use the "Add Member" button to add users to this team.</p>
							</div>
						</Table.Cell>
					</Table.Row>
				{:else}
					{#each members as m (m.user_id)}
						<Table.Row>
							<Table.Cell>
								{#if userById(m.user_id)}
									{@const u = userById(m.user_id)!}
									<div class="flex items-center gap-3">
										<UserAvatar firstName={u.first_name} lastName={u.last_name} />
										<div class="flex flex-col">
											<div class="font-medium">
												{u.first_name}
												{u.last_name}
											</div>
											<div class="text-sm text-muted-foreground">{u.email}</div>
										</div>
									</div>
								{:else}
									<div class="font-medium">{m.user_id}</div>
								{/if}
							</Table.Cell>
							<Table.Cell>
								<select
									class="flex h-8 rounded-md border border-input bg-transparent px-2 text-sm shadow-sm"
									onchange={(e) =>
										changeRole(m.user_id, (e.target as HTMLSelectElement).value as any)}
									disabled={busy}
								>
									<option value="member" selected={m.role === 'member'}>Member</option>
									<option value="manager" selected={m.role === 'manager'}>Manager</option>
									<option value="owner" selected={m.role === 'owner'}>Owner</option>
								</select>
							</Table.Cell>
							<Table.Cell class="text-right">
								<Button variant="outline" onclick={() => remove(m.user_id)} disabled={busy}>
									<UserMinus class="mr-2 h-4 w-4" />
									Remove
								</Button>
							</Table.Cell>
						</Table.Row>
					{/each}
				{/if}
			</Table.Body>
		</Table.Root>
	</div>
</div>
