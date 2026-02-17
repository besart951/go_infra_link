<script lang="ts">
	import { onDestroy } from 'svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import AsyncMultiSelect from '$lib/components/ui/combobox/AsyncMultiSelect.svelte';
	import { addToast } from '$lib/components/toast.svelte';
	import { fieldDeviceRepository } from '$lib/infrastructure/api/fieldDeviceRepository.js';
	import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
	import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
	import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
	import { getProject, listProjects } from '$lib/infrastructure/api/project.adapter.js';
	import type {
		Building,
		ControlCabinet,
		FieldDeviceExportJobResponse,
		SPSController
	} from '$lib/domain/facility/index.js';
	import type { Project } from '$lib/domain/project/index.js';
	import { Download, FileSpreadsheet, LoaderCircle, Play } from '@lucide/svelte';

	interface Props {
		projectId?: string;
	}

	type OptionItem = { id: string; label: string };

	let { projectId }: Props = $props();

	let selectedProjectIds = $state<string[]>([]);
	let selectedBuildingIds = $state<string[]>([]);
	let selectedControlCabinetIds = $state<string[]>([]);
	let selectedSPSControllerIds = $state<string[]>([]);
	let forceAsync = $state(false);

	let submitting = $state(false);
	let polling = $state(false);
	let activeJob = $state<FieldDeviceExportJobResponse | null>(null);
	let pollingTimer: ReturnType<typeof setInterval> | null = null;

	$effect(() => {
		if (projectId && selectedProjectIds.length === 0) {
			selectedProjectIds = [projectId];
		}
	});

	const canStartExport = $derived(
		selectedProjectIds.length > 0 ||
			selectedBuildingIds.length > 0 ||
			selectedControlCabinetIds.length > 0 ||
			selectedSPSControllerIds.length > 0
	);

	const progressWidth = $derived(`${Math.min(100, Math.max(0, activeJob?.progress ?? 0))}%`);
	const isRunning = $derived(activeJob?.status === 'queued' || activeJob?.status === 'processing');
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
		const res = await listProjects({ page: 1, limit: 100, search });
		return res.items.map(toProjectOption);
	}

	async function fetchProjectsByIds(ids: string[]): Promise<OptionItem[]> {
		const items = await Promise.all(ids.map((id) => getProject(id)));
		return items.map(toProjectOption);
	}

	async function fetchBuildings(search: string): Promise<OptionItem[]> {
		const res = await buildingRepository.list({
			pagination: { page: 1, pageSize: 100 },
			search: { text: search }
		});
		return res.items.map(toBuildingOption);
	}

	async function fetchBuildingsByIds(ids: string[]): Promise<OptionItem[]> {
		const items = await buildingRepository.getBulk(ids);
		return items.map(toBuildingOption);
	}

	async function fetchControlCabinets(search: string): Promise<OptionItem[]> {
		const res = await controlCabinetRepository.list({
			pagination: { page: 1, pageSize: 100 },
			search: { text: search }
		});
		return res.items.map(toCabinetOption);
	}

	async function fetchControlCabinetsByIds(ids: string[]): Promise<OptionItem[]> {
		const items = await controlCabinetRepository.getBulk(ids);
		return items.map(toCabinetOption);
	}

	async function fetchSpsControllers(search: string): Promise<OptionItem[]> {
		const res = await spsControllerRepository.list({
			pagination: { page: 1, pageSize: 100 },
			search: { text: search }
		});
		return res.items.map(toControllerOption);
	}

	async function fetchSpsControllersByIds(ids: string[]): Promise<OptionItem[]> {
		const items = await spsControllerRepository.getBulk(ids);
		return items.map(toControllerOption);
	}

	function stopPolling() {
		if (pollingTimer) {
			clearInterval(pollingTimer);
			pollingTimer = null;
		}
		polling = false;
	}

	async function refreshJobStatus() {
		if (!activeJob?.job_id) return;
		try {
			const next = await fieldDeviceRepository.getExportJob(activeJob.job_id);
			activeJob = next;
			if (next.status === 'completed') {
				stopPolling();
				addToast('Export completed. You can download the file now.', 'success');
			}
			if (next.status === 'failed') {
				stopPolling();
				addToast(next.error || 'Export failed', 'error');
			}
		} catch (error) {
			stopPolling();
			addToast(error instanceof Error ? error.message : 'Failed to refresh export status', 'error');
		}
	}

	function startPolling() {
		stopPolling();
		polling = true;
		pollingTimer = setInterval(() => {
			void refreshJobStatus();
		}, 2000);
	}

	async function handleStartExport() {
		if (!canStartExport) {
			addToast('Select at least one filter before export.', 'error');
			return;
		}
		submitting = true;
		try {
			const job = await fieldDeviceRepository.createExport({
				project_ids: selectedProjectIds,
				buildings_id: selectedBuildingIds,
				control_cabinet_id: selectedControlCabinetIds,
				sps_controller_id: selectedSPSControllerIds,
				force_async: forceAsync
			});
			activeJob = job;
			if (job.status === 'queued' || job.status === 'processing') {
				startPolling();
				addToast('Export gestartet. Status wird aktualisiert…', 'success');
			}
			if (job.status === 'completed') {
				addToast('Export completed. Ready for download.', 'success');
			}
		} catch (error) {
			addToast(error instanceof Error ? error.message : 'Failed to start export', 'error');
		} finally {
			submitting = false;
		}
	}

	function handleDownload() {
		if (!activeJob?.job_id) return;
		window.location.href = fieldDeviceRepository.getExportDownloadUrl(activeJob.job_id);
	}

	onDestroy(() => {
		stopPolling();
	});
</script>

<Card.Root>
	<Card.Header>
		<Card.Title class="flex items-center gap-2">
			<FileSpreadsheet class="size-4" />
			Excel Export
		</Card.Title>
		<Card.Description>
			Wähle einzelne Projekte, Gebäude, Schaltschränke oder SPS-Controller für den Export.
		</Card.Description>
	</Card.Header>
	<Card.Content class="space-y-4">
		<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
			<div class="space-y-2">
				<label class="text-sm font-medium" for="export-projects">Projects</label>
				<AsyncMultiSelect
					id="export-projects"
					bind:value={selectedProjectIds}
					fetcher={fetchProjects}
					fetchByIds={fetchProjectsByIds}
					labelKey="label"
					idKey="id"
					placeholder="Select projects"
					searchPlaceholder="Search projects..."
				/>
			</div>
			<div class="space-y-2">
				<label class="text-sm font-medium" for="export-buildings">Buildings</label>
				<AsyncMultiSelect
					id="export-buildings"
					bind:value={selectedBuildingIds}
					fetcher={fetchBuildings}
					fetchByIds={fetchBuildingsByIds}
					labelKey="label"
					idKey="id"
					placeholder="Select buildings"
					searchPlaceholder="Search buildings..."
				/>
			</div>
			<div class="space-y-2">
				<label class="text-sm font-medium" for="export-cabinets">Control Cabinets</label>
				<AsyncMultiSelect
					id="export-cabinets"
					bind:value={selectedControlCabinetIds}
					fetcher={fetchControlCabinets}
					fetchByIds={fetchControlCabinetsByIds}
					labelKey="label"
					idKey="id"
					placeholder="Select control cabinets"
					searchPlaceholder="Search control cabinets..."
				/>
			</div>
			<div class="space-y-2">
				<label class="text-sm font-medium" for="export-controllers">SPS Controllers</label>
				<AsyncMultiSelect
					id="export-controllers"
					bind:value={selectedSPSControllerIds}
					fetcher={fetchSpsControllers}
					fetchByIds={fetchSpsControllersByIds}
					labelKey="label"
					idKey="id"
					placeholder="Select SPS controllers"
					searchPlaceholder="Search SPS controllers..."
				/>
			</div>
		</div>

		<div class="flex items-center gap-2">
			<input id="force-async" type="checkbox" bind:checked={forceAsync} class="h-4 w-4" />
			<label for="force-async" class="text-sm">Always queue export (async)</label>
		</div>

		<div class="flex gap-2">
			<Button onclick={handleStartExport} disabled={!canStartExport || submitting || isRunning}>
				{#if submitting || isRunning}
					<LoaderCircle class="mr-2 size-4 animate-spin" />
				{:else}
					<Play class="mr-2 size-4" />
				{/if}
				Start Export
			</Button>

			{#if isCompleted}
				<Button variant="outline" onclick={handleDownload}>
					<Download class="mr-2 size-4" />
					Download {activeJob?.output_type === 'zip' ? 'ZIP' : 'Excel'}
				</Button>
			{/if}
		</div>

		{#if activeJob}
			<div class="space-y-2 rounded-md border p-3">
				<div class="flex items-center justify-between text-sm">
					<div>Status: <span class="font-medium">{activeJob.status}</span></div>
					<div>{activeJob.progress}%</div>
				</div>
				<div class="h-2 w-full overflow-hidden rounded bg-muted">
					<div class="h-full bg-primary transition-all" style={`width: ${progressWidth};`}></div>
				</div>
				<p class="text-xs text-muted-foreground">{activeJob.message}</p>
				{#if isFailed && activeJob.error}
					<p class="text-sm text-destructive">{activeJob.error}</p>
				{/if}
			</div>
		{/if}
	</Card.Content>
</Card.Root>
