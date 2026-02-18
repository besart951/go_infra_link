<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
	import type { ControlCabinet } from '$lib/domain/facility/index.js';
	import { createTranslator } from '$lib/i18n/translator.js';

	type Props = {
		value?: string;
		width?: string;
	};

	let { value = $bindable(''), width = 'w-[250px]' }: Props = $props();

	const t = createTranslator();

	async function fetcher(search: string): Promise<ControlCabinet[]> {
		const res = await controlCabinetRepository.list({
			pagination: { page: 1, pageSize: 20 },
			search: { text: search }
		});
		return res.items;
	}

	async function fetchById(id: string): Promise<ControlCabinet> {
		return controlCabinetRepository.get(id);
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="control_cabinet_nr"
	placeholder={$t('facility.selects.control_cabinet')}
	{width}
/>
