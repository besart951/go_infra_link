<script lang="ts">
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import TrashIcon from '@lucide/svelte/icons/trash-2';
	import type { PageData } from './$types.js';
	import { ManageBuildingUseCase } from '$lib/application/useCases/facility/manageBuildingUseCase.js';
	import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
	const manageBuilding = new ManageBuildingUseCase(buildingRepository);
	import { goto, invalidateAll } from '$app/navigation';
	import { useLiveValidation } from '$lib/hooks/useLiveValidation.svelte.js';
	import { getFieldError } from '$lib/api/client.js';
	import { createTranslator } from '$lib/i18n/translator.js';
	import { t as translate } from '$lib/i18n/index.js';

	let { data }: { data: PageData } = $props();

	let isSubmitting = $state(false);
	let errors: Record<string, string> = $state({});
	let successMessage = $state('');
	let iwsCode = $state('');
	let buildingGroup = $state(0);

	const t = createTranslator();

	$effect(() => {
		iwsCode = data.building.iws_code;
		buildingGroup = data.building.building_group;
	});
	const liveValidation = useLiveValidation((data: any) => manageBuilding.validate(data), {
		debounceMs: 400
	});

	function triggerValidation() {
		liveValidation.trigger({
			id: data.building.id,
			iws_code: iwsCode,
			building_group: Number(buildingGroup)
		});
	}

	async function handleDeleteClick() {
		if (!confirm(translate('facility.building_detail.delete_confirm'))) {
			return;
		}

		try {
			isSubmitting = true;
			await manageBuilding.delete(data.building.id);
			await goto('/facility/buildings');
		} catch (e: any) {
			console.error('Delete failed', e);
			alert(e.message || translate('facility.building_detail.delete_failed'));
		} finally {
			isSubmitting = false;
		}
	}

	async function handleUpdate(e: SubmitEvent) {
		e.preventDefault();
		isSubmitting = true;
		errors = {};
		successMessage = '';

		const formData = new FormData(e.currentTarget as HTMLFormElement);
		const iws_code = iwsCode?.toString().trim();
		const building_group = String(buildingGroup).trim();

		// Validation
		if (!iws_code) {
			errors.iws_code = translate('facility.building_detail.iws_required');
		}

		if (!building_group) {
			errors.building_group = translate('facility.building_detail.group_required');
		} else if (isNaN(Number(building_group))) {
			errors.building_group = translate('facility.building_detail.group_number');
		}

		if (Object.keys(errors).length > 0) {
			isSubmitting = false;
			return;
		}

		try {
			await manageBuilding.update(data.building.id, {
				iws_code,
				building_group: Number(building_group)
			});
			successMessage = translate('facility.building_detail.updated');
			await invalidateAll();
		} catch (e: any) {
			console.error('Update failed', e);
			errors.form = e.message || translate('facility.building_detail.update_failed');
		} finally {
			isSubmitting = false;
		}
	}
</script>

<svelte:head>
	<title>{data.building.iws_code} | {$t('facility.buildings')} | {$t('app.brand')}</title>
</svelte:head>

<div class="mx-auto max-w-2xl space-y-6">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-4">
			<Button variant="ghost" size="icon" href="/facility/buildings">
				<ArrowLeftIcon class="size-4" />
			</Button>
			<div>
				<h1 class="text-2xl font-semibold tracking-tight">{data.building.iws_code}</h1>
				<p class="text-sm text-muted-foreground">
					{$t('facility.building_detail.edit_description')}
				</p>
			</div>
		</div>
		<div>
			<Button
				variant="destructive"
				size="sm"
				type="button"
				onclick={handleDeleteClick}
				disabled={isSubmitting}
			>
				<TrashIcon class="mr-2 size-4" />
				{$t('common.delete')}
			</Button>
		</div>
	</div>

	{#if errors.form}
		<div class="rounded-md border border-destructive bg-destructive/10 p-4 text-destructive">
			{errors.form}
		</div>
	{/if}

	{#if liveValidation.error}
		<div class="rounded-md border border-destructive bg-destructive/10 p-4 text-destructive">
			{liveValidation.error}
		</div>
	{/if}

	{#if successMessage}
		<div
			class="rounded-md border border-green-500 bg-green-500/10 p-4 text-green-700 dark:text-green-400"
		>
			{successMessage}
		</div>
	{/if}

	<form onsubmit={handleUpdate} class="space-y-6">
		<div class="rounded-lg border bg-card p-6">
			<Field.Set>
				<Field.Legend>{$t('facility.building_detail.legend')}</Field.Legend>

				<Field.Field>
					<Field.Label for="iws_code">{$t('facility.building_detail.iws_label')}</Field.Label>
					<Field.Content>
						<Input
							id="iws_code"
							name="iws_code"
							placeholder={$t('facility.building_detail.iws_placeholder')}
							bind:value={iwsCode}
							aria-invalid={!!errors.iws_code ||
								!!getFieldError(liveValidation.fieldErrors, 'iws_code', ['building'])}
							oninput={triggerValidation}
						/>
						{#if errors.iws_code || getFieldError( liveValidation.fieldErrors, 'iws_code', ['building'] )}
							<Field.Error>
								{errors.iws_code ||
									getFieldError(liveValidation.fieldErrors, 'iws_code', ['building'])}
							</Field.Error>
						{/if}
					</Field.Content>
					<Field.Description>{$t('facility.building_detail.iws_help')}</Field.Description>
				</Field.Field>

				<Field.Field>
					<Field.Label for="building_group"
						>{$t('facility.building_detail.group_label')}</Field.Label
					>
					<Field.Content>
						<Input
							id="building_group"
							name="building_group"
							type="number"
							placeholder={$t('facility.building_detail.group_placeholder')}
							bind:value={buildingGroup}
							aria-invalid={!!errors.building_group ||
								!!getFieldError(liveValidation.fieldErrors, 'building_group', ['building'])}
							oninput={triggerValidation}
						/>
						{#if errors.building_group || getFieldError( liveValidation.fieldErrors, 'building_group', ['building'] )}
							<Field.Error>
								{errors.building_group ||
									getFieldError(liveValidation.fieldErrors, 'building_group', ['building'])}
							</Field.Error>
						{/if}
					</Field.Content>
					<Field.Description>{$t('facility.building_detail.group_help')}</Field.Description>
				</Field.Field>
			</Field.Set>
		</div>

		<div class="flex justify-end gap-4">
			<Button variant="outline" href="/facility/buildings">{$t('common.cancel')}</Button>
			<Button type="submit" disabled={isSubmitting}>
				{isSubmitting ? $t('common.saving') : $t('common.save_changes')}
			</Button>
		</div>
	</form>
</div>
