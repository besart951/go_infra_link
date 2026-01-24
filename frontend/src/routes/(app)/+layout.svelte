<script lang="ts">
	import AppSidebar from '$lib/components/app-sidebar.svelte';
	import * as Breadcrumb from '$lib/components/ui/breadcrumb/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { page } from '$app/stores';
	import type { LayoutData } from './$types.js';

	const { children, data } = $props<{ children: any; data: LayoutData }>();

	const titleForPath = (pathname: string) => {
		if (pathname === '/') return 'Dashboard';
		if (pathname.startsWith('/projects')) return 'Projects';
		if (pathname.startsWith('/users')) return 'Users';
		if (pathname.startsWith('/teams')) return 'Teams';
		if (pathname.startsWith('/settings')) return 'Settings';
		if (pathname.startsWith('/facility/buildings')) return 'Buildings';
		if (pathname.startsWith('/facility/control-cabinets')) return 'Control Cabinets';
		if (pathname.startsWith('/facility/sps-controllers')) return 'SPS Controllers';
		if (pathname.startsWith('/facility/field-devices')) return 'Field Devices';
		if (pathname.startsWith('/facility')) return 'Facility';
		return 'App';
	};

	// Provide a default user if not loaded
	const defaultUser = {
		id: '',
		first_name: 'User',
		last_name: '',
		email: '',
		role: 'user' as const,
		is_active: true,
		failed_login_attempts: 0,
		created_at: '',
		updated_at: ''
	};
</script>

<Sidebar.Provider>
	<AppSidebar
		user={data.user ?? defaultUser}
		teams={data.teams ?? []}
		projects={data.projects ?? []}
	/>
	<Sidebar.Inset>
		<header class="flex h-16 shrink-0 items-center gap-2">
			<div class="flex items-center gap-2 px-4">
				<Sidebar.Trigger class="-ms-1" />
				<Separator orientation="vertical" class="me-2 data-[orientation=vertical]:h-4" />
				<Breadcrumb.Root>
					<Breadcrumb.List>
						<Breadcrumb.Item class="hidden md:block">
							<Breadcrumb.Link href="/">Infrastructure Link</Breadcrumb.Link>
						</Breadcrumb.Item>
						<Breadcrumb.Separator class="hidden md:block" />
						<Breadcrumb.Item>
							<Breadcrumb.Page>{titleForPath($page.url.pathname)}</Breadcrumb.Page>
						</Breadcrumb.Item>
					</Breadcrumb.List>
				</Breadcrumb.Root>
			</div>
		</header>
		{#if data && data.backendAvailable === false}
			<div class="px-4">
				<div class="rounded-md border bg-muted px-3 py-2 text-sm text-muted-foreground">
					Backend service is currently unavailable. Some actions may fail.
				</div>
			</div>
		{/if}
		<div class="flex flex-1 flex-col gap-4 p-4 pt-0">{@render children?.()}</div>
	</Sidebar.Inset>
</Sidebar.Provider>
