<script lang="ts">
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import TrashIcon from '@lucide/svelte/icons/trash-2';
	import type { PageData, ActionData } from './$types.js';
	import { enhance } from '$app/forms';

	let { data, form }: { data: PageData; form: ActionData } = $props();

	let deleteFormEl: HTMLFormElement | null = $state(null);

	function handleDeleteClick(e: Event) {
		e.preventDefault();
		if (confirm('Are you sure you want to delete this building? This action cannot be undone.')) {
			deleteFormEl?.submit();
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
		<form method="POST" action="?/delete" bind:this={deleteFormEl} use:enhance>
			<Button variant="destructive" size="sm" type="button" onclick={handleDeleteClick}>
				<TrashIcon class="mr-2 size-4" />
				Delete
			</Button>
		</form>
	</div>

	{#if form?.errors?.form}
		<div class="rounded-md border border-destructive bg-destructive/10 p-4 text-destructive">
			{form.errors.form}
		</div>
	{/if}

	{#if form?.success}
		<div
			class="rounded-md border border-green-500 bg-green-500/10 p-4 text-green-700 dark:text-green-400"
		>
			Building updated successfully!
		</div>
	{/if}

	<form method="POST" action="?/update" use:enhance class="space-y-6">
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
							value={form?.values?.iws_code ?? data.building.iws_code}
							aria-invalid={!!form?.errors?.iws_code}
						/>
						{#if form?.errors?.iws_code}
							<Field.Error>{form.errors.iws_code}</Field.Error>
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
							value={form?.values?.building_group ?? data.building.building_group}
							aria-invalid={!!form?.errors?.building_group}
						/>
						{#if form?.errors?.building_group}
							<Field.Error>{form.errors.building_group}</Field.Error>
						{/if}
					</Field.Content>
					<Field.Description>The group number this building belongs to.</Field.Description>
				</Field.Field>
			</Field.Set>
		</div>

		<div class="flex justify-end gap-4">
			<Button variant="outline" href="/facility/buildings">Cancel</Button>
			<Button type="submit">Save Changes</Button>
		</div>
	</form>

	<div class="rounded-lg border bg-card p-6">
		<h2 class="mb-4 text-lg font-medium">Additional Information</h2>
		<dl class="grid gap-4 text-sm">
			<div class="flex justify-between">
				<dt class="text-muted-foreground">ID</dt>
				<dd class="font-mono">{data.building.id}</dd>
			</div>
			<div class="flex justify-between">
				<dt class="text-muted-foreground">Created</dt>
				<dd>{new Date(data.building.created_at).toLocaleString()}</dd>
			</div>
			<div class="flex justify-between">
				<dt class="text-muted-foreground">Last Updated</dt>
				<dd>{new Date(data.building.updated_at).toLocaleString()}</dd>
			</div>
		</dl>
	</div>
</div>
