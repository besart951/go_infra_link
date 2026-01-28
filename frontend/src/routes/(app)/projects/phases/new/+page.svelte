<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { ArrowLeft } from '@lucide/svelte';
	import PhaseForm from '$lib/components/project/PhaseForm.svelte';

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
			Back
		</Button>
		<div>
			<h1 class="text-3xl font-bold tracking-tight">Create Phase</h1>
			<p class="mt-1 text-muted-foreground">Add a new phase to a project.</p>
		</div>
	</div>

	<PhaseForm on:success={handleSuccess} on:cancel={handleCancel} />
</div>
