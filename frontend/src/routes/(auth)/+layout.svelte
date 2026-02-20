<script lang="ts">
	import AnimatedAirplane from '$lib/components/ui/animated-airplane/AnimatedAirplane.svelte';
	const { children } = $props();
</script>

<div class="auth-wrapper">
	<!-- Animated airplane background layer -->
	<div class="airplane-bg" aria-hidden="true">
		<AnimatedAirplane />
	</div>

	<!-- Login form layer -->
	<div class="flex min-h-svh items-center justify-center p-6 relative z-10">
		<div class="w-full max-w-sm bg-background/85 backdrop-blur-md rounded-2xl p-8 shadow-xl border border-border/40">
			{@render children?.()}
		</div>
	</div>
</div>

<style>
	.auth-wrapper {
		position: relative;
		min-height: 100svh;
		overflow: hidden;
		background: var(--background);
	}

	.airplane-bg {
		position: absolute;
		inset: 0;
		z-index: 0;
		pointer-events: none;
	}
</style>
