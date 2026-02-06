<script lang="ts">
	/**
	 * EditableSelectCell Component
	 * Inline editable table cell with click-to-edit select dropdown
	 */
	import * as Tooltip from '$lib/components/ui/tooltip/index.js';

	interface SelectOption {
		value: string;
		label: string;
	}

	interface Props {
		value: string;
		options: SelectOption[];
		pendingValue?: string;
		isDirty?: boolean;
		error?: string;
		disabled?: boolean;
		emptyText?: string;
		onSave: (value: string) => void;
	}

	let {
		value,
		options,
		pendingValue,
		isDirty = false,
		error,
		disabled = false,
		emptyText = '-',
		onSave
	}: Props = $props();

	let isEditing = $state(false);
	let selectElement: HTMLSelectElement | null = $state(null);

	const displayValue = $derived(pendingValue !== undefined ? pendingValue : value);
	const displayLabel = $derived(
		options.find((o) => o.value === displayValue)?.label || displayValue || emptyText
	);
	const hasError = $derived(!!error);

	function startEditing() {
		if (disabled) return;
		isEditing = true;
		setTimeout(() => selectElement?.focus(), 0);
	}

	function handleChange(e: Event) {
		const newValue = (e.target as HTMLSelectElement).value;
		isEditing = false;
		if (newValue !== displayValue) {
			onSave(newValue);
		}
	}

	function handleBlur() {
		isEditing = false;
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			e.preventDefault();
			isEditing = false;
		}
	}
</script>

{#if isEditing}
	<select
		bind:this={selectElement}
		value={displayValue}
		onchange={handleChange}
		onblur={handleBlur}
		onkeydown={handleKeydown}
		class={[
			'h-7 w-full min-w-20 rounded-sm border bg-background px-1.5 py-0.5 text-sm focus:outline-none focus:ring-1 focus:ring-ring',
			hasError ? 'border-destructive focus:ring-destructive' : 'border-input'
		]
			.filter(Boolean)
			.join(' ')}
	>
		{#each options as opt (opt.value)}
			<option value={opt.value}>{opt.label}</option>
		{/each}
	</select>
{:else if hasError}
	<Tooltip.Provider>
		<Tooltip.Root>
			<Tooltip.Trigger>
				{#snippet child({ props })}
					<button
						{...props}
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
						<span class="truncate">{displayLabel}</span>
					</button>
				{/snippet}
			</Tooltip.Trigger>
			<Tooltip.Content side="top" class="text-destructive-foreground max-w-xs bg-destructive">
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
		<span class="truncate">{displayLabel}</span>
	</button>
{/if}
