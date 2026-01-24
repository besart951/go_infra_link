<script lang="ts">
	import Building2Icon from '@lucide/svelte/icons/building-2';
</script>

<svelte:head>
	<title>Facility | Infra Link</title>
</svelte:head>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-semibold tracking-tight">Facility Management</h1>
		<p class="text-sm text-muted-foreground">
			Select a category from the sidebar to manage your facility infrastructure.
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
					<h3 class="font-semibold">Buildings</h3>
					<p class="text-sm text-muted-foreground">Manage building infrastructure</p>
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
					<h3 class="font-semibold">Control Cabinets</h3>
					<p class="text-sm text-muted-foreground">Manage cabinet configurations</p>
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
					<h3 class="font-semibold">SPS Controllers</h3>
					<p class="text-sm text-muted-foreground">Manage SPS devices</p>
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
					<h3 class="font-semibold">Field Devices</h3>
					<p class="text-sm text-muted-foreground">Manage field devices</p>
				</div>
			</div>
		</a>
	</div>
</div>
