<script lang="ts">
	import { cn, type WithElementRef } from "$lib/utils.js";
	import type { HTMLLiAttributes } from "svelte/elements";

	let {
		ref = $bindable(null),
		class: className,
		children,
		...rest
	}: WithElementRef<HTMLLiAttributes, HTMLLIElement> = $props();
</script>

<li bind:this={ref} class={cn("inline-flex items-center gap-1.5", className)} {...rest}>
	{@render children?.()}
</li>
