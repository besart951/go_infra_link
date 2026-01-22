<script lang="ts">
	import { cn, type WithElementRef } from "$lib/utils.js";
	import type { HTMLAnchorAttributes } from "svelte/elements";

	let {
		ref = $bindable(null),
		class: className,
		href,
		children,
		...rest
	}: WithElementRef<HTMLAnchorAttributes, HTMLAnchorElement> = $props();
</script>

<a
	bind:this={ref}
	href={href}
	class={cn("hover:text-foreground transition-colors", className)}
	{...rest}
>
	{@render children?.()}
</a>
