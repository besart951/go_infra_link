<script lang="ts">
  import * as Card from '@ui-svelte/components/ui/card/index.js';
  import { Badge } from '@ui-svelte/components/ui/badge/index.js';
  import Building2Icon from '@lucide/svelte/icons/building-2';
  import CalendarIcon from '@lucide/svelte/icons/calendar';
  import PanelTopIcon from '@lucide/svelte/icons/panel-top';
  import type { Building, ControlCabinet } from '$lib/domain/facility/index.js';
  import { createTranslator } from '@i18n/translator.js';

  type Props = {
    cabinet: ControlCabinet;
    building: Building | null;
  };

  let { cabinet, building }: Props = $props();
  const t = createTranslator();

  function buildingLabel(): string {
    if (!building) return cabinet.building_id;
    return `${building.iws_code}-${building.building_group}`;
  }

  function fmtDate(date: string): string {
    return new Date(date).toLocaleString();
  }
</script>

<Card.Root class="border-primary/20 bg-card">
  <Card.Header>
    <Card.Title class="flex items-center gap-2">
      <PanelTopIcon class="size-5 text-primary" />
      {$t('facility.control_cabinet_detail.overview_title')}
    </Card.Title>
    <Card.Description>{$t('facility.control_cabinet_detail.overview_desc')}</Card.Description>
  </Card.Header>
  <Card.Content>
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <div class="rounded-md border border-border/70 bg-muted/30 p-3">
        <p class="text-xs text-muted-foreground">
          {$t('facility.control_cabinet_detail.cabinet_number')}
        </p>
        <div class="mt-2 flex items-center gap-2">
          <Badge variant="default">#{cabinet.control_cabinet_nr}</Badge>
        </div>
      </div>

      <div class="rounded-md border border-border/70 bg-muted/30 p-3">
        <p class="text-xs text-muted-foreground">{$t('facility.building')}</p>
        <div class="mt-2 flex items-center gap-2 text-sm font-medium">
          <Building2Icon class="size-4 text-primary" />
          <span>{buildingLabel()}</span>
        </div>
      </div>

      <div class="rounded-md border border-border/70 bg-muted/30 p-3">
        <p class="text-xs text-muted-foreground">
          {$t('facility.control_cabinet_detail.created_at')}
        </p>
        <div class="mt-2 flex items-center gap-2 text-sm font-medium">
          <CalendarIcon class="size-4 text-primary" />
          <span>{fmtDate(cabinet.created_at)}</span>
        </div>
      </div>

      <div class="rounded-md border border-border/70 bg-muted/30 p-3">
        <p class="text-xs text-muted-foreground">
          {$t('facility.control_cabinet_detail.updated_at')}
        </p>
        <div class="mt-2 flex items-center gap-2 text-sm font-medium">
          <CalendarIcon class="size-4 text-primary" />
          <span>{fmtDate(cabinet.updated_at)}</span>
        </div>
      </div>
    </div>
  </Card.Content>
</Card.Root>
