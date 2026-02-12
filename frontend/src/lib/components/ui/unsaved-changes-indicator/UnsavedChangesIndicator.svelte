<!--
	UnsavedChangesIndicator - Reusable component for displaying unsaved changes status
	
	A clean, minimal indicator that can be placed anywhere in the UI.
	Follows shadcn principles: unstyled by default, easily customizable.
	
	@component
	@example
	```svelte
	<UnsavedChangesIndicator 
		count={3} 
		variant="badge"
		message="Changes saved locally" 
	/>
	```
-->
<script lang="ts">
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { AlertCircle, Save } from '@lucide/svelte';

	interface Props {
		/**
		 * Number of unsaved changes
		 */
		count: number;

		/**
		 * Display variant
		 * - 'badge': Compact badge with count
		 * - 'inline': Inline text with icon
		 * - 'card': Card-style with message
		 */
		variant?: 'badge' | 'inline' | 'card';

		/**
		 * Optional custom message to display
		 */
		message?: string;

		/**
		 * Additional CSS classes
		 */
		class?: string;
	}

	let { count, variant = 'badge', message, class: className = '' }: Props = $props();

	const defaultMessage = $derived(count === 1 ? '1 unsaved change' : `${count} unsaved changes`);
	const displayMessage = $derived(message || defaultMessage);
</script>

{#if count > 0}
	{#if variant === 'badge'}
		<Badge variant="secondary" class={`flex items-center gap-1 ${className}`}>
			<AlertCircle class="h-3 w-3" />
			<span class="text-xs">{count}</span>
		</Badge>
	{:else if variant === 'inline'}
		<div class={`inline-flex items-center gap-2 text-sm text-muted-foreground ${className}`}>
			<AlertCircle class="h-4 w-4 text-amber-500" />
			<span>{displayMessage}</span>
		</div>
	{:else if variant === 'card'}
		<div
			class={`flex items-start gap-3 rounded-lg border border-amber-200 bg-amber-50 p-3 dark:border-amber-900 dark:bg-amber-950 ${className}`}
		>
			<AlertCircle class="mt-0.5 h-4 w-4 flex-shrink-0 text-amber-600 dark:text-amber-400" />
			<div class="flex-1 space-y-1">
				<p class="text-sm font-medium text-amber-900 dark:text-amber-100">{displayMessage}</p>
				{#if message}
					<p class="text-xs text-amber-700 dark:text-amber-300">
						Changes are saved locally and persist across page navigation
					</p>
				{/if}
			</div>
		</div>
	{/if}
{/if}
