<script lang="ts">
	/**
	 * EditableCell Component
	 * Inline editable table cell with click-to-edit behavior
	 * Supports pending values display and error states
	 */
	import { Input } from '$lib/components/ui/input/index.js';
	import * as Tooltip from '$lib/components/ui/tooltip/index.js';

	interface Props {
		value: string;
		pendingValue?: string; // Value to display when there's a pending edit
		type?: 'text' | 'number';
		placeholder?: string;
		maxlength?: number;
		min?: number;
		max?: number;
		isDirty?: boolean;
		error?: string; // Error message to display
		disabled?: boolean;
		emptyText?: string;
		onSave: (value: string) => void;
	}

	let {
		value,
		pendingValue,
		type = 'text',
		placeholder = '',
		maxlength,
		min,
		max,
		isDirty = false,
		error,
		disabled = false,
		emptyText = '-',
		onSave
	}: Props = $props();

	let isEditing = $state(false);
	let editValue = $state(value);
	let inputElement: HTMLInputElement | null = $state(null);

	// Display value: use pending value if available, otherwise original value
	const displayValue = $derived(pendingValue !== undefined ? pendingValue : value);
	const hasError = $derived(!!error);

	function startEditing() {
		if (disabled) return;
		// Start with the display value (pending or original)
		editValue = displayValue;
		isEditing = true;
		// Focus after DOM update
		setTimeout(() => inputElement?.focus(), 0);
	}

	function handleSave() {
		isEditing = false;
		if (editValue !== displayValue) {
			onSave(editValue);
		}
	}

	function handleCancel() {
		isEditing = false;
		editValue = displayValue;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			e.preventDefault();
			handleSave();
		} else if (e.key === 'Escape') {
			e.preventDefault();
			handleCancel();
		}
	}

	function handleBlur() {
		handleSave();
	}

	// Update editValue when display value changes
	$effect(() => {
		if (!isEditing) {
			editValue = displayValue;
		}
	});
</script>

{#if isEditing}
	<Input
		bind:ref={inputElement}
		{type}
		bind:value={editValue}
		{placeholder}
		{maxlength}
		{min}
		{max}
		onkeydown={handleKeydown}
		onblur={handleBlur}
		class={[
			'h-7 w-full min-w-16 px-2 py-1 text-sm',
			hasError ? 'border-destructive focus-visible:ring-destructive' : ''
		]
			.filter(Boolean)
			.join(' ')}
	/>
{:else}
	{#if hasError}
		<Tooltip.Provider>
			<Tooltip.Root>
				<Tooltip.Trigger asChild>
					<button
						type="button"
						onclick={startEditing}
						{disabled}
						class={[
							'flex h-7 min-h-7 w-full cursor-pointer items-center rounded-sm border px-2 py-1 text-left text-sm transition-colors',
							'border-destructive bg-destructive/10 hover:bg-destructive/20',
							disabled ? 'cursor-not-allowed opacity-50' : ''
						]
							.filter(Boolean)
							.join(' ')}
					>
						{#if displayValue}
							{#if type === 'number'}
								<code class="rounded bg-muted px-1.5 py-0.5 text-sm">{displayValue}</code>
							{:else}
								<span class="truncate">{displayValue}</span>
							{/if}
						{:else}
							<span class="text-muted-foreground">{emptyText}</span>
						{/if}
					</button>
				</Tooltip.Trigger>
				<Tooltip.Content side="top" class="max-w-xs bg-destructive text-destructive-foreground">
					<p>{error}</p>
				</Tooltip.Content>
			</Tooltip.Root>
		</Tooltip.Provider>
	{:else}
		<button
			type="button"
			onclick={startEditing}
			{disabled}
			class={[
				'flex h-7 min-h-7 w-full cursor-pointer items-center rounded-sm px-2 py-1 text-left text-sm transition-colors',
				'hover:bg-muted/50 focus:bg-muted/50 focus:outline-none',
				isDirty ? 'bg-yellow-50 dark:bg-yellow-950/30' : '',
				disabled ? 'cursor-not-allowed opacity-50' : ''
			]
				.filter(Boolean)
				.join(' ')}
		>
			{#if displayValue}
				{#if type === 'number'}
					<code class="rounded bg-muted px-1.5 py-0.5 text-sm">{displayValue}</code>
				{:else}
					<span class="truncate">{displayValue}</span>
				{/if}
			{:else}
				<span class="text-muted-foreground">{emptyText}</span>
			{/if}
		</button>
	{/if}
{/if}
