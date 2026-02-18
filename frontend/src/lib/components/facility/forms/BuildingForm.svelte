<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import type { Building } from '$lib/domain/facility/index.js';
	import { ManageBuildingUseCase } from '$lib/application/useCases/facility/manageBuildingUseCase.js';
	import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
	const manageBuilding = new ManageBuildingUseCase(buildingRepository);
	import { useFormState } from '$lib/hooks/useFormState.svelte.js';
	import { useLiveValidation } from '$lib/hooks/useLiveValidation.svelte.js';
	import { getFieldError } from '$lib/api/client.js';
	import { createTranslator } from '$lib/i18n/translator.js';

	interface BuildingFormProps {
		initialData?: Building;
		onSuccess?: (building: Building) => void;
		onCancel?: () => void;
	}

	let { initialData, onSuccess, onCancel }: BuildingFormProps = $props();

	const t = createTranslator();

	let iws_code = $state('');
	let building_group = $state(0);

	$effect(() => {
		iws_code = initialData?.iws_code ?? '';
		building_group = initialData?.building_group ?? 0;
	});

	const formState = useFormState({
		onSuccess: (result: Building) => {
			onSuccess?.(result);
		}
	});

	const liveValidation = useLiveValidation(
		(data: { id?: string; iws_code: string; building_group: number }) =>
			manageBuilding.validate(data),
		{ debounceMs: 400 }
	);

	function triggerValidation() {
		liveValidation.trigger({
			id: initialData?.id,
			iws_code,
			building_group: Number(building_group)
		});
	}

	function getError(field: string) {
		return (
			getFieldError(liveValidation.fieldErrors, field, ['building']) ??
			formState.getFieldError(field, ['building'])
		);
	}

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();
		await formState.handleSubmit(async () => {
			if (initialData) {
				return await manageBuilding.update(initialData.id, {
					iws_code,
					building_group: Number(building_group)
				});
			} else {
				return await manageBuilding.create({
					iws_code,
					building_group: Number(building_group)
				});
			}
		});
	}
</script>

<form onsubmit={handleSubmit} class="space-y-4 rounded-md border bg-muted/20 p-4">
	<div class="mb-4 flex items-center justify-between">
		<h3 class="text-lg font-medium">
			{initialData
				? $t('facility.forms.building.title_edit')
				: $t('facility.forms.building.title_new')}
		</h3>
	</div>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
		<div class="space-y-2">
			<Label for="iws_code">{$t('facility.forms.building.iws_label')}</Label>
			<Input
				id="iws_code"
				bind:value={iws_code}
				required
				placeholder={$t('facility.forms.building.iws_placeholder')}
				minlength={4}
				maxlength={4}
				oninput={triggerValidation}
			/>
			<p class="text-xs text-muted-foreground">{$t('facility.forms.building.iws_exact')}</p>
			{#if getError('iws_code')}
				<p class="text-sm text-red-500">{getError('iws_code')}</p>
			{/if}
		</div>

		<div class="space-y-2">
			<Label for="building_group">{$t('facility.forms.building.group_label')}</Label>
			<Input
				id="building_group"
				type="number"
				bind:value={building_group}
				required
				placeholder={$t('facility.forms.building.group_placeholder')}
				oninput={triggerValidation}
			/>
			{#if getError('building_group')}
				<p class="text-sm text-red-500">{getError('building_group')}</p>
			{/if}
		</div>
	</div>

	{#if formState.error || liveValidation.error}
		<p class="text-sm text-red-500">{formState.error || liveValidation.error}</p>
	{/if}

	<div class="flex justify-end gap-2 pt-2">
		<Button type="button" variant="ghost" onclick={onCancel}>{$t('common.cancel')}</Button>
		<Button type="submit" disabled={formState.loading}>
			{initialData ? $t('common.update') : $t('common.create')}
		</Button>
	</div>
</form>
