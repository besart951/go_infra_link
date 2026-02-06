<script lang="ts">
	import FieldDeviceListView from '$lib/components/facility/FieldDeviceListView.svelte';
</script>

<svelte:head>
	<title>Field Devices | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
	<div>
		<h1 class="text-2xl font-semibold tracking-tight">Field Devices</h1>
		<p class="text-sm text-muted-foreground">
			Manage field devices, BMK identifiers, and specifications.
		</p>
	</div>

	<FieldDeviceListView />
</div>
