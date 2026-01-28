<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { listProjects } from '$lib/infrastructure/api/project.adapter.js';

	type PhaseOption = {
		id: string;
		name: string;
	};

	export let value: string = '';
	export let width: string = 'w-[260px]';
	export let id: string | undefined = undefined;

	const MAX_PHASE_SAMPLES = 200;

	async function fetcher(search: string): Promise<PhaseOption[]> {
		const res = await listProjects({ page: 1, limit: MAX_PHASE_SAMPLES });
		const unique = new Map<string, PhaseOption>();
		for (const project of res.items ?? []) {
			const phaseId = project.phase_id?.trim();
			if (!phaseId) continue;
			if (!unique.has(phaseId)) {
				unique.set(phaseId, { id: phaseId, name: phaseId });
			}
		}
		const items = Array.from(unique.values());
		const normalized = search.trim().toLowerCase();
		if (!normalized) return items;
		return items.filter((item) => item.name.toLowerCase().includes(normalized));
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
