<script lang="ts">
	import { cn, type WithElementRef } from "$lib/utils.js";
	import type { HTMLOlAttributes } from "svelte/elements";

	let {
		ref = $bindable(null),
		class: className,
		children,
		...rest
	}: WithElementRef<HTMLOlAttributes, HTMLOListElement> = $props();
</script>

<ol
	bind:this={ref}
	class={cn("text-muted-foreground flex flex-wrap items-center gap-1.5 text-sm", className)}
	{...rest}
>
	{@render children?.()}
</ol>
