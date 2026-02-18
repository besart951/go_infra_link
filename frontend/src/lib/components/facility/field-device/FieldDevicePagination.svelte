<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { ChevronLeft, ChevronRight } from '@lucide/svelte';
	import { createTranslator } from '$lib/i18n/translator.js';

	interface Props {
		page: number;
		totalPages: number;
		total: number;
		loading: boolean;
		onPrevious: () => void;
		onNext: () => void;
	}

	let { page, totalPages, total, loading, onPrevious, onNext }: Props = $props();

	const t = createTranslator();
</script>

{#if totalPages > 1}
	<div class="flex items-center justify-between">
		<div class="text-sm text-muted-foreground">
			{$t('messages.page_of', { page, total: totalPages })}
			&bull; {$t('messages.total_items', { count: total })}
		</div>
		<div class="flex items-center gap-2">
			<Button variant="outline" size="sm" disabled={page <= 1 || loading} onclick={onPrevious}>
				<ChevronLeft class="mr-1 h-4 w-4" />
				{$t('common.previous')}
			</Button>
			<Button variant="outline" size="sm" disabled={page >= totalPages || loading} onclick={onNext}>
				{$t('common.next')}
				<ChevronRight class="ml-1 h-4 w-4" />
			</Button>
		</div>
	</div>
{/if}
