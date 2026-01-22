<script lang="ts">
	import { cn } from '$lib/utils.js';

	let { class: className, ...restProps } = $props<{ class?: string; [key: string]: any }>();
</script>

<thead class={cn('[&_tr]:border-b', className)} {...restProps}>
	{@render children?.()}
</thead>
