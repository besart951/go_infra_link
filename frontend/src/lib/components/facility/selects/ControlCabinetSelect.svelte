<script lang="ts">
  import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
  import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
  import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
  import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';
  import type { Building, ControlCabinet } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { fetchAllPages } from '$lib/components/facility/shared/paginatedListFetcher.js';

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
  const effectiveRefreshKey = $derived(
    projectId !== undefined || buildingId !== undefined || refreshKey !== undefined
      ? `${projectId ?? ''}|${buildingId ?? ''}|${refreshKey ?? ''}`
      : undefined
  );
  const projectControlCabinetsCache = new Map<string, Promise<ControlCabinet[]>>();

  function formatBuildingLabel(building: Building): string {
    return `${building.iws_code}-${building.building_group}`;
  }

  function matchesSearch(
    cabinet: ControlCabinet,
    search: string,
    buildingLabels: Map<string, string>
  ): boolean {
    const query = search.trim().toLowerCase();
    if (!query) return true;

    return [cabinet.control_cabinet_nr, formatControlCabinetLabel(cabinet, buildingLabels)]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(query));
  }

  function formatControlCabinetLabel(cabinet: ControlCabinet, labels: Map<string, string>): string {
    const buildingLabel = labels.get(cabinet.building_id) ?? cabinet.building_id;
    return `${buildingLabel} ${cabinet.control_cabinet_nr}`.trim();
  }

  const buildingLabelCache = new Map<string, Promise<Map<string, string>>>();

  async function ensureBuildingLabels(cabinets: ControlCabinet[]): Promise<Map<string, string>> {
    const buildingIds = [
      ...new Set(cabinets.map((cabinet) => cabinet.building_id).filter(Boolean))
    ];
    const cacheKey = buildingIds.sort().join(',');
    const cached = buildingLabelCache.get(cacheKey);
    if (cached) return cached;

    const load = (async () => {
      if (buildingIds.length === 0) return new Map<string, string>();
      const buildings = await buildingRepository.getBulk(buildingIds);
      const next = new Map<string, string>();
      for (const building of buildings) {
        next.set(building.id, formatBuildingLabel(building));
      }
      return next;
    })();

    buildingLabelCache.set(cacheKey, load);
    load.catch(() => {
      if (buildingLabelCache.get(cacheKey) === load) {
        buildingLabelCache.delete(cacheKey);
      }
    });

    return load;
  }

  async function fetchProjectControlCabinets(search: string): Promise<ControlCabinet[]> {
    if (!projectId) return [];

    const cabinets = await loadProjectControlCabinets(projectId);
    const buildingLabels = await ensureBuildingLabels(cabinets);
    const scoped = buildingId
      ? cabinets.filter((cabinet) => cabinet.building_id === buildingId)
      : cabinets;
    return scoped.filter((cabinet) => {
      const label = formatControlCabinetLabel(cabinet, buildingLabels);
      return [cabinet.control_cabinet_nr, label]
        .filter(Boolean)
        .some((value) => value!.toLowerCase().includes(search.trim().toLowerCase()));
    });
  }

  async function loadProjectControlCabinets(projectId: string): Promise<ControlCabinet[]> {
    const cached = projectControlCabinetsCache.get(projectId);
    if (cached) return cached;

    const load = (async () => {
      const links = await fetchAllPages((page, pageSize) =>
        projectRepository.listControlCabinets(projectId, { page, limit: pageSize })
      );
      const cabinetIds = [...new Set(links.map((link) => link.control_cabinet_id).filter(Boolean))];
      if (cabinetIds.length === 0) return [];
      return controlCabinetRepository.getBulk(cabinetIds);
    })();

    projectControlCabinetsCache.set(projectId, load);
    load.catch(() => {
      if (projectControlCabinetsCache.get(projectId) === load) {
        projectControlCabinetsCache.delete(projectId);
      }
    });

    return load;
  }

  async function fetcher(search: string): Promise<ControlCabinet[]> {
    if (projectId) {
      return fetchProjectControlCabinets(search);
    }

    const res = await controlCabinetRepository.list({
      pagination: { page: 1, pageSize: 20 },
      search: { text: search }
    });
    const labels = await ensureBuildingLabels(res.items);
    return res.items.filter((cabinet) => {
      const label = formatControlCabinetLabel(cabinet, labels);
      return [cabinet.control_cabinet_nr, label]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(search.trim().toLowerCase()));
    });
  }

  async function fetchById(id: string): Promise<ControlCabinet> {
    const cabinet = await controlCabinetRepository.get(id);
    return cabinet;
  }
</script>

<AsyncCombobox
  bind:value
  {fetcher}
  {fetchById}
  refreshKey={effectiveRefreshKey}
  labelKey="control_cabinet_nr"
  labelFormatter={(cabinet) => cabinet.control_cabinet_nr}
  placeholder={$t('facility.selects.control_cabinet')}
  {disabled}
  {width}
  {onValueChange}
/>
