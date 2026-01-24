<script lang="ts">
	import { cn, type WithElementRef } from '$lib/utils.js';
	import type { HTMLOlAttributes } from 'svelte/elements';

	let {
		ref = $bindable(null),
		class: className,
		children,
		...rest
	}: WithElementRef<HTMLOlAttributes, HTMLOListElement> = $props();
</script>

<ol
	bind:this={ref}
	class={cn('flex flex-wrap items-center gap-1.5 text-sm text-muted-foreground', className)}
	{...rest}
>
	{@render children?.()}
</ol>
