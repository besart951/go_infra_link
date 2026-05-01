<script lang="ts">
  import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
  import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
  import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
  import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';
  import type { Building, ControlCabinet } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';

  type BuildingOption = Building & { display_name: string };

  type Props = {
    value?: string;
    width?: string;
    projectId?: string;
    disabled?: boolean;
    refreshKey?: string | number;
    onValueChange?: (value: string) => void;
  };

  let {
    value = $bindable(''),
    width = 'w-[250px]',
    projectId,
    disabled = false,
    refreshKey,
    onValueChange
  }: Props = $props();

  const t = createTranslator();
  const effectiveRefreshKey = $derived(
    projectId !== undefined || refreshKey !== undefined
      ? `${projectId ?? ''}|${refreshKey ?? ''}`
      : undefined
  );

  function toOption(item: Building): BuildingOption {
    return {
      ...item,
      display_name: `${item.iws_code}-${item.building_group}`
    };
  }

  function matchesSearch(item: BuildingOption, search: string): boolean {
    const query = search.trim().toLowerCase();
    if (!query) return true;
    return item.display_name.toLowerCase().includes(query);
  }

  async function fetchProjectBuildings(search: string): Promise<BuildingOption[]> {
    if (!projectId) return [];

    const links = await projectRepository.listControlCabinets(projectId, {
      page: 1,
      limit: 1000
    });
    const cabinetIds = Array.from(
      new Set(links.items.map((link) => link.control_cabinet_id).filter(Boolean))
    );
    if (cabinetIds.length === 0) return [];

    const cabinets = await controlCabinetRepository.getBulk(cabinetIds);
    const buildingIds = Array.from(
      new Set(cabinets.map((cabinet: ControlCabinet) => cabinet.building_id).filter(Boolean))
    );
    if (buildingIds.length === 0) return [];

    const buildings = await buildingRepository.getBulk(buildingIds);
    return buildings.map(toOption).filter((item) => matchesSearch(item, search));
  }

  async function fetcher(search: string): Promise<BuildingOption[]> {
    if (projectId) {
      return fetchProjectBuildings(search);
    }

    const res = await buildingRepository.list({
      pagination: { page: 1, pageSize: 20 },
      search: { text: search }
    });
    return res.items.map(toOption);
  }

  async function fetchById(id: string): Promise<BuildingOption> {
    const building = await buildingRepository.get(id);
    return toOption(building);
  }
</script>

<AsyncCombobox
  bind:value
  {fetcher}
  {fetchById}
  refreshKey={effectiveRefreshKey}
  labelKey="display_name"
  placeholder={$t('facility.selects.building')}
  {disabled}
  {width}
  {onValueChange}
/>
