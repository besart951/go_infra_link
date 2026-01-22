<script lang="ts">
	import { cn } from '$lib/utils.js';

	let { class: className, ...restProps } = $props<{ class?: string; [key: string]: any }>();
</script>

<tbody class={cn('[&_tr:last-child]:border-0', className)} {...restProps}>
	{@render children?.()}
</tbody>
