<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import Toasts, { addToast } from '$lib/components/toast.svelte';
	import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
	import { confirm } from '$lib/stores/confirm-dialog.js';
	import { ArrowLeft, Trash2 } from '@lucide/svelte';
	import type { Phase } from '$lib/domain/phase/index.js';
	import { deletePhase, getPhase } from '$lib/infrastructure/api/phase.adapter.js';

	const phaseId = $derived($page.params.id ?? '');

	let phase = $state<Phase | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let busy = $state(false);

	async function load() {
		if (!phaseId) {
			error = 'Missing phase id';
			loading = false;
			return;
		}

		loading = true;
		error = null;
		try {
			phase = await getPhase(phaseId);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load phase';
		} finally {
			loading = false;
		}
	}

	async function handleDelete() {
		if (!phase) return;
		const ok = await confirm({
			title: 'Delete Phase',
			message: `Delete ${phase.name}? This action cannot be undone.`,
			confirmText: 'Delete',
			cancelText: 'Cancel',
			variant: 'destructive'
		});

		if (!ok) return;
		busy = true;
		try {
			await deletePhase(phase.id);
			addToast('Phase deleted successfully', 'success');
			goto('/projects/phases');
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to delete phase', 'error');
		} finally {
			busy = false;
		}
	}

	onMount(() => {
		load();
	});
</script>

<Toasts />
<ConfirmDialog />

<div class="flex flex-col gap-6">
	<div class="flex items-start gap-3">
		<Button variant="outline" onclick={() => goto('/projects/phases')}>
			<ArrowLeft class="mr-2 h-4 w-4" />
			Back
		</Button>
		<div class="flex-1">
			<h1 class="text-3xl font-bold tracking-tight">{phase?.name ?? 'Phase'}</h1>
			<p class="mt-1 text-muted-foreground">Phase details and metadata.</p>
		</div>
		{#if phase}
			<Button variant="destructive" size="sm" onclick={handleDelete} disabled={busy}>
				<Trash2 class="mr-2 h-4 w-4" />
				Delete
			</Button>
		{/if}
	</div>

	{#if error}
		<div class="rounded-md border bg-muted px-4 py-3 text-muted-foreground">
			<p class="font-medium">Could not load phase</p>
			<p class="text-sm">{error}</p>
		</div>
	{:else if loading}
		<div class="rounded-md border bg-muted px-4 py-3 text-muted-foreground">Loading...</div>
	{:else if phase}
		<div class="rounded-lg border bg-card p-6">
			<dl class="grid gap-4 text-sm">
				<div class="flex justify-between">
					<dt class="text-muted-foreground">Name</dt>
					<dd class="font-medium">{phase.name}</dd>
				</div>
				<div class="flex justify-between">
					<dt class="text-muted-foreground">ID</dt>
					<dd class="font-mono">{phase.id}</dd>
				</div>
				<div class="flex justify-between">
					<dt class="text-muted-foreground">Created</dt>
					<dd>{new Date(phase.created_at).toLocaleString()}</dd>
				</div>
				<div class="flex justify-between">
					<dt class="text-muted-foreground">Updated</dt>
					<dd>{new Date(phase.updated_at).toLocaleString()}</dd>
				</div>
			</dl>
		</div>
	{/if}
</div>
