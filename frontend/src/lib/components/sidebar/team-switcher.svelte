<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { useSidebar } from '$lib/components/ui/sidebar/index.js';
	import ChevronsUpDownIcon from '@lucide/svelte/icons/chevrons-up-down';
	import PlusIcon from '@lucide/svelte/icons/plus';
	import type { Team } from '$lib/domain/team/index.js';
	import { createTranslator } from '$lib/i18n/translator.js';

	let {
		teams,
		activeTeam = $bindable(),
		onCreateTeam
	}: {
		teams: Team[];
		activeTeam?: Team;
		onCreateTeam?: () => void;
	} = $props();

	const t = createTranslator();

	const sidebar = useSidebar();
</script>

<Sidebar.Menu>
	<Sidebar.MenuItem>
		<DropdownMenu.Root>
			<DropdownMenu.Trigger>
				{#snippet child({ props })}
					<Sidebar.MenuButton
						size="lg"
						class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
						{...props}
					>
						<div
							class="flex size-8 items-center justify-center rounded-md bg-primary font-semibold text-primary-foreground"
						>
							{activeTeam?.name?.[0]?.toUpperCase() ?? $t('app.brand_short')}
						</div>
						<div class="grid flex-1 text-start text-sm leading-tight">
							<span class="truncate font-semibold">{activeTeam?.name ?? $t('app.brand')}</span>
							<span class="truncate text-xs text-muted-foreground">
								{activeTeam
									? $t('sidebar.team_switcher.team')
									: $t('sidebar.team_switcher.console')}
							</span>
						</div>
						<ChevronsUpDownIcon class="ms-auto" />
					</Sidebar.MenuButton>
				{/snippet}
			</DropdownMenu.Trigger>
			<DropdownMenu.Content
				class="w-(--bits-dropdown-menu-anchor-width) min-w-56 rounded-lg"
				side={sidebar.isMobile ? 'bottom' : 'right'}
				align="start"
				sideOffset={4}
			>
				<DropdownMenu.Label class="text-xs text-muted-foreground">
					{$t('sidebar.team_switcher.label')}
				</DropdownMenu.Label>
				{#each teams as team (team.id)}
					<DropdownMenu.Item onclick={() => (activeTeam = team)} class="gap-2 p-2">
						<div
							class="flex size-6 items-center justify-center rounded-sm bg-primary text-xs font-semibold text-primary-foreground"
						>
							{team.name?.[0]?.toUpperCase() ?? 'T'}
						</div>
						{team.name}
					</DropdownMenu.Item>
				{/each}
				{#if onCreateTeam}
					<DropdownMenu.Separator />
					<DropdownMenu.Item onclick={onCreateTeam} class="gap-2 p-2">
						<div class="flex size-6 items-center justify-center rounded-md border bg-background">
							<PlusIcon class="size-4" />
						</div>
						<span class="font-medium text-muted-foreground">
							{$t('sidebar.team_switcher.add_team')}
						</span>
					</DropdownMenu.Item>
				{/if}
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</Sidebar.MenuItem>
</Sidebar.Menu>
