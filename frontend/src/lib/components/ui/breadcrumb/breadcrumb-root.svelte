<script lang="ts">
	import { cn, type WithElementRef } from '$lib/utils.js';
	import type { HTMLAttributes } from 'svelte/elements';

	let {
		ref = $bindable(null),
		class: className,
		children,
		...rest
	}: WithElementRef<HTMLAttributes<HTMLElement>, HTMLElement> = $props();
</script>

<nav bind:this={ref} aria-label="breadcrumb" class={cn('', className)} {...rest}>
	{@render children?.()}
</nav>
