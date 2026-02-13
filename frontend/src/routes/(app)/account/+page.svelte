<script lang="ts">
	import { createTranslator } from '$lib/i18n/translator';

	const t = createTranslator();
</script>

<svelte:head>
	<title>{$t('navigation.account')} | Infra Link</title>
</svelte:head>

<h1 class="text-2xl font-bold">{$t('messages.account_page')}</h1>
