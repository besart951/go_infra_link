<script lang="ts">
	import { Input } from '$lib/components/ui/input/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import type { ActionData } from './$types.js';

	let { form }: { form: ActionData } = $props();
</script>

<svelte:head>
	<title>New Building | Infra Link</title>
</svelte:head>

<div class="mx-auto max-w-2xl space-y-6">
	<div class="flex items-center gap-4">
		<Button variant="ghost" size="icon" href="/facility/buildings">
			<ArrowLeftIcon class="size-4" />
		</Button>
		<div>
			<h1 class="text-2xl font-semibold tracking-tight">New Building</h1>
			<p class="text-sm text-muted-foreground">Create a new building in your facility.</p>
		</div>
	</div>

	{#if form?.errors?.form}
		<div class="rounded-md border border-destructive bg-destructive/10 p-4 text-destructive">
			{form.errors.form}
		</div>
	{/if}

	<form method="POST" class="space-y-6">
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
							value={form?.values?.iws_code ?? ''}
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
							value={form?.values?.building_group ?? ''}
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
			<Button type="submit">Create Building</Button>
		</div>
	</form>
</div>
