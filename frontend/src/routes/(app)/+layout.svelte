<script lang="ts">
	import AppSidebar from "$lib/components/app-sidebar.svelte";
	import * as Breadcrumb from "$lib/components/ui/breadcrumb/index.js";
	import { Separator } from "$lib/components/ui/separator/index.js";
	import * as Sidebar from "$lib/components/ui/sidebar/index.js";
	import { page } from "$app/stores";
	import type { LayoutData } from "./$types.js";

	const { children, data } = $props<{ children: any; data: LayoutData }>();

	const titleForPath = (pathname: string) => {
		if (pathname === "/") return "Dashboard";
		if (pathname.startsWith("/projects")) return "Projects";
		if (pathname.startsWith("/users")) return "Users";
		return "App";
	};
</script>

<Sidebar.Provider>
	<AppSidebar />
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
				<div class="bg-muted text-muted-foreground rounded-md border px-3 py-2 text-sm">
					Backend service is currently unavailable. Some actions may fail.
				</div>
			</div>
		{/if}
		<div class="flex flex-1 flex-col gap-4 p-4 pt-0">{@render children?.()}</div>
	</Sidebar.Inset>
</Sidebar.Provider>
