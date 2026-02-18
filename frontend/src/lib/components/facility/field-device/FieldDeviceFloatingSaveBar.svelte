<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Save, Undo } from '@lucide/svelte';
	import type { useFieldDeviceEditing } from '$lib/hooks/useFieldDeviceEditing.svelte.js';
	import { createTranslator } from '$lib/i18n/translator.js';

	interface Props {
		editing: ReturnType<typeof useFieldDeviceEditing>;
		onSave: () => void;
		onDiscard: () => void;
	}

	let { editing, onSave, onDiscard }: Props = $props();

	const t = createTranslator();
</script>

{#if editing.hasUnsavedChanges}
	<div
		class="fixed bottom-4 left-1/2 z-50 flex -translate-x-1/2 flex-col gap-2 rounded-lg border bg-card px-4 py-3 shadow-lg"
	>
		<!-- Action bar -->
		<div class="flex items-center gap-3">
			<span class="text-sm font-medium">
				{$t('field_device.save_bar.unsaved', { count: editing.pendingCount })}
			</span>
			<Button size="sm" onclick={onSave}>
				<Save class="mr-1 h-4 w-4" />
				{$t('field_device.save_bar.save_all')}
			</Button>
			<Button variant="ghost" size="sm" onclick={onDiscard}>
				<Undo class="mr-1 h-4 w-4" />
				{$t('field_device.save_bar.discard')}
			</Button>
		</div>

		<!-- Warning message -->
		<p class="text-xs text-muted-foreground">
			{$t('field_device.save_bar.notice')}
		</p>
	</div>
{/if}
