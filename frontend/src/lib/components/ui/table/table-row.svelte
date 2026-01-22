<script lang="ts">
	import { cn } from '$lib/utils.js';

	let { class: className, ...restProps } = $props<{ class?: string; [key: string]: any }>();
</script>

<tr
	class={cn('border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted', className)}
	{...restProps}
>
	{@render children?.()}
</tr>
