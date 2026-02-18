<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { addToast } from '$lib/components/toast.svelte';
	import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
	import { confirm } from '$lib/stores/confirm-dialog.js';
	import { createTranslator } from '$lib/i18n/translator.js';
	import { t as translate } from '$lib/i18n/index.js';
	import { ArrowLeft, Trash2 } from '@lucide/svelte';
	import type { Phase } from '$lib/domain/phase/index.js';
	import { deletePhase, getPhase } from '$lib/infrastructure/api/phase.adapter.js';

	const t = createTranslator();

	const phaseId = $derived($page.params.id ?? '');

	let phase = $state<Phase | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let busy = $state(false);

	async function load() {
		if (!phaseId) {
			error = translate('phases.errors.missing_id');
			loading = false;
			return;
		}

		loading = true;
		error = null;
		try {
			phase = await getPhase(phaseId);
		} catch (err) {
			error = err instanceof Error ? err.message : translate('phases.errors.load_failed');
		} finally {
			loading = false;
		}
	}

	async function handleDelete() {
		if (!phase) return;
		const ok = await confirm({
			title: translate('phases.confirm.delete_title'),
			message: translate('phases.confirm.delete_message', { name: phase.name }),
			confirmText: translate('common.delete'),
			cancelText: translate('common.cancel'),
			variant: 'destructive'
		});

		if (!ok) return;
		busy = true;
		try {
			await deletePhase(phase.id);
			addToast(translate('phases.toasts.deleted'), 'success');
			goto('/projects/phases');
		} catch (err) {
			addToast(
				err instanceof Error ? err.message : translate('phases.toasts.delete_failed'),
				'error'
			);
		} finally {
			busy = false;
		}
	}

	onMount(() => {
		load();
	});
</script>

<ConfirmDialog />

<div class="flex flex-col gap-6">
	<div class="flex items-start gap-3">
		<Button variant="outline" onclick={() => goto('/projects/phases')}>
			<ArrowLeft class="mr-2 h-4 w-4" />
			{$t('common.back')}
		</Button>
		<div class="flex-1">
			<h1 class="text-3xl font-bold tracking-tight">
				{phase?.name ?? $t('phases.detail.fallback')}
			</h1>
			<p class="mt-1 text-muted-foreground">{$t('phases.detail.description')}</p>
		</div>
		{#if phase}
			<Button variant="destructive" size="sm" onclick={handleDelete} disabled={busy}>
				<Trash2 class="mr-2 h-4 w-4" />
				{$t('common.delete')}
			</Button>
		{/if}
	</div>

	{#if error}
		<div class="rounded-md border bg-muted px-4 py-3 text-muted-foreground">
			<p class="font-medium">{$t('phases.errors.load_title')}</p>
			<p class="text-sm">{error}</p>
		</div>
	{:else if loading}
		<div class="rounded-md border bg-muted px-4 py-3 text-muted-foreground">
			{$t('common.loading')}
		</div>
	{:else if phase}
		<div class="rounded-lg border bg-card p-6">
			<dl class="grid gap-4 text-sm">
				<div class="flex justify-between">
					<dt class="text-muted-foreground">Name</dt>
					<dd class="font-medium">{phase.name}</dd>
				</div>
				<div class="flex justify-between">
					<dt class="text-muted-foreground">{$t('common.id')}</dt>
					<dd class="font-mono">{phase.id}</dd>
				</div>
				<div class="flex justify-between">
					<dt class="text-muted-foreground">{$t('common.created')}</dt>
					<dd>{new Date(phase.created_at).toLocaleString()}</dd>
				</div>
				<div class="flex justify-between">
					<dt class="text-muted-foreground">{$t('common.modified')}</dt>
					<dd>{new Date(phase.updated_at).toLocaleString()}</dd>
				</div>
			</dl>
		</div>
	{/if}
</div>
