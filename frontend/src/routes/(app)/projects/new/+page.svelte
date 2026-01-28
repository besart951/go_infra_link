<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	onMount(() => {
		goto('/projects');
	});
</script>

<div class="p-6 text-sm text-muted-foreground">Redirecting to Projects…</div>
