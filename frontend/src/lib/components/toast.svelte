<script lang="ts" context="module">
	import { writable } from 'svelte/store';

	export type ToastType = 'success' | 'error' | 'warning' | 'info';

	export interface Toast {
		id: string;
		message: string;
		type: ToastType;
	}

	const toasts = writable<Toast[]>([]);

	export function addToast(message: string, type: ToastType = 'info', duration = 5000) {
		const id = Math.random().toString(36).substring(2, 9);
		toasts.update((all) => [...all, { id, message, type }]);

		if (duration > 0) {
			setTimeout(() => {
				removeToast(id);
			}, duration);
		}

		return id;
	}

	export function removeToast(id: string) {
		toasts.update((all) => all.filter((t) => t.id !== id));
	}

	export { toasts };
</script>

<script lang="ts">
	import { fly } from 'svelte/transition';
	import { X, CheckCircle, AlertCircle, Info, AlertTriangle } from '@lucide/svelte';

	function getIcon(type: ToastType) {
		switch (type) {
			case 'success':
				return CheckCircle;
			case 'error':
				return AlertCircle;
			case 'warning':
				return AlertTriangle;
			default:
				return Info;
		}
	}

	function getColorClasses(type: ToastType): string {
		switch (type) {
			case 'success':
				return 'bg-green-50 border-green-200 text-green-900 dark:bg-green-950 dark:border-green-800 dark:text-green-100';
			case 'error':
				return 'bg-red-50 border-red-200 text-red-900 dark:bg-red-950 dark:border-red-800 dark:text-red-100';
			case 'warning':
				return 'bg-yellow-50 border-yellow-200 text-yellow-900 dark:bg-yellow-950 dark:border-yellow-800 dark:text-yellow-100';
			default:
				return 'bg-blue-50 border-blue-200 text-blue-900 dark:bg-blue-950 dark:border-blue-800 dark:text-blue-100';
		}
	}
</script>

{#if $toasts.length > 0}
	<div class="fixed right-4 bottom-4 z-50 flex max-w-md flex-col gap-2">
		{#each $toasts as toast (toast.id)}
			{@const Icon = getIcon(toast.type)}
			<div
				transition:fly={{ y: 50, duration: 200 }}
				class="flex items-start gap-3 rounded-lg border p-4 shadow-lg {getColorClasses(toast.type)}"
			>
				<Icon class="mt-0.5 h-5 w-5 shrink-0" />
				<p class="flex-1 text-sm">{toast.message}</p>
				<button
					onclick={() => removeToast(toast.id)}
					class="shrink-0 rounded-sm opacity-70 transition-opacity hover:opacity-100"
					aria-label="Close notification"
				>
					<X class="h-4 w-4" />
				</button>
			</div>
		{/each}
	</div>
{/if}
