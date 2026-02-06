<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import BuildingSelect from './BuildingSelect.svelte';
	import {
		createControlCabinet,
		updateControlCabinet
	} from '$lib/infrastructure/api/facility.adapter.js';
	import { getErrorMessage, getFieldError, getFieldErrors } from '$lib/api/client.js';
	import type { ControlCabinet } from '$lib/domain/facility/index.js';

	interface Props {
		initialData?: ControlCabinet;
		onSuccess?: (cabinet: ControlCabinet) => void;
		onCancel?: () => void;
	}

	let { initialData, onSuccess, onCancel }: Props = $props();

	let control_cabinet_nr = $state('');
	let building_id = $state('');
	let loading = $state(false);
	let error = $state('');
	let fieldErrors = $state<Record<string, string>>({});

	// React to initialData changes
	$effect(() => {
		if (initialData) {
			control_cabinet_nr = initialData.control_cabinet_nr;
			building_id = initialData.building_id;
		}
	});

	const fieldError = (name: string) => getFieldError(fieldErrors, name, ['controlcabinet']);

	async function handleSubmit() {
		loading = true;
		error = '';
		fieldErrors = {};

		if (!building_id) {
			error = 'Please select a building';
			loading = false;
			return;
		}

		try {
			if (initialData) {
				const res = await updateControlCabinet(initialData.id, {
					id: initialData.id,
					control_cabinet_nr,
					building_id
				});
				onSuccess?.(res);
			} else {
				const res = await createControlCabinet({
					control_cabinet_nr,
					building_id
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

<form
	onsubmit={(e) => {
		e.preventDefault();
		handleSubmit();
	}}
	class="space-y-4 rounded-md border bg-muted/20 p-4"
>
	<div class="mb-4 flex items-center justify-between">
		<h3 class="text-lg font-medium">
			{initialData ? 'Edit Control Cabinet' : 'New Control Cabinet'}
		</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="control_cabinet_nr">Control Cabinet Nr</Label>
			<Input id="control_cabinet_nr" bind:value={control_cabinet_nr} required maxlength={11} />
			{#if fieldError('control_cabinet_nr')}
				<p class="text-sm text-red-500">{fieldError('control_cabinet_nr')}</p>
			{/if}
		</div>

		<div class="space-y-2">
			<Label>Building</Label>
			<div class="block">
				<BuildingSelect bind:value={building_id} width="w-full" />
			</div>
			{#if fieldError('building_id')}
				<p class="text-sm text-red-500">{fieldError('building_id')}</p>
			{/if}
		</div>
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
