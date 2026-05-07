<script lang="ts">
  import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
  import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';
  import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
  import type { SPSController } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { fetchAllPages } from '$lib/components/facility/shared/paginatedListFetcher.js';

  interface Props {
    value?: string;
    width?: string;
    projectId?: string;
    controlCabinetId?: string;
    disabled?: boolean;
    refreshKey?: string | number;
    onValueChange?: (value: string) => void;
  }

  let {
    value = $bindable(''),
    width = 'w-[250px]',
    projectId,
    controlCabinetId,
    disabled = false,
    refreshKey,
    onValueChange
  }: Props = $props();

  const t = createTranslator();
  const effectiveRefreshKey = $derived(
    projectId !== undefined || controlCabinetId !== undefined || refreshKey !== undefined
      ? `${projectId ?? ''}|${controlCabinetId ?? ''}|${refreshKey ?? ''}`
      : undefined
  );
  const projectSpsControllersCache = new Map<string, Promise<SPSController[]>>();

  function matchesSearch(controller: SPSController, search: string): boolean {
    const query = search.trim().toLowerCase();
    if (!query) return true;

    return [controller.device_name, controller.ga_device, controller.ip_address]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(query));
  }

  async function loadProjectSPSControllers(projectId: string): Promise<SPSController[]> {
    const cached = projectSpsControllersCache.get(projectId);
    if (cached) return cached;

    const load = (async () => {
      const links = await fetchAllPages((page, pageSize) =>
        projectRepository.listSPSControllers(projectId, { page, limit: pageSize })
      );
      const controllerIds = [
        ...new Set(links.map((link) => link.sps_controller_id).filter(Boolean))
      ];
      if (controllerIds.length === 0) return [];
      return spsControllerRepository.getBulk(controllerIds);
    })();

    projectSpsControllersCache.set(projectId, load);
    load.catch(() => {
      if (projectSpsControllersCache.get(projectId) === load) {
        projectSpsControllersCache.delete(projectId);
      }
    });

    return load;
  }

  async function fetchProjectSpsControllers(search: string): Promise<SPSController[]> {
    if (!projectId) return [];

    const controllers = await loadProjectSPSControllers(projectId);
    const scopedControllers = controlCabinetId
      ? controllers.filter((controller) => controller.control_cabinet_id === controlCabinetId)
      : controllers;

    return scopedControllers.filter((controller) => matchesSearch(controller, search));
  }

  async function fetcher(search: string): Promise<SPSController[]> {
    if (projectId) {
      return fetchProjectSpsControllers(search);
    }

    const res = await spsControllerRepository.list({
      pagination: { page: 1, pageSize: 20 },
      search: { text: search },
      filters: controlCabinetId ? { control_cabinet_id: controlCabinetId } : undefined
    });
    return res.items.filter((controller) => matchesSearch(controller, search));
  }

  async function fetchById(id: string): Promise<SPSController> {
    return spsControllerRepository.get(id);
  }
</script>

<AsyncCombobox
  bind:value
  {fetcher}
  {fetchById}
  refreshKey={effectiveRefreshKey}
  labelKey="device_name"
  placeholder={$t('facility.selects.sps_controller')}
  {disabled}
  {width}
  {onValueChange}
/>
