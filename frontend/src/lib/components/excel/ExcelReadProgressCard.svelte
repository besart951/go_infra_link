<script lang="ts">
	import { X } from '@lucide/svelte';
	import { Button } from '$lib/components/ui/button/index.js';

	interface Props {
		progressPercent: number;
		progressMessage: string;
		isReading: boolean;
		onCancel: () => void;
	}

	let { progressPercent, progressMessage, isReading, onCancel }: Props = $props();

	let progressWidth = $derived(`${Math.min(100, Math.max(0, progressPercent))}%`);
</script>

<div class="rounded-lg border bg-background p-4">
	<div class="mb-2 flex items-center justify-between text-sm">
		<span class="font-medium">Read progress</span>
		<span>{progressPercent}%</span>
	</div>
	<div class="h-2 w-full overflow-hidden rounded-full bg-muted">
		<div class="h-full bg-primary transition-all" style={`width: ${progressWidth};`}></div>
	</div>
	<p class="mt-2 text-xs text-muted-foreground">{progressMessage}</p>

	{#if isReading}
		<div class="mt-3">
			<Button type="button" variant="outline" onclick={onCancel}>
				<X class="mr-2 size-4" />
				Cancel
			</Button>
		</div>
	{/if}
</div>
