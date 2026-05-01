<script lang="ts">
  import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
  import * as Popover from '$lib/components/ui/popover/index.js';
  import { Separator } from '$lib/components/ui/separator/index.js';
  import { ToggleButton } from '$lib/components/ui/toggle-button/index.js';
  import SlidersHorizontal from '@lucide/svelte/icons/sliders-horizontal';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { useFieldDeviceState } from './state/context.svelte.js';

  const t = createTranslator();
  const fieldDeviceState = useFieldDeviceState();
  let open = $state(false);
</script>

<Popover.Root bind:open>
  <Popover.Trigger
    class={`${buttonVariants({
      variant: open || fieldDeviceState.view.isCustomized ? 'secondary' : 'outline',
      size: 'icon-sm'
    })} relative`}
    title={$t('field_device.search.view')}
    aria-label={$t('field_device.search.view')}
  >
    <SlidersHorizontal />
    {#if fieldDeviceState.view.isCustomized}
      <span
        class="pointer-events-none absolute -top-0.5 -right-0.5 h-2.5 w-2.5 rounded-full bg-success ring-2 ring-background"
      ></span>
    {/if}
  </Popover.Trigger>

  <Popover.Content align="end" class="w-80 space-y-4 p-3">
    <div class="space-y-1">
      <p class="text-sm font-medium">{$t('field_device.view.title')}</p>
      <p class="text-xs text-muted-foreground">{$t('field_device.view.description')}</p>
    </div>

    <div class="space-y-2">
      <p class="text-xs font-medium text-muted-foreground uppercase">
        {$t('field_device.view.density')}
      </p>
      <div class="grid grid-cols-3 gap-1 rounded-md bg-muted p-1">
        {#each fieldDeviceState.view.density.options as option (option.value)}
          <Button
            type="button"
            variant="ghost"
            class={[
              'h-auto rounded-md px-2 py-1.5 text-sm font-medium transition-colors',
              fieldDeviceState.view.density.value === option.value
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:bg-background/70 hover:text-foreground'
            ]
              .filter(Boolean)
              .join(' ')}
            aria-pressed={fieldDeviceState.view.density.value === option.value}
            onclick={() => fieldDeviceState.view.density.set(option.value)}
          >
            {$t(option.labelKey)}
          </Button>
        {/each}
      </div>
    </div>

    <Separator />

    <div class="space-y-2">
      <div>
        <p class="text-xs font-medium text-muted-foreground uppercase">
          {$t('field_device.view.grouping')}
        </p>
        <p class="text-xs text-muted-foreground">{$t('field_device.view.grouping_hint')}</p>
      </div>

      <div class="grid grid-cols-2 gap-2">
        {#each fieldDeviceState.view.groupOptions as option (option.key)}
          <ToggleButton
            pressed={fieldDeviceState.view.grouping.isActive(option.key)}
            class="min-h-9 w-full"
            onclick={() => void fieldDeviceState.toggleGrouping(option.key)}
          >
            {$t(option.labelKey)}
          </ToggleButton>
        {/each}
      </div>

      {#if fieldDeviceState.loadingGroupingLookups}
        <p class="text-xs text-muted-foreground">{$t('field_device.view.loading_groups')}</p>
      {/if}
    </div>
  </Popover.Content>
</Popover.Root>
