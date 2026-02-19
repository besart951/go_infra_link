<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { ManageObjectDataUseCase } from '$lib/application/useCases/facility/manageObjectDataUseCase.js';
	import { objectDataRepository } from '$lib/infrastructure/api/objectDataRepository.js';
	const manageObjectData = new ManageObjectDataUseCase(objectDataRepository);
	import { getErrorMessage, getFieldError, getFieldErrors } from '$lib/api/client.js';
	import type { ObjectData, BacnetObjectInput } from '$lib/domain/facility/index.js';

	import { Plus } from '@lucide/svelte';
	import BacnetObjectRow from '../bacnet/BacnetObjectRow.svelte';
	import ApparatMultiSelect from '../selects/ApparatMultiSelect.svelte';
	import { createTranslator } from '$lib/i18n/translator.js';

	interface Props {
		initialData?: ObjectData;
		onSuccess?: (objectData: ObjectData) => void;
		onCancel?: () => void;
	}

	let { initialData, onSuccess, onCancel }: Props = $props();

	const t = createTranslator();

	let description = $state('');
	let version = $state('1.0');
	let is_active = $state(true);
	let apparat_ids: string[] = $state([]);
	let bacnetObjects: BacnetObjectInput[] = $state([]);
	let loading = $state(false);
	let error = $state('');
	let fieldErrors = $state<Record<string, string>>({});

	type BacnetRowErrors = Partial<
		Record<
			| 'text_fix'
			| 'description'
			| 'software_type'
			| 'software_number'
			| 'hardware_type'
			| 'hardware_quantity',
			string
		>
	>;

	$effect(() => {
		if (initialData) {
			description = initialData.description ?? '';
			version = initialData.version ?? '1.0';
			is_active = initialData.is_active ?? true;
			apparat_ids = (initialData.apparats ?? []).map((a) => a.id);
			bacnetObjects = (initialData.bacnet_objects ?? []).map((obj) => ({
				text_fix: obj.text_fix ?? '',
				description: obj.description ?? '',
				gms_visible: obj.gms_visible ?? false,
				optional: obj.optional ?? false,
				text_individual: obj.text_individual ?? '',
				software_type: obj.software_type ?? 'ai',
				software_number: obj.software_number ?? 1,
				hardware_type: obj.hardware_type ?? 'ai',
				hardware_quantity: obj.hardware_quantity ?? 1,
				alarm_type_id: obj.alarm_type_id ?? ''
			}));
		} else {
			description = '';
			version = '1.0';
			is_active = true;
			apparat_ids = [];
			bacnetObjects = [];
		}
	});

	const fieldError = (name: string) => getFieldError(fieldErrors, name, ['objectdata']);

	function getBacnetIndexedFieldError(index: number, name: string): string | undefined {
		const baseCandidates = [
			`bacnetobjects[${index}].${name}`,
			`bacnet_objects[${index}].${name}`,
			`bacnetobjects.${index}.${name}`,
			`bacnet_objects.${index}.${name}`,
			`objectdata.bacnetobjects[${index}].${name}`,
			`objectdata.bacnet_objects[${index}].${name}`,
			`objectdata.bacnetobjects.${index}.${name}`,
			`objectdata.bacnet_objects.${index}.${name}`,
			`error.bacnetobjects[${index}].${name}`,
			`errors.bacnetobjects[${index}].${name}`
		];

		for (const key of baseCandidates) {
			if (fieldErrors[key]) {
				return fieldErrors[key];
			}
		}

		if (name === 'text_fix') {
			return (
				fieldErrors['objectdata.bacnetobject.textfix'] ??
				fieldErrors['objectdata.bacnetobject.text_fix'] ??
				fieldErrors['bacnetobject.textfix'] ??
				fieldErrors['bacnetobject.text_fix']
			);
		}

		if (name === 'software_type') {
			return fieldErrors['objectdata.bacnetobject.software_type'] ?? fieldErrors['bacnetobject.software_type'];
		}

		if (name === 'software_number') {
			return (
				fieldErrors['objectdata.bacnetobject.software_number'] ??
				fieldErrors['bacnetobject.software_number'] ??
				fieldErrors['objectdata.bacnetobject.software'] ??
				fieldErrors['bacnetobject.software']
			);
		}

		return undefined;
	}

	function getBacnetRowErrors(index: number): BacnetRowErrors {
		return {
			text_fix: getBacnetIndexedFieldError(index, 'text_fix'),
			software_type: getBacnetIndexedFieldError(index, 'software_type'),
			software_number: getBacnetIndexedFieldError(index, 'software_number'),
			hardware_type: getBacnetIndexedFieldError(index, 'hardware_type'),
			hardware_quantity: getBacnetIndexedFieldError(index, 'hardware_quantity')
		};
	}

	function validateBacnetObjects(): boolean {
		const nextFieldErrors: Record<string, string> = {};
		const seenSoftware = new Set<string>();

		for (const [index, obj] of bacnetObjects.entries()) {
			if (!obj.text_fix?.trim()) {
				nextFieldErrors[`bacnetobjects[${index}].text_fix`] = 'textfix is required';
			}

			const softwareType = obj.software_type?.trim().toLowerCase() ?? '';
			if (!softwareType) {
				nextFieldErrors[`bacnetobjects[${index}].software_type`] = 'software_type is required';
				continue;
			}

			const softwareNumber = Number(obj.software_number);
			if (!Number.isFinite(softwareNumber) || softwareNumber < 0 || softwareNumber > 65535) {
				nextFieldErrors[`bacnetobjects[${index}].software_number`] =
					'software_number must be between 0 and 65535';
				continue;
			}

			const softwareKey = `${softwareType}:${softwareNumber}`;
			if (seenSoftware.has(softwareKey)) {
				nextFieldErrors[`bacnetobjects[${index}].software_number`] =
					'software_type + software_number must be unique within the object data';
			} else {
				seenSoftware.add(softwareKey);
			}
		}

		if (Object.keys(nextFieldErrors).length > 0) {
			fieldErrors = nextFieldErrors;
			error = '';
			return false;
		}

		return true;
	}

	function addBacnetObject() {
		bacnetObjects = [
			...bacnetObjects,
			{
				text_fix: '',
				description: '',
				gms_visible: false,
				optional: false,
				text_individual: '',
				software_type: 'ai',
				software_number: 1,
				hardware_type: 'ai',
				hardware_quantity: 1,
				alarm_type_id: ''
			}
		];
	}

	function removeBacnetObject(index: number) {
		bacnetObjects = bacnetObjects.filter((_, i) => i !== index);
	}

	function updateBacnetObject(index: number, field: string, value: any) {
		bacnetObjects = bacnetObjects.map((obj, i) => {
			if (i === index) {
				return { ...obj, [field]: value };
			}
			return obj;
		});
	}

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		loading = true;
		error = '';
		fieldErrors = {};

		if (!validateBacnetObjects()) {
			loading = false;
			return;
		}

		try {
			if (initialData) {
				const res = await manageObjectData.update(initialData.id, {
					description,
					version,
					is_active,
					apparat_ids,
					bacnet_objects: bacnetObjects
				});
				onSuccess?.(res);
			} else {
				const res = await manageObjectData.create({
					description,
					version,
					is_active,
					apparat_ids,
					bacnet_objects: bacnetObjects
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
			{initialData
				? $t('facility.forms.object_data.title_edit')
				: $t('facility.forms.object_data.title_new')}
		</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-3">
		<div class="space-y-2 md:col-span-2">
			<Label for="object_data_description">{$t('common.description')}</Label>
			<Input id="object_data_description" bind:value={description} required maxlength={250} />
			{#if fieldError('description')}
				<p class="text-sm text-red-500">{fieldError('description')}</p>
			{/if}
		</div>
		<div class="space-y-2">
			<Label for="object_data_version">{$t('facility.forms.object_data.version_label')}</Label>
			<Input id="object_data_version" bind:value={version} required maxlength={100} />
			{#if fieldError('version')}
				<p class="text-sm text-red-500">{fieldError('version')}</p>
			{/if}
		</div>
		<div class="flex items-center gap-2 md:col-span-3">
			<input id="object_data_active" type="checkbox" bind:checked={is_active} class="h-4 w-4" />
			<Label for="object_data_active">{$t('common.active')}</Label>
		</div>
		{#if fieldError('is_active')}
			<p class="text-sm text-red-500 md:col-span-3">{fieldError('is_active')}</p>
		{/if}
		<div class="space-y-2 md:col-span-3">
			<Label for="object_data_apparats">{$t('facility.forms.object_data.apparats_label')}</Label>
			<ApparatMultiSelect id="object_data_apparats" bind:value={apparat_ids} />
			{#if fieldError('apparat_ids')}
				<p class="text-sm text-red-500">{fieldError('apparat_ids')}</p>
			{/if}
		</div>
	</div>

	<!-- BACnet Objects Section -->
	<div class="space-y-3 pt-4">
		<div class="flex items-center justify-between border-t pt-4">
			<div>
				<h4 class="text-base font-medium">{$t('facility.forms.object_data.bacnet_title')}</h4>
				<p class="text-sm text-muted-foreground">
					{$t('facility.forms.object_data.bacnet_description')}
				</p>
			</div>
			<Button type="button" variant="outline" size="sm" onclick={addBacnetObject}>
				<Plus class="mr-2 size-4" />
				{$t('facility.forms.object_data.bacnet_add')}
			</Button>
		</div>

		{#if bacnetObjects.length === 0}
			<div class="rounded-md border border-dashed p-8 text-center">
				<p class="text-sm text-muted-foreground">
					{$t('facility.forms.object_data.bacnet_empty')}
				</p>
			</div>
		{:else}
			<div class="space-y-3">
				{#each bacnetObjects as obj, index (index)}
					{@const rowErrors = getBacnetRowErrors(index)}
					<BacnetObjectRow
						{index}
						bind:textFix={obj.text_fix}
						bind:description={obj.description}
						bind:gmsVisible={obj.gms_visible}
						bind:optional={obj.optional}
						bind:textIndividual={obj.text_individual}
						bind:softwareType={obj.software_type}
						bind:softwareNumber={obj.software_number}
						bind:hardwareType={obj.hardware_type}
						bind:hardwareQuantity={obj.hardware_quantity}
						bind:alarmTypeId={obj.alarm_type_id}
						errors={rowErrors}
						onRemove={() => removeBacnetObject(index)}
						onUpdate={(field, value) => updateBacnetObject(index, field, value)}
					/>
				{/each}
			</div>
		{/if}
	</div>

	{#if error}
		<p class="text-sm text-red-500">{error}</p>
	{/if}

	<div class="flex justify-end gap-2 pt-2">
		<Button type="button" variant="ghost" onclick={onCancel}>{$t('common.cancel')}</Button>
		<Button type="submit" disabled={loading}>
			{initialData ? $t('common.update') : $t('common.create')}
		</Button>
	</div>
</form>
