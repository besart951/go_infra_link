<script lang="ts">
	import { cn } from '$lib/utils.js';

	let { class: className, ...restProps } = $props<{ class?: string; [key: string]: any }>();
</script>

<th
	class={cn(
		'h-12 px-4 text-left align-middle font-medium text-muted-foreground [&:has([role=checkbox])]:pr-0',
		className
	)}
	{...restProps}
>
	{@render children?.()}
</th>
