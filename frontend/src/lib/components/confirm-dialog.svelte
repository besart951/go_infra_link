<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { AlertTriangle } from '@lucide/svelte';
	import { confirmDialogState } from '$lib/stores/confirm-dialog.js';
</script>

{#if $confirmDialogState.open}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
		role="button"
		tabindex="0"
		aria-label="Close dialog"
		onclick={(e) => {
			if (e.target === e.currentTarget) {
				$confirmDialogState.onCancel?.();
			}
		}}
		onkeydown={(e) => {
			if (e.key === 'Enter' || e.key === ' ') {
				$confirmDialogState.onCancel?.();
			}
		}}
	>
		<div
			class="bg-background w-full max-w-md rounded-lg border p-6 shadow-lg"
			role="dialog"
			aria-modal="true"
		>
			<div class="flex items-start gap-4">
				{#if $confirmDialogState.variant === 'destructive'}
					<div class="bg-destructive/10 text-destructive rounded-full p-2">
						<AlertTriangle class="h-6 w-6" />
					</div>
				{/if}
				<div class="flex-1">
					<h2 class="text-lg font-semibold">{$confirmDialogState.title}</h2>
					<p class="text-muted-foreground mt-2 text-sm">{$confirmDialogState.message}</p>
				</div>
			</div>

			<div class="mt-6 flex justify-end gap-3">
				<Button variant="outline" onclick={() => $confirmDialogState.onCancel?.()}>
					{$confirmDialogState.cancelText}
				</Button>
				<Button
					variant={$confirmDialogState.variant === 'destructive' ? 'destructive' : 'default'}
					onclick={() => $confirmDialogState.onConfirm?.()}
				>
					{$confirmDialogState.confirmText}
				</Button>
			</div>
		</div>
	</div>
{/if}

<svelte:window
	onkeydown={(e) => {
		if ($confirmDialogState.open && e.key === 'Escape') {
			$confirmDialogState.onCancel?.();
		}
	}}
/>
