<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import {
		getNotificationClass,
		listNotificationClasses
	} from '$lib/infrastructure/api/facility.adapter.js';
	import type { NotificationClass } from '$lib/domain/facility/index.js';

	export let value: string = '';
	export let width: string = 'w-[250px]';

	async function fetcher(search: string): Promise<NotificationClass[]> {
		const res = await listNotificationClasses({ search, limit: 20 });
		return res.items || [];
	}

	async function fetchById(id: string): Promise<NotificationClass> {
		return getNotificationClass(id);
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
