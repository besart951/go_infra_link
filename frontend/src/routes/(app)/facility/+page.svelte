<script lang="ts">
	import Building2Icon from '@lucide/svelte/icons/building-2';
	import { createTranslator } from '$lib/i18n/translator';

	const t = createTranslator();
</script>

<svelte:head>
	<title>{$t('facility.facility_overview')} | Infra Link</title>
</svelte:head>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-semibold tracking-tight">{$t('facility.facility_overview')}</h1>
		<p class="text-sm text-muted-foreground">
			{$t('facility.facility_overview_desc')}
		</p>
	</div>

	<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
		<a
			href="/facility/buildings"
			class="group rounded-lg border bg-card p-6 text-card-foreground shadow-sm transition-colors hover:bg-accent"
		>
			<div class="flex items-center gap-4">
				<div class="rounded-md bg-primary/10 p-3">
					<Building2Icon class="size-6 text-primary" />
				</div>
				<div>
					<h3 class="font-semibold">{$t('facility.buildings')}</h3>
					<p class="text-sm text-muted-foreground">{$t('facility.manage_building_infrastructure')}</p>
				</div>
			</div>
		</a>

		<a
			href="/facility/control-cabinets"
			class="group rounded-lg border bg-card p-6 text-card-foreground shadow-sm transition-colors hover:bg-accent"
		>
			<div class="flex items-center gap-4">
				<div class="rounded-md bg-primary/10 p-3">
					<Building2Icon class="size-6 text-primary" />
				</div>
				<div>
					<h3 class="font-semibold">{$t('facility.control_cabinets')}</h3>
					<p class="text-sm text-muted-foreground">{$t('facility.manage_cabinet_configurations')}</p>
				</div>
			</div>
		</a>

		<a
			href="/facility/sps-controllers"
			class="group rounded-lg border bg-card p-6 text-card-foreground shadow-sm transition-colors hover:bg-accent"
		>
			<div class="flex items-center gap-4">
				<div class="rounded-md bg-primary/10 p-3">
					<Building2Icon class="size-6 text-primary" />
				</div>
				<div>
					<h3 class="font-semibold">{$t('facility.sps_controllers')}</h3>
					<p class="text-sm text-muted-foreground">{$t('facility.manage_sps_devices')}</p>
				</div>
			</div>
		</a>

		<a
			href="/facility/field-devices"
			class="group rounded-lg border bg-card p-6 text-card-foreground shadow-sm transition-colors hover:bg-accent"
		>
			<div class="flex items-center gap-4">
				<div class="rounded-md bg-primary/10 p-3">
					<Building2Icon class="size-6 text-primary" />
				</div>
				<div>
					<h3 class="font-semibold">{$t('facility.field_devices')}</h3>
					<p class="text-sm text-muted-foreground">{$t('facility.field_devices_desc')}</p>
				</div>
			</div>
		</a>
	</div>
</div>
