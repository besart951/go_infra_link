<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import ProjectMultiSelectFilter from '$lib/components/facility/shared/ProjectMultiSelectFilter.svelte';
  import {
    decodeMultiFilter,
    encodeMultiFilter,
    sortMultiFilterOptions,
    type MultiFilterOption
  } from '$lib/components/facility/shared/projectFacilityListFilters.js';
  import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
  import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
  import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';
  import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
  import { spsControllerSystemTypeRepository } from '$lib/infrastructure/api/spsControllerSystemTypeRepository.js';
  import type {
    Building,
    ControlCabinet,
    SPSController,
    SPSControllerSystemType
  } from '$lib/domain/facility/index.js';
  import ProjectSelect from '$lib/components/project/ProjectSelect.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import type { FieldDeviceFilters } from './state/types.js';
  import { useFieldDeviceState } from './state/context.svelte.js';
  import { fetchAllPages } from '$lib/components/facility/shared/paginatedListFetcher.js';
  import { X } from '@lucide/svelte';

  interface Props {
    showProjectFilter?: boolean;
  }

  let { showProjectFilter = false }: Props = $props();

  const t = createTranslator();
  const fieldDeviceState = useFieldDeviceState();
  const OPTION_PAGE_SIZE = 1000;

  let buildingIds = $state(
    selectedIdsFromFilters(
      fieldDeviceState.filters.buildingIds,
      fieldDeviceState.filters.buildingId
    )
  );
  let controlCabinetIds = $state(
    selectedIdsFromFilters(
      fieldDeviceState.filters.controlCabinetIds,
      fieldDeviceState.filters.controlCabinetId
    )
  );
  let spsControllerIds = $state(
    selectedIdsFromFilters(
      fieldDeviceState.filters.spsControllerIds,
      fieldDeviceState.filters.spsControllerId
    )
  );
  let spsControllerSystemTypeIds = $state(
    selectedIdsFromFilters(
      fieldDeviceState.filters.spsControllerSystemTypeIds,
      fieldDeviceState.filters.spsControllerSystemTypeId
    )
  );
  let projectId = $state(fieldDeviceState.filters.projectId ?? '');

  let buildingOptions = $state<MultiFilterOption[]>([]);
  let controlCabinetOptions = $state<MultiFilterOption[]>([]);
  let spsControllerOptions = $state<MultiFilterOption[]>([]);
  let spsControllerSystemTypeOptions = $state<MultiFilterOption[]>([]);
  let optionsRequestVersion = 0;
  let lastFilterContext = '';
  let inFlightFilterContext = $state<string | null>(null);

  interface ProjectScopeData {
    controlCabinets: ControlCabinet[];
    spsControllers: SPSController[];
  }

  const projectScopeCache = new Map<string, Promise<ProjectScopeData>>();
  const buildingLabelsCache = new Map<string, Promise<Map<string, string>>>();

  const showBuildingFilter = $derived(!fieldDeviceState.isFilterFixed('buildingId'));
  const showControlCabinetFilter = $derived(!fieldDeviceState.isFilterFixed('controlCabinetId'));
  const showSpsControllerFilter = $derived(!fieldDeviceState.isFilterFixed('spsControllerId'));
  const showSpsSystemTypeFilter = $derived(
    !fieldDeviceState.isFilterFixed('spsControllerSystemTypeId')
  );
  const showProjectSelect = $derived(
    showProjectFilter && !fieldDeviceState.isFilterFixed('projectId')
  );

  const fixedBuildingIds = $derived(toFixedIds(fieldDeviceState.fixedFilterValue('buildingId')));
  const fixedControlCabinetIds = $derived(
    toFixedIds(fieldDeviceState.fixedFilterValue('controlCabinetId'))
  );
  const fixedSpsControllerIds = $derived(
    toFixedIds(fieldDeviceState.fixedFilterValue('spsControllerId'))
  );
  const effectiveProjectId = $derived(fieldDeviceState.fixedFilterValue('projectId') ?? projectId);
  const effectiveBuildingIds = $derived(
    fixedBuildingIds.length > 0 ? fixedBuildingIds : buildingIds
  );
  const effectiveControlCabinetIds = $derived(
    fixedControlCabinetIds.length > 0 ? fixedControlCabinetIds : controlCabinetIds
  );
  const effectiveSpsControllerIds = $derived(
    fixedSpsControllerIds.length > 0 ? fixedSpsControllerIds : spsControllerIds
  );
  const scopedProjectId = $derived(effectiveProjectId || undefined);

  const hasActiveFilters = $derived(
    (showBuildingFilter && buildingIds.length > 0) ||
      (showControlCabinetFilter && controlCabinetIds.length > 0) ||
      (showSpsControllerFilter && spsControllerIds.length > 0) ||
      (showSpsSystemTypeFilter && spsControllerSystemTypeIds.length > 0) ||
      (showProjectSelect && projectId)
  );

  $effect(() => {
    if (!fieldDeviceState.showFilterPanel) {
      lastFilterContext = '';
      inFlightFilterContext = null;
      return;
    }

    const context = {
      projectId: scopedProjectId,
      buildingIds: [...effectiveBuildingIds],
      controlCabinetIds: [...effectiveControlCabinetIds],
      spsControllerIds: [...effectiveSpsControllerIds]
    };

    const contextKey = buildFilterContextKey(context);
    if (contextKey === lastFilterContext || contextKey === inFlightFilterContext) return;

    inFlightFilterContext = contextKey;
    void loadFilterOptions(context, contextKey);
  });

  function selectedIdsFromFilters(
    encoded: string | undefined,
    fallback: string | undefined
  ): string[] {
    const ids = decodeMultiFilter(encoded);
    if (ids.length > 0) return ids;
    return fallback ? [fallback] : [];
  }

  function toFixedIds(value: string | undefined): string[] {
    return value ? [value] : [];
  }

  function retainAvailableIds(ids: string[], options: MultiFilterOption[]): string[] {
    const availableIds = new Set(options.map((option) => option.id));
    return ids.filter((id) => availableIds.has(id));
  }

  function sameIds(a: string[], b: string[]): boolean {
    return a.length === b.length && a.every((id, index) => id === b[index]);
  }

  function normalizeIdList(ids: string[]): string {
    return [...new Set(ids)]
      .filter((id) => id.length > 0)
      .sort()
      .join(',');
  }

  function buildFilterContextKey(context: {
    projectId?: string;
    buildingIds: string[];
    controlCabinetIds: string[];
    spsControllerIds: string[];
  }): string {
    return [
      context.projectId ?? '',
      normalizeIdList(context.buildingIds),
      normalizeIdList(context.controlCabinetIds),
      normalizeIdList(context.spsControllerIds)
    ].join('|');
  }

  function formatBuildingLabel(building: Building): string {
    return `${building.iws_code}-${building.building_group}`;
  }

  function formatControlCabinetLabel(
    cabinet: ControlCabinet,
    buildingLabels: Map<string, string>
  ): string {
    const buildingLabel = buildingLabels.get(cabinet.building_id) ?? cabinet.building_id;
    return `${buildingLabel} ${cabinet.control_cabinet_nr}`.trim();
  }

  function formatSpsControllerLabel(controller: SPSController): string {
    return [controller.device_name, controller.ga_device].filter(Boolean).join(' | ');
  }

  function formatSystemTypeLabel(systemType: SPSControllerSystemType): string {
    const systemTypeLabel = [systemType.number, systemType.document_name]
      .filter((value) => value !== undefined && value !== '')
      .join(' - ');
    return [systemType.sps_controller_name, systemTypeLabel || systemType.id]
      .filter(Boolean)
      .join(' | ');
  }

  async function loadFilterOptions(
    context: {
      projectId?: string;
      buildingIds: string[];
      controlCabinetIds: string[];
      spsControllerIds: string[];
    },
    contextKey: string
  ): Promise<void> {
    const requestVersion = ++optionsRequestVersion;

    try {
      const [buildings, cabinets, controllers, systemTypes] = await Promise.all([
        loadBuildingOptions(context.projectId),
        loadControlCabinetOptions(context.projectId, context.buildingIds),
        loadSpsControllerOptions(context.projectId, context.buildingIds, context.controlCabinetIds),
        loadSpsControllerSystemTypeOptions(
          context.projectId,
          context.buildingIds,
          context.controlCabinetIds,
          context.spsControllerIds
        )
      ]);

      if (requestVersion !== optionsRequestVersion) return;

      buildingOptions = buildings;
      controlCabinetOptions = cabinets;
      spsControllerOptions = controllers;
      spsControllerSystemTypeOptions = systemTypes;
      syncSelectionsToOptions();
    } catch (error) {
      console.error('Failed to load field device filter options:', error);
    } finally {
      if (requestVersion === optionsRequestVersion) {
        lastFilterContext = contextKey;
        if (inFlightFilterContext === contextKey) {
          inFlightFilterContext = null;
        }
      }
    }
  }

  async function loadBuildingOptions(projectId: string | undefined): Promise<MultiFilterOption[]> {
    if (projectId) {
      const cabinets = await loadProjectControlCabinets(projectId);
      const buildingIds = [
        ...new Set(cabinets.map((cabinet) => cabinet.building_id).filter(Boolean))
      ];
      if (buildingIds.length === 0) return [];

      const buildings = await buildingRepository.getBulk(buildingIds);
      return sortMultiFilterOptions(
        buildings.map((building) => ({
          id: building.id,
          label: formatBuildingLabel(building)
        }))
      );
    }

    const response = await buildingRepository.list({
      pagination: { page: 1, pageSize: OPTION_PAGE_SIZE },
      search: { text: '' }
    });
    return sortMultiFilterOptions(
      response.items.map((building) => ({
        id: building.id,
        label: formatBuildingLabel(building)
      }))
    );
  }

  async function loadControlCabinetOptions(
    projectId: string | undefined,
    buildingIds: string[]
  ): Promise<MultiFilterOption[]> {
    const cabinets = projectId
      ? await loadProjectControlCabinets(projectId)
      : (
          await controlCabinetRepository.list({
            pagination: { page: 1, pageSize: OPTION_PAGE_SIZE },
            search: { text: '' }
          })
        ).items;

    const scopedCabinets =
      buildingIds.length > 0
        ? cabinets.filter((cabinet) => buildingIds.includes(cabinet.building_id))
        : cabinets;

    const buildingLabels = await loadBuildingLabels(scopedCabinets);
    return sortMultiFilterOptions(
      scopedCabinets.map((cabinet) => ({
        id: cabinet.id,
        label: formatControlCabinetLabel(cabinet, buildingLabels)
      }))
    );
  }

  async function loadSpsControllerOptions(
    projectId: string | undefined,
    buildingIds: string[],
    controlCabinetIds: string[]
  ): Promise<MultiFilterOption[]> {
    const controllers = projectId
      ? await loadProjectSpsControllers(projectId)
      : (
          await spsControllerRepository.list({
            pagination: { page: 1, pageSize: OPTION_PAGE_SIZE },
            search: { text: '' }
          })
        ).items;

    const scopedControllers = await filterControllersByAncestors(
      controllers,
      buildingIds,
      controlCabinetIds
    );

    return sortMultiFilterOptions(
      scopedControllers.map((controller) => ({
        id: controller.id,
        label: formatSpsControllerLabel(controller)
      }))
    );
  }

  async function loadSpsControllerSystemTypeOptions(
    projectId: string | undefined,
    buildingIds: string[],
    controlCabinetIds: string[],
    spsControllerIds: string[]
  ): Promise<MultiFilterOption[]> {
    const fetchedSystemTypes = await fetchAllPages((page, pageSize) =>
      spsControllerSystemTypeRepository.list({
        pagination: { page, pageSize },
        search: { text: '' },
        filters: projectId ? { project_id: projectId } : undefined
      })
    );
    let systemTypes = fetchedSystemTypes;
    if (spsControllerIds.length > 0) {
      systemTypes = systemTypes.filter((item) => spsControllerIds.includes(item.sps_controller_id));
    } else if (buildingIds.length > 0 || controlCabinetIds.length > 0) {
      const controllerIds = [
        ...new Set(systemTypes.map((item) => item.sps_controller_id).filter(Boolean))
      ];
      const controllers =
        controllerIds.length > 0 ? await spsControllerRepository.getBulk(controllerIds) : [];
      const allowedControllers = await filterControllersByAncestors(
        controllers,
        buildingIds,
        controlCabinetIds
      );
      const allowedControllerIds = new Set(allowedControllers.map((controller) => controller.id));
      systemTypes = systemTypes.filter((item) => allowedControllerIds.has(item.sps_controller_id));
    }

    return sortMultiFilterOptions(
      systemTypes.map((systemType) => ({
        id: systemType.id,
        label: formatSystemTypeLabel(systemType)
      }))
    );
  }

  async function loadProjectControlCabinets(projectId: string): Promise<ControlCabinet[]> {
    const scope = await loadProjectScope(projectId);
    return scope.controlCabinets;
  }

  async function loadProjectSpsControllers(projectId: string): Promise<SPSController[]> {
    const scope = await loadProjectScope(projectId);
    return scope.spsControllers;
  }

  async function loadProjectScope(projectId: string): Promise<ProjectScopeData> {
    const cached = projectScopeCache.get(projectId);
    if (cached) return cached;

    const load = (async () => {
      const [cabinetLinks, spsLinks] = await Promise.all([
        fetchAllPages((page, pageSize) =>
          projectRepository.listControlCabinets(projectId, { page, limit: pageSize })
        ),
        fetchAllPages((page, pageSize) =>
          projectRepository.listSPSControllers(projectId, { page, limit: pageSize })
        )
      ]);

      const [controlCabinetIds, spsControllerIds] = [
        [...new Set(cabinetLinks.map((link) => link.control_cabinet_id).filter(Boolean))],
        [...new Set(spsLinks.map((link) => link.sps_controller_id).filter(Boolean))]
      ];

      const [controlCabinets, spsControllers] = await Promise.all([
        controlCabinetIds.length > 0 ? controlCabinetRepository.getBulk(controlCabinetIds) : [],
        spsControllerIds.length > 0 ? spsControllerRepository.getBulk(spsControllerIds) : []
      ]);

      return { controlCabinets, spsControllers };
    })();

    projectScopeCache.set(projectId, load);
    load.catch(() => {
      if (projectScopeCache.get(projectId) === load) {
        projectScopeCache.delete(projectId);
      }
    });

    return load;
  }

  async function loadBuildingLabels(cabinets: ControlCabinet[]): Promise<Map<string, string>> {
    const buildingIds = [
      ...new Set(cabinets.map((cabinet) => cabinet.building_id).filter(Boolean))
    ];
    if (buildingIds.length === 0) return new Map();

    const cacheKey = normalizeIdList(buildingIds);
    const cached = buildingLabelsCache.get(cacheKey);
    if (cached) return cached;

    const load = (async () => {
      const buildings = await buildingRepository.getBulk(buildingIds);
      return new Map(buildings.map((building) => [building.id, formatBuildingLabel(building)]));
    })();

    buildingLabelsCache.set(cacheKey, load);
    load.catch(() => {
      if (buildingLabelsCache.get(cacheKey) === load) {
        buildingLabelsCache.delete(cacheKey);
      }
    });

    return load;
  }

  async function filterControllersByAncestors(
    controllers: SPSController[],
    buildingIds: string[],
    controlCabinetIds: string[]
  ): Promise<SPSController[]> {
    let scopedControllers =
      controlCabinetIds.length > 0
        ? controllers.filter((controller) =>
            controlCabinetIds.includes(controller.control_cabinet_id)
          )
        : controllers;

    if (buildingIds.length === 0) {
      return scopedControllers;
    }

    const cabinetIds = [
      ...new Set(
        scopedControllers.map((controller) => controller.control_cabinet_id).filter(Boolean)
      )
    ];
    const cabinets =
      cabinetIds.length > 0 ? await controlCabinetRepository.getBulk(cabinetIds) : [];
    const allowedCabinetIds = new Set(
      cabinets
        .filter((cabinet) => buildingIds.includes(cabinet.building_id))
        .map((cabinet) => cabinet.id)
    );

    scopedControllers = scopedControllers.filter((controller) =>
      allowedCabinetIds.has(controller.control_cabinet_id)
    );
    return scopedControllers;
  }

  function syncSelectionsToOptions(): void {
    const nextBuildingIds = retainAvailableIds(buildingIds, buildingOptions);
    const nextControlCabinetIds = retainAvailableIds(controlCabinetIds, controlCabinetOptions);
    const nextSpsControllerIds = retainAvailableIds(spsControllerIds, spsControllerOptions);
    const nextSpsControllerSystemTypeIds = retainAvailableIds(
      spsControllerSystemTypeIds,
      spsControllerSystemTypeOptions
    );

    if (!sameIds(buildingIds, nextBuildingIds)) buildingIds = nextBuildingIds;
    if (!sameIds(controlCabinetIds, nextControlCabinetIds)) {
      controlCabinetIds = nextControlCabinetIds;
    }
    if (!sameIds(spsControllerIds, nextSpsControllerIds)) spsControllerIds = nextSpsControllerIds;
    if (!sameIds(spsControllerSystemTypeIds, nextSpsControllerSystemTypeIds)) {
      spsControllerSystemTypeIds = nextSpsControllerSystemTypeIds;
    }
  }

  function handleProjectChange(value: string) {
    projectId = value;
    buildingIds = [];
    controlCabinetIds = [];
    spsControllerIds = [];
    spsControllerSystemTypeIds = [];
  }

  function handleBuildingChange(value: string[]) {
    buildingIds = value;
    controlCabinetIds = [];
    spsControllerIds = [];
    spsControllerSystemTypeIds = [];
  }

  function handleControlCabinetChange(value: string[]) {
    controlCabinetIds = value;
    spsControllerIds = [];
    spsControllerSystemTypeIds = [];
  }

  function handleSpsControllerChange(value: string[]) {
    spsControllerIds = value;
    spsControllerSystemTypeIds = [];
  }

  function applyFilters() {
    const filters: FieldDeviceFilters = {
      buildingIds: showBuildingFilter ? encodeMultiFilter(buildingIds) : undefined,
      controlCabinetIds: showControlCabinetFilter
        ? encodeMultiFilter(controlCabinetIds)
        : undefined,
      spsControllerIds: showSpsControllerFilter ? encodeMultiFilter(spsControllerIds) : undefined,
      spsControllerSystemTypeIds: showSpsSystemTypeFilter
        ? encodeMultiFilter(spsControllerSystemTypeIds)
        : undefined,
      projectId: showProjectSelect && projectId ? projectId : undefined
    };

    void fieldDeviceState.applyFilters(filters);
  }

  function clearFilters() {
    buildingIds = [];
    controlCabinetIds = [];
    spsControllerIds = [];
    spsControllerSystemTypeIds = [];
    projectId = '';
    void fieldDeviceState.clearFilters();
  }
</script>

<Card.Root>
  <Card.Content>
    <div class="grid grid-cols-1 gap-x-6 gap-y-4 lg:grid-cols-2">
      {#if showProjectSelect}
        <div class="flex flex-col gap-2">
          <label for="project-filter" class="text-sm font-medium">
            {$t('field_device.filters.project')}
          </label>
          <ProjectSelect
            bind:value={projectId}
            width="w-full"
            onValueChange={handleProjectChange}
          />
        </div>
      {/if}
      {#if showBuildingFilter}
        <ProjectMultiSelectFilter
          items={buildingOptions}
          value={buildingIds}
          label={$t('field_device.filters.building')}
          placeholder={$t('field_device.filters.all_buildings')}
          searchPlaceholder={$t('field_device.filters.search_buildings')}
          emptyText={$t('field_device.filters.no_buildings')}
          selectedText={$t('field_device.filters.buildings_selected', {
            count: buildingIds.length
          })}
          clearText={$t('field_device.filters.clear_selection')}
          width="w-full"
          popupWidth="w-(--bits-popover-anchor-width)"
          onValueChange={handleBuildingChange}
        />
      {/if}
      {#if showControlCabinetFilter}
        <ProjectMultiSelectFilter
          items={controlCabinetOptions}
          value={controlCabinetIds}
          label={$t('field_device.filters.control_cabinet')}
          placeholder={$t('field_device.filters.all_control_cabinets')}
          searchPlaceholder={$t('field_device.filters.search_control_cabinets')}
          emptyText={$t('field_device.filters.no_control_cabinets')}
          selectedText={$t('field_device.filters.control_cabinets_selected', {
            count: controlCabinetIds.length
          })}
          clearText={$t('field_device.filters.clear_selection')}
          width="w-full"
          popupWidth="w-(--bits-popover-anchor-width)"
          onValueChange={handleControlCabinetChange}
        />
      {/if}
      {#if showSpsControllerFilter}
        <ProjectMultiSelectFilter
          items={spsControllerOptions}
          value={spsControllerIds}
          label={$t('field_device.filters.sps_controller')}
          placeholder={$t('field_device.filters.all_sps_controllers')}
          searchPlaceholder={$t('field_device.filters.search_sps_controllers')}
          emptyText={$t('field_device.filters.no_sps_controllers')}
          selectedText={$t('field_device.filters.sps_controllers_selected', {
            count: spsControllerIds.length
          })}
          clearText={$t('field_device.filters.clear_selection')}
          width="w-full"
          popupWidth="w-(--bits-popover-anchor-width)"
          onValueChange={handleSpsControllerChange}
        />
      {/if}
      {#if showSpsSystemTypeFilter}
        <ProjectMultiSelectFilter
          items={spsControllerSystemTypeOptions}
          value={spsControllerSystemTypeIds}
          label={$t('field_device.filters.sps_system_type')}
          placeholder={$t('field_device.filters.all_sps_system_types')}
          searchPlaceholder={$t('field_device.filters.search_sps_system_types')}
          emptyText={$t('field_device.filters.no_sps_system_types')}
          selectedText={$t('field_device.filters.sps_system_types_selected', {
            count: spsControllerSystemTypeIds.length
          })}
          clearText={$t('field_device.filters.clear_selection')}
          width="w-full"
          popupWidth="w-(--bits-popover-anchor-width)"
          onValueChange={(ids) => (spsControllerSystemTypeIds = ids)}
        />
      {/if}
    </div>
    <div class="mt-4 flex justify-end gap-2">
      <Button onclick={applyFilters}>{$t('field_device.filters.apply')}</Button>
      {#if hasActiveFilters}
        <Button variant="outline" onclick={clearFilters}>
          <X class="mr-2 size-4" />
          {$t('field_device.filters.clear')}
        </Button>
      {/if}
    </div>
  </Card.Content>
</Card.Root>
