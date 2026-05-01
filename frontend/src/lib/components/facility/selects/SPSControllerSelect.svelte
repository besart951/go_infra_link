<script lang="ts">
  import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
  import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';
  import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
  import type { SPSController } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';

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

  function matchesSearch(controller: SPSController, search: string): boolean {
    const query = search.trim().toLowerCase();
    if (!query) return true;

    return [controller.device_name, controller.ga_device, controller.ip_address]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(query));
  }

  async function fetchProjectSpsControllers(search: string): Promise<SPSController[]> {
    if (!projectId) return [];

    const links = await projectRepository.listSPSControllers(projectId, {
      page: 1,
      limit: 1000
    });
    const controllerIds = Array.from(
      new Set(links.items.map((link) => link.sps_controller_id).filter(Boolean))
    );
    if (controllerIds.length === 0) return [];

    let controllers = await spsControllerRepository.getBulk(controllerIds);
    if (controlCabinetId) {
      controllers = controllers.filter(
        (controller) => controller.control_cabinet_id === controlCabinetId
      );
    }

    return controllers.filter((controller) => matchesSearch(controller, search));
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
    return res.items;
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
