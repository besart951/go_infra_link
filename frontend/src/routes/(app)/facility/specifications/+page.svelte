<script lang="ts">
	import { createTranslator } from '$lib/i18n/translator';

	const t = createTranslator();
</script>

<svelte:head>
	<title>{$t('facility.specifications')} | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-4">
	<div>
		<h1 class="text-2xl font-semibold tracking-tight">{$t('facility.specifications')}</h1>
		<p class="text-sm text-muted-foreground">
			{$t('facility.specifications_desc')}
		</p>
	</div>
	<a href="/facility/field-devices" class="text-sm font-medium underline"> {$t('facility.go_to_field_devices')} </a>
</div>
