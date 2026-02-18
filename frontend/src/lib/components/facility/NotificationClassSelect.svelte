<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { notificationClassRepository } from '$lib/infrastructure/api/notificationClassRepository.js';
	import type { NotificationClass } from '$lib/domain/facility/index.js';

	type Props = {
		value?: string;
		width?: string;
	};

	let { value = $bindable(''), width = 'w-[250px]' }: Props = $props();

	async function fetcher(search: string): Promise<NotificationClass[]> {
		const res = await notificationClassRepository.list({
			pagination: { page: 1, pageSize: 20 },
			search: { text: search }
		});
		return res.items;
	}

	async function fetchById(id: string): Promise<NotificationClass> {
		return notificationClassRepository.get(id);
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="meaning"
	placeholder="Select Notification Class..."
	{width}
/>
