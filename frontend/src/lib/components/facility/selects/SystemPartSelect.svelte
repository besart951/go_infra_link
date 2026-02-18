<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { systemPartRepository } from '$lib/infrastructure/api/systemPartRepository.js';
	import type { SystemPart } from '$lib/domain/facility/index.js';
	import { createTranslator } from '$lib/i18n/translator.js';

	type Props = {
		value?: string;
		width?: string;
		onValueChange?: (value: string) => void;
	};

	let { value = $bindable(''), width = 'w-[250px]', onValueChange }: Props = $props();

	const t = createTranslator();

	async function fetcher(search: string): Promise<SystemPart[]> {
		const res = await systemPartRepository.list({
			pagination: { page: 1, pageSize: 20 },
			search: { text: search }
		});
		return res.items;
	}

	async function fetchById(id: string): Promise<SystemPart> {
		return systemPartRepository.get(id);
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="name"
	placeholder={$t('facility.selects.system_part')}
	{width}
	{onValueChange}
/>
