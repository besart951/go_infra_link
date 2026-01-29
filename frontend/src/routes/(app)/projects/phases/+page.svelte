<script lang="ts">
	import { onMount } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Plus, Pencil, Trash2, Eye } from '@lucide/svelte';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { addToast } from '$lib/components/toast.svelte';
	import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
	import { confirm } from '$lib/stores/confirm-dialog.js';
	import { phaseListStore } from '$lib/stores/phases/phaseListStore.js';
	import type { Phase } from '$lib/domain/phase/index.js';
	import PhaseForm from '$lib/components/project/PhaseForm.svelte';
	import { deletePhase } from '$lib/infrastructure/api/phase.adapter.js';

	let showForm = $state(false);
	let editingPhase: Phase | undefined = $state(undefined);
	let deleting = $state(false);

	function handleEdit(phase: Phase) {
		editingPhase = phase;
		showForm = true;
	}

	function handleCreate() {
		editingPhase = undefined;
		showForm = true;
	}

	function handleSuccess() {
		showForm = false;
		editingPhase = undefined;
		phaseListStore.reload();
	}

	function handleCancel() {
		showForm = false;
		editingPhase = undefined;
	}

	async function handleDelete(phase: Phase) {
		const ok = await confirm({
			title: 'Delete Phase',
			message: `Delete ${phase.name}? This action cannot be undone.`,
			confirmText: 'Delete',
			cancelText: 'Cancel',
			variant: 'destructive'
		});

		if (!ok) return;
		deleting = true;
		try {
			await deletePhase(phase.id);
			addToast('Phase deleted successfully', 'success');
			phaseListStore.reload();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to delete phase', 'error');
		} finally {
			deleting = false;
		}
	}

	onMount(() => {
		phaseListStore.load();
	});
</script>

<ConfirmDialog />

<svelte:head>
	<title>Phases | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">Phases</h1>
			<p class="text-sm text-muted-foreground">Manage project phases and their assignments.</p>
		</div>
		{#if !showForm}
			<Button onclick={handleCreate}>
				<Plus class="mr-2 size-4" />
				New Phase
			</Button>
		{/if}
	</div>

	{#if showForm}
		<PhaseForm initialData={editingPhase} on:success={handleSuccess} on:cancel={handleCancel} />
	{/if}

	<PaginatedList
		state={$phaseListStore}
		columns={[
			{ key: 'name', label: 'Name' },
			{ key: 'created', label: 'Created' },
			{ key: 'actions', label: 'Actions', width: 'w-[140px]' }
		]}
		searchPlaceholder="Search phases..."
		emptyMessage="No phases found. Create your first phase to get started."
		onSearch={(text) => phaseListStore.search(text)}
		onPageChange={(page) => phaseListStore.goToPage(page)}
		onReload={() => phaseListStore.reload()}
	>
		{#snippet rowSnippet(phase: Phase)}
			<Table.Cell class="font-medium">
				<a href="/projects/phases/{phase.id}" class="hover:underline">
					{phase.name}
				</a>
			</Table.Cell>
			<Table.Cell>
				{phase.created_at ? new Date(phase.created_at).toLocaleDateString() : '—'}
			</Table.Cell>
			<Table.Cell>
				<div class="flex items-center gap-2">
					<Button variant="ghost" size="icon" onclick={() => handleEdit(phase)}>
						<Pencil class="size-4" />
					</Button>
					<Button variant="ghost" size="icon" href="/projects/phases/{phase.id}">
						<Eye class="size-4" />
					</Button>
					<Button
						variant="ghost"
						size="icon"
						disabled={deleting}
						onclick={() => handleDelete(phase)}
					>
						<Trash2 class="size-4 text-destructive" />
					</Button>
				</div>
			</Table.Cell>
		{/snippet}
	</PaginatedList>
</div>
