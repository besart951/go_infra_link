<script lang="ts">
	const { children } = $props();
</script>

<div class="min-h-svh bg-background flex items-center justify-center p-6">
	<div class="w-full max-w-sm">{@render children?.()}</div>
</div>
