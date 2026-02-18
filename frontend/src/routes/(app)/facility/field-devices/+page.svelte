<script lang="ts">
	import FieldDeviceListView from '$lib/components/facility/field-device/FieldDeviceListView.svelte';
	import { createTranslator } from '$lib/i18n/translator';

	const t = createTranslator();
</script>

<svelte:head>
	<title>{$t('facility.field_devices_title')} | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div>
		<h1 class="text-2xl font-semibold tracking-tight">{$t('facility.field_devices_title')}</h1>
		<p class="text-sm text-muted-foreground">
			{$t('facility.field_devices_desc')}
		</p>
	</div>

	<FieldDeviceListView />
</div>
