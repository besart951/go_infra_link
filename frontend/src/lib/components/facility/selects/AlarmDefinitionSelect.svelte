<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { alarmDefinitionRepository } from '$lib/infrastructure/api/alarmDefinitionRepository.js';
	import type { AlarmDefinition } from '$lib/domain/facility/index.js';
	import { createTranslator } from '$lib/i18n/translator.js';

	type Props = {
		value?: string;
		width?: string;
	};

	let { value = $bindable(), width = 'w-[300px]' }: Props = $props();

	const t = createTranslator();

	async function fetcher(search: string): Promise<AlarmDefinition[]> {
		const res = await alarmDefinitionRepository.list({
			pagination: { page: 1, pageSize: 20 },
			search: { text: search }
		});
		return res.items;
	}

	async function fetchById(id: string): Promise<AlarmDefinition> {
		return alarmDefinitionRepository.get(id);
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="name"
	placeholder={$t('facility.selects.alarm_definition')}
	{width}
/>
