<script lang="ts">
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import TrashIcon from '@lucide/svelte/icons/trash-2';
	import type { PageData } from './$types.js';
	import { updateBuilding, deleteBuilding } from '$lib/infrastructure/api/facility.adapter';
	import { goto, invalidateAll } from '$app/navigation';
	import { page } from '$app/stores';

	let { data }: { data: PageData } = $props();

	let isSubmitting = $state(false);
	let errors: Record<string, string> = $state({});
	let successMessage = $state('');

	async function handleDeleteClick() {
		if (!confirm('Are you sure you want to delete this building? This action cannot be undone.')) {
			return;
		}

		try {
			isSubmitting = true;
			await deleteBuilding(data.building.id);
			await goto('/facility/buildings');
		} catch (e: any) {
			console.error('Delete failed', e);
			alert(e.message || 'Failed to delete building');
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
		const iws_code = formData.get('iws_code')?.toString().trim();
		const building_group = formData.get('building_group')?.toString().trim();

		// Validation
		if (!iws_code) {
			errors.iws_code = 'IWS Code is required';
		}

		if (!building_group) {
			errors.building_group = 'Building Group is required';
		} else if (isNaN(Number(building_group))) {
			errors.building_group = 'Building Group must be a number';
		}

		if (Object.keys(errors).length > 0) {
			isSubmitting = false;
			return;
		}

		try {
			await updateBuilding(data.building.id, {
				iws_code,
				building_group: Number(building_group)
			});
			successMessage = 'Building updated successfully!';
			await invalidateAll();
		} catch (e: any) {
			console.error('Update failed', e);
			errors.form = e.message || 'An unexpected error occurred';
		} finally {
			isSubmitting = false;
		}
	}
</script>

<svelte:head>
	<title>{data.building.iws_code} | Buildings | Infra Link</title>
</svelte:head>

<div class="mx-auto max-w-2xl space-y-6">
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-4">
			<Button variant="ghost" size="icon" href="/facility/buildings">
				<ArrowLeftIcon class="size-4" />
			</Button>
			<div>
				<h1 class="text-2xl font-semibold tracking-tight">{data.building.iws_code}</h1>
				<p class="text-sm text-muted-foreground">Edit building details</p>
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
				Delete
			</Button>
		</div>
	</div>

	{#if errors.form}
		<div class="rounded-md border border-destructive bg-destructive/10 p-4 text-destructive">
			{errors.form}
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
				<Field.Legend>Building Details</Field.Legend>

				<Field.Field>
					<Field.Label for="iws_code">IWS Code</Field.Label>
					<Field.Content>
						<Input
							id="iws_code"
							name="iws_code"
							placeholder="e.g. ABCD"
							value={data.building.iws_code}
							aria-invalid={!!errors.iws_code}
						/>
						{#if errors.iws_code}
							<Field.Error>{errors.iws_code}</Field.Error>
						{/if}
					</Field.Content>
					<Field.Description>The unique IWS code identifier for this building.</Field.Description>
				</Field.Field>

				<Field.Field>
					<Field.Label for="building_group">Building Group</Field.Label>
					<Field.Content>
						<Input
							id="building_group"
							name="building_group"
							type="number"
							placeholder="e.g. 1"
							value={data.building.building_group}
							aria-invalid={!!errors.building_group}
						/>
						{#if errors.building_group}
							<Field.Error>{errors.building_group}</Field.Error>
						{/if}
					</Field.Content>
					<Field.Description>The group number this building belongs to.</Field.Description>
				</Field.Field>
			</Field.Set>
		</div>

		<div class="flex justify-end gap-4">
			<Button variant="outline" href="/facility/buildings">Cancel</Button>
			<Button type="submit" disabled={isSubmitting}>
				{isSubmitting ? 'Saving...' : 'Save Changes'}
			</Button>
		</div>
	</form>
</div>
