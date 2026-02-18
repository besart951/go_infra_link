<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import PaginatedList from '$lib/components/list/PaginatedList.svelte';
	import { addToast } from '$lib/components/toast.svelte';
	import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
	import { confirm } from '$lib/stores/confirm-dialog.js';
	import ControlCabinetForm from '$lib/components/facility/forms/ControlCabinetForm.svelte';
	import SPSControllerForm from '$lib/components/facility/forms/SPSControllerForm.svelte';
	import FieldDeviceListView from '$lib/components/facility/field-device/FieldDeviceListView.svelte';
	import {
		getProject,
		listProjectControlCabinets,
		addProjectControlCabinet,
		removeProjectControlCabinet,
		listProjectSPSControllers,
		addProjectSPSController
	} from '$lib/infrastructure/api/project.adapter.js';
	import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
	import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
	import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
	import type { Project } from '$lib/domain/project/index.js';
	import type {
		ProjectControlCabinetLink,
		ProjectSPSControllerLink
	} from '$lib/domain/project/index.js';
	import type { Building, ControlCabinet, SPSController } from '$lib/domain/facility/index.js';
	import { ArrowLeft, Plus, Pencil } from '@lucide/svelte';

	const projectId = $derived($page.params.id ?? '');

	let project = $state<Project | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let controlCabinetLinks = $state<ProjectControlCabinetLink[]>([]);
	let controlCabinetOptions = $state<ControlCabinet[]>([]);
	let controlCabinetLoading = $state(false);
	let showControlCabinetForm = $state(false);
	let controlCabinetSearch = $state('');
	let buildingMap = $state(new Map<string, string>());
	const buildingRequests = new Set<string>();

	let spsControllerLinks = $state<ProjectSPSControllerLink[]>([]);
	let spsControllerOptions = $state<SPSController[]>([]);
	let spsControllerLoading = $state(false);
	let showSpsControllerForm = $state(false);
	let editingSpsController: SPSController | undefined = $state(undefined);
	let spsControllerSearchText = $state('');
	let spsControllerPage = $state(1);
	const spsControllerPageSize = 10;

	const filteredControlCabinetLinks = $derived(
		controlCabinetSearch.trim()
			? controlCabinetLinks.filter((link) =>
					controlCabinetLabel(link.control_cabinet_id)
						.toLowerCase()
						.includes(controlCabinetSearch.trim().toLowerCase())
				)
			: controlCabinetLinks
	);

	function uniqueIds(ids: string[]): string[] {
		return Array.from(new Set(ids.filter(Boolean)));
	}

	async function fetchControlCabinetsByIds(ids: string[]): Promise<ControlCabinet[]> {
		const unique = uniqueIds(ids);
		if (unique.length === 0) return [];
		return controlCabinetRepository.getBulk(unique);
	}

	async function fetchSpsControllersByIds(ids: string[]): Promise<SPSController[]> {
		const unique = uniqueIds(ids);
		if (unique.length === 0) return [];
		return spsControllerRepository.getBulk(unique);
	}

	const spsControllerLinkMap = $derived.by(
		() => new Map(spsControllerLinks.map((link) => [link.sps_controller_id, link]))
	);

	const linkedSpsControllers = $derived.by(() =>
		spsControllerOptions.filter((controller) => spsControllerLinkMap.has(controller.id))
	);

	const filteredSpsControllers = $derived.by(() => {
		const query = spsControllerSearchText.trim().toLowerCase();
		if (!query) return linkedSpsControllers;
		return linkedSpsControllers.filter((controller) =>
			[
				controller.device_name,
				controller.ga_device,
				controller.ip_address,
				controlCabinetLabel(controller.control_cabinet_id)
			]
				.filter(Boolean)
				.some((value) => value!.toLowerCase().includes(query))
		);
	});

	const spsControllerTotalPages = $derived.by(() =>
		filteredSpsControllers.length === 0
			? 0
			: Math.ceil(filteredSpsControllers.length / spsControllerPageSize)
	);

	const spsControllerPageItems = $derived.by(() => {
		const start = (spsControllerPage - 1) * spsControllerPageSize;
		return filteredSpsControllers.slice(start, start + spsControllerPageSize);
	});

	const spsControllerListState = $derived.by(() => ({
		items: spsControllerPageItems,
		total: filteredSpsControllers.length,
		page: spsControllerTotalPages === 0 ? 1 : Math.min(spsControllerPage, spsControllerTotalPages),
		pageSize: spsControllerPageSize,
		totalPages: spsControllerTotalPages,
		searchText: spsControllerSearchText,
		loading: spsControllerLoading,
		error: null
	}));

	function formatBuildingLabel(building: Building): string {
		return `${building.iws_code}-${building.building_group}`;
	}

	function getBuildingLabel(buildingId: string | undefined | null): string {
		if (!buildingId) return '-';
		return buildingMap.get(buildingId) ?? buildingId;
	}

	function updateBuildingMap(buildings: Building[]) {
		const next = new Map(buildingMap);
		for (const building of buildings) {
			next.set(building.id, formatBuildingLabel(building));
		}
		buildingMap = next;
	}

	async function ensureBuildingLabels(items: ControlCabinet[]) {
		const uniqueIds = new Set(
			items.map((item) => item.building_id).filter((id): id is string => Boolean(id))
		);
		const missingIds = Array.from(uniqueIds).filter(
			(id) => !buildingMap.has(id) && !buildingRequests.has(id)
		);

		if (missingIds.length === 0) return;

		missingIds.forEach((id) => buildingRequests.add(id));

		try {
			const items = await buildingRepository.getBulk(missingIds);
			updateBuildingMap(items);
		} catch (err) {
			console.error('Failed to load buildings:', err);
		} finally {
			missingIds.forEach((id) => buildingRequests.delete(id));
		}
	}

	function controlCabinetLabel(id: string): string {
		const item = controlCabinetOptions.find((c) => c.id === id);
		return item?.control_cabinet_nr || item?.id || id;
	}

	$effect(() => {
		if (spsControllerTotalPages === 0) {
			spsControllerPage = 1;
			return;
		}
		if (spsControllerPage > spsControllerTotalPages) {
			spsControllerPage = spsControllerTotalPages;
		}
	});

	async function loadProject() {
		if (!projectId) return;
		loading = true;
		error = null;
		try {
			project = await getProject(projectId);
		} catch (err) {
			const message = err instanceof Error ? err.message : 'Failed to load project';
			error = message;
			addToast(message, 'error');
		} finally {
			loading = false;
		}
	}

	async function loadControlCabinets() {
		if (!projectId) return;
		controlCabinetLoading = true;
		try {
			const linksRes = await listProjectControlCabinets(projectId, { page: 1, limit: 100 });
			controlCabinetLinks = linksRes.items;

			// Hydrate labels by fetching the exact linked cabinets (avoid global list limits).
			const cabinetIds = linksRes.items.map((l) => l.control_cabinet_id);
			controlCabinetOptions = await fetchControlCabinetsByIds(cabinetIds);
			await ensureBuildingLabels(controlCabinetOptions);
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to load control cabinets', 'error');
		} finally {
			controlCabinetLoading = false;
		}
	}

	async function loadSpsControllers() {
		if (!projectId) return;
		spsControllerLoading = true;
		try {
			const linksRes = await listProjectSPSControllers(projectId, { page: 1, limit: 100 });
			spsControllerLinks = linksRes.items;

			// Hydrate linked SPS controllers by ID (avoid global list limits so newly created
			// controllers always appear).
			const controllerIds = linksRes.items.map((l) => l.sps_controller_id);
			spsControllerOptions = await fetchSpsControllersByIds(controllerIds);

			// Ensure control cabinet labels are available for the cabinet column.
			const cabinetIds = spsControllerOptions.map((c) => c.control_cabinet_id);
			const existing = new Set(controlCabinetOptions.map((c) => c.id));
			const missing = uniqueIds(cabinetIds).filter((id) => !existing.has(id));
			if (missing.length > 0) {
				const fetched = await fetchControlCabinetsByIds(missing);
				controlCabinetOptions = [...controlCabinetOptions, ...fetched];
				await ensureBuildingLabels(fetched);
			}
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to load SPS controllers', 'error');
		} finally {
			spsControllerLoading = false;
		}
	}

	async function handleControlCabinetCreated(item: ControlCabinet) {
		if (!projectId) return;
		try {
			await addProjectControlCabinet(projectId, item.id);
			addToast('Control cabinet created', 'success');
			showControlCabinetForm = false;
			await loadControlCabinets();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to link control cabinet', 'error');
		}
	}

	async function removeControlCabinet(linkId: string) {
		if (!projectId) return;
		const ok = await confirm({
			title: 'Remove control cabinet',
			message: 'Remove this control cabinet from the project?',
			confirmText: 'Remove',
			cancelText: 'Cancel',
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await removeProjectControlCabinet(projectId, linkId);
			addToast('Control cabinet removed', 'success');
			await loadControlCabinets();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to remove control cabinet', 'error');
		}
	}

	function handleSpsControllerEdit(item: SPSController) {
		editingSpsController = item;
		showSpsControllerForm = true;
	}

	function handleSpsControllerCreate() {
		editingSpsController = undefined;
		showSpsControllerForm = true;
	}

	function handleSpsControllerCancel() {
		showSpsControllerForm = false;
		editingSpsController = undefined;
	}

	function handleSpsControllerSearch(text: string) {
		spsControllerSearchText = text;
		spsControllerPage = 1;
	}

	function handleSpsControllerPageChange(page: number) {
		spsControllerPage = page;
	}

	async function handleSpsControllerSuccess(item: SPSController) {
		if (!projectId) return;
		try {
			if (!editingSpsController) {
				await addProjectSPSController(projectId, item.id);
				addToast('SPS controller created', 'success');
			} else {
				addToast('SPS controller updated', 'success');
			}
			showSpsControllerForm = false;
			editingSpsController = undefined;
			await loadSpsControllers();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to save SPS controller', 'error');
		}
	}

	async function handleDeleteSpsController(item: SPSController) {
		const ok = await confirm({
			title: 'Delete SPS controller',
			message: `Delete ${item.device_name}?`,
			confirmText: 'Delete',
			cancelText: 'Cancel',
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await spsControllerRepository.delete(item.id);
			addToast('SPS controller deleted', 'success');
			await loadSpsControllers();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to delete SPS controller', 'error');
		}
	}

	onMount(() => {
		loadProject();
		loadControlCabinets();
		loadSpsControllers();
	});
</script>

<ConfirmDialog />

<div class="flex flex-col gap-6">
	<div class="flex items-start gap-3">
		<Button variant="outline" onclick={() => goto('/projects')}>
			<ArrowLeft class="mr-2 h-4 w-4" />
			Back
		</Button>
		<div>
			<h1 class="text-3xl font-bold tracking-tight">{project?.name ?? 'Project'}</h1>
			<p class="mt-1 text-muted-foreground">Manage project assets and assignments.</p>
		</div>
		<div class="ml-auto">
			<Button variant="outline" href={`/projects/${projectId}/settings`}>Settings</Button>
		</div>
	</div>

	{#if error}
		<div class="rounded-md border bg-muted px-4 py-3 text-muted-foreground">
			<p class="font-medium">Could not load project</p>
			<p class="text-sm">{error}</p>
		</div>
	{/if}

	{#if loading}
		<div class="rounded-lg border bg-background p-6">
			<div class="grid gap-4 md:grid-cols-2">
				{#each Array(6) as _}
					<Skeleton class="h-6 w-full" />
				{/each}
			</div>
		</div>
	{:else if !project}
		<div class="rounded-lg border bg-background p-6 text-sm text-muted-foreground">
			Project not found.
		</div>
	{:else}
		<div class="grid gap-6">
			<div class="rounded-lg border bg-background p-6">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<div>
						<h2 class="text-lg font-semibold">Control Cabinets</h2>
						<p class="text-sm text-muted-foreground">Create and assign control cabinets.</p>
					</div>
					<div class="flex items-center gap-2">
						<Input
							class="w-64"
							placeholder="Search control cabinets..."
							bind:value={controlCabinetSearch}
						/>
						<Button variant="outline" onclick={loadControlCabinets} disabled={controlCabinetLoading}
							>Refresh</Button
						>
						{#if !showControlCabinetForm}
							<Button onclick={() => (showControlCabinetForm = true)}>
								<Plus class="mr-2 size-4" />
								New Control Cabinet
							</Button>
						{/if}
					</div>
				</div>

				{#if showControlCabinetForm}
					<ControlCabinetForm
						onSuccess={handleControlCabinetCreated}
						onCancel={() => (showControlCabinetForm = false)}
					/>
				{/if}

				<div class="mt-6 rounded-lg border bg-background">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Control Cabinet</Table.Head>
								<Table.Head>Building</Table.Head>
								<Table.Head class="w-32"></Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#if controlCabinetLoading}
								{#each Array(4) as _}
									<Table.Row>
										<Table.Cell><Skeleton class="h-4 w-60" /></Table.Cell>
										<Table.Cell><Skeleton class="h-4 w-40" /></Table.Cell>
										<Table.Cell><Skeleton class="h-8 w-20" /></Table.Cell>
									</Table.Row>
								{/each}
							{:else if filteredControlCabinetLinks.length === 0}
								<Table.Row>
									<Table.Cell colspan={3} class="h-20 text-center text-sm text-muted-foreground">
										No control cabinets found.
									</Table.Cell>
								</Table.Row>
							{:else}
								{#each filteredControlCabinetLinks as link (link.id)}
									<Table.Row>
										<Table.Cell class="font-medium">
											{controlCabinetLabel(link.control_cabinet_id)}
										</Table.Cell>
										<Table.Cell>
											{getBuildingLabel(
												controlCabinetOptions.find((c) => c.id === link.control_cabinet_id)
													?.building_id
											)}
										</Table.Cell>
										<Table.Cell class="text-right">
											<Button variant="outline" onclick={() => removeControlCabinet(link.id)}>
												Remove
											</Button>
										</Table.Cell>
									</Table.Row>
								{/each}
							{/if}
						</Table.Body>
					</Table.Root>
				</div>
			</div>

			<div class="rounded-lg border bg-background p-6">
				<div class="flex flex-wrap items-center justify-between gap-3">
					<div>
						<h2 class="text-lg font-semibold">SPS Controllers</h2>
						<p class="text-sm text-muted-foreground">Create and assign SPS controllers.</p>
					</div>
					<div class="flex items-center gap-2">
						{#if !showSpsControllerForm}
							<Button onclick={handleSpsControllerCreate}>
								<Plus class="mr-2 size-4" />
								New SPS Controller
							</Button>
						{/if}
					</div>
				</div>

				{#if showSpsControllerForm}
					<SPSControllerForm
						initialData={editingSpsController}
						onSuccess={handleSpsControllerSuccess}
						onCancel={handleSpsControllerCancel}
					/>
				{/if}

				<PaginatedList
					state={spsControllerListState}
					columns={[
						{ key: 'device_name', label: 'Device Name' },
						{ key: 'ga_device', label: 'GA Device' },
						{ key: 'ip_address', label: 'IP Address' },
						{ key: 'cabinet', label: 'Cabinet' },
						{ key: 'created', label: 'Created' },
						{ key: 'actions', label: 'Actions', width: 'w-[120px]' }
					]}
					searchPlaceholder="Search SPS controllers..."
					emptyMessage="No SPS controllers found. Create your first SPS controller to get started."
					onSearch={handleSpsControllerSearch}
					onPageChange={handleSpsControllerPageChange}
					onReload={loadSpsControllers}
				>
					{#snippet rowSnippet(controller: SPSController)}
						<Table.Cell class="font-medium">
							<a href="/facility/sps-controllers/{controller.id}" class="hover:underline">
								{controller.device_name}
							</a>
						</Table.Cell>
						<Table.Cell>{controller.ga_device ?? '-'}</Table.Cell>
						<Table.Cell>
							{#if controller.ip_address}
								<code class="rounded bg-muted px-1.5 py-0.5 text-sm">
									{controller.ip_address}
								</code>
							{:else}
								-
							{/if}
						</Table.Cell>
						<Table.Cell>{controlCabinetLabel(controller.control_cabinet_id)}</Table.Cell>
						<Table.Cell>
							{new Date(controller.created_at).toLocaleDateString()}
						</Table.Cell>
						<Table.Cell>
							<div class="flex items-center gap-2">
								<Button
									variant="ghost"
									size="icon"
									onclick={() => handleSpsControllerEdit(controller)}
								>
									<Pencil class="size-4" />
								</Button>
								<Button variant="ghost" size="sm" href="/facility/sps-controllers/{controller.id}">
									View
								</Button>
								<Button
									variant="ghost"
									size="sm"
									onclick={() => handleDeleteSpsController(controller)}
								>
									Delete
								</Button>
							</div>
						</Table.Cell>
					{/snippet}
				</PaginatedList>
			</div>

			<div class="rounded-lg border bg-background p-6">
				<div class="mb-4">
					<h2 class="text-lg font-semibold">Field Devices</h2>
					<p class="text-sm text-muted-foreground">
						Manage field devices linked to this project with full inline editing.
					</p>
				</div>
				<FieldDeviceListView {projectId} />
			</div>
		</div>
	{/if}
</div>

<ConfirmDialog />
