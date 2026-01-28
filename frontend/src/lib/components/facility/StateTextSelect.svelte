<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { getStateText, listStateTexts } from '$lib/infrastructure/api/facility.adapter.js';
	import type { StateText } from '$lib/domain/facility/index.js';

	export let value: string = '';
	export let width: string = 'w-[250px]';

	async function fetcher(search: string): Promise<StateText[]> {
		const res = await listStateTexts({ search, limit: 20 });
		return res.items || [];
	}

	async function fetchById(id: string): Promise<StateText> {
		return getStateText(id);
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="state_text1"
	placeholder="Select State Text..."
	{width}
/>
