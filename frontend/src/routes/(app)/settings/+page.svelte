<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { setThemePreference, themePreference, type ThemePreference } from '$lib/stores/theme.js';
	import { LaptopMinimal, Moon, Sun } from '@lucide/svelte';

	type ThemeOption = {
		value: ThemePreference;
		label: string;
		description: string;
		icon: typeof LaptopMinimal;
	};

	const options: ThemeOption[] = [
		{
			value: 'system',
			label: 'System',
			description: 'Follow your OS appearance.',
			icon: LaptopMinimal
		},
		{
			value: 'light',
			label: 'Light',
			description: 'Always use light theme.',
			icon: Sun
		},
		{
			value: 'dark',
			label: 'Dark',
			description: 'Always use dark theme.',
			icon: Moon
		}
	];
</script>

<div class="flex flex-col gap-6">
	<div>
		<h1 class="text-3xl font-bold tracking-tight">Settings</h1>
		<p class="text-muted-foreground mt-2 text-sm">Customize your console preferences.</p>
	</div>

	<div class="bg-card rounded-lg border p-4">
		<div class="flex flex-col gap-1">
			<h2 class="text-base font-semibold">Appearance</h2>
			<p class="text-muted-foreground text-sm">Choose how Infra Link looks on this device.</p>
		</div>

		<div class="mt-4 grid gap-2 sm:grid-cols-3">
			{#each options as opt (opt.value)}
				{@const active = $themePreference === opt.value}
				<Button
					variant={active ? 'default' : 'outline'}
					class="h-auto justify-start gap-3 px-4 py-3"
					onclick={() => setThemePreference(opt.value)}
				>
					<svelte:component this={opt.icon} />
					<span class="flex flex-col items-start gap-0.5 text-left">
						<span class="leading-tight">{opt.label}</span>
						<span class={active ? 'text-primary-foreground/80 text-xs' : 'text-muted-foreground text-xs'}>
							{opt.description}
						</span>
					</span>
				</Button>
			{/each}
		</div>
	</div>
</div>
