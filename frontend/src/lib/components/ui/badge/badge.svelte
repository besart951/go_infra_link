<script lang="ts">
	import { cn } from '$lib/utils.js';
	import type { Snippet } from 'svelte';

	type BadgeVariant = 'default' | 'secondary' | 'destructive' | 'outline' | 'success' | 'warning';

	let {
		class: className,
		variant = 'default',
		children,
		...restProps
	}: {
		class?: string;
		variant?: BadgeVariant;
		children?: Snippet;
		[key: string]: any;
	} = $props();

	const variants = {
		default: 'border-transparent bg-primary text-primary-foreground hover:bg-primary/80',
		secondary: 'border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80',
		destructive:
			'border-transparent bg-destructive text-destructive-foreground hover:bg-destructive/80',
		outline: 'text-foreground',
		success: 'border-transparent bg-green-500 text-white hover:bg-green-600',
		warning: 'border-transparent bg-yellow-500 text-white hover:bg-yellow-600'
	};
</script>

<div
	class={cn(
		'inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:ring-2 focus:ring-ring focus:ring-offset-2 focus:outline-none',
		variants[variant],
		className
	)}
	{...restProps}
>
	{@render children?.()}
</div>
