<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import Toasts, { addToast } from '$lib/components/toast.svelte';
	import { createTeam, listTeams, type Team } from '$lib/api/teams.js';
	import { Plus, Search } from '@lucide/svelte';

	type CreateTeamForm = {
		name: string;
		description: string;
	};

	let loading = $state(true);
	let error = $state<string | null>(null);
	let teams = $state<Team[]>([]);
	let search = $state('');
	let searchDebounce = $state<ReturnType<typeof setTimeout> | null>(null);

	let createOpen = $state(false);
	let createBusy = $state(false);
	let form = $state<CreateTeamForm>({ name: '', description: '' });

	async function loadTeams() {
		loading = true;
		error = null;
		try {
			const res = await listTeams({ page: 1, limit: 50, search });
			teams = res.items;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load teams';
		} finally {
			loading = false;
		}
	}

	function onSearchInput(e: Event) {
		search = (e.target as HTMLInputElement).value;
		if (searchDebounce) clearTimeout(searchDebounce);
		searchDebounce = setTimeout(() => loadTeams(), 250);
	}

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
			await loadTeams();
			goto(`/teams/${t.id}`);
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to create team', 'error');
		} finally {
			createBusy = false;
		}
	}

	onMount(() => {
		loadTeams();
	});
</script>

<Toasts />

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

	<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
		<div class="relative max-w-sm flex-1">
			<Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
			<Input
				type="search"
				placeholder="Search teams..."
				class="pl-9"
				value={search}
				oninput={onSearchInput}
			/>
		</div>
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

	{#if error}
		<div class="rounded-md border bg-muted px-4 py-3 text-muted-foreground">
			<p class="font-medium">Could not load teams</p>
			<p class="text-sm">{error}</p>
		</div>
	{/if}

	<div class="rounded-lg border bg-background">
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head>Name</Table.Head>
					<Table.Head>Description</Table.Head>
					<Table.Head class="w-30"></Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#if loading}
					{#each Array(6) as _}
						<Table.Row>
							<Table.Cell><Skeleton class="h-4 w-60" /></Table.Cell>
							<Table.Cell><Skeleton class="h-4 w-70" /></Table.Cell>
							<Table.Cell><Skeleton class="h-8 w-24" /></Table.Cell>
						</Table.Row>
					{/each}
				{:else if teams.length === 0}
					<Table.Row>
						<Table.Cell colspan={3}>
							<div class="flex flex-col items-center justify-center gap-2 py-10 text-center">
								<div class="text-sm font-medium">No teams yet</div>
								<div class="text-sm text-muted-foreground">
									Create your first team to start assigning access.
								</div>
							</div>
						</Table.Cell>
					</Table.Row>
				{:else}
					{#each teams as t (t.id)}
						<Table.Row>
							<Table.Cell class="font-medium">{t.name}</Table.Cell>
							<Table.Cell class="text-muted-foreground">{t.description ?? ''}</Table.Cell>
							<Table.Cell class="text-right">
								<Button variant="outline" onclick={() => goto(`/teams/${t.id}`)}>Manage</Button>
							</Table.Cell>
						</Table.Row>
					{/each}
				{/if}
			</Table.Body>
		</Table.Root>
	</div>
</div>
