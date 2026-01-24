<script lang="ts">
	import { ChevronRight } from '@lucide/svelte';
	import { cn, type WithElementRef } from '$lib/utils.js';
	import type { HTMLLiAttributes } from 'svelte/elements';

	let {
		ref = $bindable(null),
		class: className,
		...rest
	}: WithElementRef<HTMLLiAttributes, HTMLLIElement> = $props();
</script>

<li
	bind:this={ref}
	aria-hidden="true"
	class={cn('text-muted-foreground [&>svg]:size-3.5', className)}
	{...rest}
>
	<ChevronRight />
</li>
