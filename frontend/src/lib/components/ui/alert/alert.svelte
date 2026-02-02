<script lang="ts">
	import { cn } from '$lib/utils.js';
	import type { HTMLAttributes } from 'svelte/elements';

	interface Props extends HTMLAttributes<HTMLDivElement> {
		variant?: 'default' | 'destructive';
	}

	let { variant = 'default', class: className, children, ...restProps }: Props = $props();

	const variantClasses = {
		default: 'bg-background text-foreground border-border',
		destructive: 'border-destructive/50 text-destructive dark:border-destructive [&>svg]:text-destructive'
	};
</script>

<div
	class={cn(
		'relative w-full rounded-lg border px-4 py-3 text-sm [&>svg+div]:translate-y-[-3px] [&>svg]:absolute [&>svg]:left-4 [&>svg]:top-4 [&>svg]:text-foreground [&>svg~*]:pl-7',
		variantClasses[variant],
		className
	)}
	{...restProps}
	role="alert"
>
	{@render children?.()}
</div>
