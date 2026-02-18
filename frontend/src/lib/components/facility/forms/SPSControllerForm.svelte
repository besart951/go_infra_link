<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import ControlCabinetSelect from '../selects/ControlCabinetSelect.svelte';
	import SystemTypeSelect from '../selects/SystemTypeSelect.svelte';
	import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
	import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
	import { systemTypeRepository } from '$lib/infrastructure/api/systemTypeRepository.js';
	import { spsControllerSystemTypeRepository } from '$lib/infrastructure/api/spsControllerSystemTypeRepository.js';
	import { ManageSPSControllerUseCase } from '$lib/application/useCases/facility/manageSPSControllerUseCase.js';
	import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
	const manageSPSController = new ManageSPSControllerUseCase(spsControllerRepository);
	import { getErrorMessage, getFieldError, getFieldErrors } from '$lib/api/client.js';
	import type {
		Building,
		ControlCabinet,
		SPSController,
		SPSControllerSystemType,
		SPSControllerSystemTypeInput,
		SystemType
	} from '$lib/domain/facility/index.js';
	import { useLiveValidation } from '$lib/hooks/useLiveValidation.svelte.js';

	interface Props {
		initialData?: SPSController;
		onSuccess?: (controller: SPSController) => void;
		onCancel?: () => void;
	}

	let { initialData, onSuccess, onCancel }: Props = $props();

	let ga_device = $state('');
	let device_name = $state('');
	let ip_address = $state('');
	let subnet = $state('');
	let gateway = $state('');
	let vlan = $state('');
	let control_cabinet_id = $state('');
	let system_type_id = $state('');
	let systemTypes: SPSControllerSystemTypeInput[] = $state([]);
	let systemTypeLabels: Record<string, string> = $state({});
	let systemTypeDetails: Record<string, SystemType> = $state({});
	let systemTypeDetailsLoading = $state(false);
	let systemTypesLoading = $state(false);
	let lastLoadedSystemTypesFor: string | null = $state(null);
	let controlCabinet: ControlCabinet | null = $state(null);
	let building: Building | null = $state(null);
	let cabinetLoading = $state(false);
	let buildingLoading = $state(false);
	let lastLoadedCabinetId: string | null = $state(null);
	let gaDeviceTouched = $state(false);
	let lastGADeviceControlCabinetId: string | null = $state(null);
	let nextGADevice = $state<string | null>(null);
	let gaDeviceSuggestionLoading = $state(false);
	let gaDeviceCheckTimer: ReturnType<typeof setTimeout> | null = null;

	let loading = $state(false);
	let error = $state('');
	let fieldErrors = $state<Record<string, string>>({});
	const liveValidation = useLiveValidation((data: { id?: string; control_cabinet_id: string; ga_device?: string; device_name: string; ip_address?: string; subnet?: string; gateway?: string; vlan?: string }) => manageSPSController.validate(data), { debounceMs: 400 });

	$effect(() => {
		if (!initialData) {
			return;
		}
		ga_device = initialData.ga_device ?? '';
		device_name = initialData.device_name;
		ip_address = initialData.ip_address ?? '';
		subnet = initialData.subnet ?? '';
		gateway = initialData.gateway ?? '';
		vlan = initialData.vlan ?? '';
		control_cabinet_id = initialData.control_cabinet_id;
		gaDeviceTouched = false;
		nextGADevice = null;
	});

	$effect(() => {
		if (initialData?.id && lastLoadedSystemTypesFor !== initialData.id) {
			lastLoadedSystemTypesFor = initialData.id;
			loadSystemTypes();
		}
		if (!initialData?.id && lastLoadedSystemTypesFor !== null) {
			lastLoadedSystemTypesFor = null;
			systemTypes = [];
			systemTypeLabels = {};
		}
	});

	$effect(() => {
		if (!control_cabinet_id) {
			nextGADevice = null;
			controlCabinet = null;
			building = null;
			device_name = '';
			if (gaDeviceCheckTimer) {
				clearTimeout(gaDeviceCheckTimer);
				gaDeviceCheckTimer = null;
			}
			return;
		}
		triggerValidation();
		if (lastGADeviceControlCabinetId !== control_cabinet_id) {
			lastGADeviceControlCabinetId = control_cabinet_id;
			if (!initialData) {
				ga_device = '';
				gaDeviceTouched = false;
				void refreshNextGADevice(true);
			} else {
				void refreshNextGADevice(false);
			}
		} else if (!nextGADevice && !gaDeviceSuggestionLoading) {
			void refreshNextGADevice(false);
		}
	});

	$effect(() => {
		if (!control_cabinet_id) return;
		if (lastLoadedCabinetId === control_cabinet_id) return;
		lastLoadedCabinetId = control_cabinet_id;
		void loadCabinetAndBuilding(control_cabinet_id);
	});

	$effect(() => {
		const nextName = buildDeviceName(controlCabinet, building, ga_device);
		if (nextName === null) return;
		if (nextName !== device_name) {
			device_name = nextName;
			triggerValidation();
		}
	});

	const fieldError = (name: string) => getFieldError(fieldErrors, name, ['spscontroller']);
	const liveFieldError = (name: string) =>
		getFieldError(liveValidation.fieldErrors, name, ['spscontroller']);
	const combinedFieldError = (name: string) => liveFieldError(name) || fieldError(name);
	const gaDeviceIsSuboptimal = $derived.by(() => {
		if (!gaDeviceTouched || !nextGADevice) return false;
		const current = ga_device.trim().toUpperCase();
		return current !== '' && current !== nextGADevice;
	});

	function triggerValidation() {
		if (!control_cabinet_id) return;
		liveValidation.trigger({
			id: initialData?.id,
			control_cabinet_id,
			ga_device: ga_device || undefined,
			device_name,
			ip_address: ip_address || undefined,
			subnet: subnet || undefined,
			gateway: gateway || undefined,
			vlan: vlan || undefined
		});
	}

	function scheduleNextGADeviceCheck() {
		if (!control_cabinet_id) return;
		if (gaDeviceCheckTimer) {
			clearTimeout(gaDeviceCheckTimer);
		}
		gaDeviceCheckTimer = setTimeout(() => {
			gaDeviceCheckTimer = null;
			void refreshNextGADevice(false);
		}, 400);
	}

	async function fetchNextGADevice(): Promise<string | null> {
		if (!control_cabinet_id) return null;
		gaDeviceSuggestionLoading = true;
		try {
			const res = await manageSPSController.getNextGADevice(control_cabinet_id, initialData?.id);
			nextGADevice = res?.ga_device ?? null;
			return nextGADevice;
		} catch (e) {
			console.error(e);
			return null;
		} finally {
			gaDeviceSuggestionLoading = false;
		}
	}

	async function refreshNextGADevice(applyIfEmpty: boolean) {
		const next = await fetchNextGADevice();
		if (applyIfEmpty && next) {
			ga_device = next;
			gaDeviceTouched = false;
			triggerValidation();
		}
	}

	async function applySuggestedGADevice() {
		if (!control_cabinet_id) return;
		if (!nextGADevice) {
			await refreshNextGADevice(false);
		}
		if (nextGADevice) {
			ga_device = nextGADevice;
			gaDeviceTouched = false;
			triggerValidation();
		}
	}

	async function loadSystemTypes() {
		if (!initialData?.id) return;
		systemTypesLoading = true;
		try {
			const res = await spsControllerSystemTypeRepository.list({
				pagination: { page: 1, pageSize: 100 },
				search: { text: '' },
				filters: { sps_controller_id: initialData.id }
			});
			const items = res.items;
			const labelFallbacks: Record<string, string> = {};
			const uniqueIds = Array.from(
				new Set(items.map((item) => item.system_type_id).filter(Boolean))
			);
			items.forEach((item) => {
				if (item.system_type_name) {
					labelFallbacks[item.system_type_id] = item.system_type_name;
				}
			});
			systemTypes = items.map((item: SPSControllerSystemType) => ({
				system_type_id: item.system_type_id,
				number: item.number ?? undefined,
				document_name: item.document_name ?? undefined
			}));
			systemTypeLabels = await buildSystemTypeLabels(uniqueIds, labelFallbacks);
			await Promise.all(uniqueIds.map((id) => ensureSystemTypeDetails(id)));
		} catch (e) {
			console.error(e);
		} finally {
			systemTypesLoading = false;
		}
	}

	$effect(() => {
		if (!system_type_id) return;
		if (systemTypeDetails[system_type_id]) return;
		void ensureSystemTypeDetails(system_type_id);
	});

	async function addSystemType() {
		if (!system_type_id) {
			return;
		}
		try {
			const systemType = await ensureSystemTypeDetails(system_type_id);
			if (!systemType) return;
			systemTypeLabels = {
				...systemTypeLabels,
				[system_type_id]: buildSystemTypeLabel(
					systemType.name,
					systemType.number_min,
					systemType.number_max
				)
			};
			const nextNumber = getNextAvailableSystemTypeNumber(
				system_type_id,
				systemType.number_min,
				systemType.number_max
			);
			if (nextNumber == null) {
				return;
			}
			systemTypes = [
				...systemTypes,
				{
					system_type_id,
					number: nextNumber
				}
			];
		} catch (e) {
			console.error(e);
		}
	}

	function formatNumber(value: number): string {
		return String(value).padStart(4, '0');
	}

	function buildSystemTypeLabel(name: string, min: number, max: number): string {
		return `${name} (${formatNumber(min)}-${formatNumber(max)})`;
	}

	async function ensureSystemTypeDetails(id: string): Promise<SystemType | null> {
		if (systemTypeDetails[id]) return systemTypeDetails[id];
		systemTypeDetailsLoading = true;
		try {
			const systemType = await systemTypeRepository.get(id);
			systemTypeDetails = { ...systemTypeDetails, [id]: systemType };
			return systemType;
		} catch (e) {
			console.error(e);
			return null;
		} finally {
			systemTypeDetailsLoading = false;
		}
	}

	function buildDeviceName(
		cabinet: ControlCabinet | null,
		cabinetBuilding: Building | null,
		gaDeviceValue: string
	): string | null {
		const ga = gaDeviceValue.trim();
		if (!ga) return '';
		const iwsCode = cabinetBuilding?.iws_code?.trim();
		const cabinetNr = cabinet?.control_cabinet_nr?.trim();
		if (!iwsCode || !cabinetNr) return null;
		return `${iwsCode}_${cabinetNr}_${ga}`;
	}

	async function loadCabinetAndBuilding(cabinetId: string) {
		cabinetLoading = true;
		buildingLoading = true;
		try {
			const cabinet = await controlCabinetRepository.get(cabinetId);
			if (control_cabinet_id !== cabinetId) return;
			controlCabinet = cabinet;
			if (!cabinet?.building_id) {
				building = null;
				return;
			}
			const b = await buildingRepository.get(cabinet.building_id);
			if (control_cabinet_id !== cabinetId) return;
			building = b;
		} catch (e) {
			console.error(e);
			controlCabinet = null;
			building = null;
		} finally {
			cabinetLoading = false;
			buildingLoading = false;
		}
	}

	async function buildSystemTypeLabels(
		ids: string[],
		fallbacks: Record<string, string>
	): Promise<Record<string, string>> {
		const entries = await Promise.all(
			ids.map(async (id) => {
				try {
					const systemType = await systemTypeRepository.get(id);
					return [
						id,
						buildSystemTypeLabel(systemType.name, systemType.number_min, systemType.number_max)
					] as const;
				} catch (error) {
					console.error('Failed to load system type details:', error);
					return [id, fallbacks[id] ?? id] as const;
				}
			})
		);
		return Object.fromEntries(entries);
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

	function getNextAvailableSystemTypeNumber(
		systemTypeId: string,
		min: number,
		max: number
	): number | null {
		const usedNumbers = new Set(
			systemTypes
				.filter((item) => item.system_type_id === systemTypeId)
				.map((item) => item.number)
				.filter((value): value is number => typeof value === 'number')
		);
		for (let number = min; number <= max; number += 1) {
			if (!usedNumbers.has(number)) return number;
		}
		return null;
	}

	const systemTypeAddState = $derived.by(() => {
		if (!system_type_id) {
			return { disabled: true, tooltip: 'Select a system type first.' };
		}
		const details = systemTypeDetails[system_type_id];
		if (!details) {
			return {
				disabled: true,
				tooltip: systemTypeDetailsLoading
					? 'Loading system type details...'
					: 'Loading system type.'
			};
		}
		const rangeSize = details.number_max - details.number_min + 1;
		const usedNumbers = new Set(
			systemTypes
				.filter(
					(item) =>
						item.system_type_id === system_type_id &&
						typeof item.number === 'number' &&
						item.number >= details.number_min &&
						item.number <= details.number_max
				)
				.map((item) => item.number as number)
		);
		if (usedNumbers.size >= rangeSize) {
			return { disabled: true, tooltip: 'All numbers are used for this system type.' };
		}
		return { disabled: false, tooltip: 'Add next available number.' };
	});

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
				const res = await manageSPSController.update(initialData.id, {
					id: initialData.id,
					ga_device,
					device_name,
					ip_address: ip_address || undefined,
					subnet: subnet || undefined,
					gateway: gateway || undefined,
					vlan: vlan || undefined,
					control_cabinet_id,
					system_types: systemTypes
				});
				onSuccess?.(res);
			} else {
				const res = await manageSPSController.create({
					ga_device,
					device_name,
					ip_address: ip_address || undefined,
					subnet: subnet || undefined,
					gateway: gateway || undefined,
					vlan: vlan || undefined,
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

	<div class="space-y-2">
		<Label>Control Cabinet</Label>
		<div class="block">
			<ControlCabinetSelect bind:value={control_cabinet_id} width="w-full" />
		</div>
		{#if combinedFieldError('control_cabinet_id')}
			<p class="text-sm text-red-500">{combinedFieldError('control_cabinet_id')}</p>
		{:else if !control_cabinet_id}
			<p class="text-sm text-muted-foreground">
				Select a control cabinet to unlock the remaining fields.
			</p>
		{/if}
	</div>

	{#if control_cabinet_id}
		<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
			<div class="space-y-2">
				<Label for="ga_device" class="flex items-center gap-2">
					GA Device
					<span
						class="inline-flex h-4 w-4 items-center justify-center rounded-full border text-[10px] text-muted-foreground"
						title="Yellow means there is a lower unused GA device available. You can keep your value or press 'Use next'."
					>
						i
					</span>
				</Label>
				<div class="flex items-start gap-2">
					<Input
						id="ga_device"
						value={ga_device}
						required
						maxlength={3}
						class={`flex-1 ${
							gaDeviceIsSuboptimal
								? 'border-yellow-400 bg-yellow-50/60 focus-visible:ring-yellow-300'
								: ''
						}`}
						oninput={(e) => {
							ga_device = (e.target as HTMLInputElement).value.toUpperCase();
							gaDeviceTouched = true;
							triggerValidation();
							scheduleNextGADeviceCheck();
						}}
					/>
					{#if gaDeviceIsSuboptimal}
						<Button
							type="button"
							variant="outline"
							size="sm"
							class="mt-1"
							onclick={applySuggestedGADevice}
						>
							Use next
						</Button>
					{/if}
				</div>
				{#if gaDeviceIsSuboptimal && nextGADevice}
					<p class="text-xs text-yellow-700">
						Lowest available is {nextGADevice}.
					</p>
				{/if}
				{#if gaDeviceSuggestionLoading}
					<p class="text-xs text-muted-foreground">Checking lowest available GA device...</p>
				{/if}
				{#if combinedFieldError('ga_device')}
					<p class="text-sm text-red-500">{combinedFieldError('ga_device')}</p>
				{/if}
			</div>
			<div class="space-y-2">
				<Label for="device_name">Device Name</Label>
				<Input
					id="device_name"
					bind:value={device_name}
					disabled
					readonly
					required
					maxlength={100}
				/>
				{#if cabinetLoading || buildingLoading}
					<p class="text-xs text-muted-foreground">Loading device name...</p>
				{/if}
				{#if combinedFieldError('device_name')}
					<p class="text-sm text-red-500">{combinedFieldError('device_name')}</p>
				{/if}
			</div>
			<div class="space-y-2">
				<Label for="ip_address">IP Address</Label>
				<Input id="ip_address" bind:value={ip_address} maxlength={50} oninput={triggerValidation} />
				{#if combinedFieldError('ip_address')}
					<p class="text-sm text-red-500">{combinedFieldError('ip_address')}</p>
				{/if}
			</div>
			<div class="space-y-2">
				<Label for="subnet">Subnet Mask</Label>
				<Input id="subnet" bind:value={subnet} maxlength={50} oninput={triggerValidation} />
				{#if combinedFieldError('subnet')}
					<p class="text-sm text-red-500">{combinedFieldError('subnet')}</p>
				{/if}
			</div>
			<div class="space-y-2">
				<Label for="gateway">Gateway</Label>
				<Input id="gateway" bind:value={gateway} maxlength={50} oninput={triggerValidation} />
				{#if combinedFieldError('gateway')}
					<p class="text-sm text-red-500">{combinedFieldError('gateway')}</p>
				{/if}
			</div>
			<div class="space-y-2">
				<Label for="vlan">VLAN</Label>
				<Input id="vlan" bind:value={vlan} maxlength={50} oninput={triggerValidation} />
				{#if combinedFieldError('vlan')}
					<p class="text-sm text-red-500">{combinedFieldError('vlan')}</p>
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
						disabled={systemTypeAddState.disabled}
						title={systemTypeAddState.tooltip}
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
									{systemTypeLabels[st.system_type_id] ?? st.system_type_id}
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
			{#if combinedFieldError('system_types')}
				<p class="text-sm text-red-500">{combinedFieldError('system_types')}</p>
			{/if}
		</div>
	{/if}

	{#if error || liveValidation.error}
		<p class="text-sm text-red-500">{error || liveValidation.error}</p>
	{/if}

	<div class="flex justify-end gap-2 pt-2">
		<Button type="button" variant="ghost" onclick={onCancel}>Cancel</Button>
		<Button type="submit" disabled={loading}>
			{initialData ? 'Update' : 'Create'}
		</Button>
	</div>
</form>
