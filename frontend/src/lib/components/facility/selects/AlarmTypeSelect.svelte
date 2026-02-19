<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { alarmTypeRepository } from '$lib/infrastructure/api/alarmTypeRepository.js';
	import type { AlarmType } from '$lib/domain/facility/index.js';

	type Props = {
		value?: string;
		width?: string;
	};

	let { value = $bindable(), width = 'w-[300px]' }: Props = $props();

	async function fetcher(search: string): Promise<AlarmType[]> {
		const res = await alarmTypeRepository.list({
			page: 1,
			pageSize: 20,
			search
		});
		return res.items;
	}

	async function fetchById(id: string): Promise<AlarmType> {
		return alarmTypeRepository.get(id);
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="name"
	placeholder="Alarmtyp auswählen"
	{width}
/>
