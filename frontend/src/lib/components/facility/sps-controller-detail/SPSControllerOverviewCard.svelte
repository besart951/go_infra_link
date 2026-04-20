<script lang="ts">
  import * as Card from '$lib/components/ui/card/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { useSPSControllerDetailState } from './state/context.svelte.js';

  const state = useSPSControllerDetailState();
  const t = createTranslator();
</script>

<Card.Root class="border-border/70 bg-card/80">
  <Card.Header>
    <div class="space-y-1.5">
      <Card.Title class="text-xl">{$t('facility.sps_controller_detail.overview_title')}</Card.Title>
      <Card.Description>{$t('facility.sps_controller_detail.overview_desc')}</Card.Description>
    </div>
  </Card.Header>

  <Card.Content>
    <div class="space-y-4">
      {#each state.overviewItems as item, index (item.label)}
        <div
          class={`grid gap-3 ${index < state.overviewItems.length - 1 ? 'border-b border-border/50 pb-4' : ''} sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]`}
        >
          <div class="text-xs font-medium tracking-[0.18em] text-muted-foreground uppercase">
            {item.label}
          </div>
          <div
            class={`text-sm font-medium text-foreground sm:text-right ${item.monospace ? 'font-mono' : ''}`}
          >
            {item.value}
          </div>
        </div>
      {/each}
    </div>
  </Card.Content>
</Card.Root>
