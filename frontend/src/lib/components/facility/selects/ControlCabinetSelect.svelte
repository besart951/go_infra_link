<script lang="ts">
  import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
  import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
  import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
  import type { Building, ControlCabinet } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';

  type Props = {
    value?: string;
    width?: string;
  };

  let { value = $bindable(''), width = 'w-[250px]' }: Props = $props();

  const t = createTranslator();
  let buildingLabels = $state(new Map<string, string>());

  function formatBuildingLabel(building: Building): string {
    return `${building.iws_code}-${building.building_group}`;
  }

  async function ensureBuildingLabels(cabinets: ControlCabinet[]) {
    const missingIds = Array.from(
      new Set(cabinets.map((cabinet) => cabinet.building_id).filter(Boolean))
    ).filter((id) => !buildingLabels.has(id));

    if (missingIds.length === 0) return;

    try {
      const buildings = await buildingRepository.getBulk(missingIds);
      const next = new Map(buildingLabels);
      for (const building of buildings) {
        next.set(building.id, formatBuildingLabel(building));
      }
      buildingLabels = next;
    } catch (error) {
      console.error('Failed to load building labels for control cabinets:', error);
    }
  }

  function formatControlCabinetLabel(cabinet: ControlCabinet): string {
    const buildingLabel = buildingLabels.get(cabinet.building_id) ?? cabinet.building_id;
    return `${buildingLabel} ${cabinet.control_cabinet_nr}`.trim();
  }

  async function fetcher(search: string): Promise<ControlCabinet[]> {
    const res = await controlCabinetRepository.list({
      pagination: { page: 1, pageSize: 20 },
      search: { text: search }
    });

    await ensureBuildingLabels(res.items);
    return res.items;
  }

  async function fetchById(id: string): Promise<ControlCabinet> {
    const cabinet = await controlCabinetRepository.get(id);
    await ensureBuildingLabels([cabinet]);
    return cabinet;
  }
</script>

<AsyncCombobox
  bind:value
  {fetcher}
  {fetchById}
  labelKey="control_cabinet_nr"
  labelFormatter={formatControlCabinetLabel}
  placeholder={$t('facility.selects.control_cabinet')}
  {width}
/>
