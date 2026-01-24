<script lang="ts">
	const { children } = $props();
</script>

<div class="flex min-h-svh items-center justify-center bg-background p-6">
	<div class="w-full max-w-sm">{@render children?.()}</div>
</div>
