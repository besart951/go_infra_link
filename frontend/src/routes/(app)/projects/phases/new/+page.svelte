<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { ArrowLeft } from '@lucide/svelte';
	import PhaseForm from '$lib/components/project/PhaseForm.svelte';
	import { createTranslator } from '$lib/i18n/translator.js';

	const t = createTranslator();

	function handleSuccess() {
		goto('/projects/phases');
	}

	function handleCancel() {
		goto('/projects/phases');
	}
</script>

<div class="flex flex-col gap-6">
	<div class="flex items-start gap-3">
		<Button variant="outline" onclick={() => goto('/projects/phases')}>
			<ArrowLeft class="mr-2 h-4 w-4" />
			{$t('common.back')}
		</Button>
		<div>
			<h1 class="text-3xl font-bold tracking-tight">{$t('phases.new.title')}</h1>
			<p class="mt-1 text-muted-foreground">{$t('phases.new.description')}</p>
		</div>
	</div>

	<PhaseForm onSuccess={handleSuccess} onCancel={handleCancel} />
</div>
