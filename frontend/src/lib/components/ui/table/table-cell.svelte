<script lang="ts">
	import { cn } from '$lib/utils.js';

	let { class: className, ...restProps } = $props<{ class?: string; [key: string]: any }>();
</script>

<td class={cn('p-4 align-middle [&:has([role=checkbox])]:pr-0', className)} {...restProps}>
	{@render children?.()}
</td>
