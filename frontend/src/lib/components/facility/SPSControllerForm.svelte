<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import ControlCabinetSelect from './ControlCabinetSelect.svelte';
	import SystemTypeSelect from './SystemTypeSelect.svelte';
	import {
		createSPSController,
		getSystemType,
		listSPSControllerSystemTypes,
		updateSPSController
	} from '$lib/infrastructure/api/facility.adapter.js';
	import { getErrorMessage, getFieldError, getFieldErrors } from '$lib/api/client.js';
	import type {
		SPSController,
		SPSControllerSystemType,
		SPSControllerSystemTypeInput
	} from '$lib/domain/facility/index.js';

	interface Props {
		initialData?: SPSController;
		onSuccess?: (controller: SPSController) => void;
		onCancel?: () => void;
	}

	let { initialData, onSuccess, onCancel }: Props = $props();

	let ga_device = $state('');
	let device_name = $state('');
	let ip_address = $state('');
	let control_cabinet_id = $state('');
	let system_type_id = $state('');
	let systemTypes: SPSControllerSystemTypeInput[] = $state([]);
	let systemTypeNames: Record<string, string> = $state({});
	let systemTypesLoading = $state(false);
	let lastLoadedSystemTypesFor: string | null = $state(null);

	let loading = $state(false);
	let error = $state('');
	let fieldErrors = $state<Record<string, string>>({});

	$effect(() => {
		if (!initialData) {
			return;
		}
		ga_device = initialData.ga_device ?? '';
		device_name = initialData.device_name;
		ip_address = initialData.ip_address ?? '';
		control_cabinet_id = initialData.control_cabinet_id;
	});

	$effect(() => {
		if (initialData?.id && lastLoadedSystemTypesFor !== initialData.id) {
			lastLoadedSystemTypesFor = initialData.id;
			loadSystemTypes();
		}
		if (!initialData?.id && lastLoadedSystemTypesFor !== null) {
			lastLoadedSystemTypesFor = null;
			systemTypes = [];
			systemTypeNames = {};
		}
	});

	const fieldError = (name: string) => getFieldError(fieldErrors, name, ['spscontroller']);

	async function loadSystemTypes() {
		if (!initialData?.id) return;
		systemTypesLoading = true;
		try {
			const res = await listSPSControllerSystemTypes({
				page: 1,
				limit: 100,
				sps_controller_id: initialData.id
			});
			const names: Record<string, string> = {};
			systemTypes = (res.items ?? []).map((item: SPSControllerSystemType) => {
				if (item.system_type_name) {
					names[item.system_type_id] = item.system_type_name;
				}
				return {
					system_type_id: item.system_type_id,
					number: item.number ?? undefined,
					document_name: item.document_name ?? undefined
				};
			});
			systemTypeNames = names;
		} catch (e) {
			console.error(e);
		} finally {
			systemTypesLoading = false;
		}
	}

	async function addSystemType() {
		if (!system_type_id) {
			return;
		}
		try {
			const systemType = await getSystemType(system_type_id);
			systemTypeNames = {
				...systemTypeNames,
				[system_type_id]: systemType.name
			};
		} catch (e) {
			console.error(e);
		}
		systemTypes = [
			...systemTypes,
			{
				system_type_id
			}
		];
	}

	function removeSystemType(index: number) {
		systemTypes = systemTypes.filter((_, i) => i !== index);
	}

	function updateSystemTypeField(
		index: number,
		field: keyof SPSControllerSystemTypeInput,
		value: string
	) {
		systemTypes = systemTypes.map((item, i) => {
			if (i !== index) return item;
			if (field === 'number') {
				const parsed = value === '' ? undefined : Number(value);
				return { ...item, number: Number.isNaN(parsed) ? undefined : parsed };
			}
			return { ...item, document_name: value || undefined };
		});
	}

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		loading = true;
		error = '';
		fieldErrors = {};

		if (!control_cabinet_id) {
			error = 'Please select a control cabinet';
			loading = false;
			return;
		}

		try {
			if (initialData) {
				const res = await updateSPSController(initialData.id, {
					id: initialData.id,
					ga_device,
					device_name,
					ip_address,
					control_cabinet_id,
					system_types: systemTypes
				});
				onSuccess?.(res);
			} else {
				const res = await createSPSController({
					ga_device,
					device_name,
					ip_address,
					control_cabinet_id,
					system_types: systemTypes
				});
				onSuccess?.(res);
			}
		} catch (e) {
			console.error(e);
			fieldErrors = getFieldErrors(e);
			error = Object.keys(fieldErrors).length ? '' : getErrorMessage(e);
		} finally {
			loading = false;
		}
	}
</script>

<form onsubmit={handleSubmit} class="space-y-4 rounded-md border bg-muted/20 p-4">
	<div class="mb-4 flex items-center justify-between">
		<h3 class="text-lg font-medium">
			{initialData ? 'Edit SPS Controller' : 'New SPS Controller'}
		</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="ga_device">GA Device</Label>
			<Input id="ga_device" bind:value={ga_device} required maxlength={10} />
			{#if fieldError('ga_device')}
				<p class="text-sm text-red-500">{fieldError('ga_device')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="device_name">Device Name</Label>
			<Input id="device_name" bind:value={device_name} required maxlength={100} />
			{#if fieldError('device_name')}
				<p class="text-sm text-red-500">{fieldError('device_name')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="ip_address">IP Address</Label>
			<Input id="ip_address" bind:value={ip_address} required maxlength={50} />
			{#if fieldError('ip_address')}
				<p class="text-sm text-red-500">{fieldError('ip_address')}</p>
			{/if}
		</div>

		<div class="space-y-2">
			<Label>Control Cabinet</Label>
			<div class="block">
				<ControlCabinetSelect bind:value={control_cabinet_id} width="w-full" />
			</div>
			{#if fieldError('control_cabinet_id')}
				<p class="text-sm text-red-500">{fieldError('control_cabinet_id')}</p>
			{/if}
		</div>
	</div>

	<div class="space-y-3 pt-4">
		<div class="flex items-center justify-between border-t pt-4">
			<div>
				<h4 class="text-base font-medium">System Types</h4>
				<p class="text-sm text-muted-foreground">Assign system types to this SPS controller</p>
			</div>
			<div class="flex items-center gap-2">
				<SystemTypeSelect bind:value={system_type_id} width="w-[250px]" />
				<Button
					type="button"
					variant="outline"
					size="sm"
					onclick={addSystemType}
					disabled={!system_type_id}
				>
					Add
				</Button>
			</div>
		</div>

		{#if systemTypesLoading}
			<p class="text-sm text-muted-foreground">Loading system types...</p>
		{:else if systemTypes.length === 0}
			<div class="rounded-md border border-dashed p-6 text-center">
				<p class="text-sm text-muted-foreground">No system types added yet.</p>
			</div>
		{:else}
			<div class="max-h-80 space-y-2 overflow-y-auto pr-1">
				{#each systemTypes as st, index (index)}
					<div class="grid grid-cols-1 gap-3 rounded-md border p-3 md:grid-cols-12">
						<div class="md:col-span-4">
							<div class="text-xs text-muted-foreground">System Type</div>
							<div class="text-sm font-medium">
								{systemTypeNames[st.system_type_id] ?? st.system_type_id}
							</div>
						</div>
						<div class="md:col-span-3">
							<Label class="text-xs">Number</Label>
							<Input
								type="number"
								value={st.number ?? ''}
								oninput={(e) =>
									updateSystemTypeField(index, 'number', (e.target as HTMLInputElement).value)}
							/>
						</div>
						<div class="md:col-span-4">
							<Label class="text-xs">Document name</Label>
							<Input
								value={st.document_name ?? ''}
								oninput={(e) =>
									updateSystemTypeField(
										index,
										'document_name',
										(e.target as HTMLInputElement).value
									)}
								maxlength={250}
							/>
						</div>
						<div class="flex items-end justify-end md:col-span-1">
							<Button type="button" variant="ghost" onclick={() => removeSystemType(index)}>
								Remove
							</Button>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>

	{#if error}
		<p class="text-sm text-red-500">{error}</p>
	{/if}

	<div class="flex justify-end gap-2 pt-2">
		<Button type="button" variant="ghost" onclick={onCancel}>Cancel</Button>
		<Button type="submit" disabled={loading}>
			{initialData ? 'Update' : 'Create'}
		</Button>
	</div>
</form>
