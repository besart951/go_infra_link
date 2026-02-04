<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { addToast } from '$lib/components/toast.svelte';
	import { confirm } from '$lib/stores/confirm-dialog.js';
	import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
	import { createTeam, deleteTeam, listTeamMembers } from '$lib/api/teams.js';
	import { Plus, Trash2 } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { teamsStore } from '$lib/stores/list/entityStores.js';
	import type { Team } from '$lib/domain/entities/team.js';

	type CreateTeamForm = {
		name: string;
		description: string;
	};

	let createOpen = $state(false);
	let createBusy = $state(false);
	let form = $state<CreateTeamForm>({ name: '', description: '' });
	let memberCounts = $state<Map<string, number>>(new Map());

	function canSubmitCreate(): boolean {
		return form.name.trim().length > 0 && !createBusy;
	}

	async function submitCreate() {
		if (!canSubmitCreate()) return;
		createBusy = true;
		try {
			const t = await createTeam({
				name: form.name.trim(),
				description: form.description.trim() ? form.description.trim() : null
			});
			addToast('Team created', 'success');
			form = { name: '', description: '' };
			createOpen = false;
			teamsStore.reload();
			goto(`/teams/${t.id}`);
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to create team', 'error');
		} finally {
			createBusy = false;
		}
	}

	async function handleDeleteTeam(team: Team) {
		const confirmed = await confirm({
			title: 'Delete Team',
			message: `Are you sure you want to delete "${team.name}"? This action cannot be undone.`,
			confirmText: 'Delete',
			cancelText: 'Cancel',
			variant: 'destructive'
		});

		if (confirmed) {
			try {
				await deleteTeam(team.id);
				teamsStore.reload();
				addToast('Team deleted successfully', 'success');
			} catch (err) {
				addToast(err instanceof Error ? err.message : 'Failed to delete team', 'error');
			}
		}
	}

	async function loadMemberCounts(teams: Team[]) {
		const counts = new Map<string, number>();
		await Promise.all(
			teams.map(async (t) => {
				try {
					const res = await listTeamMembers(t.id, { page: 1, limit: 1 });
					counts.set(t.id, res.total);
				} catch {
					counts.set(t.id, 0);
				}
			})
		);
		memberCounts = counts;
	}

	$effect(() => {
		const items = $teamsStore.items;
		if (items.length > 0) {
			loadMemberCounts(items);
		}
	});

	onMount(() => {
		teamsStore.load();
	});
</script>

<div class="flex flex-col gap-6">
	<div class="flex items-start justify-between gap-4">
		<div>
			<h1 class="text-3xl font-bold tracking-tight">Teams</h1>
			<p class="mt-1 text-muted-foreground">Create teams and manage access.</p>
		</div>
		<Button variant="outline" onclick={() => (createOpen = !createOpen)}>
			<Plus class="mr-2 h-4 w-4" />
			Create team
		</Button>
	</div>

	{#if createOpen}
		<div class="rounded-lg border bg-background p-4">
			<div class="grid gap-3 md:grid-cols-3">
				<div class="md:col-span-1">
					<label class="text-sm font-medium" for="team_name">Name</label>
					<Input
						id="team_name"
						placeholder="Operations"
						bind:value={form.name}
						disabled={createBusy}
					/>
				</div>
				<div class="md:col-span-2">
					<label class="text-sm font-medium" for="team_desc">Description (optional)</label>
					<Input
						id="team_desc"
						placeholder="Optional"
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

	<PaginatedList
		state={$teamsStore}
		columns={[
			{ key: 'name', label: 'Name' },
			{ key: 'description', label: 'Description' },
			{ key: 'members', label: 'Members', width: 'w-24' },
			{ key: 'actions', label: '', width: 'w-40' }
		]}
		searchPlaceholder="Search teams..."
		emptyMessage="No teams yet. Create your first team to start assigning access."
		onSearch={(text) => teamsStore.search(text)}
		onPageChange={(page) => teamsStore.goToPage(page)}
		onReload={() => teamsStore.reload()}
	>
		{#snippet rowSnippet(team: Team)}
			<Table.Cell class="font-medium">{team.name}</Table.Cell>
			<Table.Cell class="text-muted-foreground">{team.description ?? ''}</Table.Cell>
			<Table.Cell>
				{@const count = memberCounts.get(team.id)}
				{#if count !== undefined}
					<Badge variant="secondary">{count}</Badge>
				{:else}
					<span class="text-sm text-muted-foreground">&mdash;</span>
				{/if}
			</Table.Cell>
			<Table.Cell class="text-right">
				<div class="flex items-center justify-end gap-2">
					<Button variant="outline" onclick={() => goto(`/teams/${team.id}`)}>Manage</Button>
					<Button
						variant="outline"
						size="icon"
						onclick={() => handleDeleteTeam(team)}
					>
						<Trash2 class="h-4 w-4 text-destructive" />
					</Button>
				</div>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>

<ConfirmDialog />
