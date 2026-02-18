<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { spsControllerSystemTypeRepository } from '$lib/infrastructure/api/spsControllerSystemTypeRepository.js';
	import type { SPSControllerSystemType } from '$lib/domain/facility/index.js';
	import { createTranslator } from '$lib/i18n/translator.js';

	type Props = {
		value?: string;
		width?: string;
	};

	let { value = $bindable(''), width = 'w-[250px]' }: Props = $props();

	const t = createTranslator();

	async function fetcher(search: string): Promise<SPSControllerSystemType[]> {
		const res = await spsControllerSystemTypeRepository.list({
			pagination: { page: 1, pageSize: 20 },
			search: { text: search }
		});
		return res.items;
	}

	async function fetchById(id: string): Promise<SPSControllerSystemType> {
		return spsControllerSystemTypeRepository.get(id);
	}

	function formatLabel(item: SPSControllerSystemType): string {
		const number = item.number ?? '';
		const documentName = item.document_name ?? '';
		return `${number} - ${documentName}`;
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="document_name"
	labelFormatter={formatLabel}
	placeholder={$t('facility.selects.sps_controller_system_type')}
	{width}
/>
