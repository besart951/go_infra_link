<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { ChevronLeft, ChevronRight } from '@lucide/svelte';

	interface Props {
		page: number;
		totalPages: number;
		total: number;
		loading: boolean;
		onPrevious: () => void;
		onNext: () => void;
	}

	let { page, totalPages, total, loading, onPrevious, onNext }: Props = $props();
</script>

{#if totalPages > 1}
	<div class="flex items-center justify-between">
		<div class="text-sm text-muted-foreground">
			Page {page} of {totalPages} &bull; {total}
			{total === 1 ? 'item' : 'items'} total
		</div>
		<div class="flex items-center gap-2">
			<Button
				variant="outline"
				size="sm"
				disabled={page <= 1 || loading}
				onclick={onPrevious}
			>
				<ChevronLeft class="mr-1 h-4 w-4" />
				Previous
			</Button>
			<Button
				variant="outline"
				size="sm"
				disabled={page >= totalPages || loading}
				onclick={onNext}
			>
				Next
				<ChevronRight class="ml-1 h-4 w-4" />
			</Button>
		</div>
	</div>
{/if}
