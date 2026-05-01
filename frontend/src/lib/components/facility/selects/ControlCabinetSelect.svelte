<script lang="ts">
  import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
  import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
  import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
  import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';
  import type { Building, ControlCabinet } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';

  type Props = {
    value?: string;
    width?: string;
    refreshKey?: string | number;
    buildingId?: string;
    projectId?: string;
    disabled?: boolean;
    onValueChange?: (value: string) => void;
  };

  let {
    value = $bindable(''),
    width = 'w-[250px]',
    refreshKey,
    buildingId,
    projectId,
    disabled = false,
    onValueChange
  }: Props = $props();

  const t = createTranslator();
  let buildingLabels = $state(new Map<string, string>());
  const effectiveRefreshKey = $derived(
    projectId !== undefined || buildingId !== undefined || refreshKey !== undefined
      ? `${projectId ?? ''}|${buildingId ?? ''}|${refreshKey ?? ''}`
      : undefined
  );

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

  function matchesSearch(cabinet: ControlCabinet, search: string): boolean {
    const query = search.trim().toLowerCase();
    if (!query) return true;

    return [cabinet.control_cabinet_nr, formatControlCabinetLabel(cabinet)]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(query));
  }

  async function fetchProjectControlCabinets(search: string): Promise<ControlCabinet[]> {
    if (!projectId) return [];

    const links = await projectRepository.listControlCabinets(projectId, {
      page: 1,
      limit: 1000
    });
    const cabinetIds = Array.from(
      new Set(links.items.map((link) => link.control_cabinet_id).filter(Boolean))
    );
    if (cabinetIds.length === 0) return [];

    let cabinets = await controlCabinetRepository.getBulk(cabinetIds);
    if (buildingId) {
      cabinets = cabinets.filter((cabinet) => cabinet.building_id === buildingId);
    }

    await ensureBuildingLabels(cabinets);
    return cabinets.filter((cabinet) => matchesSearch(cabinet, search));
  }

  async function fetcher(search: string): Promise<ControlCabinet[]> {
    if (projectId) {
      return fetchProjectControlCabinets(search);
    }

    const res = await controlCabinetRepository.list({
      pagination: { page: 1, pageSize: 20 },
      search: { text: search },
      filters: buildingId ? { building_id: buildingId } : undefined
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
  refreshKey={effectiveRefreshKey}
  labelKey="control_cabinet_nr"
  labelFormatter={formatControlCabinetLabel}
  placeholder={$t('facility.selects.control_cabinet')}
  {disabled}
  {width}
  {onValueChange}
/>
