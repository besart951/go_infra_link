<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { createObjectData, updateObjectData } from '$lib/infrastructure/api/facility.adapter.js';
	import { getErrorMessage } from '$lib/api/client.js';
	import type { ObjectData, BacnetObjectInput } from '$lib/domain/facility/index.js';
	import { createEventDispatcher } from 'svelte';
	import { Plus } from '@lucide/svelte';
	import BacnetObjectRow from './BacnetObjectRow.svelte';

	interface Props {
		initialData?: ObjectData;
	}

	let { initialData }: Props = $props();

	let description = $state('');
	let version = $state('1.0');
	let is_active = $state(true);
	let bacnetObjects: BacnetObjectInput[] = $state([]);
	let loading = $state(false);
	let error = $state('');

	$effect(() => {
		if (initialData) {
			description = initialData.description ?? '';
			version = initialData.version ?? '1.0';
			is_active = initialData.is_active ?? true;
			bacnetObjects = (initialData.bacnet_objects ?? []).map((obj) => ({
				text_fix: obj.text_fix ?? '',
				description: obj.description ?? '',
				gms_visible: obj.gms_visible ?? false,
				optional: obj.optional ?? false,
				text_individual: obj.text_individual ?? '',
				software_type: obj.software_type ?? 'ai',
				software_number: obj.software_number ?? 1,
				hardware_type: obj.hardware_type ?? 'ai',
				hardware_quantity: obj.hardware_quantity ?? 1
			}));
		} else {
			description = '';
			version = '1.0';
			is_active = true;
			bacnetObjects = [];
		}
	});

	const dispatch = createEventDispatcher();

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
				hardware_quantity: 1
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

		try {
			if (initialData) {
				const res = await updateObjectData(initialData.id, {
					description,
					version,
					is_active
				});
				dispatch('success', res);
			} else {
				const res = await createObjectData({
					description,
					version,
					is_active,
					bacnet_objects: bacnetObjects
				});
				dispatch('success', res);
			}
		} catch (e) {
			console.error(e);
			error = getErrorMessage(e);
		} finally {
			loading = false;
		}
	}
</script>

<form onsubmit={handleSubmit} class="space-y-4 rounded-md border bg-muted/20 p-4">
	<div class="mb-4 flex items-center justify-between">
		<h3 class="text-lg font-medium">{initialData ? 'Edit Object Data' : 'New Object Data'}</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-3">
		<div class="space-y-2 md:col-span-2">
			<Label for="object_data_description">Description</Label>
			<Input id="object_data_description" bind:value={description} required maxlength={250} />
		</div>
		<div class="space-y-2">
			<Label for="object_data_version">Version</Label>
			<Input id="object_data_version" bind:value={version} required maxlength={100} />
		</div>
		<div class="flex items-center gap-2 md:col-span-3">
			<input id="object_data_active" type="checkbox" bind:checked={is_active} class="h-4 w-4" />
			<Label for="object_data_active">Active</Label>
		</div>
	</div>

	<!-- BACnet Objects Section -->
	<div class="space-y-3 pt-4">
		<div class="flex items-center justify-between border-t pt-4">
			<div>
				<h4 class="text-base font-medium">BACnet Objects</h4>
				<p class="text-sm text-muted-foreground">Add and configure BACnet objects for this template</p>
			</div>
			<Button type="button" variant="outline" size="sm" onclick={addBacnetObject}>
				<Plus class="mr-2 size-4" />
				Add Object
			</Button>
		</div>

		{#if bacnetObjects.length === 0}
			<div class="rounded-md border border-dashed p-8 text-center">
				<p class="text-sm text-muted-foreground">
					No BACnet objects added yet. Click "Add Object" to get started.
				</p>
			</div>
		{:else}
			<div class="space-y-3">
				{#each bacnetObjects as obj, index (index)}
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
		<Button type="button" variant="ghost" onclick={() => dispatch('cancel')}>Cancel</Button>
		<Button type="submit" disabled={loading}>{initialData ? 'Update' : 'Create'}</Button>
	</div>
</form>
