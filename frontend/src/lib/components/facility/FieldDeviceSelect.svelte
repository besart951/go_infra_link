<script lang="ts">
	import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
	import { getFieldDevice, listFieldDevices } from '$lib/infrastructure/api/facility.adapter.js';
	import type { FieldDevice } from '$lib/domain/facility/index.js';

	export let value: string = '';
	export let width: string = 'w-[250px]';

	type FieldDeviceItem = FieldDevice & { display_name: string };

	function toItem(device: FieldDevice): FieldDeviceItem {
		return {
			...device,
			display_name: device.bmk || device.apparat_nr?.toString() || device.id
		};
	}

	async function fetcher(search: string): Promise<FieldDeviceItem[]> {
		const res = await listFieldDevices({ search, limit: 20 });
		return (res.items || []).map(toItem);
	}

	async function fetchById(id: string): Promise<FieldDeviceItem> {
		const device = await getFieldDevice(id);
		return toItem(device);
	}
</script>

<AsyncCombobox
	bind:value
	{fetcher}
	{fetchById}
	labelKey="display_name"
	placeholder="Select Field Device..."
	{width}
/>
