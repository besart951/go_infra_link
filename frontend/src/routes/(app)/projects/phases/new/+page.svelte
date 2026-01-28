<script lang="ts">
	import { createPhase } from '$lib/infrastructure/api/phase.adapter.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import Toasts, { addToast } from '$lib/components/toast.svelte';
	import { goto } from '$app/navigation';

	let id = '';
	let name = '';
	let busy = false;

	async function submit() {
		if (!id.trim()) return addToast('Phase id is required', 'error');
		busy = true;
		try {
			// Create a minimal project to register the phase id in backend
			await createPhase(id, name);
			addToast('Phase created', 'success');
			goto('/projects/phases');
		} catch (e) {
			addToast(e instanceof Error ? e.message : 'Failed to create phase', 'error');
		} finally {
			busy = false;
		}
	}
</script>

<Toasts />

<div class="max-w-md space-y-4">
	<h1 class="text-2xl font-semibold">Create Phase</h1>

	<div class="space-y-2">
		<label class="text-sm font-medium" for="phase_id">Phase ID</label>
		<Input id="phase_id" bind:value={id} placeholder="e.g. ALPHA-1" />
	</div>

	<div class="space-y-2">
		<label class="text-sm font-medium" for="phase_name">Display Name (optional)</label>
		<Input id="phase_name" bind:value={name} placeholder="Optional friendly name" />
	</div>

	<div class="flex justify-end gap-2">
		<Button variant="outline" onclick={() => goto('/projects/phases')}>Cancel</Button>
		<Button onclick={submit} disabled={busy}>Create Phase</Button>
	</div>
</div>
