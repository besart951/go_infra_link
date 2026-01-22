<script lang="ts" context="module">
	import { writable } from 'svelte/store';

	export interface ConfirmDialogOptions {
		title: string;
		message: string;
		confirmText?: string;
		cancelText?: string;
		variant?: 'default' | 'destructive';
	}

	interface ConfirmDialogState extends ConfirmDialogOptions {
		open: boolean;
		onConfirm?: () => void;
		onCancel?: () => void;
	}

	const dialogState = writable<ConfirmDialogState>({
		open: false,
		title: '',
		message: '',
		confirmText: 'Confirm',
		cancelText: 'Cancel',
		variant: 'default'
	});

	export function confirm(options: ConfirmDialogOptions): Promise<boolean> {
		return new Promise((resolve) => {
			dialogState.set({
				...options,
				open: true,
				confirmText: options.confirmText || 'Confirm',
				cancelText: options.cancelText || 'Cancel',
				variant: options.variant || 'default',
				onConfirm: () => {
					closeDialog();
					resolve(true);
				},
				onCancel: () => {
					closeDialog();
					resolve(false);
				}
			});
		});
	}

	function closeDialog() {
		dialogState.update((state) => ({ ...state, open: false }));
	}
</script>

<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { AlertTriangle } from 'lucide-svelte';

	let { $dialogState: state } = $props();
</script>

{#if state.open}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
		onclick={() => state.onCancel?.()}
	>
		<div
			class="bg-background w-full max-w-md rounded-lg border p-6 shadow-lg"
			onclick={(e) => e.stopPropagation()}
		>
			<div class="flex items-start gap-4">
				{#if state.variant === 'destructive'}
					<div class="bg-destructive/10 text-destructive rounded-full p-2">
						<AlertTriangle class="h-6 w-6" />
					</div>
				{/if}
				<div class="flex-1">
					<h2 class="text-lg font-semibold">{state.title}</h2>
					<p class="text-muted-foreground mt-2 text-sm">{state.message}</p>
				</div>
			</div>
			<div class="mt-6 flex justify-end gap-3">
				<Button variant="outline" onclick={() => state.onCancel?.()}>
					{state.cancelText}
				</Button>
				<Button variant={state.variant === 'destructive' ? 'destructive' : 'default'} onclick={() => state.onConfirm?.()}>
					{state.confirmText}
				</Button>
			</div>
		</div>
	</div>
{/if}

<svelte:window
	onkeydown={(e) => {
		if (state.open && e.key === 'Escape') {
			state.onCancel?.();
		}
	}}
/>
