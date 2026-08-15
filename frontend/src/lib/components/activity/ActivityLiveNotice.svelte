<script lang="ts">
  import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
  import { Button } from '$lib/components/ui/button/index.js';
  import { t as translate } from '$lib/i18n/index.js';

  interface Props {
    count: number;
    loading?: boolean;
    onShow: () => void;
  }

  let { count, loading = false, onShow }: Props = $props();
</script>

{#if count > 0}
  <div class="sticky top-3 z-10 flex justify-center">
    <Button size="sm" variant="secondary" disabled={loading} onclick={onShow} class="shadow-sm">
      <RefreshCwIcon class={['size-4', loading ? 'animate-spin' : ''].join(' ')} />
      {translate('history.activity.live_changes', {
        count,
        label: translate(
          count === 1 ? 'history.activity.activity_singular' : 'history.activity.activity_plural'
        )
      })}
    </Button>
  </div>
{/if}
