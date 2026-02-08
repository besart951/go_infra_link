<script lang="ts">
	import { onMount } from 'svelte';
	import AppSidebar from '$lib/components/app-sidebar.svelte';
	import Toasts from '$lib/components/toast.svelte';
	import * as Breadcrumb from '$lib/components/ui/breadcrumb/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { page } from '$app/stores';
	import type { LayoutData } from './$types.js';
	import { loadAuth } from '$lib/stores/auth.svelte.js';
	import { goto } from '$app/navigation';

	const { children, data } = $props<{ children: any; data: LayoutData }>();

	onMount(async () => {
		await loadAuth();
	});

	$effect(() => {
		if (!data.user && data.backendAvailable !== false) {
			goto('/login');
		}
	});

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
		if (pathname.startsWith('/facility/system-types')) return 'System Types';
		if (pathname.startsWith('/facility/system-parts')) return 'System Parts';
		if (pathname.startsWith('/facility/apparats')) return 'Apparats';
		if (pathname.startsWith('/facility/object-data')) return 'Object Data';
		if (pathname.startsWith('/facility/specifications')) return 'Specifications';
		if (pathname.startsWith('/facility/state-texts')) return 'State Texts';
		if (pathname.startsWith('/facility/alarm-definitions')) return 'Alarm Definitions';
		if (pathname.startsWith('/facility/notification-classes')) return 'Notification Classes';
		if (pathname.startsWith('/facility')) return 'Facility';
		return 'App';
	};
</script>

{#if data.user}
<Sidebar.Provider>
	<AppSidebar
		user={data.user}
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
	<Toasts />
</Sidebar.Provider>
{:else if data.backendAvailable === false}
	<div class="flex h-screen w-full items-center justify-center p-4">
		<div class="w-full max-w-md rounded-lg border bg-card p-6 shadow-sm">
			<h2 class="mb-2 text-lg font-semibold text-destructive">Backend Unavailable</h2>
			<p class="text-sm text-muted-foreground">
				The backend service is currently unreachable. Please check your connection or contact support.
			</p>
		</div>
	</div>
{/if}
