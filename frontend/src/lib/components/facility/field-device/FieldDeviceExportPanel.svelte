<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import AsyncMultiSelect from '$lib/components/ui/combobox/AsyncMultiSelect.svelte';
  import { addToast } from '$lib/components/toast.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { t as translate } from '$lib/i18n/index.js';

  // Use Cases
  import { ExportFieldDevicesUseCase } from '$lib/application/useCases/facility/exportFieldDevicesUseCase.js';
  import { ListEntityUseCase } from '$lib/application/useCases/listEntityUseCase.js';
  import { fieldDeviceRepository } from '$lib/infrastructure/api/fieldDeviceRepository.js';
  import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
  import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
  import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
  import { facilityJobState } from '$lib/state/copyOperation.svelte.js';
  import type { FieldDeviceFilters } from './state/types.js';
  import { buildFieldDeviceExportRequest, hasExportScope } from './exportRequest.js';

  import { projectRepository } from '$lib/infrastructure/api/projectRepository.js';

  import type { Building, ControlCabinet, SPSController } from '$lib/domain/facility/index.js';
  import type { Project } from '$lib/domain/project/index.js';
  import { Download, FileSpreadsheet, LoaderCircle, Play } from '@lucide/svelte';

  interface Props {
    projectId?: string;
    filters?: FieldDeviceFilters;
    searchText?: string;
  }

  type OptionItem = { id: string; label: string };

  let { projectId, filters, searchText }: Props = $props();

  const t = createTranslator();

  // Instantiate Use Cases
  const exportUseCase = new ExportFieldDevicesUseCase(fieldDeviceRepository);
  const listBuildingsUseCase = new ListEntityUseCase(buildingRepository);
  const listControlCabinetsUseCase = new ListEntityUseCase(controlCabinetRepository);
  const listSPSControllersUseCase = new ListEntityUseCase(spsControllerRepository);

  let selectedProjectIds = $state<string[]>([]);
  let selectedBuildingIds = $state<string[]>([]);
  let selectedControlCabinetIds = $state<string[]>([]);
  let selectedSPSControllerIds = $state<string[]>([]);
  let forceAsync = $state(false);

  let submitting = $state(false);
  let activeJobId = $state<string | null>(null);

  $effect(() => {
    if (projectId && selectedProjectIds.length === 0) {
      selectedProjectIds = [projectId];
    }
  });

  const exportRequest = $derived(
    buildFieldDeviceExportRequest({
      projectId,
      filters,
      search: searchText,
      selectedProjectIds,
      selectedBuildingIds,
      selectedControlCabinetIds,
      selectedSPSControllerIds
    })
  );
  const canStartExport = $derived(hasExportScope(exportRequest));

  const activeJob = $derived(
    facilityJobState.jobs.find((job) => job.jobId === activeJobId) ??
      facilityJobState.exportJobs[0] ??
      null
  );
  const progressWidth = $derived(`${Math.min(100, Math.max(0, activeJob?.progress ?? 0))}%`);
  const isRunning = $derived(activeJob?.status === 'queued' || activeJob?.status === 'running');
  const isCompleted = $derived(activeJob?.status === 'completed');
  const isFailed = $derived(activeJob?.status === 'failed');

  function toProjectOption(project: Project): OptionItem {
    return { id: project.id, label: project.name || project.id };
  }

  function toBuildingOption(building: Building): OptionItem {
    return { id: building.id, label: `${building.iws_code}-${building.building_group}` };
  }

  function toCabinetOption(cabinet: ControlCabinet): OptionItem {
    return { id: cabinet.id, label: cabinet.control_cabinet_nr || cabinet.id };
  }

  function toControllerOption(controller: SPSController): OptionItem {
    const ga = controller.ga_device || '-';
    return { id: controller.id, label: `${ga} - ${controller.device_name || controller.id}` };
  }

  async function fetchProjects(search: string): Promise<OptionItem[]> {
    const res = await projectRepository.list({
      pagination: { page: 1, pageSize: 100 },
      search: { text: search }
    });
    return res.items.map(toProjectOption);
  }

  async function fetchProjectsByIds(ids: string[]): Promise<OptionItem[]> {
    const items = await Promise.all(ids.map((id) => projectRepository.get(id)));
    return items.map(toProjectOption);
  }

  async function fetchBuildings(search: string): Promise<OptionItem[]> {
    const res = await listBuildingsUseCase.execute({
      pagination: { page: 1, pageSize: 100 },
      search: { text: search }
    });
    return res.items.map(toBuildingOption);
  }

  async function fetchBuildingsByIds(ids: string[]): Promise<OptionItem[]> {
    const items = await listBuildingsUseCase.getBulk(ids);
    return items.map(toBuildingOption);
  }

  async function fetchControlCabinets(search: string): Promise<OptionItem[]> {
    const res = await listControlCabinetsUseCase.execute({
      pagination: { page: 1, pageSize: 100 },
      search: { text: search }
    });
    return res.items.map(toCabinetOption);
  }

  async function fetchControlCabinetsByIds(ids: string[]): Promise<OptionItem[]> {
    const items = await listControlCabinetsUseCase.getBulk(ids);
    return items.map(toCabinetOption);
  }

  async function fetchSpsControllers(search: string): Promise<OptionItem[]> {
    const res = await listSPSControllersUseCase.execute({
      pagination: { page: 1, pageSize: 100 },
      search: { text: search }
    });
    return res.items.map(toControllerOption);
  }

  async function fetchSpsControllersByIds(ids: string[]): Promise<OptionItem[]> {
    const items = await listSPSControllersUseCase.getBulk(ids);
    return items.map(toControllerOption);
  }

  async function handleStartExport() {
    if (!canStartExport) {
      addToast(translate('field_device.export.toasts.select_filter'), 'error');
      return;
    }
    submitting = true;
    try {
      const job = await exportUseCase.createExport({ ...exportRequest, force_async: forceAsync });
      activeJobId = job.job_id;
      facilityJobState.track({
        jobId: job.job_id,
        kind: 'field_device',
        type: 'export',
        class: 'export',
        status: job.status === 'processing' ? 'running' : job.status,
        progress: job.progress,
        stage: job.message
      });
      if (job.status === 'queued' || job.status === 'processing') {
        addToast(translate('field_device.export.toasts.started'), 'success');
      }
      if (job.status === 'completed') {
        addToast(translate('field_device.export.toasts.completed'), 'success');
      }
    } catch (error) {
      addToast(
        error instanceof Error
          ? error.message
          : translate('field_device.export.toasts.start_failed'),
        'error'
      );
    } finally {
      submitting = false;
    }
  }

  function handleDownload() {
    if (!activeJob) return;
    window.location.href =
      activeJob.result?.download_url ?? exportUseCase.getExportDownloadUrl(activeJob.jobId);
  }
</script>

<Card.Root>
  <Card.Header>
    <Card.Title class="flex items-center gap-2">
      <FileSpreadsheet class="size-4" />
      {$t('field_device.export.title')}
    </Card.Title>
    <Card.Description>
      {$t('field_device.export.description')}
    </Card.Description>
  </Card.Header>
  <Card.Content class="space-y-4">
    <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
      <div class="space-y-2">
        <Label class="text-sm font-medium" for="export-projects">
          {$t('field_device.export.projects')}
        </Label>
        <AsyncMultiSelect
          id="export-projects"
          bind:value={selectedProjectIds}
          fetcher={fetchProjects}
          fetchByIds={fetchProjectsByIds}
          labelKey="label"
          idKey="id"
          placeholder={$t('field_device.export.projects_placeholder')}
          searchPlaceholder={$t('field_device.export.projects_search')}
        />
      </div>
      <div class="space-y-2">
        <Label class="text-sm font-medium" for="export-buildings">
          {$t('field_device.export.buildings')}
        </Label>
        <AsyncMultiSelect
          id="export-buildings"
          bind:value={selectedBuildingIds}
          fetcher={fetchBuildings}
          fetchByIds={fetchBuildingsByIds}
          labelKey="label"
          idKey="id"
          placeholder={$t('field_device.export.buildings_placeholder')}
          searchPlaceholder={$t('field_device.export.buildings_search')}
        />
      </div>
      <div class="space-y-2">
        <Label class="text-sm font-medium" for="export-cabinets">
          {$t('field_device.export.control_cabinets')}
        </Label>
        <AsyncMultiSelect
          id="export-cabinets"
          bind:value={selectedControlCabinetIds}
          fetcher={fetchControlCabinets}
          fetchByIds={fetchControlCabinetsByIds}
          labelKey="label"
          idKey="id"
          placeholder={$t('field_device.export.control_cabinets_placeholder')}
          searchPlaceholder={$t('field_device.export.control_cabinets_search')}
        />
      </div>
      <div class="space-y-2">
        <Label class="text-sm font-medium" for="export-controllers">
          {$t('field_device.export.sps_controllers')}
        </Label>
        <AsyncMultiSelect
          id="export-controllers"
          bind:value={selectedSPSControllerIds}
          fetcher={fetchSpsControllers}
          fetchByIds={fetchSpsControllersByIds}
          labelKey="label"
          idKey="id"
          placeholder={$t('field_device.export.sps_controllers_placeholder')}
          searchPlaceholder={$t('field_device.export.sps_controllers_search')}
        />
      </div>
    </div>

    <div class="flex items-center gap-2">
      <Checkbox id="force-async" bind:checked={forceAsync} />
      <Label for="force-async" class="text-sm">{$t('field_device.export.force_async')}</Label>
    </div>

    <div class="flex gap-2">
      <Button onclick={handleStartExport} disabled={!canStartExport || submitting || isRunning}>
        {#if submitting || isRunning}
          <LoaderCircle class="mr-2 size-4 animate-spin" />
        {:else}
          <Play class="mr-2 size-4" />
        {/if}
        {$t('field_device.export.actions.start')}
      </Button>

      {#if isCompleted}
        <Button variant="outline" onclick={handleDownload}>
          <Download class="mr-2 size-4" />
          {$t('field_device.export.actions.download', {
            type: activeJob?.result?.output_type === 'zip' ? 'ZIP' : 'Excel'
          })}
        </Button>
      {/if}
    </div>

    {#if activeJob}
      <div class="space-y-2 rounded-md border p-3">
        <div class="flex items-center justify-between text-sm">
          <div>
            {$t('field_device.export.status.label')}
            <span class="font-medium">{activeJob.status}</span>
          </div>
          <div>{activeJob.progress}%</div>
        </div>
        <div class="h-2 w-full overflow-hidden rounded-md bg-muted">
          <div class="h-full bg-primary transition-all" style={`width: ${progressWidth};`}></div>
        </div>
        <p class="text-xs text-muted-foreground">{activeJob.stage}</p>
        {#if isFailed && activeJob.error}
          <p class="text-sm text-destructive">{activeJob.error}</p>
        {/if}
      </div>
    {/if}
  </Card.Content>
</Card.Root>
