<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import BuildingSelect from '../selects/BuildingSelect.svelte';
	import { ManageControlCabinetUseCase } from '$lib/application/useCases/facility/manageControlCabinetUseCase.js';
	import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
	const manageControlCabinet = new ManageControlCabinetUseCase(controlCabinetRepository);
	import { getErrorMessage, getFieldError, getFieldErrors } from '$lib/api/client.js';
	import type { ControlCabinet } from '$lib/domain/facility/index.js';
	import { useLiveValidation } from '$lib/hooks/useLiveValidation.svelte.js';

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
	const liveValidation = useLiveValidation((data: { id?: string; building_id: string; control_cabinet_nr?: string }) => manageControlCabinet.validate(data), { debounceMs: 400 });

	// React to initialData changes
	$effect(() => {
		if (initialData) {
			control_cabinet_nr = initialData.control_cabinet_nr;
			building_id = initialData.building_id;
		}
	});

	$effect(() => {
		if (!building_id) return;
		triggerValidation();
	});

	const fieldError = (name: string) => getFieldError(fieldErrors, name, ['controlcabinet']);
	const liveFieldError = (name: string) =>
		getFieldError(liveValidation.fieldErrors, name, ['controlcabinet']);
	const combinedFieldError = (name: string) => liveFieldError(name) || fieldError(name);

	function triggerValidation() {
		if (!building_id) return;
		liveValidation.trigger({
			id: initialData?.id,
			building_id,
			control_cabinet_nr: control_cabinet_nr || undefined
		});
	}

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
				const res = await manageControlCabinet.update(initialData.id, {
					id: initialData.id,
					control_cabinet_nr,
					building_id
				});
				onSuccess?.(res);
			} else {
				const res = await manageControlCabinet.create({
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
			<Input
				id="control_cabinet_nr"
				bind:value={control_cabinet_nr}
				required
				maxlength={11}
				oninput={triggerValidation}
			/>
			{#if combinedFieldError('control_cabinet_nr')}
				<p class="text-sm text-red-500">{combinedFieldError('control_cabinet_nr')}</p>
			{/if}
		</div>

		<div class="space-y-2">
			<Label>Building</Label>
			<div class="block">
				<BuildingSelect bind:value={building_id} width="w-full" />
			</div>
			{#if combinedFieldError('building_id')}
				<p class="text-sm text-red-500">{combinedFieldError('building_id')}</p>
			{/if}
		</div>
	</div>

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
