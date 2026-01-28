<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { listPhases } from '$lib/infrastructure/api/phase.adapter.js';
	import type { Phase } from '$lib/domain/phase/index.js';

	export let value: string = '';
	export let width: string = 'w-[260px]';
	export let id: string | undefined = undefined;

	const MAX_PHASE_SAMPLES = 100;

	async function fetcher(search: string): Promise<Phase[]> {
		const res = await listPhases({ page: 1, limit: MAX_PHASE_SAMPLES, search });
		return (res.items ?? []).map((phase) => ({
			...phase,
			name: phase.name || phase.id
		}));
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	labelKey="name"
	placeholder="Select phase..."
	searchPlaceholder="Search phases..."
	emptyText="No phases found."
	{width}
	{id}
/>
