<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils.js';

	let { class: className, children, ...restProps } = $props<{
		class?: string;
		children?: Snippet;
		[key: string]: any;
	}>();
</script>

<tr
	class={cn('border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted', className)}
	{...restProps}
>
	{@render children?.()}
</tr>
