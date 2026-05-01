<script lang="ts" module>
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
  import { Button } from '$lib/components/ui/button/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';

  const t = createTranslator();

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
        return 'bg-success-muted border-success-border text-success-muted-foreground dark:bg-success-muted dark:border-success-border dark:text-success-muted-foreground';
      case 'error':
        return 'bg-destructive/10 border-destructive/30 text-destructive dark:bg-destructive/20 dark:border-destructive/40 dark:text-destructive';
      case 'warning':
        return 'bg-warning-muted border-warning-border text-warning-muted-foreground dark:bg-warning-muted dark:border-warning-border dark:text-warning-muted-foreground';
      default:
        return 'bg-info-muted border-info-border text-info-muted-foreground dark:bg-info-muted dark:border-info-border dark:text-info-muted-foreground';
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
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          onclick={() => removeToast(toast.id)}
          class="size-6 shrink-0 opacity-70 hover:opacity-100"
          aria-label={$t('common.close')}
        >
          <X class="h-4 w-4" />
        </Button>
      </div>
    {/each}
  </div>
{/if}
