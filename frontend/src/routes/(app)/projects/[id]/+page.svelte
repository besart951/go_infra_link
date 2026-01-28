<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import Toasts, { addToast } from '$lib/components/toast.svelte';
	import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
	import { confirm } from '$lib/stores/confirm-dialog.js';
	import ControlCabinetForm from '$lib/components/facility/ControlCabinetForm.svelte';
	import SPSControllerForm from '$lib/components/facility/SPSControllerForm.svelte';
	import FieldDeviceForm from '$lib/components/facility/FieldDeviceForm.svelte';
	import {
		getProject,
		listProjectControlCabinets,
		addProjectControlCabinet,
		removeProjectControlCabinet,
		listProjectSPSControllers,
		addProjectSPSController,
		removeProjectSPSController,
		listProjectFieldDevices,
		addProjectFieldDevice,
		removeProjectFieldDevice
	} from '$lib/infrastructure/api/project.adapter.js';
	import {
		listControlCabinets,
		listSPSControllers,
		listFieldDevices
	} from '$lib/infrastructure/api/facility.adapter.js';
	import type { Project } from '$lib/domain/project/index.js';
	import type {
		ProjectControlCabinetLink,
		ProjectSPSControllerLink,
		ProjectFieldDeviceLink
	} from '$lib/domain/project/index.js';
	import type { ControlCabinet, SPSController, FieldDevice } from '$lib/domain/facility/index.js';
	import { ArrowLeft, Plus } from '@lucide/svelte';

	const projectId = $derived($page.params.id ?? '');

	let project = $state<Project | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let controlCabinetLinks = $state<ProjectControlCabinetLink[]>([]);
	let controlCabinetOptions = $state<ControlCabinet[]>([]);
	let controlCabinetLoading = $state(false);
	let showControlCabinetForm = $state(false);

	let spsControllerLinks = $state<ProjectSPSControllerLink[]>([]);
	let spsControllerOptions = $state<SPSController[]>([]);
	let spsControllerLoading = $state(false);
	let showSpsControllerForm = $state(false);

	let fieldDeviceLinks = $state<ProjectFieldDeviceLink[]>([]);
	let fieldDeviceOptions = $state<FieldDevice[]>([]);
	let fieldDeviceLoading = $state(false);
	let showFieldDeviceForm = $state(false);

	function controlCabinetLabel(id: string): string {
		const item = controlCabinetOptions.find((c) => c.id === id);
		return item?.control_cabinet_nr || item?.id || id;
	}

	function spsControllerLabel(id: string): string {
		const item = spsControllerOptions.find((c) => c.id === id);
		return item?.device_name || item?.id || id;
	}

	function fieldDeviceLabel(id: string): string {
		const item = fieldDeviceOptions.find((c) => c.id === id);
		return item?.bmk || item?.apparat_nr?.toString() || item?.id || id;
	}

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
			const [linksRes, optionsRes] = await Promise.all([
				listProjectControlCabinets(projectId, { page: 1, limit: 100 }),
				listControlCabinets({ page: 1, limit: 100 })
			]);
			controlCabinetLinks = linksRes.items;
			controlCabinetOptions = optionsRes.items;
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
			const [linksRes, optionsRes] = await Promise.all([
				listProjectSPSControllers(projectId, { page: 1, limit: 100 }),
				listSPSControllers({ page: 1, limit: 100 })
			]);
			spsControllerLinks = linksRes.items;
			spsControllerOptions = optionsRes.items;
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to load SPS controllers', 'error');
		} finally {
			spsControllerLoading = false;
		}
	}

	async function loadFieldDevices() {
		if (!projectId) return;
		fieldDeviceLoading = true;
		try {
			const [linksRes, optionsRes] = await Promise.all([
				listProjectFieldDevices(projectId, { page: 1, limit: 100 }),
				listFieldDevices({ page: 1, limit: 100 })
			]);
			fieldDeviceLinks = linksRes.items;
			fieldDeviceOptions = optionsRes.items;
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to load field devices', 'error');
		} finally {
			fieldDeviceLoading = false;
		}
	}

	async function handleControlCabinetCreated(event: CustomEvent<ControlCabinet>) {
		if (!projectId) return;
		const item = event.detail;
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

	async function handleSpsControllerCreated(event: CustomEvent<SPSController>) {
		if (!projectId) return;
		const item = event.detail;
		try {
			await addProjectSPSController(projectId, item.id);
			addToast('SPS controller created', 'success');
			showSpsControllerForm = false;
			await loadSpsControllers();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to link SPS controller', 'error');
		}
	}

	async function removeSpsController(linkId: string) {
		if (!projectId) return;
		const ok = await confirm({
			title: 'Remove SPS controller',
			message: 'Remove this SPS controller from the project?',
			confirmText: 'Remove',
			cancelText: 'Cancel',
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await removeProjectSPSController(projectId, linkId);
			addToast('SPS controller removed', 'success');
			await loadSpsControllers();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to remove SPS controller', 'error');
		}
	}

	async function handleFieldDeviceCreated(event: CustomEvent<FieldDevice>) {
		if (!projectId) return;
		const item = event.detail;
		try {
			await addProjectFieldDevice(projectId, item.id);
			addToast('Field device created', 'success');
			showFieldDeviceForm = false;
			await loadFieldDevices();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to link field device', 'error');
		}
	}

	async function removeFieldDevice(linkId: string) {
		if (!projectId) return;
		const ok = await confirm({
			title: 'Remove field device',
			message: 'Remove this field device from the project?',
			confirmText: 'Remove',
			cancelText: 'Cancel',
			variant: 'destructive'
		});
		if (!ok) return;
		try {
			await removeProjectFieldDevice(projectId, linkId);
			addToast('Field device removed', 'success');
			await loadFieldDevices();
		} catch (err) {
			addToast(err instanceof Error ? err.message : 'Failed to remove field device', 'error');
		}
	}

	onMount(() => {
		loadProject();
		loadControlCabinets();
		loadSpsControllers();
		loadFieldDevices();
	});
</script>

<Toasts />
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
						on:success={handleControlCabinetCreated}
						on:cancel={() => (showControlCabinetForm = false)}
					/>
				{/if}

				<div class="mt-6 rounded-lg border bg-background">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Control Cabinet</Table.Head>
								<Table.Head class="w-32"></Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#if controlCabinetLoading}
								{#each Array(4) as _}
									<Table.Row>
										<Table.Cell><Skeleton class="h-4 w-60" /></Table.Cell>
										<Table.Cell><Skeleton class="h-8 w-20" /></Table.Cell>
									</Table.Row>
								{/each}
							{:else if controlCabinetLinks.length === 0}
								<Table.Row>
									<Table.Cell colspan={2} class="h-20 text-center text-sm text-muted-foreground">
										No control cabinets assigned yet.
									</Table.Cell>
								</Table.Row>
							{:else}
								{#each controlCabinetLinks as link (link.id)}
									<Table.Row>
										<Table.Cell class="font-medium">
											{controlCabinetLabel(link.control_cabinet_id)}
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
						<Button variant="outline" onclick={loadSpsControllers} disabled={spsControllerLoading}
							>Refresh</Button
						>
						{#if !showSpsControllerForm}
							<Button onclick={() => (showSpsControllerForm = true)}>
								<Plus class="mr-2 size-4" />
								New SPS Controller
							</Button>
						{/if}
					</div>
				</div>

				{#if showSpsControllerForm}
					<SPSControllerForm
						on:success={handleSpsControllerCreated}
						on:cancel={() => (showSpsControllerForm = false)}
					/>
				{/if}

				<div class="mt-6 rounded-lg border bg-background">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>SPS Controller</Table.Head>
								<Table.Head class="w-32"></Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#if spsControllerLoading}
								{#each Array(4) as _}
									<Table.Row>
										<Table.Cell><Skeleton class="h-4 w-60" /></Table.Cell>
										<Table.Cell><Skeleton class="h-8 w-20" /></Table.Cell>
									</Table.Row>
								{/each}
							{:else if spsControllerLinks.length === 0}
								<Table.Row>
									<Table.Cell colspan={2} class="h-20 text-center text-sm text-muted-foreground">
										No SPS controllers assigned yet.
									</Table.Cell>
								</Table.Row>
							{:else}
								{#each spsControllerLinks as link (link.id)}
									<Table.Row>
										<Table.Cell class="font-medium">
											{spsControllerLabel(link.sps_controller_id)}
										</Table.Cell>
										<Table.Cell class="text-right">
											<Button variant="outline" onclick={() => removeSpsController(link.id)}>
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
						<h2 class="text-lg font-semibold">Field Devices</h2>
						<p class="text-sm text-muted-foreground">Create and assign field devices.</p>
					</div>
					<div class="flex items-center gap-2">
						<Button variant="outline" onclick={loadFieldDevices} disabled={fieldDeviceLoading}
							>Refresh</Button
						>
						{#if !showFieldDeviceForm}
							<Button onclick={() => (showFieldDeviceForm = true)}>
								<Plus class="mr-2 size-4" />
								New Field Device
							</Button>
						{/if}
					</div>
				</div>

				{#if showFieldDeviceForm}
					<FieldDeviceForm
						{projectId}
						on:success={handleFieldDeviceCreated}
						on:cancel={() => (showFieldDeviceForm = false)}
					/>
				{/if}

				<div class="mt-6 rounded-lg border bg-background">
					<Table.Root>
						<Table.Header>
							<Table.Row>
								<Table.Head>Field Device</Table.Head>
								<Table.Head class="w-32"></Table.Head>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#if fieldDeviceLoading}
								{#each Array(4) as _}
									<Table.Row>
										<Table.Cell><Skeleton class="h-4 w-60" /></Table.Cell>
										<Table.Cell><Skeleton class="h-8 w-20" /></Table.Cell>
									</Table.Row>
								{/each}
							{:else if fieldDeviceLinks.length === 0}
								<Table.Row>
									<Table.Cell colspan={2} class="h-20 text-center text-sm text-muted-foreground">
										No field devices assigned yet.
									</Table.Cell>
								</Table.Row>
							{:else}
								{#each fieldDeviceLinks as link (link.id)}
									<Table.Row>
										<Table.Cell class="font-medium">
											{fieldDeviceLabel(link.field_device_id)}
										</Table.Cell>
										<Table.Cell class="text-right">
											<Button variant="outline" onclick={() => removeFieldDevice(link.id)}>
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
		</div>
	{/if}
</div>
